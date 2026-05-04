package adx

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ADXWorkloadGroup struct {
	WorkloadGroupName string
	WorkloadGroup     string
}

func resourceADXWorkloadGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXWorkloadGroupCreateUpdate,
		ReadContext:   resourceADXWorkloadGroupRead,
		UpdateContext: resourceADXWorkloadGroupCreateUpdate,
		DeleteContext: resourceADXWorkloadGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cluster": getClusterConfigInputSchema(),
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
				Description:      "Database name used as context for the management command. Workload groups are cluster-level resources.",
			},
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
				Description:      "Name of the workload group.",
			},
			"request_limits_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validate.StringIsJSON,
				DiffSuppressFunc: suppressJSONDiff,
				Description:      "JSON representation of the request limits policy.",
			},
			"request_rate_limit_policies": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validate.StringIsJSON,
				DiffSuppressFunc: suppressJSONDiff,
				Description:      "JSON representation of the request rate limit policies array.",
			},
			"request_rate_limits_enforcement_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validate.StringIsJSON,
				DiffSuppressFunc: suppressJSONDiff,
				Description:      "JSON representation of the request rate limits enforcement policy.",
			},
			"request_queuing_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validate.StringIsJSON,
				DiffSuppressFunc: suppressJSONDiff,
				Description:      "JSON representation of the request queuing policy.",
			},
			"query_consistency_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validate.StringIsJSON,
				DiffSuppressFunc: suppressJSONDiff,
				Description:      "JSON representation of the query consistency policy.",
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXWorkloadGroupCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	name := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	policyObject := buildWorkloadGroupPolicyObject(d)

	escapedName := escapeEntityNameIfRequired(name)
	createStatement := fmt.Sprintf(".create-or-alter workload_group %s ```\n%s\n```", escapedName, policyObject)

	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	_, err = client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating/updating workload group %q (Database %q): %+v", name, databaseName, err)
	}

	d.SetId(buildADXResourceId(clusterConfig.URI, databaseName, "workload_group", name))

	return resourceADXWorkloadGroupRead(ctx, d, meta)
}

func resourceADXWorkloadGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXWorkloadGroupID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	showCommand := fmt.Sprintf(".show workload_group %s", escapeEntityNameIfRequired(id.Name))
	resultSet, diags := readADXEntity[ADXWorkloadGroup](ctx, meta, clusterConfig, id, showCommand, "workload_group")
	if diags.HasError() {
		return diags
	}

	if len(resultSet) < 1 {
		d.SetId("")
		return diags
	}

	d.Set("name", resultSet[0].WorkloadGroupName)
	d.Set("database_name", id.DatabaseName)

	flattenWorkloadGroupPolicies(d, resultSet[0].WorkloadGroup)

	return diags
}

func resourceADXWorkloadGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXWorkloadGroupID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deleteStatement := fmt.Sprintf(".drop workload_group %s", escapeEntityNameIfRequired(id.Name))
	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, deleteStatement)
}

func parseADXWorkloadGroupID(input string) (*adxResourceId, error) {
	return parseADXResourceID(input, 4, 0, 1, 2, 3)
}

func buildWorkloadGroupPolicyObject(d *schema.ResourceData) string {
	var parts []string

	if v, ok := d.GetOk("request_limits_policy"); ok {
		parts = append(parts, fmt.Sprintf(`"RequestLimitsPolicy": %s`, normalizeJSON(v.(string))))
	}
	if v, ok := d.GetOk("request_rate_limit_policies"); ok {
		parts = append(parts, fmt.Sprintf(`"RequestRateLimitPolicies": %s`, normalizeJSON(v.(string))))
	}
	if v, ok := d.GetOk("request_rate_limits_enforcement_policy"); ok {
		parts = append(parts, fmt.Sprintf(`"RequestRateLimitsEnforcementPolicy": %s`, normalizeJSON(v.(string))))
	}
	if v, ok := d.GetOk("request_queuing_policy"); ok {
		parts = append(parts, fmt.Sprintf(`"RequestQueuingPolicy": %s`, normalizeJSON(v.(string))))
	}
	if v, ok := d.GetOk("query_consistency_policy"); ok {
		parts = append(parts, fmt.Sprintf(`"QueryConsistencyPolicy": %s`, normalizeJSON(v.(string))))
	}

	return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
}

func flattenWorkloadGroupPolicies(d *schema.ResourceData, workloadGroupJSON string) {
	policies := parseWorkloadGroupJSON(workloadGroupJSON)

	policyFields := map[string]string{
		"RequestLimitsPolicy":                "request_limits_policy",
		"RequestRateLimitPolicies":           "request_rate_limit_policies",
		"RequestRateLimitsEnforcementPolicy": "request_rate_limits_enforcement_policy",
		"RequestQueuingPolicy":               "request_queuing_policy",
		"QueryConsistencyPolicy":             "query_consistency_policy",
	}

	for apiKey, tfKey := range policyFields {
		if v, ok := policies[apiKey]; ok && v != "null" {
			d.Set(tfKey, v)
		}
	}
}

// parseWorkloadGroupJSON parses the workload group JSON and extracts individual policy sections.
func parseWorkloadGroupJSON(input string) map[string]string {
	result := make(map[string]string)

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal([]byte(input), &parsed); err != nil {
		return result
	}

	for key, val := range parsed {
		if string(val) == "null" {
			result[key] = "null"
			continue
		}
		result[key] = string(val)
	}

	return result
}

// suppressJSONDiff suppresses diffs when the shared keys between config and state
// have equal values. This handles:
// - Whitespace/formatting differences (compact vs pretty JSON)
// - API returning extra server-side default fields not in the user config
// - API not returning fields that the user set (accepted but not echoed back)
// A real change (user modifies a value) is still detected because shared keys will differ.
func suppressJSONDiff(k, old, new string, d *schema.ResourceData) bool {
	var oldJSON, newJSON interface{}
	if err := json.Unmarshal([]byte(old), &oldJSON); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &newJSON); err != nil {
		return false
	}
	return jsonIsSubset(newJSON, oldJSON) || jsonIsSubset(oldJSON, newJSON)
}

// jsonIsSubset checks if 'subset' is semantically contained in 'superset'.
// For objects: every key in subset must exist in superset with a matching value.
// For arrays: must be deeply equal.
// For scalars: must be deeply equal.
func jsonIsSubset(subset, superset interface{}) bool {
	subMap, subOk := subset.(map[string]interface{})
	supMap, supOk := superset.(map[string]interface{})
	if subOk && supOk {
		for key, subVal := range subMap {
			supVal, exists := supMap[key]
			if !exists || !jsonIsSubset(subVal, supVal) {
				return false
			}
		}
		return true
	}
	return reflect.DeepEqual(subset, superset)
}

// normalizeJSON compacts JSON to a canonical form for consistent API calls.
func normalizeJSON(input string) string {
	var parsed interface{}
	if err := json.Unmarshal([]byte(input), &parsed); err != nil {
		return input
	}
	result, err := json.Marshal(parsed)
	if err != nil {
		return input
	}
	return string(result)
}
