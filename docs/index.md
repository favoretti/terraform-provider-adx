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
  table_schema  = "f1:string,f2:string,f3:int"
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

```

## Argument Reference

* `adx_endpoint` - (Optional) ADX Endpoint URI, starting with `https://`. It can also be sourced from the `ADX_ENDPOING` environment variable.

* `client_id` - (Optional) The client ID. It can also be sourced from the `ADX_CLIENT_ID` environment variable.

* `client_secret` - (Optional) The client secret. It can also be sourced from the `ADX_CLIENT_SECRET` environment variable.

* `tenant_id` - (Optional) The tenant ID. It can also be sourced from the `ADX_TENANT_ID` environment variable.
