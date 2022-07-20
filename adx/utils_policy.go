package adx

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
	client := meta.(*Meta).Kusto

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	_, err := client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating %s %s Policy %q (Database %q): %+v", entityType, policyName, entityName, databaseName, err)
	}

	id := buildADXResourceId(client.Endpoint(), databaseName, entityType, entityName, "policy", policyName)
	d.SetId(id)

	return diags
}

func readADXPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}, entityType string, policyName string) (*adxPolicyResource, []TablePolicy, diag.Diagnostics) {
	var diags diag.Diagnostics
	id, err := parseADXPolicyID(d.Id())
	if err != nil {
		return nil, nil, diag.Errorf("could not read adx policy due to error parsing ID: %+v", err)
	}

	showCommand := fmt.Sprintf(".show %s %s policy %s", entityType, id.Name, policyName)

	resultSet, diags := readADXEntity[TablePolicy](ctx, meta, &id.adxResourceId, showCommand, entityType)
	if diags.HasError() {
		return id, nil, diag.Errorf("error reading adx policy")
	}

	return id, resultSet, diags
}

func deleteADXPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}, entityType string, policyName string) diag.Diagnostics {
	id, err := parseADXPolicyID(d.Id())
	if err != nil {
		return diag.Errorf("could not delete adx policy due to error parsing ID: %+v", err)
	}

	return deleteADXEntity(ctx, d, meta, id.DatabaseName, fmt.Sprintf(".delete %s %s policy %s", entityType, id.Name, policyName))
}
