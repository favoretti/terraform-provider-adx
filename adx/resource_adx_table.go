package adx

import (
	"context"
	"fmt"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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
				ValidateDiagFunc: stringIsNotEmpty,
			},
		},
	}
}

func resourceADXTableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Meta).Kusto

	table_name := d.Get("table_name").(string)
	database_name := d.Get("database_name").(string)
	table_schema := d.Get("table_schema").(string)

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})

	create_statement := fmt.Sprintf(".create table %s (%s)", table_name, table_schema)

	_, err := client.Mgmt(ctx, database_name, kusto.NewStmt("", kStmtOpts).UnsafeAdd(create_statement))
	if err != nil {
		return diag.Errorf("error creating Table %q (Database %q): %+v", table_name, database_name, err)
	}

	id := fmt.Sprintf("%s|%s|%s", client.Endpoint(), database_name, table_name)
	d.SetId(id)

	resourceADXTableRead(ctx, d, meta)

	return diags
}

func resourceADXTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	//client := meta.(*Meta).Kusto

	id, err := parseADXTableID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("table_name", id.Name)
	d.Set("database_name", id.DatabaseName)

	return diags
}

func resourceADXTableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}
