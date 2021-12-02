package adx

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TableSchema struct {
	TableName string
	Schema string
	DatabaseName string
	Folder string
	DocString string
}

func resourceADXTable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableCreate,
		ReadContext:   resourceADXTableRead,
		DeleteContext: resourceADXTableDelete,

		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew: 		  true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"table_schema": {
				Type:             schema.TypeString,
				Optional: true,
				Computed: true,
				AtLeastOneOf: []string{"table_schema", "column"},
				ConflictsWith: []string{"column"},
				ValidateDiagFunc: stringMatch(
					regexp.MustCompile("[a-zA-Z0-9:-_,]+"),
					"Table schema must contain only letters, number, dashes, semicolons, commas and underscores and no spaces",
					),
			},

			"column": {
				Type: schema.TypeList,
				AtLeastOneOf: []string{"table_schema", "column"},
				ConflictsWith: []string{"table_schema"},
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: stringIsNotEmpty,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateDiagFunc: stringIsNotEmpty,
						},
					},
				},
			},
		},
	}
}

func resourceADXTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Meta).Kusto

	tableName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	tableDef := ""

	if tableSchema := d.Get("table_schema").(string); len(tableSchema) != 0 {
		tableDef = tableSchema
	} else {
		tableDef = expandTableColumn(d.Get("column").([]interface{}))
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	createStatement := fmt.Sprintf(".create table %s (%s)", tableName, tableDef)

	_, err := client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating Table %q (Database %q): %+v", tableName, databaseName, err)
	}

	id := fmt.Sprintf("%s|%s|%s", client.Endpoint(), databaseName, tableName)
	d.SetId(id)

	resourceADXTableRead(ctx, d, meta)

	return diags
}

func resourceADXTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*Meta).Kusto

	id, err := parseADXTableID(d.Id())
	if err != nil {
		return diag.FromErr(err)
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

	return diags
}

func resourceADXTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*Meta).Kusto

	id, err := parseADXTableID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	deleteStatement := fmt.Sprintf(".drop table %s", id.Name)

	_, err = client.Mgmt(ctx, id.DatabaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(deleteStatement))
	if err != nil {
		return diag.Errorf("error deleting Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	d.SetId("")

	return diags
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
