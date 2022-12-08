package adx

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto/data/value"
	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type ADXMaterializedView struct {
	Name              string
	SourceTable       string
	Query             string
	MaterializedTo    value.DateTime
	AutoUpdateSchema  string
	EffectiveDateTime value.DateTime
	Lookback          string
}

func resourceADXMaterializedView() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXMaterializedViewCreate,
		UpdateContext: resourceADXMaterializedViewUpdate,
		ReadContext:   resourceADXMaterializedViewRead,
		DeleteContext: resourceADXMaterializedViewDelete,

		Schema: map[string]*schema.Schema{
			"cluster": getClusterConfigInputSchema(),
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
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
				Computed: true,
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

			"allow_mv_without_rls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXMaterializedViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceADXMaterializedViewCreateUpdate(ctx, d, meta, true)
}

func resourceADXMaterializedViewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceADXMaterializedViewCreateUpdate(ctx, d, meta, false)
}

func resourceADXMaterializedViewCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, new bool) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
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

	if backfill, ok := d.GetOk("backfill"); ok && new {
		withParams = append(withParams, fmt.Sprintf("backfill=%t", backfill.(bool)))
	}
	if allowMVWithoutRLS, ok := d.GetOk("allow_mv_without_rls"); ok && new {
		withParams = append(withParams, fmt.Sprintf("allowMaterializedViewsWithoutRowLevelSecurity=%t", allowMVWithoutRLS.(bool)))
	}
	if updateExtentsCreationTime, ok := d.GetOk("update_extents_creation_time"); ok && new {
		withParams = append(withParams, fmt.Sprintf("UpdateExtentsCreationTime=%t", updateExtentsCreationTime.(bool)))
	}
	if autoUpdateSchema, ok := d.GetOk("auto_update_schema"); ok && new {
		withParams = append(withParams, fmt.Sprintf("autoUpdateSchema=%t", autoUpdateSchema.(bool)))
	}
	if effectiveDateTime, ok := d.GetOk("effective_date_time"); ok && new {
		withParams = append(withParams, fmt.Sprintf("effectiveDateTime=%s", effectiveDateTime.(string)))
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

	_, err := queryADXMgmt(ctx, meta, clusterConfig, databaseName, createStatement)
	if err != nil {
		return diag.Errorf("error creating materialized-view %s (Database %q): %+v", name, databaseName, err)
	}

	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}
	d.SetId(buildADXResourceId(client.Endpoint(), databaseName, "materializedview", name))

	return resourceADXMaterializedViewRead(ctx, d, meta)
}

func resourceADXMaterializedViewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXMaterializedViewID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	showCommand := fmt.Sprintf(".show materialized-views | where Name == '%s' | extend Lookback=tostring(Lookback), IsHealthy=tolower(tostring(IsHealthy)), IsEnabled=tolower(tostring(IsEnabled)), AutoUpdateSchema=tolower(tostring(AutoUpdateSchema)), EffectiveDateTime", id.Name)
	//showCommand := fmt.Sprintf(".show materialized-views | where Name == '%s' | extend Lookback=tostring(Lookback)", id.Name)
	resultSet, diags := readADXEntity[ADXMaterializedView](ctx, meta, clusterConfig, id, showCommand, "materialized-view")
	if diags.HasError() {
		return diags
	}

	if len(resultSet) < 1 {
		d.SetId("")
	} else {

		autoUpdateSchema, _ := strconv.ParseBool(resultSet[0].AutoUpdateSchema)

		d.Set("name", id.Name)
		d.Set("database_name", id.DatabaseName)
		d.Set("source_table_name", resultSet[0].SourceTable)
		d.Set("query", resultSet[0].Query)
		d.Set("auto_update_schema", autoUpdateSchema)
		d.Set("effective_date_time", resultSet[0].EffectiveDateTime.String())
	}

	return diags
}

func resourceADXMaterializedViewDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXMaterializedViewID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, fmt.Sprintf(".drop materialized-view %s", id.Name))
}

func parseADXMaterializedViewID(input string) (*adxResourceId, error) {
	return parseADXResourceID(input, 4, 0, 1, 2, 3)
}
