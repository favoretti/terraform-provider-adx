package adx

import (
	"context"
	"fmt"
	"strconv"

	"encoding/json"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TableRestrictedViewPolicy struct {
	IsEnabled *bool
}

// This resource intentionally has no follower_database option: ADX only supports
// overriding caching policy and authorized principals on a follower database (see
// https://learn.microsoft.com/azure/data-explorer/kusto/management/cluster-follower),
// there is no .alter follower database ... policy restricted_view_access command.
func resourceADXTableRestrictedViewPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableRestrictedViewPolicyCreateUpdate,
		ReadContext:   resourceADXTableRestrictedViewPolicyRead,
		DeleteContext: resourceADXTableRestrictedViewPolicyDelete,
		UpdateContext: resourceADXTableRestrictedViewPolicyCreateUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

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

			"enabled": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validate.StringInSlice([]string{
					"true",
					"false",
				}),
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXTableRestrictedViewPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	enabled := d.Get("enabled").(string)

	createStatement := fmt.Sprintf(".alter table %s policy restricted_view_access %s", tableName, enabled)

	if err := createADXPolicy(ctx, d, meta, "table", "restricted_view_access", databaseName, tableName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXTableRestrictedViewPolicyRead(ctx, d, meta)
}

func resourceADXTableRestrictedViewPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, resultSet, diags := readADXPolicy(ctx, d, meta, "table", "restricted_view_access")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	if resultSet[0].Policy == "null" {
		d.SetId("")
	} else {
		var policy TableRestrictedViewPolicy
		if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
			return diag.Errorf("error parsing policy restricted_view for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
		}

		if policy.IsEnabled == nil {
			return diag.Errorf("invalid object returned for policy restricted_view for table %q (Database %q): %s", id.Name, id.DatabaseName, resultSet[0])
		}

		d.Set("enabled", strconv.FormatBool(*policy.IsEnabled))
		d.Set("table_name", id.Name)
		d.Set("database_name", id.DatabaseName)
	}

	return diags
}

func resourceADXTableRestrictedViewPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "restricted_view_access")
}
