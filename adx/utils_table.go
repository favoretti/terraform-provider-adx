package adx

import (
	"strings"
	"fmt"
)

type adxTableResource struct {
	
	adxResource
}

type adxTableMappingResource struct {
	MappingName string
	Kind string

	adxResource
}

func parseADXTableID(input string) (*adxResource, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 3 {
		return nil, fmt.Errorf("error parsing ADX Table Mapping resource ID: unexpected format: %q", input)
	}

	return &adxResource{
		EndpointURI:  parts[0],
		DatabaseName: parts[1],
		EntityType:   "ingestion_mapping",
		Name:         parts[2],
	}, nil
}

func parseADXTableMappingID(input string) (*adxTableMappingResource, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 5 {
		return nil, fmt.Errorf("error parsing ADX Table Mapping resource ID: unexpected format: %q", input)
	}

	res := adxResource{
		EndpointURI:  parts[0],
		DatabaseName: parts[1],
		EntityType:   "ingestion_mapping",
		Name:         parts[2],
	}

	return &adxTableMappingResource{
		MappingName:  parts[4],
		Kind:         parts[3],
		adxResource: res,
	}, nil
}