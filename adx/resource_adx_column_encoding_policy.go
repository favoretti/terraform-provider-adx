package adx

import (
	"context"
	"fmt"

	"encoding/json"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ColumnEncodingPolicy struct {
	EntityIdentifier   string
	EncodingPolicyType string
}

func resourceADXColumnEncodingPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXColumnEncodingPolicyCreateUpdate,
		ReadContext:   resourceADXColumnEncodingPolicyRead,
		UpdateContext: resourceADXColumnEncodingPolicyCreateUpdate,
		DeleteContext: resourceADXColumnEncodingPolicyDelete,
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

			"entity_identifier": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"encoding_policy_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Null",
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXColumnEncodingPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	databaseName := d.Get("database_name").(string)
	entityIdentifier := d.Get("entity_identifier").(string)
	encodingPolicyType := d.Get("encoding_policy_type").(string)

	createStatement := fmt.Sprintf(".alter column %s policy encoding type='%s'", entityIdentifier, encodingPolicyType)

	if err := createADXPolicy(ctx, d, meta, "column", "encoding", databaseName, entityIdentifier, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXColumnEncodingPolicyRead(ctx, d, meta)
}

func resourceADXColumnEncodingPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, resultSet, diags := readADXPolicy(ctx, d, meta, "column", "encoding")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	var policy ColumnEncodingPolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing policy encoding for Column %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	d.Set("column_name", id.Name)
	d.Set("database_name", id.DatabaseName)
	d.Set("entity_identifier", policy.EntityIdentifier)
	d.Set("encoding_policy_type", policy.EncodingPolicyType)

	return diags
}

func resourceADXColumnEncodingPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	entityIdentifier := d.Get("entity_identifier").(string)

	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	id, err := parseADXPolicyID(d.Id())
	if err != nil {
		return diag.Errorf("could not delete adx policy due to error parsing ID: %+v", err)
	}

	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, fmt.Sprintf(".alter column %s policy encoding type='Null'", entityIdentifier)) //Encoding policy can't be deleted. So set it back to default.
}
