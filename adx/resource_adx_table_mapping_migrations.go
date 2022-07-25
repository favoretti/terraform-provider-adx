package adx

import (
	"context"
	"log"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/copystructure"
)

func TableMappingSchemaV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"table_name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"kind": {
				Type:     schema.TypeString,
				Required: true,
				ValidateDiagFunc: validate.StringInSlice([]string{
					"Json",
				}),
			},
			"mapping": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"column": {
							Type:     schema.TypeString,
							Required: true,
						},
						"path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"datatype": {
							Type:     schema.TypeString,
							Required: true,
						},
						"transform": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"last_updated_on": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func TableMappingV0ToV1Upgrader() schema.StateUpgrader {
	return schema.StateUpgrader{
		Version: 0,
		Type:    TableMappingSchemaV0().CoreConfigSchema().ImpliedType(),
		Upgrade: TableMappingV0ToV1UpgradeFunc,
	}
}

func TableMappingV0ToV1UpgradeFunc(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	oldId := rawState["id"].(string)
	id, err := parseADXTableMappingV0ID(oldId)
	if err != nil {
		return rawState, err
	}

	z, err := copystructure.Copy(rawState)
	if err != nil {
		return nil, err
	}

	result := z.(map[string]interface{})
	newId := buildADXResourceId(id.EndpointURI, id.DatabaseName, "table", id.Name, "tablemapping", id.Kind, id.MappingName)
	log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newId)
	result["id"] = newId

	return result, nil

}
