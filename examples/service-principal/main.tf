terraform {
  required_providers {
    adx = {
      source = "favoretti/adx"
    }
  }
}

# Configure ADX provider with Service Principal credentials
provider "adx" {
  adx_endpoint  = var.adx_endpoint
  client_id     = var.client_id
  client_secret = var.client_secret
  tenant_id     = var.tenant_id
  lazy_init     = true
}

variable "adx_endpoint" {
  description = "The Azure Data Explorer cluster endpoint (e.g., https://mycluster.eastus.kusto.windows.net)"
  type        = string
}

variable "client_id" {
  description = "The service principal client ID"
  type        = string
}

variable "client_secret" {
  description = "The service principal client secret"
  type        = string
  sensitive   = true
}

variable "tenant_id" {
  description = "The Azure tenant ID"
  type        = string
}

variable "database_name" {
  description = "The Azure Data Explorer database name"
  type        = string
  default     = "TestDatabase"
}

# Example table resource
resource "adx_table" "example" {
  name          = "ServicePrincipalTable"
  database_name = var.database_name
  table_schema  = "UserId:string,Action:string,Timestamp:datetime,Data:dynamic"
}

# Example function resource
resource "adx_function" "example" {
  name          = "GetUserActions"
  database_name = var.database_name
  body          = "${adx_table.example.name} | where UserId == userId | order by Timestamp desc"
  
  parameters {
    name = "userId"
    type = "string"
  }
}

# Example table with cluster override (different authentication for specific resource)
resource "adx_table" "override_example" {
  name          = "OverrideTable" 
  database_name = var.database_name
  table_schema  = "Id:string,Value:int"
  
  # Override cluster config for this specific resource
  cluster {
    uri       = var.adx_endpoint
    client_id = var.client_id
    client_secret = var.client_secret
    tenant_id = var.tenant_id
  }
}

# Outputs
output "table_name" {
  description = "The name of the created table"
  value       = adx_table.example.name
}

output "function_name" {
  description = "The name of the created function"
  value       = adx_function.example.name
}

output "override_table_name" {
  description = "The name of the table with cluster override"
  value       = adx_table.override_example.name
}

output "provider_auth_method" {
  description = "Authentication method used"
  value       = "Service Principal"
}