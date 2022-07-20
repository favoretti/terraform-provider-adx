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

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
