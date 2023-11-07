package adx

import (
	"fmt"
	"strings"
)

type adxTableResourceId struct {
	adxResourceId
}

type adxTableMappingResourceId struct {
	MappingName string
	Kind        string

	adxResourceId
}

type adxTableContinuousExportResourceId struct {
	ExternalTableName string
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
