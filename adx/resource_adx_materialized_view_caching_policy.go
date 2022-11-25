package adx

import (
	"context"
	"fmt"
	"regexp"

	"encoding/json"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type MaterializedViewCachingPolicy struct {
	DataHotSpan *PolicyStringValue
}

func resourceADXMaterializedViewCachingPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXMaterializedViewCachingPolicyCreateUpdate,
		ReadContext:   resourceADXMaterializedViewCachingPolicyRead,
		DeleteContext: resourceADXMaterializedViewCachingPolicyDelete,
		UpdateContext: resourceADXMaterializedViewCachingPolicyCreateUpdate,

		Schema: map[string]*schema.Schema{
			"cluster": getClusterConfigInputSchema(),
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"view_name": {
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

func resourceADXMaterializedViewCachingPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	viewName := d.Get("view_name").(string)
	databaseName := d.Get("database_name").(string)
	dataHotSpan := d.Get("data_hot_span").(string)
	followerDatabase := d.Get("follower_database").(bool)

	followerDatabaseClause := ""
	if followerDatabase {
		followerDatabaseClause = fmt.Sprintf("follower database %s", escapeEntityName(databaseName))
	}

	createStatement := fmt.Sprintf(".alter %s materialized-view %s policy caching hot = %s", followerDatabaseClause, viewName, dataHotSpan)

	if err := createADXPolicy(ctx, d, meta, "materialized-view", "caching", databaseName, viewName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXMaterializedViewCachingPolicyRead(ctx, d, meta)
}

func resourceADXMaterializedViewCachingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, resultSet, diags := readADXPolicy(ctx, d, meta, "materialized-view", "caching")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	var policy MaterializedViewCachingPolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing policy caching for materialized-view %q (Database %q): %+v", id.Name, id.DatabaseName, err)
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

	d.Set("view_name", id.Name)
	d.Set("database_name", id.DatabaseName)

	return diags
}

func resourceADXMaterializedViewCachingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "materialized-view", "caching")
}
