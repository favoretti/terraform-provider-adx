package adx

import (
	"context"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type Config struct {
	ClientID     string
	ClientSecret string
	TenantID     string
	Endpoint     string
}

type Meta struct {
	Kusto       kusto.Client
	StopContext context.Context
}

func (c *Config) Client(userAgent string) (*Meta, diag.Diagnostics) {
	meta := Meta{
		StopContext: context.Background(),
	}

	auth := kusto.Authorization{Config: auth.NewClientCredentialsConfig(c.ClientID, c.ClientSecret, c.TenantID)}
	client, err := kusto.New(c.Endpoint, auth)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	meta.Kusto = *client

	return &meta, nil
}
