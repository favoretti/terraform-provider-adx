package adx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type MergePolicyLookback struct {
	Kind         string  `json:"Kind"`
	CustomPeriod *string `json:"CustomPeriod"`
}

type MergePolicy struct {
	RowCountUpperBoundForMerge       int                 `json:"RowCountUpperBoundForMerge"`
	OriginalSizeMBUpperBoundForMerge int                 `json:"OriginalSizeMBUpperBoundForMerge"`
	MaxExtentsToMerge                int                 `json:"MaxExtentsToMerge"`
	MaxRangeInHours                  int                 `json:"MaxRangeInHours"`
	AllowRebuild                     bool                `json:"AllowRebuild"`
	AllowMerge                       bool                `json:"AllowMerge"`
	Lookback                         MergePolicyLookback `json:"Lookback"`
}

func resourceADXMergePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXMergePolicyCreateUpdate,
		ReadContext:   resourceADXMergePolicyRead,
		UpdateContext: resourceADXMergePolicyCreateUpdate,
		DeleteContext: resourceADXMergePolicyDelete,
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

			"entity_type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringInSlice([]string{"database", "table", "materialized_view"}),
			},

			"entity_name": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"row_count_upper_bound_for_merge": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  16000000,
			},

			"original_size_mb_upper_bound_for_merge": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30000,
			},

			"max_extents_to_merge": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  100,
			},

			"max_range_in_hours": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  24,
			},

			"allow_rebuild": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"allow_merge": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"lookback_kind": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "Default",
				ValidateDiagFunc: validate.StringInSlice([]string{"Default", "All", "HotCache", "Custom"}),
			},

			"lookback_custom_period": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		CustomizeDiff: mergePolicyCustomizeDiff,
	}
}

func mergePolicyCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if err := clusterConfigCustomDiff(ctx, diff, meta); err != nil {
		return err
	}

	entityType := diff.Get("entity_type").(string)
	if entityType == "table" || entityType == "materialized_view" {
		entityName, ok := diff.GetOk("entity_name")
		if !ok || entityName.(string) == "" {
			return fmt.Errorf("entity_name is required when entity_type is %q", entityType)
		}
	}

	lookbackKind := diff.Get("lookback_kind").(string)
	if lookbackKind == "Custom" {
		customPeriod, ok := diff.GetOk("lookback_custom_period")
		if !ok || customPeriod.(string) == "" {
			return fmt.Errorf("lookback_custom_period is required when lookback_kind is \"Custom\"")
		}
	}

	return nil
}

func mergePolicyToKustoEntityType(entityType string) string {
	if entityType == "materialized_view" {
		return "materialized-view"
	}
	return entityType
}

func mergePolicyFromKustoEntityType(entityType string) string {
	if entityType == "materialized-view" {
		return "materialized_view"
	}
	return entityType
}

func resourceADXMergePolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	databaseName := d.Get("database_name").(string)
	entityType := d.Get("entity_type").(string)
	kustoEntityType := mergePolicyToKustoEntityType(entityType)

	var entityName string
	if entityType == "database" {
		entityName = databaseName
	} else {
		entityName = d.Get("entity_name").(string)
	}

	policy := MergePolicy{
		RowCountUpperBoundForMerge:       d.Get("row_count_upper_bound_for_merge").(int),
		OriginalSizeMBUpperBoundForMerge: d.Get("original_size_mb_upper_bound_for_merge").(int),
		MaxExtentsToMerge:                d.Get("max_extents_to_merge").(int),
		MaxRangeInHours:                  d.Get("max_range_in_hours").(int),
		AllowRebuild:                     d.Get("allow_rebuild").(bool),
		AllowMerge:                       d.Get("allow_merge").(bool),
		Lookback: MergePolicyLookback{
			Kind: d.Get("lookback_kind").(string),
		},
	}

	if v, ok := d.GetOk("lookback_custom_period"); ok && v.(string) != "" {
		customPeriod := v.(string)
		policy.Lookback.CustomPeriod = &customPeriod
	}

	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.Errorf("error serializing merge policy: %+v", err)
	}

	createStatement := fmt.Sprintf(".alter %s %s policy merge @'%s'", kustoEntityType, escapeEntityNameIfRequired(entityName), string(policyJSON))

	if diags := createADXPolicy(ctx, d, meta, kustoEntityType, "merge", databaseName, entityName, createStatement); diags != nil {
		return diags
	}

	return resourceADXMergePolicyRead(ctx, d, meta)
}

func resourceADXMergePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, err := parseADXPolicyID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing merge policy ID: %+v", err)
	}

	kustoEntityType := id.EntityType
	terraformEntityType := mergePolicyFromKustoEntityType(kustoEntityType)

	_, resultSet, diags := readADXPolicy(ctx, d, meta, kustoEntityType, "merge")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	if resultSet[0].Policy == "null" || resultSet[0].Policy == "" {
		d.SetId("")
		return diags
	}

	var policy MergePolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing merge policy for %s %q (Database %q): %+v", kustoEntityType, id.Name, id.DatabaseName, err)
	}

	d.Set("database_name", id.DatabaseName)
	d.Set("entity_type", terraformEntityType)
	if terraformEntityType != "database" {
		d.Set("entity_name", id.Name)
	}

	d.Set("row_count_upper_bound_for_merge", policy.RowCountUpperBoundForMerge)
	d.Set("original_size_mb_upper_bound_for_merge", policy.OriginalSizeMBUpperBoundForMerge)
	d.Set("max_extents_to_merge", policy.MaxExtentsToMerge)
	d.Set("max_range_in_hours", policy.MaxRangeInHours)
	d.Set("allow_rebuild", policy.AllowRebuild)
	d.Set("allow_merge", policy.AllowMerge)
	d.Set("lookback_kind", policy.Lookback.Kind)

	if policy.Lookback.CustomPeriod != nil {
		d.Set("lookback_custom_period", *policy.Lookback.CustomPeriod)
	} else {
		d.Set("lookback_custom_period", "")
	}

	return diags
}

func resourceADXMergePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, err := parseADXPolicyID(d.Id())
	if err != nil {
		return diag.Errorf("error parsing merge policy ID: %+v", err)
	}

	return deleteADXPolicy(ctx, d, meta, id.EntityType, "merge")
}
