package adx

import (
	"context"
	"log"
	"regexp"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/copystructure"
)

func TableSchemaV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"database_name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"table_schema": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				AtLeastOneOf:  []string{"table_schema", "column", "from_query"},
				ConflictsWith: []string{"column", "from_query"},
				ValidateDiagFunc: validate.StringMatch(
					regexp.MustCompile("[a-zA-Z0-9:-_,]+"),
					"Table schema must contain only letters, number, dashes, semicolons, commas and underscores and no spaces",
				),
			},

			"column": {
				Type:          schema.TypeList,
				AtLeastOneOf:  []string{"table_schema", "column", "from_query"},
				ConflictsWith: []string{"table_schema", "from_query"},
				Optional:      true,
				Computed:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validate.StringIsNotEmpty,
						},
						"type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validate.StringIsNotEmpty,
						},
					},
				},
			},

			"from_query": {
				Type:          schema.TypeList,
				AtLeastOneOf:  []string{"table_schema", "column", "from_query"},
				ConflictsWith: []string{"table_schema", "column"},
				Optional:      true,
				Computed:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validate.StringIsNotEmpty,
						},
						"append": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"extend_schema": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"recreate_schema": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"distributed": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"force_an_update_when_value_changed": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},

			"merge_on_update": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func TableV0ToV1Upgrader() schema.StateUpgrader {
	return schema.StateUpgrader{
		Version: 0,
		Type:    TableSchemaV0().CoreConfigSchema().ImpliedType(),
		Upgrade: TableV0ToV1UpgradeFunc,
	}
}

func TableV0ToV1UpgradeFunc(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	oldId := rawState["id"].(string)
	id, err := parseADXTableV0ID(oldId)
	if err != nil {
		return rawState, err
	}

	z, err := copystructure.Copy(rawState)
	if err != nil {
		return nil, err
	}

	result := z.(map[string]interface{})
	newId := buildADXResourceId(id.EndpointURI, id.DatabaseName, "table", id.Name)
	log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newId)
	result["id"] = newId

	return result, nil

}
