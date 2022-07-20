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

type adxResourceId struct {
	EndpointURI  string
	Name         string
	DatabaseName string
	EntityType   string
}

type adxSimpleQueryResult struct {
	Result string
}

func readADXEntity[T any](ctx context.Context, meta interface{}, id *adxResourceId, query string, entityType string) ([]T, diag.Diagnostics) {
	var diags diag.Diagnostics

	resultSet, err := queryADXMgmtAndParse[T](ctx, meta, id.DatabaseName, query)
	if err != nil {
		return resultSet, diag.Errorf("error reading adx entity: %+v", err)
	}

	if len(resultSet) < 1 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to load entity state from adx",
			Detail:   fmt.Sprintf("adx returned no results for(%s) (Database %q)", query, id.DatabaseName),
		  })
		return nil, diags
	}

	return resultSet, diags
}

func queryADXMgmtAndParse[T any](ctx context.Context, meta interface{}, databaseName string, query string) ([]T, error) {
	resp, err := queryADXMgmt(ctx, meta, databaseName, query)
	if err != nil {
		return nil, err
	}
	defer resp.Stop()

	var resultSet []T
	err = resp.Do(
		func(row *table.Row) error {
			result := new(T)
			if err := row.ToStruct(result); err != nil {
				return fmt.Errorf("error parsing adx query response (Database %q): %+v", databaseName, err)
			}
			resultSet = append(resultSet, *result)
			return nil
		})

	if err != nil {
		return nil, fmt.Errorf("error parsing adx result set: %+v", err)
	}

	return resultSet, nil
}

func queryADXMgmt(ctx context.Context, meta interface{}, databaseName string, query string) (*kusto.RowIterator, error) {
	client := meta.(*Meta).Kusto

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	resp, err := client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(query))
	if err != nil {
		return nil, fmt.Errorf("error executing adx mgmt query(%s) database(%q): %+v", query, databaseName, err)
	}
	return resp, nil
}

func queryADX(ctx context.Context, meta interface{}, databaseName string, query string) (*kusto.RowIterator, error) {
	client := meta.(*Meta).Kusto

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	resp, err := client.Query(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(query))
	if err != nil {
		return nil, fmt.Errorf("error executing adx query(%s) database(%q): %+v", query, databaseName, err)
	}
	return resp, nil
}

func deleteADXEntity(ctx context.Context, d *schema.ResourceData, meta interface{}, databaseName string, deleteStatement string) diag.Diagnostics {
    var diags diag.Diagnostics
	resp, err := queryADXMgmt(ctx, meta, databaseName, deleteStatement)
	if err!=nil {
		return diag.Errorf("error deleting adx entity: %+v", err)
	}
	defer resp.Stop()

	d.SetId("")
	return diags
}

func buildADXResourceId(endpoint string, params ...string) string {
	endpoint = strings.Replace(endpoint, "https://", "", 1)
	endpoint = strings.Replace(endpoint, "http://", "", 1)
	return endpoint + "|" + strings.Join(params[:], "|")
}

func parseADXResourceID(input string, expectedParts int, uriIndex int, dbNameIndex int, entityTypeIndex int, nameIndex int) (*adxResourceId, error) {
	parts := strings.Split(input, "|")
	if len(parts) != expectedParts {
		return nil, fmt.Errorf("error parsing ADX resource ID: unexpected format: %q", input)
	}

	return &adxResourceId{
		EndpointURI:  parts[uriIndex],
		DatabaseName: parts[dbNameIndex],
		EntityType:   parts[entityTypeIndex],
		Name:         parts[nameIndex],
	}, nil
}

func toADXTimespanLiteral(ctx context.Context, meta interface{}, databaseName string, input string, expectedUnit string) (string, error) {
	// Expected unit can be d,h,m,s
	if input != "" && expectedUnit != "" {
		query := fmt.Sprintf("print Result=tostring(toint(totimespan('%s')/1%s))", input, expectedUnit)
		resultSet, err := queryADXMgmtAndParse[adxSimpleQueryResult](ctx, meta, databaseName, query)
		if err != nil {
			return input, fmt.Errorf("error converting timespan literal: %+v", err)
		}
		return fmt.Sprintf("%s%s", resultSet[0].Result, expectedUnit), nil
	}
	return input, nil
}
