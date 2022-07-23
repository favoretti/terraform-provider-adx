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

type TableRetentionPolicy struct {
	SoftDeletePeriod string
	Recoverability   string
}

func resourceADXTableRetentionPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableRetentionPolicyCreateUpdate,
		ReadContext:   resourceADXTableRetentionPolicyRead,
		DeleteContext: resourceADXTableRetentionPolicyDelete,
		UpdateContext: resourceADXTableRetentionPolicyCreateUpdate,

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

			"soft_delete_period": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validate.StringMatch(
					regexp.MustCompile("[0-9]{1,3}[dhms]"),
					"soft delete timespan must be in the format of <amount><unit> such as 1m for (one minute) or 30d (thirty days)",
				),
			},

			"recoverability": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXTableRetentionPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	softDeleteTimespan := d.Get("soft_delete_period").(string)
	recoverability := d.Get("recoverability").(bool)

	recoverabilityString := "enabled"
	if !recoverability {
		recoverabilityString = "disabled"
	}

	createStatement := fmt.Sprintf(".alter-merge table %s policy retention softdelete = %s recoverability = %s", tableName, softDeleteTimespan, recoverabilityString)

	if err := createADXPolicy(ctx, d, meta, "table", "retention", databaseName, tableName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXTableRetentionPolicyRead(ctx, d, meta)
}

func resourceADXTableRetentionPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, resultSet, diags := readADXPolicy(ctx, d, meta, "table", "retention")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	var policy TableRetentionPolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing policy retention for Table %q (Database %q): %+v", id.Name, id.DatabaseName)
	}

	originalSoftDeletePeriod := d.Get("soft_delete_period")

	if originalSoftDeletePeriod != "" {
		// return diag.Errorf(originalSoftDeletePeriod.(string))

		originalSoftDeletePeriodTimeUnit := originalSoftDeletePeriod.(string)[len(originalSoftDeletePeriod.(string))-1:]

		softDeletePeriod, err := toADXTimespanLiteral(ctx, meta, clusterConfig, id.DatabaseName, policy.SoftDeletePeriod, originalSoftDeletePeriodTimeUnit)
		if err != nil {
			return diag.Errorf("%+v", err)
		}
		d.Set("soft_delete_period", softDeletePeriod)
	} else {
		d.Set("soft_delete_period", policy.SoftDeletePeriod)
	}

	d.Set("table_name", id.Name)
	d.Set("database_name", id.DatabaseName)

	recoverability := true
	if policy.Recoverability != "Enabled" {
		recoverability = false
	}

	d.Set("recoverability", recoverability)

	return diags
}

func resourceADXTableRetentionPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "retention")
}
