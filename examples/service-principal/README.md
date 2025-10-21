# Service Principal Authentication Example

This example demonstrates how to use the ADX Terraform provider with explicit Service Principal credentials.

## Prerequisites

1. **Azure Service Principal**: Create a service principal with appropriate permissions:
   ```bash
   # Create service principal
   az ad sp create-for-rbac --name "terraform-adx-sp" --role Contributor
   
   # Note down the output:
   # - appId (this is your client_id)
   # - password (this is your client_secret)  
   # - tenant (this is your tenant_id)
   ```

2. **ADX Permissions**: Grant the service principal appropriate permissions on your ADX cluster:
   ```bash
   # Example: Grant database admin permissions
   az kusto database-principal-assignment create \
     --cluster-name "mycluster" \
     --database-name "TestDatabase" \
     --resource-group "my-rg" \
     --principal-id "service-principal-object-id" \
     --principal-type "App" \
     --role "Admin"
   ```

## Usage

1. Copy the example variables file:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your values:
   ```hcl
   adx_endpoint  = "https://mycluster.eastus.kusto.windows.net"
   client_id     = "your-service-principal-client-id"
   client_secret = "your-service-principal-client-secret"
   tenant_id     = "your-tenant-id"
   database_name = "TestDatabase"
   ```

3. Initialize and apply:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Alternative: Environment Variables

Instead of using terraform.tfvars, you can use environment variables:

```bash
export TF_VAR_adx_endpoint="https://mycluster.eastus.kusto.windows.net"
export TF_VAR_client_id="your-service-principal-client-id"
export TF_VAR_client_secret="your-service-principal-client-secret"
export TF_VAR_tenant_id="your-tenant-id"
export TF_VAR_database_name="TestDatabase"
```

## What This Example Creates

- An ADX table named `ServicePrincipalTable` for user actions
- A Kusto function `GetUserActions` that queries the table
- Another table `OverrideTable` demonstrating cluster configuration override
- Outputs showing the created resources

## Cluster Configuration Override

This example also demonstrates how to override cluster configuration at the resource level, which is useful when you need different authentication for specific resources.

## Security Best Practices

- Store sensitive values like `client_secret` in environment variables or secure key stores
- Use principle of least privilege when assigning permissions
- Rotate service principal credentials regularly
- Consider using managed identities where possible

## Cleanup

```bash
terraform destroy
```