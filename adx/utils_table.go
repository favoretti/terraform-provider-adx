package adx

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-kusto-go/kusto/data/table"
)

type adxTableResourceId struct {
	adxResourceId
}

type adxTableMappingResourceId struct {
	MappingName string
	Kind        string

	adxResourceId
}

func parseADXTableMappingID(input string) (*adxTableMappingResourceId, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 7 {
		return nil, fmt.Errorf("error parsing ADX resource ID: unexpected format: %q", input)
	}

	id := new(adxTableMappingResourceId)

	id.EndpointURI = parts[0]
	id.DatabaseName = parts[1]
	id.EntityType = parts[2]
	id.Name = parts[3]
	id.Kind = parts[5]
	id.MappingName = parts[6]

	return id, nil
}

func parseADXTableV0ID(input string) (*adxResourceId, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 3 {
		return nil, fmt.Errorf("error parsing ADX Table resource ID: unexpected format: %q", input)
	}

	return &adxResourceId{
		EndpointURI:  parts[0],
		DatabaseName: parts[1],
		EntityType:   "table",
		Name:         parts[2],
	}, nil
}

func parseADXTableMappingV0ID(input string) (*adxTableMappingResourceId, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 5 {
		return nil, fmt.Errorf("error parsing ADX Table Mapping resource ID: unexpected format: %q", input)
	}

	res := adxResourceId{
		EndpointURI:  parts[0],
		DatabaseName: parts[1],
		EntityType:   "ingestion_mapping",
		Name:         parts[2],
	}

	return &adxTableMappingResourceId{
		Kind:          parts[3],
		MappingName:   parts[4],
		adxResourceId: res,
	}, nil
}

func isTableExists(ctx context.Context, meta interface{}, clusterConfig *ClusterConfig, databaseName, tableName string) (bool, error) {
	showStatement := fmt.Sprintf(".show tables (%s) details", tableName)

	resp, err := queryADXMgmt(ctx, meta, clusterConfig, databaseName, showStatement)
	defer resp.Stop()
	if err != nil {
		return false, fmt.Errorf("error checking if table exists (%s) in database (%s): %+v", tableName, databaseName, err)
	}
	var exists bool
	err = resp.Do(
		func(row *table.Row) error {
			exists = true
			return nil
		})
	if err != nil {
		return false, fmt.Errorf("error checking if table exists (%s) in database (%s): %+v", tableName, databaseName, err)
	}
	return exists, nil
}
