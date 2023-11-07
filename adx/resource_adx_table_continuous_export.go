package adx

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ADXContinuousExport struct {
	Name string
}

func resourceADXTableContinuousExport() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXContinuousExportCreateUpdate,
		UpdateContext: resourceADXContinuousExportCreateUpdate,
		ReadContext:   resourceADXContinuousExportRead,
		DeleteContext: resourceADXContinuousExportDelete,
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"external_table_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"query": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"interval_between_runs": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "10h",
			},

			"forced_latency": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"size_limit": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  100000000, //100 MB
			},

			"distributed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"parquet_row_group_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  100000,
			},

			"use_native_parquet_writer": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"managed_identity": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"is_disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceADXContinuousExportCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	name := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)
	query := d.Get("query").(string)
	externalTableName := d.Get("external_table_name").(string)

	var withParams []string

	if intervalBetweenRuns, ok := d.GetOk("interval_between_runs"); ok {
		withParams = append(withParams, fmt.Sprintf("intervalBetweenRuns='%s'", intervalBetweenRuns.(string)))
	}
	if forcedLatency, ok := d.GetOk("forced_latency"); ok {
		withParams = append(withParams, fmt.Sprintf("forcedLatency=%s", forcedLatency.(string)))
	}
	if sizeLimit, ok := d.GetOk("size_limit"); ok {
		withParams = append(withParams, fmt.Sprintf("sizeLimit=%d", sizeLimit.(int)))
	}
	if distributed, ok := d.GetOk("distributed"); ok {
		withParams = append(withParams, fmt.Sprintf("distributed=%t", distributed.(bool)))
	}
	if parquetRowGroupSize, ok := d.GetOk("parquet_row_group_size"); ok {
		withParams = append(withParams, fmt.Sprintf("parquetRowGroupSize=%d", parquetRowGroupSize.(int)))
	}
	if useNativeParquetWriter, ok := d.GetOk("use_native_parquet_writer"); ok {
		withParams = append(withParams, fmt.Sprintf("useNativeParquetWriter='%s'", useNativeParquetWriter.(string)))
	}
	if managedIdentity, ok := d.GetOk("managed_identity"); ok {
		withParams = append(withParams, fmt.Sprintf("managedIdentity='%s'", managedIdentity.(string)))
	}
	if isDisabled, ok := d.GetOk("is_disabled"); ok {
		withParams = append(withParams, fmt.Sprintf("isDisabled=%t", isDisabled.(bool)))
	}

	withClause := ""
	if len(withParams) > 0 {
		withClause = fmt.Sprintf("with(%s)", strings.Join(withParams, ", "))
	}

	createStatement := fmt.Sprintf(".create-or-alter continuous-export %s to table %s %s <| %s", name, externalTableName, withClause, query)
	_, err := queryADXMgmt(ctx, meta, clusterConfig, databaseName, createStatement)
	if err != nil {
		return diag.Errorf("error creating continuous-export %s (Database %q): %+v", name, databaseName, err)
	}

	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}
	d.SetId(buildADXResourceId(client.Endpoint(), databaseName, "continuousexport", name))

	return resourceADXContinuousExportRead(ctx, d, meta)
}

func resourceADXContinuousExportRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXContinuousExportID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	showCommand := fmt.Sprintf(".show continuous-export %s | project Name, ExternalTableName, Query", id.Name)

	resultSet, diags := readADXEntity[ADXContinuousExport](ctx, meta, clusterConfig, id, showCommand, "continuous-export")
	if diags.HasError() {
		return diags
	}

	if len(resultSet) < 1 {
		d.SetId("")
	} else {

		d.Set("name", id.Name)
		d.Set("database_name", id.DatabaseName)

	}

	return diags
}

func resourceADXContinuousExportDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXContinuousExportID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, fmt.Sprintf(".drop continuous-export %s", id.Name))
}

func parseADXContinuousExportID(input string) (*adxResourceId, error) {
	return parseADXResourceID(input, 4, 0, 1, 2, 3)
}
