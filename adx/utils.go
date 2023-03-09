package adx

import (
	"context"
	"crypto"
	_ "crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/data/value"
	"github.com/Azure/azure-kusto-go/kusto/unsafe"

	"github.com/Azure/go-autorest/autorest/azure/auth"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

type adxAsyncOperationResp struct {
	OperationId value.GUID
}

type adxAsyncOperationsDetails struct {
	OperationId   value.GUID
	Operation     string
	NodeId        string
	StartedOn     value.DateTime
	LastUpdatedOn value.DateTime
	Duration      value.Timespan
	State         string
	Status        string
}

func readADXEntity[T any](ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, id *adxResourceId, query string, entityType string) ([]T, diag.Diagnostics) {
	var diags diag.Diagnostics

	resultSet, err := queryADXMgmtAndParse[T](ctx, meta, clusterConfig, id.DatabaseName, query)
	if err != nil {
		return resultSet, diag.Errorf("error reading adx entity: %+v", err)
	}

	return resultSet, diags
}

func parseADXResp[T any](resp *kusto.RowIterator, databaseName string) ([]T, error) {
	defer resp.Stop()

	var resultSet []T
	err := resp.Do(
		func(row *table.Row) error {
			result := new(T)
			if err := row.ToStruct(result); err != nil {
				return fmt.Errorf("error parsing adx query response (Database %q): %+v", databaseName, err)
			}
			resultSet = append(resultSet, *result)
			return nil
		})

	if err != nil {
		return nil, err
	}

	return resultSet, nil
}

func queryADXAndParse[T any](ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName string, query string) ([]T, error) {
	resp, err := queryADX(ctx, meta, clusterConfig, databaseName, query)
	if err != nil {
		return nil, err
	}
	return parseADXResp[T](resp, databaseName)
}

func queryADXMgmtAndParse[T any](ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName string, query string) ([]T, error) {
	resp, err := queryADXMgmt(ctx, meta, clusterConfig, databaseName, query)
	if err != nil {
		return nil, err
	}
	return parseADXResp[T](resp, databaseName)
}

func queryADXMgmt(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName string, query string) (*kusto.RowIterator, error) {
	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return nil, err
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	resp, err := client.Mgmt(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(query))
	if err != nil {
		return nil, fmt.Errorf("error executing adx mgmt query(%s) database(%q): %+v", query, databaseName, err)
	}
	return resp, nil
}

func queryADX(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName string, query string) (*kusto.RowIterator, error) {
	client, err := getADXClient(meta, clusterConfig)
	if err != nil {
		return nil, err
	}

	kStmtOpts := kusto.UnsafeStmt(unsafe.Stmt{Add: true})
	resp, err := client.Query(ctx, databaseName, kusto.NewStmt("", kStmtOpts).UnsafeAdd(query))
	if err != nil {
		return nil, fmt.Errorf("error executing adx query(%s) database(%q): %+v", query, databaseName, err)
	}
	return resp, nil
}

func deleteADXEntity(ctx context.Context, d *schema.ResourceData, meta interface{}, clusterConfig *ClusterConfig, databaseName string, deleteStatement string) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, err := queryADXMgmt(ctx, meta, clusterConfig, databaseName, deleteStatement)
	if err != nil {
		return diag.Errorf("error deleting adx entity: %+v", err)
	}
	defer resp.Stop()

	d.SetId("")
	return diags
}

func buildADXClient(clusterConfig *ClusterConfig) (*kusto.Client, error) {
	if len(clusterConfig.ClientID) == 0 {
		return nil, fmt.Errorf("client_id is required either in the resource or provider config")
	}
	if len(clusterConfig.ClientSecret) == 0 {
		return nil, fmt.Errorf("client_secret is required either in the resource or provider config")
	}
	if len(clusterConfig.TenantID) == 0 {
		return nil, fmt.Errorf("tenant_id is required either in the resource or provider config")
	}
	if len(clusterConfig.URI) == 0 {
		return nil, fmt.Errorf("uri is required either in the resource or provider config")
	}

	auth := kusto.Authorization{Config: auth.NewClientCredentialsConfig(clusterConfig.ClientID, clusterConfig.ClientSecret, clusterConfig.TenantID)}
	client, err := kusto.New(clusterConfig.URI, auth)
	if err != nil {
		return nil, fmt.Errorf("error creating adx client from config: %+v", err)
	}
	return client, nil
}

func getADXClient(meta interface{}, clusterConfig *ClusterConfig) (*kusto.Client, error) {

	meta.(*Meta).KustoClientsMapMU.RLock()
	client := getCachedADXClient(meta, clusterConfig)
	meta.(*Meta).KustoClientsMapMU.RUnlock()

	if client == nil {
		var err error
		client, err = buildADXClient(clusterConfig)
		if err != nil {
			return nil, err
		}
		meta.(*Meta).KustoClientsMapMU.Lock()
		setCachedADXClient(meta, clusterConfig, client)
		meta.(*Meta).KustoClientsMapMU.Unlock()
	}

	return client, nil
}

