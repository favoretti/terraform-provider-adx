package adx

import (
	"context"
	"fmt"
	"regexp"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ADXFunction struct {
	Name       string
	Parameters string
	Body       string
}

func resourceADXFunction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXFunctionCreateUpdate,
		UpdateContext: resourceADXFunctionCreateUpdate,
		ReadContext:   resourceADXFunctionRead,
		DeleteContext: resourceADXFunctionDelete,

		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateDiagFunc: stringMatch(
					regexp.MustCompile("[a-zA-Z_ .-0-9]+"),
					"function name must be between 1 and 1024 characters long and may contain letters, digits, underscores (_), spaces, dots (.), and dashes (-)",
				),
			},

			"body": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: stringMatch(
					regexp.MustCompile("{.*}"),
					"function body must include outer curly brackets {}",
				),
			},

			"parameters": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "()",
			},
		},
	}
}

func resourceADXFunctionCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)
	body := d.Get("body").(string)
	parameters := d.Get("parameters").(string)

	createStatement := fmt.Sprintf(".create-or-alter function \n%s%s\n%s", name, parameters, body)

	client := meta.(*Meta).Kusto

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	_, err := client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating function %s (Database %q): %+v", name, databaseName, err)
	}

	d.SetId(buildADXResourceId(client.Endpoint(), databaseName, "function", name))

	return resourceADXFunctionRead(ctx, d, meta)
}

func resourceADXFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	id, err := parseADXFunctionID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resultErr, resultSet := readADXEntity[ADXFunction](ctx, d, meta, id, fmt.Sprintf(".show function %s", id.Name), "function")
	if resultErr != nil {
		return diag.Errorf("%+v", resultErr)
	}

	d.Set("name", id.Name)
	d.Set("database_name", id.DatabaseName)
	d.Set("body", resultSet[0].Body)
	d.Set("parameters", resultSet[0].Parameters)

	return diags
}

func resourceADXFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, err := parseADXFunctionID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return deleteADXEntity(ctx, d, meta, id.DatabaseName, fmt.Sprintf(".drop function %s", id.Name))
}

func parseADXFunctionID(input string) (*adxResourceId, error) {
	return parseADXResourceID(input, 4, 0, 1, 2, 3)
}
