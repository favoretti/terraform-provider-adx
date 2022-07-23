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
	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	Column    string `json:"column"`
	Path      string `json:"path"`
	DataType  string `json:"datatype"`
	Transform string `json:"transform"`
}

func resourceADXTableMapping() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableMappingCreateUpdate,
		UpdateContext: resourceADXTableMappingCreateUpdate,
		ReadContext:   resourceADXTableMappingRead,
		DeleteContext: resourceADXTableMappingDelete,
		StateUpgraders: []schema.StateUpgrader{
			TableMappingV0ToV1Upgrader(),
		},
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"cluster": getClusterConfigInputSchema(),
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"table_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"kind": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validate.StringInSlice([]string{
					"Json",
				}),
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
							Required: true,
						},
						"datatype": {
							Type:     schema.TypeString,
							Required: true,
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

	if tableExists, err := isTableExists(ctx,meta,clusterConfig,id.DatabaseName,id.Name); err != nil || !tableExists{
		if err!=nil {
			return diag.Errorf("%+v",err)
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
	d.Set("kind", schemas[0].Kind)
	d.Set("mapping", flattenTableMapping(schemas[0].Mapping))
	d.Set("last_updated_on", schemas[0].LastUpdatedOn)
	//flattenAndSetClusterConfig(ctx, d, clusterConfig)

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

	mappings := make([]string, 0)
	for _, v := range input {
		block := v.(map[string]interface{})
		mapping := fmt.Sprintf(`"column":"%s","path":"%s","datatype":"%s"`, block["column"].(string), block["path"].(string), block["datatype"].(string))
		if t, ok := block["transform"].(string); ok {
			if len(t) != 0 {
				mapping = fmt.Sprintf(`%s,"transform":"%s"`, mapping, t)
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

	mappings := make([]interface{}, 0)
	for _, v := range oMappings {
		block := make(map[string]interface{})
		block["column"] = v.Column
		block["path"] = v.Path
		block["datatype"] = v.DataType
		block["transform"] = v.Transform
		mappings = append(mappings, block)
	}
	return mappings
}
