package adx

import (
	"context"
	"fmt"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type TablePolicyResult struct {
	PolicyName	string
	EntityName	string
	Policy   	string
	ChildEntities	string
	EntityType	string
}

func parseADXTablePolicyID(input string) (*adxResource, error) {
	return parseADXID(input,4,0,1,2)
}

func createADXPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}, entityType string, policyName string, databaseName string, entityName string, createStatement string) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Meta).Kusto
	
	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	_, err := client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(createStatement))
	if err != nil {
		return diag.Errorf("error creating %s %s Policy %q (Database %q): %+v", entityType, policyName, entityName, databaseName, err)
	}

	id := fmt.Sprintf("%s|%s|%s|%s", client.Endpoint(), databaseName, entityName, policyName)
	d.SetId(id)

	return diags
}

func readADXPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}, entityType string, policyName string) (diag.Diagnostics, *adxResource, []TablePolicyResult) {
	var diags diag.Diagnostics

	id, err := parseADXTablePolicyID(d.Id())
	if err != nil {
		return diag.FromErr(err),nil,nil
	}

	showCommand := fmt.Sprintf(".show %s %s policy %s", entityType, id.Name, policyName)
	
	resultErr, resultSet := readADXEntity[TablePolicyResult](ctx, d, meta, id, showCommand, entityType)
	if resultErr != nil {
		return diag.Errorf("%+v", resultErr), id, nil
	}

	return diags, id, resultSet
}

func deleteADXPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}, entityType string, policyName string) diag.Diagnostics {
	id, err := parseADXTablePolicyID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return deleteADXEntity(ctx,d,meta,id.DatabaseName, fmt.Sprintf(".delete %s %s policy %s", entityType, id.Name, policyName))
}