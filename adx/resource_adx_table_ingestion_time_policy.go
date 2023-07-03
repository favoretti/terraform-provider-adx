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

type IngestionTimePolicy struct {
	IsEnabled bool
}

func resourceADXTableIngestionTimePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTableIngestionTimeCreateUpdate,
		ReadContext:   resourceADXTableIngestionTimeRead,
		DeleteContext: resourceADXTableIngestionTimeDelete,
		UpdateContext: resourceADXTableIngestionTimeCreateUpdate,
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
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXTableIngestionTimeCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	enabled := d.Get("enabled").(bool)

	enabledString := strconv.FormatBool(enabled)

	createStatement := fmt.Sprintf(".alter table %s policy ingestiontime %s", tableName, enabledString)

	if err := createADXPolicy(ctx, d, meta, "table", "ingestiontime", databaseName, tableName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXTableIngestionTimeRead(ctx, d, meta)
}

func resourceADXTableIngestionTimeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, resultSet, diags := readADXPolicy(ctx, d, meta, "table", "ingestiontime")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	if resultSet[0].Policy == "null" {
		d.SetId("")
	} else {

		var policy IngestionTimePolicy
		if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
			return diag.Errorf("error parsing policy ingestiontime for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
		}

		d.Set("table_name", id.Name)
		d.Set("database_name", id.DatabaseName)
		d.Set("enabled", policy.IsEnabled)
	}

	return diags
}

func resourceADXTableIngestionTimeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "ingetionstime")
}
