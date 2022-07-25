package adx

import (
	"context"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type Config struct {
	ClientID     string
	ClientSecret string
	TenantID     string
	Endpoint     string
	LazyInit     bool
}

type Meta struct {
	KustoClientsMap      map[string]*kusto.Client
	DefaultClusterConfig *ClusterConfig
	StopContext          context.Context
}

func (c *Config) Client(userAgent string) (*Meta, diag.Diagnostics) {
	clusterConfig := providerConfigToClusterConfig(c)

	meta := Meta{
		StopContext:          context.Background(),
		DefaultClusterConfig: clusterConfig,
		KustoClientsMap:      make(map[string]*kusto.Client),
	}

	// Not returning an error here on missing values because the user can specify "missing" config for each resource
	if !c.LazyInit {
		//tflog.Info(ctx, "Lazy init is disabled, attempting to eagerly create ADX client")
		// Client is automatically cached in this function
		_, err := getADXClient(&meta, clusterConfig)
		if err != nil {
			return nil, diag.Errorf("%+v", err)
		}
	} else {
		//tflog.Info(ctx, "Lazy init is enabled, postponing ADX client creation")
	}

	return &meta, nil
}

func providerConfigToClusterConfig(config *Config) *ClusterConfig {
	return &ClusterConfig{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		TenantID:     config.TenantID,
		URI:          config.Endpoint,
	}
}
