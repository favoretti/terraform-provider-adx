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

			"table_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"table_schema": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: stringMatch(
					regexp.MustCompile("[a-zA-Z0-9:-_,]+"),
					"Table schema must contain only letters, number, dashes, semicolons, commas and underscores and no spaces",
					),
			},
		},
	}
}

func resourceADXTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Meta).Kusto

	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	tableSchema := d.Get("table_schema").(string)

	// strip any spaces from schema, since ADX returns it back without
	tableSchema = strings.ReplaceAll(tableSchema, " ", "")

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	createStatement := fmt.Sprintf(".create table %s (%s)", tableName, tableSchema)

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

	d.Set("table_name", schemas[0].TableName)
	d.Set("database_name", schemas[0].DatabaseName)
	d.Set("table_schema", schemas[0].Schema)

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
