# Azure Default Credentials Example

This example demonstrates how to use the ADX Terraform provider with Azure Default Credentials.

## Prerequisites

1. **Azure CLI** (for local development):
   ```bash
   az login
   az account set --subscription "your-subscription-id"
   ```

2. **Or Azure Environment Variables**:
   ```bash
   export AZURE_CLIENT_ID="your-service-principal-client-id"
   export AZURE_CLIENT_SECRET="your-service-principal-client-secret"
   export AZURE_TENANT_ID="your-tenant-id"
   ```

3. **Or Managed Identity** (when running on Azure resources)

## Usage

1. Set your ADX cluster endpoint:
   ```bash
   export TF_VAR_adx_endpoint="https://mycluster.eastus.kusto.windows.net"
   export TF_VAR_database_name="TestDatabase"
   ```

2. Initialize and apply:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## What This Example Creates

- An ADX table named `ExampleTable` with columns for Name, Age, Email, and CreatedAt
- A JSON mapping for the table to facilitate data ingestion
- Outputs showing the created resources

## Authentication Flow

The provider will attempt authentication in this order:
1. Environment variables (AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
2. Managed Identity (if running on Azure resources)
3. Azure CLI credentials (`az login`)
4. Azure Developer CLI credentials (`azd auth login`)
5. Azure PowerShell credentials (`Connect-AzAccount`)

## Cleanup

```bash
terraform destroy
```