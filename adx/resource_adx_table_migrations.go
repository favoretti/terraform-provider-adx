package adx

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TableV0ToV1Upgrader() schema.StateUpgrader {
	return schema.StateUpgrader{
		Version: 0,
		Type:    resourceADXTable().CoreConfigSchema().ImpliedType(),
		Upgrade: TableV0ToV1UpgradeFunc,
	}
}
func TableV0ToV1UpgradeFunc(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	oldId := rawState["id"].(string)
	id, err := parseADXTableID(oldId)
	if err != nil {
		return rawState, err
	}

	newId := buildADXResourceId(id.EndpointURI, id.DatabaseName, "table", id.Name)
	log.Printf("[DEBUG] Updating ID from %q to %q", oldId, newId)
	rawState["id"] = newId

	return rawState, nil

}
