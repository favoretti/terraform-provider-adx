package adx

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TableSchema struct {
	TableName    string
	Schema       string
	DatabaseName string
	Folder       string
	DocString    string
}

type tableFromQueryConfig struct {
	Query            string
	Append           bool
	ExtendSchema     bool
	RecreateSchema   bool
	Distributed      bool
	ForceUpdateValue string
}

func resourceADXTable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableCreate,
		ReadContext:   resourceADXTableRead,
		DeleteContext: resourceADXTableDelete,
		UpdateContext: resourceADXTableUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		StateUpgraders: []schema.StateUpgrader{
			TableV0ToV1Upgrader(),
		},
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"cluster": getClusterConfigInputSchema(),
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"table_schema": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				AtLeastOneOf:  []string{"table_schema", "column", "from_query"},
				ConflictsWith: []string{"column", "from_query"},
				ValidateDiagFunc: validate.StringMatch(
					regexp.MustCompile("[a-zA-Z0-9:-_,]+"),
					"Table schema must contain only letters, number, dashes, semicolons, commas and underscores and no spaces",
				),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return unescapeTableSchema(old) == unescapeTableSchema(new)
				},
			},

			"column": {
				Type:          schema.TypeList,
				AtLeastOneOf:  []string{"table_schema", "column", "from_query"},
				ConflictsWith: []string{"table_schema", "from_query"},
				Optional:      true,
				Computed:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validate.StringIsNotEmpty,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return unescapeEntityName(old) == unescapeEntityName(new)
							},
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validate.StringIsNotEmpty,
						},
					},
				},
			},

			"from_query": {
				Type:          schema.TypeList,
				AtLeastOneOf:  []string{"table_schema", "column", "from_query"},
				ConflictsWith: []string{"table_schema", "column"},
				Optional:      true,
				Computed:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validate.StringIsNotEmpty,
						},
						"append": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"extend_schema": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"recreate_schema": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"distributed": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"force_an_update_when_value_changed": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},

			"merge_on_update": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"folder": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"docstring": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	tableName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)
	createStatement := ""

	if fromQueryList, ok := d.GetOk("from_query"); ok {
		createStatement = buildTableFromQueryStatement(tableName, true, getTableFromQueryConfig(fromQueryList.([]interface{})), d)
	} else {
		withClause := buildTableWithClause(d)
		createStatement = fmt.Sprintf(".create table %s (%s) %s", tableName, getTableDefinition(d), withClause)
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}

	_, err = client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating Table %q (Database %q): %+v", tableName, databaseName, err)
	}

	d.SetId(buildADXResourceId(client.Endpoint(), databaseName, "table", tableName))

	resourceADXTableRead(ctx, d, meta)

	return diags
}

func resourceADXTableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	tableName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)
	mergeOnUpdate := d.Get("merge_on_update").(bool)
	createStatement := ""

	if fromQueryList, ok := d.GetOk("from_query"); ok {
		createStatement = buildTableFromQueryStatement(tableName, false, getTableFromQueryConfig(fromQueryList.([]interface{})), d)
	} else {
		alterCmd := ".alter"
		if mergeOnUpdate {
			alterCmd = ".alter-merge"
		}
		withClause := buildTableWithClause(d)
		createStatement = fmt.Sprintf("%s table %s (%s) %s", alterCmd, tableName, getTableDefinition(d), withClause)
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}

	_, err = client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error updating Table %q (Database %q): %+v", tableName, databaseName, err)
	}

	resourceADXTableRead(ctx, d, meta)

	return diags
}

func getTableDefinition(d *schema.ResourceData) string {
	tableDef := ""
	if tableSchema, ok := d.GetOk("table_schema"); ok && d.HasChange("table_schema") {
		tableDef = tableSchema.(string)
	} else {
		tableDef = expandTableColumn(d.Get("column").([]interface{}))
	}
	return tableDef
}

func getTableFromQueryConfig(fromQueryList []interface{}) *tableFromQueryConfig {
	fromQuery := fromQueryList[0].(map[string]interface{})
	config := tableFromQueryConfig{
		Query:            fromQuery["query"].(string),
		Append:           fromQuery["append"].(bool),
		ExtendSchema:     fromQuery["extend_schema"].(bool),
		RecreateSchema:   fromQuery["recreate_schema"].(bool),
		Distributed:      fromQuery["distributed"].(bool),
		ForceUpdateValue: fromQuery["force_an_update_when_value_changed"].(string),
	}
	return &config
}

func resourceADXTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}

	id, err := parseADXTableID(d.Id())
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
	showStatement := fmt.Sprintf(".show table %s cslschema", id.Name)

	resp, err := client.Mgmt(ctx, id.DatabaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(showStatement))
	if err != nil {
		return diag.Errorf("error reading Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}
	defer resp.Stop()

	var schemas []TableSchema
	err = resp.Do(
		func(row *table.Row) error {
			rec := TableSchema{}
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

	if len(schemas) == 0 {
		return diag.Errorf("error reading schemas for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	d.Set("name", schemas[0].TableName)
	d.Set("database_name", schemas[0].DatabaseName)
	d.Set("table_schema", schemas[0].Schema)
	d.Set("column", flattenTableColumn(schemas[0].Schema))
	d.Set("docstring", schemas[0].DocString)
	d.Set("folder", schemas[0].Folder)

	return diags
}

func resourceADXTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXTableID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, fmt.Sprintf(".drop table %s", id.Name))
}

func buildTableFromQueryStatement(tableName string, new bool, config *tableFromQueryConfig, d *schema.ResourceData) string {
	var withParams []string

	if config.Distributed {
		withParams = append(withParams, "distributed=true")
	}
	if config.ExtendSchema {
		withParams = append(withParams, "extend_schema=true")
	}
	// recreate_schema Only applies for .set-or-replace
	if config.RecreateSchema && !new {
		withParams = append(withParams, "recreate_schema=true")
	}
	if docstring, ok := d.GetOk("docstring"); ok {
		withParams = append(withParams, fmt.Sprintf("docstring='%s'", docstring))
	}
	if folder, ok := d.GetOk("folder"); ok {
		withParams = append(withParams, fmt.Sprintf("folder='%s'", folder))
	}

	withParamsString := ""
	if len(withParams) > 0 {
		withParamsString = fmt.Sprintf("with(%s)", strings.Join(withParams, ","))
	}

	cmd := ".set-or-append"
	if !config.Append {
		if new {
			cmd = ".set"
		} else {
			cmd = ".set-or-replace"
		}
	}

	return fmt.Sprintf("%s %s %s <| %s", cmd, tableName, withParamsString, config.Query)
}

func expandTableColumn(input []interface{}) string {
	if len(input) == 0 {
		return ""
	}

	columns := make([]string, 0)
	for _, v := range input {
		block := v.(map[string]interface{})
		column := fmt.Sprintf("%s:%s", block["name"].(string), block["type"].(string))
		columns = append(columns, column)
	}
	return strings.Join(columns, ",")
}

func flattenTableColumn(input string) []interface{} {
	if len(input) == 0 {
		return []interface{}{}
	}

	columns := make([]interface{}, 0)
	for _, v := range strings.Split(input, ",") {
		block := make(map[string]interface{})
		fields := strings.Split(v, ":")
		block["name"] = fields[0]
		block["type"] = fields[1]
		columns = append(columns, block)
	}
	return columns
}

func unescapeTableSchema(inlineSchema string) string {
	schemaColumns := strings.Split(inlineSchema, ",")
	for i, column := range schemaColumns {
		schemaColumns[i] = strings.ReplaceAll(strings.ReplaceAll(column, "['", ""), "']", "")
	}
	return strings.Join(schemaColumns, ",")
}

func parseADXTableID(input string) (*adxResourceId, error) {
	return parseADXResourceID(input, 4, 0, 1, 2, 3)
}

func buildTableWithClause(d *schema.ResourceData) string {
	var withParams []string

	if docstring, ok := d.GetOk("docstring"); ok {
		withParams = append(withParams, fmt.Sprintf("docstring='%s'", docstring))
	}
	if folder, ok := d.GetOk("folder"); ok {
		withParams = append(withParams, fmt.Sprintf("folder='%s'", folder))
	}

	withClause := ""
	if len(withParams) > 0 {
		withClause = fmt.Sprintf("with(%s)", strings.Join(withParams, ", "))
	}

	return withClause
}
