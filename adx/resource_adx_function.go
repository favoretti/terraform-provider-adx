package adx

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/favoretti/terraform-provider-adx/adx/validate"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type ADXFunction struct {
	Name       string
	Parameters string
	Body       string
	Folder     string
	DocString  string
}

func resourceADXFunction() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXFunctionCreate,
		UpdateContext: resourceADXFunctionUpdate,
		ReadContext:   resourceADXFunctionRead,
		DeleteContext: resourceADXFunctionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cluster": getClusterConfigInputSchema(),
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(validation.StringMatch(
					regexp.MustCompile("[a-zA-Z_ .-0-9]+"),
					"function name must be between 1 and 1024 characters long and may contain letters, digits, underscores (_), spaces, dots (.), and dashes (-)",
				), validation.StringLenBetween(1, 1024))),
			},

			"body": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validate.StringMatch(
					regexp.MustCompile("(?s)^{.*}$"),
					"function body must include outer curly brackets {}",
				),
			},

			"parameters": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "()",
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
func resourceADXFunctionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceADXFunctionCreateUpdate(ctx, d, meta, true)
}

func resourceADXFunctionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceADXFunctionCreateUpdate(ctx, d, meta, false)
}

func resourceADXFunctionCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, new bool) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	name := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)
	body := d.Get("body").(string)
	parameters := d.Get("parameters").(string)

	cmd := ".alter"
	if new {
		cmd = ".create"
	}

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

	createStatement := fmt.Sprintf("%s function %s\n%s%s\n%s", cmd, withClause, name, parameters, body)

	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	_, err = client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating function %s (Database %q): %+v", name, databaseName, err)
	}

	d.SetId(buildADXResourceId(clusterConfig.URI, databaseName, "function", name))

	return resourceADXFunctionRead(ctx, d, meta)
}

func resourceADXFunctionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	id, err := parseADXFunctionID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resultSet, diags := readADXEntity[ADXFunction](ctx, meta, clusterConfig, id, fmt.Sprintf(".show functions | where Name == '%s'", id.Name), "function")
	if diags.HasError() {
		return diags
	}

	if len(resultSet) < 1 {
		d.SetId("")
	} else {
		d.Set("name", id.Name)
		d.Set("database_name", id.DatabaseName)
		d.Set("body", resultSet[0].Body)
		d.Set("parameters", resultSet[0].Parameters)
		d.Set("docstring", resultSet[0].DocString)
		d.Set("folder", resultSet[0].Folder)
	}

	return diags
}

func resourceADXFunctionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXFunctionID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, fmt.Sprintf(".drop function %s", id.Name))
}

func parseADXFunctionID(input string) (*adxResourceId, error) {
	return parseADXResourceID(input, 4, 0, 1, 2, 3)
}