func setCachedADXClient(meta interface{}, clusterConfig *ClusterConfig, client *kusto.Client) {
	configHash := hashClusterConfig(clusterConfig)
	meta.(*Meta).KustoClientsMap[configHash] = client
}

func getCachedADXClient(meta interface{}, clusterConfig *ClusterConfig) *kusto.Client {
	configHash := hashClusterConfig(clusterConfig)
	return meta.(*Meta).KustoClientsMap[configHash]
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

func toADXTimespanLiteral(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName string, input string, expectedUnit string) (string, error) {
	// Expected unit can be d,h,m,s
	if input != "" && expectedUnit != "" {
		query := fmt.Sprintf("print Result=tostring(toint(totimespan('%s')/1%s))", input, expectedUnit)
		resultSet, err := queryADXAndParse[adxSimpleQueryResult](ctx, meta, clusterConfig, databaseName, query)
		if err != nil {
			return input, fmt.Errorf("error converting timespan literal: %+v", err)
		}
		return fmt.Sprintf("%s%s", resultSet[0].Result, expectedUnit), nil
	}
	return input, nil
}

func hashObjects(objs ...interface{}) []byte {
	digester := crypto.MD5.New()
	for _, ob := range objs {
		fmt.Fprint(digester, ob)
	}
	return digester.Sum(nil)
}

func isTableExists(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName, tableName string) (bool, error) {
	showStatement := fmt.Sprintf(".show tables (%s) details", tableName)
	return hasStatementResults(ctx, meta, clusterConfig, databaseName, showStatement, "checking if table exists")
}

func isMaterializedViewExists(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName, viewName string) (bool, error) {
	showStatement := fmt.Sprintf(".show materialized-views (%s) details", viewName)
	return hasStatementResults(ctx, meta, clusterConfig, databaseName, showStatement, "checking if materialized view exists")
}

func isFunctionExists(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName, functionName string) (bool, error) {
	showStatement := fmt.Sprintf(".show functions | where Name == '%s'", functionName)
	return hasStatementResults(ctx, meta, clusterConfig, databaseName, showStatement, "checking if function exists")
}

func isEntityExists(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName, entityType string, entityName string) (bool, error) {
	if entityType == "table" {
		return isTableExists(ctx, meta, clusterConfig, databaseName, entityName)
	} else if entityType == "materialized-view" {
		return isMaterializedViewExists(ctx, meta, clusterConfig, databaseName, entityName)
	} else if entityType == "function" {
		return isFunctionExists(ctx, meta, clusterConfig, databaseName, entityName)
	}
	return false, fmt.Errorf("checking for existance of entity type (%s) is not yet supported", entityType)
}

func hasStatementResults(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName, statement string, desc string) (bool, error) {
	resp, err := queryADXMgmt(ctx, meta, clusterConfig, databaseName, statement)
	if err != nil {
		return false, fmt.Errorf("error %s in database (%s): %+v", desc, databaseName, err)
	}
	defer resp.Stop()
	var hasResults bool
	err = resp.Do(
		func(row *table.Row) error {
			hasResults = true
			return nil
		})
	if err != nil {
		return false, fmt.Errorf("error %s in database (%s): %+v", desc, databaseName, err)
	}
	return hasResults, nil
}

func escapeEntityName(name string) string {
	escapedName := name
	if strings.Contains(name, "-") && !strings.HasPrefix(name, "[") {
		escapedName = fmt.Sprintf("['%s']", name)
	}
	return escapedName
}

func pollAsyncOperation(ctx context.Context, d *schema.ResourceData, meta interface{}, clusterConfig *ClusterConfig, databaseName string, operationId string, delay time.Duration, minTimeout time.Duration) (interface{}, error) {
	createWait := resource.StateChangeConf{
		Pending: []string{
			"Scheduled",
			"InProgress",
		},
		Target: []string{
			"Completed",
		},
		MinTimeout: minTimeout,
		Timeout:    d.Timeout(schema.TimeoutCreate),
		Delay:      delay,
		Refresh:    refreshStateAsyncOperation(ctx, meta, clusterConfig, databaseName, operationId),
	}
	return createWait.WaitForStateContext(ctx)
}

func refreshStateAsyncOperation(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName string, operationId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		query := fmt.Sprintf(".show operations %s", operationId)
		resultSet, err := queryADXMgmtAndParse[adxAsyncOperationsDetails](ctx, meta, clusterConfig, databaseName, query)
		if err != nil {
			return nil, "", fmt.Errorf("error checking status of operation %s: %+v", operationId, err)
		}
		return resultSet, resultSet[0].State, nil
	}
}
