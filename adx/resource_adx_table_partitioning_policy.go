package adx

import (
	"context"
	"fmt"

	"encoding/json"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TablePartitioningPolicy struct {
	PartitionKeys     *[]TablePartitioningPolicyKey
	EffectiveDateTime string
}

type TablePartitioningPolicyKey struct {
	ColumnName string
	Kind       string
	Properties *TablePartitioningPolicyKeyProperties
}

type TablePartitioningPolicyKeyProperties struct {
	Function                string
	MaxPartitionCount       int
	Seed                    int
	PartitionAssignmentMode string
	Reference               string
	RangeSize               string
	OverrideCreationTime    bool
}

func resourceADXTablePartitioningPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXTablePartitioningPolicyCreateUpdate,
		ReadContext:   resourceADXTablePartitioningPolicyRead,
		DeleteContext: resourceADXTablePartitioningPolicyDelete,
		UpdateContext: resourceADXTablePartitioningPolicyCreateUpdate,

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

			"effective_date_time": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"partition_key": {
				Type:     schema.TypeList,
				MaxItems: 2,
				MinItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"column_name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validate.StringIsNotEmpty,
						},
						"kind": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validate.StringIsNotEmpty,
						},
						"hash_properties": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"function": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validate.StringIsNotEmpty,
									},
									"max_partition_count": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"seed": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"partition_assignment_mode": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"uniform_range_properties": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"reference": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validate.StringIsNotEmpty,
									},
									"range_size": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validate.StringIsNotEmpty,
									},
									"override_creation_time": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXTablePartitioningPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tableName := d.Get("table_name").(string)
	databaseName := d.Get("database_name").(string)
	effectiveDateTime := d.Get("effective_date_time").(string)
	partitionKeys := d.Get("partition_key").([]interface{})

	err, adxPartitionKeys := expandPartitionKeys(partitionKeys)
	if err != nil {
		return diag.Errorf("%+v", err)
	}

	adxPolicy := &TablePartitioningPolicy{
		PartitionKeys:     adxPartitionKeys,
		EffectiveDateTime: effectiveDateTime,
	}

	policyJson, marshErr := json.Marshal(adxPolicy)
	if marshErr != nil {
		return diag.Errorf("%+v", marshErr)
	}

	createStatement := fmt.Sprintf(".alter table %s policy partitioning ```%s```", tableName, policyJson)

	if err := createADXPolicy(ctx, d, meta, "table", "partitioning", databaseName, tableName, createStatement); err != nil {
		return diag.Errorf("%+v", err)
	}

	return resourceADXTablePartitioningPolicyRead(ctx, d, meta)
}

func resourceADXTablePartitioningPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	id, resultSet, diags := readADXPolicy(ctx, d, meta, "table", "partitioning")
	if diags.HasError() || resultSet == nil || len(resultSet) == 0 {
		return diags
	}

	var policy TablePartitioningPolicy
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing policy partitioning for Table %q (Database %q): %+v", id.Name, id.DatabaseName, err)
	}

	d.Set("table_name", id.Name)
	d.Set("database_name", id.DatabaseName)
	d.Set("effective_date_time", policy.EffectiveDateTime)
	d.Set("partition_key", flattenPartitionKeys(policy.PartitionKeys))

	return diags
}

func resourceADXTablePartitioningPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteADXPolicy(ctx, d, meta, "table", "partitioning")
}

func expandPartitionKeys(partitionKeys []interface{}) (diag.Diagnostics, *[]TablePartitioningPolicyKey) {
	adxPartitionKeys := make([]TablePartitioningPolicyKey, 0)
	for _, v := range partitionKeys {
		partitionKey := v.(map[string]interface{})
		kind := partitionKey["kind"]

		var adxProperties TablePartitioningPolicyKeyProperties
		if kind == "Hash" {
			hashPropertiesArray := partitionKey["hash_properties"].([]interface{})
			if hashPropertiesArray == nil || len(hashPropertiesArray) == 0 {
				return diag.Errorf("hash_properties block is required when partition kind is (Hash)"), nil
			}
			hashProperties := hashPropertiesArray[0].(map[string]interface{})
			adxProperties.Function = hashProperties["function"].(string)
			adxProperties.MaxPartitionCount = hashProperties["max_partition_count"].(int)
			adxProperties.Seed = hashProperties["seed"].(int)
			adxProperties.PartitionAssignmentMode = hashProperties["partition_assignment_mode"].(string)
		} else if kind == "UniformRange" {
			uniformRangePropertiesArray := partitionKey["uniform_range_properties"].([]interface{})
			if uniformRangePropertiesArray == nil || len(uniformRangePropertiesArray) == 0 {
				return diag.Errorf("uniform_range_properties block is required when partition kind is (UniformRange)"), nil
			}
			uniformRangeProperties := uniformRangePropertiesArray[0].(map[string]interface{})
			adxProperties.Reference = uniformRangeProperties["reference"].(string)
			adxProperties.RangeSize = uniformRangeProperties["range_size"].(string)
			adxProperties.OverrideCreationTime = uniformRangeProperties["override_creation_time"].(bool)
		} else {
			return diag.Errorf("partition key kind (%s) is unsupported", kind), nil
		}

		adxPartitionKey := TablePartitioningPolicyKey{
			ColumnName: partitionKey["column_name"].(string),
			Kind:       partitionKey["kind"].(string),
			Properties: &adxProperties,
		}
		adxPartitionKeys = append(adxPartitionKeys, adxPartitionKey)
	}
	return nil, &adxPartitionKeys
}

func flattenPartitionKeys(adxPartitionKeys *[]TablePartitioningPolicyKey) []interface{} {

	if adxPartitionKeys != nil {
		partitionKeys := make([]interface{}, len(*adxPartitionKeys), len(*adxPartitionKeys))

		for i, adxPartitionKey := range *adxPartitionKeys {
			partitionKey := make(map[string]interface{})
			partitionKey["column_name"] = adxPartitionKey.ColumnName
			partitionKey["kind"] = adxPartitionKey.Kind

			if adxPartitionKey.Kind == "Hash" {
				hashProperties := make(map[string]interface{})
				hashProperties["function"] = adxPartitionKey.Properties.Function
				hashProperties["max_partition_count"] = adxPartitionKey.Properties.MaxPartitionCount
				hashProperties["seed"] = adxPartitionKey.Properties.Seed
				hashProperties["partition_assignment_mode"] = adxPartitionKey.Properties.PartitionAssignmentMode
				partitionKey["hash_properties"] = [1]map[string]interface{}{hashProperties}
			} else if adxPartitionKey.Kind == "UniformRange" {
				uniformRangeProperties := make(map[string]interface{})
				uniformRangeProperties["reference"] = adxPartitionKey.Properties.Reference
				uniformRangeProperties["range_size"] = adxPartitionKey.Properties.RangeSize
				uniformRangeProperties["override_creation_time"] = adxPartitionKey.Properties.OverrideCreationTime
				partitionKey["uniform_range_properties"] = [1]map[string]interface{}{uniformRangeProperties}
			}

			partitionKeys[i] = partitionKey
		}

		return partitionKeys
	}

	return make([]interface{}, 0)
}
