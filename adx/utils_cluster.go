package adx

import (
	"context"
	"encoding/hex"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ClusterConfig struct {
	ClientID              string
	ClientSecret          string
	TenantID              string
	URI                   string
	UseDefaultCredentials bool
}

func getClusterConfigInputSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uri": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: true,
				},
				"client_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"client_secret": {
					Type:      schema.TypeString,
					Sensitive: true,
					Optional:  true,
					Computed:  true,
				},
				"tenant_id": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				"use_default_credentials": {
					Type:     schema.TypeBool,
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}

func clusterConfigCustomDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	//id := diff.Id()
	//log.Printf("[DEBUG] Prcoessing resource with id(%s): %s", id)

	oldCluster, newCluster := diff.GetChange("cluster")
	var newClusterMap map[string]interface{}
	var oldClusterMap map[string]interface{}

	if newCluster == nil || len(newCluster.([]interface{})) == 0 {
		newClusterMap = make(map[string]interface{})
	} else {
		newClusterMap = newCluster.([]interface{})[0].(map[string]interface{})
	}

	if oldCluster == nil || len(oldCluster.([]interface{})) == 0 {
		oldClusterMap = make(map[string]interface{})
	} else {
		oldClusterMap = oldCluster.([]interface{})[0].(map[string]interface{})
	}

	defaultConfig := meta.(*Meta).DefaultClusterConfig

	log.Printf("[TRACE] diff: default cluster uri (%s)", defaultConfig.URI)
	log.Printf("[TRACE] diff: new cluster uri (%s)", newClusterMap["uri"])
	log.Printf("[TRACE] diff: old cluster uri (%s)", oldClusterMap["uri"])

	newClusterConfig := expandClusterConfig(newClusterMap)
	oldClusterConfig := expandClusterConfig(oldClusterMap)

	if oldClusterConfig.URI != newClusterConfig.URI && oldClusterConfig.URI == "" && newClusterConfig.URI == defaultConfig.URI {
		diff.Clear("cluster")
	} else {
		if oldClusterConfig.URI != "" && newClusterConfig.URI == "" {
			newClusterMap["uri"] = defaultConfig.URI
			log.Printf("[DEBUG] Defaulting cluster[0].uri diff to provider config: %s", defaultConfig.URI)
		}
		if oldClusterConfig.ClientID != "" && newClusterConfig.ClientID == "" {
			newClusterMap["client_id"] = defaultConfig.ClientID
			log.Printf("[DEBUG] Defaulting cluster[0].client_id diff to provider config: %s", defaultConfig.ClientID)
		}
		if oldClusterConfig.ClientSecret != "" && newClusterConfig.ClientSecret == "" {
			newClusterMap["client_secret"] = defaultConfig.ClientSecret
			log.Printf("[DEBUG] Defaulting cluster[0].client_secret diff to provider config: %s", defaultConfig.ClientSecret)
		}
		if oldClusterConfig.TenantID != "" && newClusterConfig.TenantID == "" {
			newClusterMap["tenant_id"] = defaultConfig.TenantID
			log.Printf("[DEBUG] Defaulting cluster[0].tenant_id diff to provider config: %s", defaultConfig.TenantID)
		}
		diff.SetNew("cluster", newCluster)
	}

	return nil
}

func applyClusterConfigDefaults(clusterConfig *ClusterConfig, defaultConfig *ClusterConfig) {
	if len(clusterConfig.ClientID) == 0 {
		log.Printf("[DEBUG] Using default ClientID from provider for cluster config")
		clusterConfig.ClientID = defaultConfig.ClientID
	}
	if len(clusterConfig.ClientSecret) == 0 {
		log.Printf("[DEBUG] Using default ClientSecret from provider for cluster config")
		clusterConfig.ClientSecret = defaultConfig.ClientSecret
	}
	if len(clusterConfig.TenantID) == 0 {
		log.Printf("[DEBUG] Using default TenantID from provider for cluster config")
		clusterConfig.TenantID = defaultConfig.TenantID
	}
	if len(clusterConfig.URI) == 0 {
		log.Printf("[DEBUG] Using default URI from provider for cluster config")
		clusterConfig.URI = defaultConfig.URI
	}
	if !clusterConfig.UseDefaultCredentials {
		log.Printf("[DEBUG] Using default UseDefaultCredentials from provider for cluster config")
		clusterConfig.UseDefaultCredentials = defaultConfig.UseDefaultCredentials
	}
}

func getAndExpandClusterConfigWithDefaults(ctx context.Context, d *schema.ResourceData, meta interface{}) *ClusterConfig {
	clusterConfig := getAndExpandClusterConfig(ctx, d)
	defaultConfig := meta.(*Meta).DefaultClusterConfig
	applyClusterConfigDefaults(clusterConfig, defaultConfig)
	return clusterConfig
}

func getAndExpandClusterConfig(ctx context.Context, d *schema.ResourceData) *ClusterConfig {
	cluster, ok := d.GetOk("cluster")
	log.Printf("[DEBUG] Cluster configuration block ok: %t", ok)
	if !ok || len(cluster.([]interface{})) == 0 {
		log.Printf("[DEBUG] No cluster configuration block found in resource definition, defaulting to cluster specified in provider config")
		return expandClusterConfig(nil)
	}
	return expandClusterConfig(cluster.([]interface{})[0])
}

func expandClusterConfig(input interface{}) *ClusterConfig {
	if input == nil {
		return &ClusterConfig{}
	}
	clusterInputMap := input.(map[string]interface{})

	return &ClusterConfig{
		ClientID:              getAttributeOrDefault(clusterInputMap, "client_id", ""),
		ClientSecret:          getAttributeOrDefault(clusterInputMap, "client_secret", ""),
		TenantID:              getAttributeOrDefault(clusterInputMap, "tenant_id", ""),
		URI:                   getAttributeOrDefault(clusterInputMap, "uri", ""),
		UseDefaultCredentials: getBoolAttributeOrDefault(clusterInputMap, "use_default_credentials", false),
	}
}

func getAttributeOrDefault(d map[string]interface{}, name string, defaultString string) string {
	if val := d[name]; val != nil {
		return val.(string)
	}
	return defaultString
}

func getBoolAttributeOrDefault(d map[string]interface{}, name string, defaultBool bool) bool {
	if val := d[name]; val != nil {
		return val.(bool)
	}
	return defaultBool
}

func flattenAndSetClusterConfig(ctx context.Context, d *schema.ResourceData, clusterConfig *ClusterConfig) {
	d.Set("cluster", flattenClusterConfig(clusterConfig))
}

func flattenClusterConfig(clusterConfig *ClusterConfig) []map[string]interface{} {
	cluster := make([]map[string]interface{}, 1, 1)
	cluster[0] = make(map[string]interface{})
	cluster[0]["client_id"] = clusterConfig.ClientID
	cluster[0]["client_secret"] = clusterConfig.ClientSecret
	cluster[0]["tenant_id"] = clusterConfig.TenantID
	cluster[0]["uri"] = clusterConfig.URI
	cluster[0]["use_default_credentials"] = clusterConfig.UseDefaultCredentials
	return cluster
}

func hashClusterConfig(clusterConfig *ClusterConfig) string {
	hash := hashObjects([]interface{}{clusterConfig.ClientID, clusterConfig.ClientSecret, clusterConfig.TenantID, clusterConfig.URI, clusterConfig.UseDefaultCredentials})
	return hex.EncodeToString(hash)
}
