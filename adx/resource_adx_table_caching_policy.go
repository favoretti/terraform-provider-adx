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
		},
	}
}

func resourceADXTableCachingPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	dataHotSpan := d.Get("data_hot_span").(string)

	createStatement := fmt.Sprintf(".alter table %s policy caching hot = %s", tableName, dataHotSpan)

	if err := createADXPolicy(ctx, d, meta, "table", "caching", databaseName, tableName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXTableCachingPolicyRead(ctx, d, meta)
}

func resourceADXTableCachingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	id, resultSet, err := readADXPolicy(ctx, d, meta, "table", "caching")
	if err != nil {
		return diag.Errorf("%+v", err)
	}

	var policy TableCachingPolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing policy caching for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	originalDataHotSpan := d.Get("data_hot_span")

	if originalDataHotSpan != "" {
		originalDataHotSpanTimeUnit := originalDataHotSpan.(string)[len(originalDataHotSpan.(string))-1:]

		dataHotSpan, err := toADXTimespanLiteral(ctx, meta, id.DatabaseName, policy.DataHotSpan.Value, originalDataHotSpanTimeUnit)
		if err != nil {
			return diag.Errorf("%+v", err)
		}
		d.Set("data_hot_span", dataHotSpan)
	} else {
		d.Set("data_hot_span", policy.DataHotSpan.Value)
	}

	d.Set("table_name", id.Name)
	d.Set("database_name", id.DatabaseName)

	return diags
}

func resourceADXTableCachingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "caching")
}
