package adx

import (
	"strings"
	"fmt"
)

type adxTableResource struct {
	
	adxResource
}

type adxTableMappingResource struct {
	TableName string
	Kind string

	adxResource
}

func parseADXTableID(input string) (*adxResource, error) {
	return parseADXID(input,3,0,1,2)
}

func parseADXTableMappingID(input string) (*adxTableMappingResource, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 5 {
		return nil, fmt.Errorf("error parsing ADX Table Mapping resource ID: unexpected format: %q", input)
	}

	res := adxResource{
		EndpointURI:  parts[0],
		DatabaseName: parts[1],
		Name:         parts[4],
	}

	return &adxTableMappingResource{
		TableName:    parts[2],
		Kind:         parts[3],
		adxResource: res,
	}, nil
}