terraform {
  required_providers {
    adx = {
      source = "favoretti/adx"
    }
  }
}

# Test provider with default credentials
provider "adx" {
  adx_endpoint            = var.adx_endpoint
  use_default_credentials = true
  lazy_init               = true
}

variable "adx_endpoint" {
  description = "The Azure Data Explorer cluster endpoint"
  type        = string
}

# This is just a test to validate provider initialization
output "provider_test" {
  value = "Provider configured with Azure Default Credentials"
}
