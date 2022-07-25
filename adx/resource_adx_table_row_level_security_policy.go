package adx

import (
	"context"
	"fmt"

	"encoding/json"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TableRowLevelSecurityPolicy struct {
	IsEnabled bool
	Query     string
}

func resourceADXTableRowLevelSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableRowLevelSecurityPolicyCreateUpdate,
		ReadContext:   resourceADXTableRowLevelSecurityPolicyRead,
		DeleteContext: resourceADXTableRowLevelSecurityPolicyDelete,
		UpdateContext: resourceADXTableRowLevelSecurityPolicyCreateUpdate,

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

func resourceADXTableRowLevelSecurityPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	query := d.Get("query").(string)
	enabled := d.Get("enabled").(bool)

	enabledString := "enable"
	if !enabled {
		enabledString = "disable"
	}

	createStatement := fmt.Sprintf(".alter table %s policy row_level_security %s \"%s\"", tableName, enabledString, query)

	if err := createADXPolicy(ctx, d, meta, "table", "row_level_security", databaseName, tableName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXTableRowLevelSecurityPolicyRead(ctx, d, meta)
}

func resourceADXTableRowLevelSecurityPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, resultSet, diags := readADXPolicy(ctx, d, meta, "table", "row_level_security")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	var policy TableRowLevelSecurityPolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing policy retention for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	d.Set("table_name", id.Name)
	d.Set("database_name", id.DatabaseName)
	d.Set("query", policy.Query)
	d.Set("enabled", policy.IsEnabled)

	return diags
}

func resourceADXTableRowLevelSecurityPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "row_level_security")
}
