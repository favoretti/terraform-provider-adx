package adx

import (
	"strings"
	"context"
	"fmt"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
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

type adxTableResource struct {
	EndpointURI  string
	Name string
	DatabaseName string
}

type adxTableMappingResource struct {
	EndpointURI  string
	Name         string
	TableName string
	Kind string
	DatabaseName string
}

func parseADXTableID(input string) (*adxTableResource, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 3 {
		return nil, fmt.Errorf("error parsing ADX Table resource ID: unexpected format: %q", input)
	}

	return &adxTableResource{
		EndpointURI:  parts[0],
		DatabaseName: parts[1],
		Name:    parts[2],
	}, nil
}

func parseADXTableMappingID(input string) (*adxTableMappingResource, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 5 {
		return nil, fmt.Errorf("error parsing ADX Table resource ID: unexpected format: %q", input)
	}

	return &adxTableMappingResource{
		EndpointURI:  parts[0],
		DatabaseName: parts[1],
		TableName:    parts[2],
		Kind: parts[3],
		Name:         parts[4],
	}, nil
}

func parseADXTablePolicyID(input string) (*adxTableResource, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 4 {
		return nil, fmt.Errorf("error parsing ADX Table policy resource ID: unexpected format: %q", input)
	}

	return &adxTableResource{
		EndpointURI:  parts[0],
		DatabaseName: parts[1],
		Name:    parts[2],
	}, nil
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

func readADXPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}, entityType string, policyName string) (diag.Diagnostics, *adxTableResource, []TablePolicyResult) {
	var diags diag.Diagnostics

	client := meta.(*Meta).Kusto

	id, err := parseADXTablePolicyID(d.Id())
	if err != nil {
		return diag.FromErr(err),nil,nil
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	showStatement := fmt.Sprintf(".show %s %s policy %s", entityType, id.Name, policyName)

	resp, err := client.Mgmt(ctx, id.DatabaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(showStatement))
	if err != nil {
		return diag.Errorf("error reading %s %q policy %s (Database %q): %+v", entityType, id.Name, policyName, id.DatabaseName, err),id,nil
	}
	defer resp.Stop()

	var resultSet []TablePolicyResult
	err = resp.Do(
		func(row *table.Row) error {
			rec := TablePolicyResult{}
			if err := row.ToStruct(&rec); err != nil {
				return fmt.Errorf("error parsing %s %s for %s %q (Database %q): %+v", entityType, policyName, entityType, id.Name, id.DatabaseName, err)
			}
			resultSet = append(resultSet, rec)
			return nil
		},
	)

	if err != nil {
		return diag.Errorf("%+v", err), id, resultSet
	}

	return diags, id, resultSet
}

func deleteADXPolicy(ctx context.Context, d *schema.ResourceData, meta interface{}, entityType string, policyName string) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*Meta).Kusto

	id, err := parseADXTablePolicyID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	deleteStatement := fmt.Sprintf(".drop %s %s policy %s", entityType, id.Name, policyName)

	_, err = client.Mgmt(ctx, id.DatabaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(deleteStatement))
	if err != nil {
		return diag.Errorf("error deleting %s %q policy %s (Database %q): %+v", entityType, id.Name, policyName, id.DatabaseName, err)
	}

	d.SetId("")

	return diags
}