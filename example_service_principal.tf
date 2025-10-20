terraform {
  required_providers {
    adx = {
      source = "favoretti/adx"
    }
  }
}

# Test provider with service principal credentials
provider "adx" {
  adx_endpoint  = var.adx_endpoint
  client_id     = var.client_id
  client_secret = var.client_secret
  tenant_id     = var.tenant_id
  lazy_init     = true
}

variable "adx_endpoint" {
  description = "The Azure Data Explorer cluster endpoint"
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

# This is just a test to validate provider initialization
output "provider_test" {
  value = "Provider configured with Service Principal credentials"
}
