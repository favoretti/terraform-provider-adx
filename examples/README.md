# ADX Terraform Provider Examples

This directory contains examples demonstrating different authentication methods and use cases for the Azure Data Explorer (ADX) Terraform provider.

## Available Examples

### 🔐 [Default Credentials](./default-credentials/)
Demonstrates using Azure Default Credentials for authentication. This is the **recommended approach** for most use cases as it provides seamless authentication across different environments.

**Best for:**
- Local development (with `az login`)
- Production workloads running on Azure (Managed Identity)
- CI/CD pipelines (with environment variables)
- AKS workloads (with Workload Identity)

### 🔑 [Service Principal](./service-principal/)
Demonstrates using explicit Service Principal credentials for authentication. This is useful when you need fine-grained control over authentication or are running outside of Azure.

**Best for:**
- On-premises deployments
- Non-Azure cloud environments
- Legacy systems
- Scenarios requiring explicit credential management

## Quick Start

1. **Choose your authentication method** based on your environment
2. **Navigate to the appropriate example directory**
3. **Follow the README instructions** in that directory
4. **Copy and customize** the example for your needs

## Authentication Comparison

| Method | Security | Ease of Use | Environment Support |
|--------|----------|-------------|-------------------|
| **Default Credentials** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Azure, Local Dev, CI/CD |
| **Service Principal** | ⭐⭐⭐⭐ | ⭐⭐⭐ | Universal |

## Common Configuration

All examples require these basic settings:

- `adx_endpoint`: Your Azure Data Explorer cluster endpoint
- `database_name`: The database where resources will be created

## Local Development Setup

For local testing, we recommend using Azure Default Credentials with Azure CLI:

```bash
# Login to Azure
az login

# Set your subscription
az account set --subscription "your-subscription-id"

# Set environment variables
export TF_VAR_adx_endpoint="https://mycluster.eastus.kusto.windows.net"
export TF_VAR_database_name="TestDatabase"

# Navigate to an example
cd default-credentials

# Run Terraform
terraform init
terraform plan
terraform apply
```

## Contributing

When adding new examples:

1. Create a new directory with a descriptive name
2. Include a comprehensive `README.md`
3. Provide a `terraform.tfvars.example` file
4. Document prerequisites and setup steps
5. Include cleanup instructions

## Support

For issues with these examples or the provider itself, please check:

- [Provider Documentation](https://registry.terraform.io/providers/favoretti/adx/latest/docs)
- [GitHub Issues](https://github.com/favoretti/terraform-provider-adx/issues)
- [Azure Data Explorer Documentation](https://docs.microsoft.com/en-us/azure/data-explorer/)