package adx

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type adxResource struct {
	EndpointURI  string
	Name string
	DatabaseName string
}

type adxSimpleQueryResult struct {
	Result string
}

func parseADXID(input string, expectedParts int, uriIndex int, dbNameIndex int, nameIndex int) (*adxResource, error) {
	parts := strings.Split(input, "|")
	if len(parts) != expectedParts {
		return nil, fmt.Errorf("error parsing ADX resource ID: unexpected format: %q", input)
	}

	return &adxResource{
		EndpointURI:  parts[uriIndex],
		DatabaseName: parts[dbNameIndex],
		Name:         parts[nameIndex],
	}, nil
}

func readADXEntity[T any](ctx context.Context, d *schema.ResourceData, meta interface{}, id *adxResource, query string, entityType string) (diag.Diagnostics, []T) {
	var diags diag.Diagnostics

	client := meta.(*Meta).Kusto

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})

	resp, err := client.Mgmt(ctx, id.DatabaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(query))
	if err != nil {
		return diag.Errorf("error reading %s %q (Database %q): %+v", entityType, id.Name, id.DatabaseName, err),nil
	}
	defer resp.Stop()

	var resultSet []T
	err = resp.Do(
		func(row *table.Row) error {
			result := new(T)
			if err := row.ToStruct(result); err != nil {
				return fmt.Errorf("error parsing %s %s (Database %q): %+v", entityType, id.Name, id.DatabaseName, err)
			}
			resultSet = append(resultSet, *result)
			return nil
	})

	if err != nil {
		return diag.Errorf("%+v", err), resultSet
	}

	if len(resultSet)<1 {
		return diag.Errorf("unable to load state from adx. adx returned no results for (%s) (Database %q)", query, id.DatabaseName), nil
	}

	return diags, resultSet
}

func queryADX[T any](ctx context.Context, d *schema.ResourceData, meta interface{}, databaseName string, query string) (diag.Diagnostics, []T) {
	var diags diag.Diagnostics

	client := meta.(*Meta).Kusto

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	resp, err := client.Query(ctx, databaseName,  kusto.NewStmt("", kStmtOpts).UnsafeAdd(query))
	if err != nil {
		return diag.Errorf("error executing adx query (Database %q): %+v", databaseName, err),nil
	}
	defer resp.Stop()

	var resultSet []T
	err = resp.Do(
		func(row *table.Row) error {
			result := new(T)
			if err := row.ToStruct(result); err != nil {
				return fmt.Errorf("error parsing query response (Database %q): %+v", databaseName, err)
			}
			resultSet = append(resultSet, *result)
			return nil
	})

	if err != nil {
		return diag.Errorf("%+v", err), nil
	}

	return diags, resultSet
}

func deleteADXEntity(ctx context.Context, d *schema.ResourceData, meta interface{}, databaseName string, deleteStatement string) diag.Diagnostics {
	var diags diag.Diagnostics

	client := meta.(*Meta).Kusto
	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})

	_, err := client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(deleteStatement))
	if err != nil {
		return diag.Errorf("error deleting (%s) (Database %q): %+v", deleteStatement, databaseName, err)
	}

	d.SetId("")

	return diags
}

func toADXTimespanLiteral(ctx context.Context, d *schema.ResourceData, meta interface{}, databaseName string, input string, expectedUnit string) (diag.Diagnostics, string) {
	// Expected unit can be d,h,m,s
	if input!= "" && expectedUnit!="" {
		query := fmt.Sprintf("print Result=tostring(toint(totimespan('%s')/1%s))",input,expectedUnit)
		resultErr, resultSet := queryADX[adxSimpleQueryResult](ctx, d, meta, databaseName, query)
		if resultErr != nil {
			return diag.Errorf("%+v", resultErr), ""
		}
		return nil,fmt.Sprintf("%s%s",resultSet[0].Result,expectedUnit)
	}
	return nil,input
}