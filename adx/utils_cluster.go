package adx

import (
	"context"
	"encoding/hex"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type ClusterConfig struct {
	ClientID     string
	ClientSecret string
	TenantID     string
	ClusterURI   string
}

func getClusterConfigInputSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cluster_uri": {
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
			},
		},
	}
}

func clusterConfigCustomDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
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

	log.Printf("[DEBUG] diff: default cluster_uri (%s)", defaultConfig.ClusterURI)
	log.Printf("[DEBUG] diff: new cluster_uri (%s)", newClusterMap["cluster_uri"])
	log.Printf("[DEBUG] diff: old cluster_uri (%s)", oldClusterMap["cluster_uri"])

	newClusterConfig := expandClusterConfig(newClusterMap)
	oldClusterConfig := expandClusterConfig(oldClusterMap)

	if oldClusterConfig.ClusterURI != newClusterConfig.ClusterURI && oldClusterConfig.ClusterURI == "" && newClusterConfig.ClusterURI == defaultConfig.ClusterURI {
		diff.Clear("cluster")
	}

	//applyClusterConfigDefaults(newClusterConfig,defaultConfig)

	//diff.SetNew("cluster",newClusterConfig)

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
	if len(clusterConfig.ClusterURI) == 0 {
		log.Printf("[DEBUG] Using default ClusterURI from provider for cluster config")
		clusterConfig.ClusterURI = defaultConfig.ClusterURI
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
	log.Printf("[DEBUG] Cluster configuration block ok: %s", ok)
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
		ClientID:     getAttributeOrDefault(clusterInputMap, "client_id", ""),
		ClientSecret: getAttributeOrDefault(clusterInputMap, "client_secret", ""),
		TenantID:     getAttributeOrDefault(clusterInputMap, "tenant_id", ""),
		ClusterURI:   getAttributeOrDefault(clusterInputMap, "cluster_uri", ""),
	}
}

func getAttributeOrDefault(d map[string]interface{}, name string, defaultString string) string {
	if val := d[name]; val != nil {
		return val.(string)
	}
	return defaultString
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
	cluster[0]["cluster_uri"] = clusterConfig.ClusterURI
	return cluster
}

func hashClusterConfig(clusterConfig *ClusterConfig) string {
	hash := hashObjects([]interface{}{clusterConfig.ClientID, clusterConfig.ClientSecret, clusterConfig.TenantID, clusterConfig.ClusterURI})
	return hex.EncodeToString(hash)
}
