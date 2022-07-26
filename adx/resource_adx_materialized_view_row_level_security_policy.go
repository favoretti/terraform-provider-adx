package adx

import (
	"context"
	"fmt"

	"encoding/json"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type MaterializedViewRowLevelSecurityPolicy struct {
	IsEnabled bool
	Query     string
}

func resourceADXMaterializedViewRowLevelSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXMaterializedViewRowLevelSecurityPolicyCreateUpdate,
		ReadContext:   resourceADXMaterializedViewRowLevelSecurityPolicyRead,
		DeleteContext: resourceADXMaterializedViewRowLevelSecurityPolicyDelete,
		UpdateContext: resourceADXMaterializedViewRowLevelSecurityPolicyCreateUpdate,

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

			"query": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXMaterializedViewRowLevelSecurityPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	viewName := d.Get("view_name").(string)
	databaseName := d.Get("database_name").(string)
	query := d.Get("query").(string)
	enabled := d.Get("enabled").(bool)

	enabledString := "enable"
	if !enabled {
		enabledString = "disable"
	}

	createStatement := fmt.Sprintf(".alter materialized-view %s policy row_level_security %s \"%s\"", viewName, enabledString, query)

	if err := createADXPolicy(ctx, d, meta, "materialized-view", "row_level_security", databaseName, viewName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXMaterializedViewRowLevelSecurityPolicyRead(ctx, d, meta)
}

func resourceADXMaterializedViewRowLevelSecurityPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, resultSet, diags := readADXPolicy(ctx, d, meta, "materialized-view", "row_level_security")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	var policy MaterializedViewRowLevelSecurityPolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing policy retention for MaterializedView %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	d.Set("view_name", id.Name)
	d.Set("database_name", id.DatabaseName)
	d.Set("query", policy.Query)
	d.Set("enabled", policy.IsEnabled)

	return diags
}

func resourceADXMaterializedViewRowLevelSecurityPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "materialized-view", "row_level_security")
}
