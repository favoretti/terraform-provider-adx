package adx

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/data/value"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type TableMapping struct {
	Name          string
	Kind          string
	Mapping       string
	LastUpdatedOn value.DateTime
	Table         string
	Database      string
}

type Mapping struct {
	Column     string `json:"column,omitempty"`
	Path       string `json:"path,omitempty"`
	Ordinal    string `json:"ordinal,omitempty"`
	ConstValue string `json:"constvalue,omitempty"`
	DataType   string `json:"datatype,omitempty"`
	Transform  string `json:"transform,omitempty"`
	Field      string `json:"field,omitempty"`
	Name       string `json:"name,omitempty"`
}

func resourceADXTableMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableMappingCreateUpdate,
		UpdateContext: resourceADXTableMappingCreateUpdate,
		ReadContext:   resourceADXTableMappingRead,
		DeleteContext: resourceADXTableMappingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		StateUpgraders: []schema.StateUpgrader{
			TableMappingV0ToV1Upgrader(),
		},
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"cluster": getClusterConfigInputSchema(),
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			},
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			},

			"table_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringIsNotEmpty),
			},

			"kind": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
					"json", "csv", "parquet", "avro", "orc", "w3clogfile",
				}, true)),
			},
			"mapping": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"column": {
							Type:     schema.TypeString,
							Required: true,
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"field": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ordinal": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"constvalue": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"datatype": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"transform": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"last_updated_on": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXTableMappingCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}

	name := d.Get("name").(string)
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	kind := d.Get("kind").(string)
	mapping := expandTableMapping(d.Get("mapping").([]interface{}))
	entityType := "table"

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	createStatement := fmt.Sprintf(".create-or-alter table %s ingestion %s mapping '%s' '[%s]'", tableName, strings.ToLower(kind), name, mapping)

	_, err = client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating Mapping %q (Table %q, Database %q): %+v", name, tableName, databaseName, err)
	}

	id := buildADXResourceId(client.Endpoint(), databaseName, entityType, tableName, "tablemapping", kind, name)
	d.SetId(id)

	resourceADXTableMappingRead(ctx, d, meta)

	return diags
}

func resourceADXTableMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}

	id, err := parseADXTableMappingID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if tableExists, err := isTableExists(ctx, meta, clusterConfig, id.DatabaseName, id.Name); err != nil || !tableExists {
		if err != nil {
			return diag.Errorf("%+v", err)
		}
		d.SetId("")
		return diags
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	showStatement := fmt.Sprintf(".show table %s ingestion %s mapping '%s'", id.Name, strings.ToLower(id.Kind), id.MappingName)

	resp, err := client.Mgmt(ctx, id.DatabaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(showStatement))
	if err != nil {
		return diag.Errorf("error reading Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}
	defer resp.Stop()

	var schemas []TableMapping
	err = resp.Do(
		func(row *table.Row) error {
			rec := TableMapping{}
			if err := row.ToStruct(&rec); err != nil {
				return fmt.Errorf("error parsing Table schema for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
			}
			schemas = append(schemas, rec)
			return nil
		},
	)

	if err != nil {
		return diag.Errorf("%+v", err)
	}

	d.Set("name", schemas[0].Name)
	d.Set("table_name", schemas[0].Table)
	d.Set("database_name", schemas[0].Database)
	d.Set("kind", strings.ToLower(schemas[0].Kind))
	d.Set("mapping", flattenTableMapping(schemas[0].Mapping))
	d.Set("last_updated_on", schemas[0].LastUpdatedOn.String())

	return diags
}

func resourceADXTableMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXTableMappingID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deleteStatement := fmt.Sprintf(".drop table %s ingestion %s mapping '%s'", id.Name, strings.ToLower(id.Kind), id.MappingName)
	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, deleteStatement)
}

func expandTableMapping(input []interface{}) string {
	if len(input) == 0 {
		return ""
	}

	optionalFields := [6]string{"path", "datatype", "transform", "ordinal", "constvalue", "field"}

	// TODO Convert this to use json.Marshall
	mappings := make([]string, 0)
	for _, v := range input {
		block := v.(map[string]interface{})
		mapping := fmt.Sprintf(`"column":"%s"`, block["column"].(string))
		for _, field := range optionalFields {
			if t, ok := block[field].(string); ok {
				if len(t) != 0 {
					mapping = fmt.Sprintf(`%s,"%s":"%s"`, mapping, field, t)
				}
			}
		}
		mapping = fmt.Sprintf("{%s}", mapping)
		mappings = append(mappings, mapping)
	}
	return strings.Join(mappings, ",")
}

func flattenTableMapping(input string) []interface{} {
	if len(input) == 0 {
		return []interface{}{}
	}

	var oMappings []Mapping
	json.Unmarshal([]byte(input), &oMappings)

	// For certain mapping types, kusto returns very inconsistent data models (variations in both case and parameter name)
	// Example: For CSV, the "column" param in input is reflected as "Name" in output

	mappings := make([]interface{}, 0)
	for _, v := range oMappings {
		block := make(map[string]interface{})
		block["column"] = v.Column
		block["path"] = v.Path
		block["ordinal"] = v.Ordinal
		block["datatype"] = v.DataType
		block["transform"] = v.Transform
		block["constvalue"] = v.ConstValue

		if block["column"] == "" {
			block["column"] = v.Name
		}
		mappings = append(mappings, block)
	}
	return mappings
}
