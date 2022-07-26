package adx

import (
	"context"

	"github.com/favoretti/terraform-provider-adx/adx/validate"
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
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"client_secret": {
				Type:             schema.TypeString,
				Optional:         true,
				Sensitive:        true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"ADX_CLIENT_SECRET"}, nil),
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"adx_endpoint": {
				Type:             schema.TypeString,
				Optional:         true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"ADX_ENDPOINT"}, nil),
				ForceNew:         true,
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"tenant_id": {
				Type:             schema.TypeString,
				Optional:         true,
				DefaultFunc:      schema.MultiEnvDefaultFunc([]string{"ADX_TENANT_ID"}, nil),
				ValidateDiagFunc: validate.StringIsNotEmpty,
			},

			"lazy_init": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},

		DataSourcesMap: map[string]*schema.Resource{},

		ResourcesMap: map[string]*schema.Resource{
			"adx_table":                              resourceADXTable(),
			"adx_table_mapping":                      resourceADXTableMapping(),
			"adx_table_ingestion_batching_policy":    resourceADXTableIngestionBatchingPolicy(),
			"adx_table_retention_policy":             resourceADXTableRetentionPolicy(),
			"adx_table_row_level_security_policy":    resourceADXTableRowLevelSecurityPolicy(),
			"adx_table_partitioning_policy":          resourceADXTablePartitioningPolicy(),
			"adx_table_caching_policy":               resourceADXTableCachingPolicy(),
			"adx_table_update_policy":                resourceADXTableUpdatePolicy(),
			"adx_function":                           resourceADXFunction(),
			"adx_materialized_view":                  resourceADXMaterializedView(),
			"adx_materialized_view_caching_policy":   resourceADXMaterializedViewCachingPolicy(),
			"adx_materialized_view_retention_policy": resourceADXMaterializedViewRetentionPolicy(),
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
			LazyInit:     d.Get("lazy_init").(bool),
		}

		ua := p.UserAgent(TerraformProviderUserAgent, p.TerraformVersion)

		return config.Client(ua)
	}
}
