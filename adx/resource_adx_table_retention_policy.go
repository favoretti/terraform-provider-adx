package adx

import (
	"context"
	"fmt"
	"regexp"

	"encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TableRetentionPolicy struct {
	SoftDeletePeriod string
	Recoverability bool
}

func resourceADXTableRetentionPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableRetentionPolicyCreate,
		ReadContext:   resourceADXTableRetentionPolicyRead,
		DeleteContext: resourceADXTableRetentionPolicyDelete,

		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"table_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew: 		  true,
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"soft_delete_period": {
				Type:             schema.TypeString,
				Optional: false,
				ValidateDiagFunc: stringMatch(
					regexp.MustCompile("[0-9]{1,3}[dhms]"),
					"soft delete timespan must be in the format of <amount><unit> such as 1m for (one minute) or 30d (thirty days)",
					),
			},

			"recoverability": {
				Type:             schema.TypeBool,
				Optional: false,
			},
		},
	}
}

func resourceADXTableRetentionPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	sotDeleteTimespan := d.Get("soft_delete_period").(string)
	recoverability := d.Get("recoverability").(bool)

	createStatement := fmt.Sprintf(".alter-merge table %s policy retention softdelete = %s recoverability = disabled", tableName, sotDeleteTimespan, recoverability)

	if err := createADXPolicy(ctx, d, meta, "table","retention", databaseName, tableName, createStatement); err != nil {
		return err
	}

	return resourceADXTableRetentionPolicyRead(ctx, d, meta)
}

func resourceADXTableRetentionPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	err, id, resultSet:= readADXPolicy(ctx,d,meta,"table","retention"); 
	if err != nil {
		return diag.Errorf("%+v", err)
	}

	var policy TableRetentionPolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing policy retention for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	d.Set("table_name", id.Name)
	d.Set("database_name", id.DatabaseName)
	d.Set("soft_delete_period", policy.SoftDeletePeriod)
	d.Set("recoverability", policy.Recoverability)

	return diags
}

func resourceADXTableRetentionPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "retention")
}