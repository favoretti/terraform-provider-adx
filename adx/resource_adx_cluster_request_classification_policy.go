package adx

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ADXRequestClassificationPolicy struct {
	PolicyName    string
	EntityName    string
	Policy        string
	ChildEntities string
	EntityType    string
}

type requestClassificationPolicyObject struct {
	IsEnabled                bool     `json:"IsEnabled"`
	ClassificationFunction   string   `json:"ClassificationFunction,omitempty"`
	ClassificationProperties []string `json:"ClassificationProperties,omitempty"`
}

func resourceADXClusterRequestClassificationPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceADXClusterRequestClassificationPolicyCreateUpdate,
		ReadContext:   resourceADXClusterRequestClassificationPolicyRead,
		UpdateContext: resourceADXClusterRequestClassificationPolicyCreateUpdate,
		DeleteContext: resourceADXClusterRequestClassificationPolicyDelete,
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
				Description:      "Database name used as context for the management command. The policy is cluster-level.",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether the request classification policy is enabled.",
			},
			"classification_function": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
				DiffSuppressFunc: suppressClassificationFunctionDiff,
				Description:      "The body of the KQL function used for classifying requests into workload groups.",
			},
		},
		CustomizeDiff: clusterConfigCustomDiff,
	}
}

func resourceADXClusterRequestClassificationPolicyCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	databaseName := d.Get("database_name").(string)
	isEnabled := d.Get("is_enabled").(bool)
	classificationFunction := d.Get("classification_function").(string)

	policyJSON := fmt.Sprintf(`{"IsEnabled":%t}`, isEnabled)
	createStatement := fmt.Sprintf(".alter cluster policy request_classification '%s' <|\n%s", policyJSON, classificationFunction)

	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	_, err = client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating/updating cluster request classification policy (Database %q): %+v", databaseName, err)
	}

	d.SetId(buildADXResourceId(clusterConfig.URI, databaseName, "cluster", "request_classification"))

	return resourceADXClusterRequestClassificationPolicyRead(ctx, d, meta)
}

func resourceADXClusterRequestClassificationPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXClusterRequestClassificationPolicyID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	showCommand := ".show cluster policy request_classification"
	resultSet, diags := readADXEntity[ADXRequestClassificationPolicy](ctx, meta, clusterConfig, id, showCommand, "cluster")
	if diags.HasError() {
		return diags
	}

	if len(resultSet) < 1 || resultSet[0].Policy == "" || resultSet[0].Policy == "null" {
		d.SetId("")
		return diags
	}

	var policy requestClassificationPolicyObject
	if err := json.Unmarshal([]byte(resultSet[0].Policy), &policy); err != nil {
		return diag.Errorf("error parsing cluster request classification policy: %+v", err)
	}

	d.Set("database_name", id.DatabaseName)
	d.Set("is_enabled", policy.IsEnabled)
	d.Set("classification_function", policy.ClassificationFunction)

	return diags
}

func resourceADXClusterRequestClassificationPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXClusterRequestClassificationPolicyID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	deleteStatement := ".delete cluster policy request_classification"
	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, deleteStatement)
}

func parseADXClusterRequestClassificationPolicyID(input string) (*adxResourceId, error) {
	return parseADXResourceID(input, 4, 0, 1, 2, 3)
}

// suppressClassificationFunctionDiff compares classification functions by normalizing whitespace.
func suppressClassificationFunctionDiff(k, old, new string, d *schema.ResourceData) bool {
	return normalizeWhitespace(old) == normalizeWhitespace(new)
}

// normalizeWhitespace collapses all runs of whitespace into single spaces and trims.
func normalizeWhitespace(s string) string {
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}
