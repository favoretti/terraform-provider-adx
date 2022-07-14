package adx

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const TerraformProviderUserAgent = "terraform-provider-adx"

func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:             schema.TypeString,
				Optional:         true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"ADX_CLIENT_ID"}, nil),
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"client_secret": {
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"ADX_CLIENT_SECRET"}, nil),
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"adx_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"ADX_ENDPOINT"}, nil),
				ValidateDiagFunc: stringIsNotEmpty,
			},

			"tenant_id": {
				Type:             schema.TypeString,
				Optional:         true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"ADX_TENANT_ID"}, nil),
				ValidateDiagFunc: stringIsNotEmpty,
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
		},

		ResourcesMap: map[string]*schema.Resource{
			"adx_table": 							resourceADXTable(),
			"adx_table_mapping":    				resourceADXTableMapping(),
			"adx_table_ingestion_batching_policy": 	resourceADXTableIngestionBatchingPolicy(),
			"adx_table_retention_policy": 			resourceADXTableRetentionPolicy(),
			"adx_function": 						resourceADXFunction(),
		},
	}

	p.ConfigureContextFunc = providerConfigure(p)

	return p
}

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		config := &Config{
			ClientID:     d.Get("client_id").(string),
			ClientSecret: d.Get("client_secret").(string),
			TenantID:     d.Get("tenant_id").(string),
			Endpoint:     d.Get("adx_endpoint").(string),
		}

		ua := p.UserAgent(TerraformProviderUserAgent, p.TerraformVersion)

		return config.Client(ua)
	}
}
