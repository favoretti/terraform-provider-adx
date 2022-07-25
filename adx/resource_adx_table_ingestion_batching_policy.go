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

type TableIngestionBatchingPolicy struct {
	MaximumBatchingTimeSpan string
	MaximumNumberOfItems    int
	MaximumRawDataSizeMB    int
}

func resourceADXTableIngestionBatchingPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableIngestionBatchingPolicyCreateUpdate,
		ReadContext:   resourceADXTableIngestionBatchingPolicyRead,
		DeleteContext: resourceADXTableIngestionBatchingPolicyDelete,
		UpdateContext: resourceADXTableIngestionBatchingPolicyCreateUpdate,

		Schema: map[string]*schema.Schema{
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

			"max_batching_timespan": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validate.StringMatch(
					regexp.MustCompile("\\d\\d:\\d\\d:\\d\\d"),
					"batching timespan must be in the format HH:MM:SS of ex. 00:10:00 for 10 minutes",
				),
			},

			"max_items": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"max_raw_size_mb": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceADXTableIngestionBatchingPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	maxBatchingTimespan := d.Get("max_batching_timespan").(string)
	maxNumberItems := d.Get("max_items").(int)
	maxRawSizeMb := d.Get("max_raw_size_mb").(int)

	createStatement := fmt.Sprintf(".alter tables (%s) policy ingestionbatching @'{\"MaximumBatchingTimeSpan\": \"%s\",\"MaximumNumberOfItems\": %d, \"MaximumRawDataSizeMB\": %d}'", tableName, maxBatchingTimespan, maxNumberItems, maxRawSizeMb)

	if err := createADXPolicy(ctx, d, meta, "table", "ingestionbatching", databaseName, tableName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXTableIngestionBatchingPolicyRead(ctx, d, meta)
}

func resourceADXTableIngestionBatchingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, resultSet, diags := readADXPolicy(ctx, d, meta, "table", "ingestionbatching")
	if diags.HasError() {
		return diags
	}

	var policy TableIngestionBatchingPolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing policy ingestionbatching for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	d.Set("table_name", id.Name)
	d.Set("database_name", id.DatabaseName)
	d.Set("max_batching_timespan", policy.MaximumBatchingTimeSpan)
	d.Set("max_items", policy.MaximumNumberOfItems)
	d.Set("max_raw_size_mb", policy.MaximumRawDataSizeMB)

	return diags
}

func resourceADXTableIngestionBatchingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "ingestionbatching")
}
