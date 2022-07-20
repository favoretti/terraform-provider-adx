package adx

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/Azure/azure-kusto-go/kusto/data/errors"
	"github.com/Azure/azure-kusto-go/kusto/data/value"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type ADXMaterializedView struct {
	Name string
	SourceTable string
	Query       string
	MaterializedTo value.DateTime
	AutoUpdateSchema string
    EffectiveDateTime value.DateTime
	Lookback string
}

func resourceADXMaterializedView() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXMaterializedViewCreate,
		UpdateContext: resourceADXMaterializedViewUpdate,
		ReadContext:   resourceADXMaterializedViewRead,
		DeleteContext: resourceADXMaterializedViewDelete,

		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"source_table_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(validation.StringMatch(
					regexp.MustCompile("[a-zA-Z_ .-0-9]+"),
					"source table name must be between 1 and 1024 characters long and may contain letters, digits, underscores (_), spaces, dots (.), and dashes (-)",
				), validation.StringLenBetween(1, 1024))),
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.All(validation.StringMatch(
					regexp.MustCompile("[a-zA-Z_ .-0-9]+"),
					"name must be between 1 and 1024 characters long and may contain letters, digits, underscores (_), spaces, dots (.), and dashes (-)",
				), validation.StringLenBetween(1, 1024))),
			},

			"query": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"async": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"backfill": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"effective_date_time": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"auto_update_schema": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"update_extents_creation_time": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceADXMaterializedViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceADXMaterializedViewCreateUpdate(ctx,d,meta,true)
}

func resourceADXMaterializedViewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceADXMaterializedViewCreateUpdate(ctx,d,meta,false)
}

func resourceADXMaterializedViewCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, new bool) diag.Diagnostics {
	name := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)
	query := d.Get("query").(string)
	sourceTableName := d.Get("source_table_name").(string)
	async := d.Get("async").(bool)

	asyncString := ""
	if async {
		asyncString = "async"
	}

	var withParams []string

	if backfill, ok := d.GetOk("backfill"); ok  {
		withParams = append(withParams, fmt.Sprintf("backfill=%t",backfill.(bool)))
	}
	if updateExtentsCreationTime, ok := d.GetOk("update_extents_creation_time"); ok {
		withParams = append(withParams, fmt.Sprintf("UpdateExtentsCreationTime=%t",updateExtentsCreationTime.(bool)))
	}
	if autoUpdateSchema, ok := d.GetOk("auto_update_schema"); ok {
		withParams = append(withParams, fmt.Sprintf("autoUpdateSchema=%t",autoUpdateSchema.(bool)))
	}
	if effectiveDateTime, ok := d.GetOk("effective_date_time"); ok  {
		withParams = append(withParams,fmt.Sprintf( "effectiveDateTime=%s",effectiveDateTime.(string)))
	}

	withClause := ""
	if len(withParams) > 0 {
		withClause = fmt.Sprintf("with(%s)", strings.Join(withParams, ", "))
	}

	cmd := ".alter"
	if new {
		cmd = ".create"
	}
	
	createStatement := fmt.Sprintf("%s %s materialized-view %s %s on table %s \n{\n%s\n}", cmd, asyncString, withClause, name, sourceTableName, query)

	client := meta.(*Meta).Kusto
	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	_, err := client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating materialized-view %s (Database %q): %+v", name, databaseName, err)
	}

	d.SetId(buildADXResourceId(client.Endpoint(), databaseName, "materializedview", name))

	return resourceADXMaterializedViewRead(ctx, d, meta)
}

func resourceADXMaterializedViewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, err := parseADXMaterializedViewID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resultSet, diags := readADXEntity[ADXMaterializedView](ctx, meta, id, fmt.Sprintf(".show materialized-view %s | extend Lookback=tostring(Lookback), IsHealthy=tolower(tostring(IsHealthy)), IsEnabled=tolower(tostring(IsEnabled)), AutoUpdateSchema=tolower(tostring(AutoUpdateSchema)), EffectiveDateTime", id.Name), "materialized-view")
	if diags.HasError() {
		return diags
	}
	
	d.Set("name", id.Name)
	d.Set("database_name", id.DatabaseName)
	d.Set("source_table_name", resultSet[0].SourceTable)
	d.Set("query", resultSet[0].Query)
	d.Set("auto_update_schema", resultSet[0].AutoUpdateSchema)
	d.Set("effective_date_time", resultSet[0].EffectiveDateTime)

	return diags
}

func resourceADXMaterializedViewDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, err := parseADXMaterializedViewID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return deleteADXEntity(ctx, d, meta, id.DatabaseName, fmt.Sprintf(".drop materialized-view %s", id.Name))
}

func parseADXMaterializedViewID(input string) (*adxResourceId, error) {
	return parseADXResourceID(input, 4, 0, 1, 2, 3)
}