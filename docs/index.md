# Azure Data Explorer Provider

Use this provider to manage Azure Data Explorer tables and mappings.

## Example Usage

```hcl
provider "adx" {
  # adx_endpoint    = "..."
  # client_id       = "..."
  # client_secret   = "..."
  # tenant_id       = "..."
}

resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"

  # either
  table_schema  = "f1:string,f2:string,f3:int"

  # or
  column {
    name = "f1"
    type = "string"
  }
  column {
    name = "f2"
    type = "string"
  }
  column {
    name = "f3"
    type = "int"
  }
}

resource "adx_table_mapping" "test" {
  name          = "TestMapping"
  database_name = "test-db"
  table_name    = adx_table.test.name
  kind          = "Json"
  mapping {
    column   = "f1"
    path     = "$.f1"
    datatype = "string"
  }
  mapping {
    column   = "f2"
    path     = "$.f2"
    datatype = "string"
  }
}

resource "adx_table_retention_policy" "test" {
  database_name      = "test-db"
  table_name         = adx_table.test.name
  soft_delete_period = "500m"
  recoverability     = false
}

resource "adx_function" "test" {
  database_name = "test-db"
  name          = "test_function"
  body          = format("{%s | limit 10}", adx_table.test.name)
}

resource "adx_table_ingestion_batching_policy" "test" {
  database_name         = "test-db"
  table_name            = adx_table.test.name
  max_batching_timespan = "00:10:00"
  max_items      = 501
  max_raw_size_mb       = 129
}

resource "adx_table_row_level_security_policy" "test" {
  database_name = "test-db"
  table_name    = adx_table.test.name
  query         = adx_function.test.name
}

resource "adx_table_caching_policy" "test" {
  database_name      = "test-db"
  table_name         = adx_table.test.name
  data_hot_span      = "30d"
}

resource "adx_table_partitioning_policy" "test" {
  database_name       = "test-db"
  table_name          = adx_table.test.name
  effective_date_time = "2022-07-19T13:56:45Z"

  partition_key {
    column_name = "f1"
    kind        = "Hash"

    hash_properties {
      function                  = "XxHash64"
      max_partition_count       = 64
      seed                      = 2
      partition_assignment_mode = "Uniform"
    }
  }
}

resource "adx_table" "test_update" {
  name          = "test_update"
  database_name = "test-db"

  table_schema = "f1:string,f2:string,f3:int"
}

resource "adx_table_update_policy" "test_update" {
  database_name = "test-db"
  table_name    = adx_table.test_update.name
  query         = adx_table.test.name
  source_table  = adx_table.test.name
  transactional = true
}

```

## Argument Reference

* `adx_endpoint` - (Optional) ADX Endpoint URI, starting with `https://`. It can also be sourced from the `ADX_ENDPOINT` environment variable.

* `client_id` - (Optional) The client ID. It can also be sourced from the `ADX_CLIENT_ID` environment variable.

* `client_secret` - (Optional) The client secret. It can also be sourced from the `ADX_CLIENT_SECRET` environment variable.

* `tenant_id` - (Optional) The tenant ID. It can also be sourced from the `ADX_TENANT_ID` environment variable.
