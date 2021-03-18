package adx

import (
	"fmt"
	"strings"
)

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
