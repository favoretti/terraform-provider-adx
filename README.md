# Terraform Provider for Azure Data Explorer

* [Terraform Website](https://www.terraform.io)

## Usage Example

```hcl
terraform {
  required_providers {
    adx = {
      source = "favoretti/adx"
    }
  }
}

# Configure the Azure Data Explorer provider
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

## Alternative authentication
Above configuration parameters can also be overriden with following environment variables:
```
ADX_ENDPOINT
ADX_CLIENT_ID
ADX_CLIENT_SECRET
ADX_TENANT_ID
```

