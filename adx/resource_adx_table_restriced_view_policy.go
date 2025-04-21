package adx

import (
	"context"
	"fmt"
	"strings"

	"encoding/json"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

type TableRestrictedViewPolicy struct {
	Enabled *PolicyStringValue
}

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

			"follower_database": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXTableRestrictedViewPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	enabled := d.Get("enabled").(string)
	followerDatabase := d.Get("follower_database").(bool)

	followerDatabaseClause := ""
	if followerDatabase {
		followerDatabaseClause = fmt.Sprintf("follower database %s", escapeEntityNameIfRequired(databaseName))
	}

	createStatement := fmt.Sprintf(".alter %s table %s policy restricted_view_access %s",
		followerDatabaseClause, tableName, enabled)

	if err := createADXPolicy(ctx, d, meta, "table", "restricted_view_access", databaseName, tableName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	// Setting policy for follower database appears to be eventually consistent.
	// Delay is sometimes up to 10 seconds before API returns new value
	if followerDatabase {
		clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
		createWait := resource.StateChangeConf{
			Target: []string{
				enabled,
			},
			MinTimeout: 5 * time.Second,
			Timeout:    d.Timeout(schema.TimeoutCreate) - time.Minute,
			Delay:      1 * time.Second,
			Refresh:    policyRestrictedViewStateRefresh(ctx, meta, clusterConfig, databaseName, "table", tableName),
		}
		if _, err := createWait.WaitForStateContext(ctx); err != nil {
			return diag.Errorf("waiting for the create/update of table %s policy restricted_view_access: %+v", tableName, err)
		}
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

		if policy.Enabled == nil {
			return diag.Errorf("invalid object returned for policy restricted_view for table %q (Database %q): %s", id.Name, id.DatabaseName, resultSet[0])
		}

		// Convert the policy value to lowercase to ensure it matches "true" or "false"
		enabled := strings.ToLower(policy.Enabled.Value)
		if enabled != "true" && enabled != "false" {
			enabled = "false" // Default to false if the value is unexpected
		}

		d.Set("enabled", enabled)
		d.Set("table_name", id.Name)
		d.Set("database_name", id.DatabaseName)
	}

	return diags
}

func resourceADXTableRestrictedViewPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "restricted_view_access")
}
