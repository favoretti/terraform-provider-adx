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

func parseADXTableID(input string) (*adxResourceId, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 3 {
		return nil, fmt.Errorf("error parsing ADX Table Mapping resource ID: unexpected format: %q", input)
	}

	return &adxResourceId{
		EndpointURI:  parts[0],
		DatabaseName: parts[1],
		EntityType:   "ingestion_mapping",
		Name:         parts[2],
	}, nil
}

func parseADXTableMappingID(input string) (*adxTableMappingResourceId, error) {
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
		MappingName:   parts[4],
		Kind:          parts[3],
		adxResourceId: res,
	}, nil
}
