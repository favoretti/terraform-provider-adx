package adx

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TablePolicy struct {
	PolicyName    string
	EntityName    string
	Policy        string
	ChildEntities string
	EntityType    string
}

type PolicyStringValue struct {
	Value string
}

type adxPolicyResource struct {
	PolicyName string
	adxResourceId
}

func parseADXPolicyID(input string) (*adxPolicyResource, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 6 {
		return nil, fmt.Errorf("error parsing ADX resource ID: unexpected format: %q", input)
	}

	id := new(adxPolicyResource)

	id.EndpointURI = parts[0]
	id.DatabaseName = parts[1]
	id.EntityType = parts[2]
	id.Name = parts[3]
	id.PolicyName = parts[5]

	return id, nil
}

func createADXPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}, entityType string, policyName string, databaseName string, entityName string, createStatement string) diag.Diagnostics {
	var diags diag.Diagnostics
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return diag.Errorf("error creating adx client connection: %+v", err)
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	_, err = client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating %s %s Policy %q (Database %q): %+v", entityType, policyName, entityName, databaseName, err)
	}

	id := buildADXResourceId(client.Endpoint(), databaseName, entityType, entityName, "policy", policyName)
	d.SetId(id)

	return diags
}

func readADXPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}, entityType string, policyName string) (*adxPolicyResource, []TablePolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)

	id, err := parseADXPolicyID(d.Id())
	if err != nil {
		return nil, nil, diag.Errorf("could not read adx policy due to error parsing ID: %+v", err)
	}

	if entityExists, err := isEntityExists(ctx, meta, clusterConfig, id.DatabaseName, entityType, id.Name); err != nil || !entityExists {
		if err != nil {
			return id, nil, diag.Errorf("%+v", err)
		}
		d.SetId("")
		return id, nil, diags
	}

	showCommand := fmt.Sprintf(".show %s %s policy %s", entityType, id.Name, policyName)

	resultSet, diags := readADXEntity[TablePolicy](ctx, meta, clusterConfig, &id.adxResourceId, showCommand, entityType)
	if diags.HasError() {
		return id, nil, diag.Errorf("error reading adx policy")
	}
	if len(resultSet) == 0 {
		return id, nil, diag.Errorf("error: no results returned for policy %s for %s %q (Database %q)", policyName, entityType, id.Name, id.DatabaseName)
	}

	return id, resultSet, diags
}

func deleteADXPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}, entityType string, policyName string) diag.Diagnostics {
	clusterConfig := getAndExpandClusterConfigWithDefaults(ctx, d, meta)
	id, err := parseADXPolicyID(d.Id())
	if err != nil {
		return diag.Errorf("could not delete adx policy due to error parsing ID: %+v", err)
	}

	followerDatabaseClause := ""
	if followerDatabase, ok := d.GetOk("follower_database"); ok {
		if followerDatabase.(bool) {
			followerDatabaseClause = fmt.Sprintf("follower database %s", escapeEntityNameIfRequired(id.DatabaseName))
		}
	}

	return deleteADXEntity(ctx, d, meta, clusterConfig, id.DatabaseName, fmt.Sprintf(".delete %s %s %s policy %s", followerDatabaseClause, entityType, id.Name, policyName))
}

func policyCacheValueStateRefresh(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName string, entityType string, entityName string, expectedUnit string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cacheValue, err := getPolicyHotCacheValue(ctx, meta, clusterConfig, databaseName, entityType, entityName, expectedUnit)
		if err != nil {
			return "", "", err
		}
		return cacheValue, string(cacheValue), nil
	}
}

func getPolicyHotCacheValue(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName string, entityType string, entityName string, expectedUnit string) (string, error) {
	// Expected unit can be d,h,m,s
	query := fmt.Sprintf(".show %s %s policy caching | project Result=tostring(toint(totimespan(todynamic(Policy).DataHotSpan.Value)/1%s))", entityType, entityName, expectedUnit)
	resultSet, err := queryADXMgmtAndParse[adxSimpleQueryResult](ctx, meta, clusterConfig, databaseName, query)
	if err != nil {
		return "", fmt.Errorf("error checking hot cache value for %s %s: %+v", entityType, entityName, err)
	}
	return fmt.Sprintf("%s%s", resultSet[0].Result, expectedUnit), nil
}
