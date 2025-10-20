# Terraform Provider for Azure Data Explorer

* [Terraform Website](https://www.terraform.io)
* [ADX Provider Documentation](https://registry.terraform.io/providers/favoretti/adx/latest/docs)

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

## Authentication Methods

The ADX provider supports two authentication methods:

### Method 1: Azure Default Credentials (Recommended)

Use Azure Default Credentials for seamless authentication across different environments:

```hcl
provider "adx" {
  adx_endpoint               = "https://adxcluster123.eastus.kusto.windows.net"
  use_default_credentials   = true
}
```

When `use_default_credentials` is set to `true`, the provider uses DefaultAzureCredential which supports:
- **Managed Identity** (when running on Azure resources like VMs, App Service, etc.)
- **Azure CLI** (when authenticated via `az login`)
- **Azure Developer CLI** (when authenticated via `azd auth login`) 
- **Environment variables** (AZURE_CLIENT_ID, AZURE_TENANT_ID, AZURE_CLIENT_SECRET)
- **Azure PowerShell** (when authenticated via `Connect-AzAccount`)

### Method 2: Service Principal (Client Credentials)

Use explicit service principal credentials:

```hcl
provider "adx" {
  adx_endpoint  = "https://adxcluster123.eastus.kusto.windows.net"
  client_id     = "your-service-principal-client-id"
  client_secret = "your-service-principal-client-secret"
  tenant_id     = "your-azure-tenant-id"
}
```

## Environment Variables

Configuration parameters can be overridden with environment variables:

```bash
# For Default Credentials (recommended)
export ADX_ENDPOINT="https://adxcluster123.eastus.kusto.windows.net"

# For Service Principal authentication
export ADX_ENDPOINT="https://adxcluster123.eastus.kusto.windows.net"
export ADX_CLIENT_ID="your-service-principal-client-id"
export ADX_CLIENT_SECRET="your-service-principal-client-secret"
export ADX_TENANT_ID="your-azure-tenant-id"
```

## Examples

See the [examples directory](./examples/) for complete working examples demonstrating different authentication methods:

- **[Default Credentials](./examples/default-credentials/)** - Recommended approach using Azure Default Credentials
- **[Service Principal](./examples/service-principal/)** - Traditional approach with explicit credentials

## Resource Examples

```hcl
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

```hcl
terraform {
  required_providers {
    adx = {
      source = "favoretti/adx"
    }
  }
}

# Configure the Azure Data Explorer provider with Default Credentials
provider "adx" {
  adx_endpoint             = "https://adxcluster123.eastus.kusto.windows.net"
  use_default_credentials  = true
}

## Lazy provider initialization

```

## Lazy Provider Initialization

```hcl
provider "adx" {
  adx_endpoint             = "https://adxcluster123.eastus.kusto.windows.net"
  use_default_credentials  = true
  lazy_init                = true
}
```

If `lazy_init` is set to true, no connection will be attempted to the ADX cluster until the first resource state load.

## Cluster Configuration Per Resource

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
    use_default_credentials = true
  }
}
```
