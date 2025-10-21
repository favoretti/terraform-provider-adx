terraform {
  required_providers {
    adx = {
      source = "favoretti/adx"
    }
  }
}

# Configure ADX provider with Azure Default Credentials
provider "adx" {
  adx_endpoint            = var.adx_endpoint
  use_default_credentials = true
  lazy_init               = true
}

variable "adx_endpoint" {
  description = "The Azure Data Explorer cluster endpoint (e.g., https://mycluster.eastus.kusto.windows.net)"
  type        = string
}

variable "database_name" {
  description = "The Azure Data Explorer database name"
  type        = string
  default     = "TestDatabase"
}

# Example table resource
resource "adx_table" "example" {
  name          = "ExampleTable"
  database_name = var.database_name
  table_schema  = "Name:string,Age:int,Email:string,CreatedAt:datetime"
}

# Example table mapping for JSON ingestion
resource "adx_table_mapping" "example_json" {
  name          = "ExampleJsonMapping"
  database_name = var.database_name
  table_name    = adx_table.example.name
  kind          = "Json"

  mapping {
    column   = "Name"
    path     = "$.name"
    datatype = "string"
  }

  mapping {
    column   = "Age"
    path     = "$.age"
    datatype = "int"
  }

  mapping {
    column   = "Email"
    path     = "$.email"
    datatype = "string"
  }

  mapping {
    column   = "CreatedAt"
    path     = "$.created_at"
    datatype = "datetime"
  }
}

# Output the table information
output "table_name" {
  description = "The name of the created table"
  value       = adx_table.example.name
}

output "table_mapping_name" {
  description = "The name of the created table mapping"
  value       = adx_table_mapping.example_json.name
}

output "provider_auth_method" {
  description = "Authentication method used"
  value       = "Azure Default Credentials"
}
