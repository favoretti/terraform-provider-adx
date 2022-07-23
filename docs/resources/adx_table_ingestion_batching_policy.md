---
page_title: "adx_table_ingestion_batching_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the ingestionbatching policy of a table in ADX.
---

# Resource `adx_table_ingestion_batching_policy`

Manages ingestion batching policy for a table in ADX.

See: [ADX - Ingestion Batching Policy](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/batchingpolicy)

## Example Usage

```terraform

resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}

resource "adx_table_ingestion_batching_policy" "test" {
  database_name         = "test-db"
  table_name            = adx_table.test.name
  max_batching_timespan = "00:10:00"
  max_items             = 500
  max_raw_size_mb       = 128
}

```

### Argument Reference

- **table_name** (String, Required) Name of the table containing the policy to modify
- **database_name** (String, Required) Database name that the target table is in
- **max_batching_timespan** (String, Required) Timespan after which an ingested blob will be sealed (Format: hh:mm:ss)
- **max_items** (String, Required) Max items to ingest before sealing a blob
- **max_raw_size_mb** (String, Required) Max size of ingested blob in MB
- **cluster** (Optional) `cluster` Configuration block (defined below) for the target cluster (overrides any config specified in the provider)

`cluster` Configuration block for connection details about the target ADX cluster 

<<<<<<< HEAD
*Note*: Any attributes specified here override the cluster config specified in the provider. Once a resource overrides an attribute specified in the provider, it will be stored explicitly as state for that resource and will not be possible to go back to the provider config unless explicitly unset.
=======
*Note*: Any attributes specified here override the cluster config specified in the provider. Once a resource overrides an attribute specified in the provider, it will be stored explicitly as state for that resource and will not be possible to go back to the provider config.
>>>>>>> 7529cf8 (doc updates for cluster config)

- **cluster_uri** - (String, Optional) Target ADX cluster endpoint URI, starting with `https://`
- **client_id** - (String, Optional) The client ID for a service principal having admin access to this cluster/database. 
- **client_secret** - (String, Optional) The client secret for a service principal having admin access to this cluster/database
- **tenant_id** - (String, Optional) Id for the tenant to which the service principal belongs

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
