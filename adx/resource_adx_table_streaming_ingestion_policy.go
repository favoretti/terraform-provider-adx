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

type TableStreamingIngestionPolicy struct {
	IsEnabled         bool
	HintAllocatedRate string
}

func resourceADXTableStreamingIngestionPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableStreamingIngestionPolicyCreateUpdate,
		ReadContext:   resourceADXTableStreamingIngestionPolicyRead,
		DeleteContext: resourceADXTableStreamingIngestionPolicyDelete,
		UpdateContext: resourceADXTableStreamingIngestionPolicyCreateUpdate,
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
				Type:     schema.TypeBool,
				Required: true,
			},

			"hint_allocated_rate": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					oldFloat, err := strconv.ParseFloat(old, 64)
					if err != nil {
						return false
					}
					newFloat, err := strconv.ParseFloat(new, 64)
					if err != nil {
						return false
					}
					return oldFloat == newFloat
				},
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXTableStreamingIngestionPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	enabled := d.Get("enabled").(bool)

	hintAllocatedRateString := "null"
	if hintAllocatedRate, ok := d.GetOk("hint_allocated_rate"); ok {
		hintAllocatedRateString = fmt.Sprintf("\"%s\"", hintAllocatedRate)
	}

	enabledString := strconv.FormatBool(enabled)

	createStatement := fmt.Sprintf(".alter table %s policy streamingingestion '{\"IsEnabled\": %s, \"HintAllocatedRate\": %s}'", tableName, enabledString, hintAllocatedRateString)

	if err := createADXPolicy(ctx, d, meta, "table", "streamingingestion", databaseName, tableName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXTableStreamingIngestionPolicyRead(ctx, d, meta)
}

func resourceADXTableStreamingIngestionPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, resultSet, diags := readADXPolicy(ctx, d, meta, "table", "streamingingestion")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	if resultSet[0].Policy == "null" {
		d.SetId("")
	} else {

		var policy TableStreamingIngestionPolicy
		if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
			return diag.Errorf("error parsing policy streamingingestion for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
		}

		d.Set("table_name", id.Name)
		d.Set("database_name", id.DatabaseName)
		d.Set("enabled", policy.IsEnabled)
		d.Set("hint_allocated_rate", policy.HintAllocatedRate)
	}

	return diags
}

func resourceADXTableStreamingIngestionPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "streamingingestion")
}
