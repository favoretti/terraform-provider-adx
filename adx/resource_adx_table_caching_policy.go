package adx

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"encoding/json"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TableCachingPolicy struct {
	DataHotSpan *PolicyStringValue
}

func resourceADXTableCachingPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableCachingPolicyCreateUpdate,
		ReadContext:   resourceADXTableCachingPolicyRead,
		DeleteContext: resourceADXTableCachingPolicyDelete,
		UpdateContext: resourceADXTableCachingPolicyCreateUpdate,

		Schema: map[string]*schema.Schema{
			"cluster": getClusterConfigInputSchema(),
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"table_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"data_hot_span": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validate.StringMatch(
					regexp.MustCompile("[0-9]{1,3}[dhms]"),
					"data_hot_span must be in the format of <amount><unit> such as 1m for (one minute) or 30d (thirty days)",
				),
			},

			"follower_database": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXTableCachingPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	dataHotSpan := d.Get("data_hot_span").(string)
	followerDatabase := d.Get("follower_database").(bool)

	followerDatabaseClause := ""
	if followerDatabase {
		followerDatabaseClause = fmt.Sprintf("follower database %s", escapeEntityName(databaseName))
	}

	createStatement := fmt.Sprintf(".alter %s table %s policy caching hot = %s", followerDatabaseClause, tableName, dataHotSpan)

	if err := createADXPolicy(ctx, d, meta, "table", "caching", databaseName, tableName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	// Setting cache for follower database appears to be eventually consistent.
	// Delay is sometimes up to 10 seconds before API returns new value
	if followerDatabase {
		dataHotSpanTimeUnit := dataHotSpan[len(dataHotSpan)-1:]
		clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
		createWait := resource.StateChangeConf{
			Target: []string{
				dataHotSpan,
			},
			MinTimeout: 5 * time.Second,
			Timeout:    d.Timeout(schema.TimeoutCreate) - time.Minute,
			Delay:      1 * time.Second,
			Refresh:    policyCacheValueStateRefresh(ctx, meta, clusterConfig, databaseName, "table", tableName, dataHotSpanTimeUnit),
		}
		if _, err := createWait.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("waiting for the create/update of table %s policy caching: %+v", tableName, err)
		}
	}

	return resourceADXTableCachingPolicyRead(ctx, d, meta)
}

func resourceADXTableCachingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, resultSet, diags := readADXPolicy(ctx, d, meta, "table", "caching")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	if resultSet[0].Policy == "null" {
		d.SetId("")
	} else {

		var policy TableCachingPolicy
		if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
			return diag.Errorf("error parsing policy caching for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
		}

		if policy.DataHotSpan == nil {
			return diag.Errorf("invalid object returned for policy caching for materialized-view %q (Database %q): %s", id.Name, id.DatabaseName, resultSet[0])
		}

		originalDataHotSpan := d.Get("data_hot_span")

		if originalDataHotSpan != "" {
			originalDataHotSpanTimeUnit := originalDataHotSpan.(string)[len(originalDataHotSpan.(string))-1:]

			dataHotSpan, err := toADXTimespanLiteral(ctx, meta, clusterConfig, id.DatabaseName, policy.DataHotSpan.Value, originalDataHotSpanTimeUnit)
			if err != nil {
				return diag.Errorf("%+v", err)
			}
			d.Set("data_hot_span", dataHotSpan)
		} else {
			d.Set("data_hot_span", policy.DataHotSpan.Value)
		}

		d.Set("table_name", id.Name)
		d.Set("database_name", id.DatabaseName)
	}

	return diags
}

func resourceADXTableCachingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "caching")
}
