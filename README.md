# Terraform Provider for Azure Data Explorer

* [Terraform Website](https://www.terraform.io)
* [ADX Provider Documentation](https://registry.terraform.io/providers/favoretti/adx/latest/docs)

[![release](https://github.com/favoretti/terraform-provider-adx/actions/workflows/release.yml/badge.svg)](https://github.com/favoretti/terraform-provider-adx/actions/workflows/release.yml)

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

```

## Alternative authentication
Above configuration parameters can also be overriden with following environment variables:
```
ADX_ENDPOINT
ADX_CLIENT_ID
ADX_CLIENT_SECRET
ADX_TENANT_ID
```

## Lazy provider initialization
```hcl
provider "adx" {
  adx_endpoint  = "https://adxcluster123.eastus.kusto.windows.net"
  client_id     = "clientId"
  client_secret = "secret"
  tenant_id     = "tenantId"
  lazy_init     = true
}
```

If `lazy_init` is set to true, no connection will be attempted to the ADX cluster until the first resource state load.

## Cluster config per resource

Resources allow overriding any of the cluster attributes specified in the provider config.

The provider config is the "default" config for each resource unless overridden.

*NOTE:* Once a resource overrides an attribute specified in the provider, it will be stored explicitly as state for that resource (instead of computed) and will not be possible to go back to the provider config.

```hcl
resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f4:string,f3:int"
  cluster {
    uri = "https://adxcluster456.eastus.kusto.windows.net"
  }
}
```
