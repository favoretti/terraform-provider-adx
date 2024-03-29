---
page_title: "adx_table_update_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages an update policy for a table in ADX.
---

# Resource `adx_table_update_policy`

Manages an update policy for a table in ADX.

See: [ADX - Update Policy](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/updatepolicy)


## Example Usage

```terraform

resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}

resource "adx_function" "test" {
  database_name = "test-db"
  name          = "test_function"
  body          = format("{%s | extend timestamp=ingestion_time()}", adx_table.test.name)
}

resource "adx_table" "test_update" {
  name          = "test_update"
  database_name = "test-db"

  table_schema = "f1:string,f2:string,f3:int"
}

resource "adx_table_update_policy" "test_update" {
  database_name = "test-db"
  table_name    = adx_table.test_update.name
  query         = adx_function.test.name
  source_table  = adx_table.test.name
  transactional = true
}

```

## Argument Reference

- **table_name** (String, Required) Name of the table containing the policy to modify
- **database_name** (String, Required) Database name that the target table is in
- **source_table** (String, Required) Name of the table that represents the source for the update policy
- **query** (String, Required) A query used to produce data for the update
- **transactional** (Boolean, Required) States if the update policy is transactional or not, default is false). If transactional and the update policy fails, the source table is not updated.
- **propagate_ingestion_properties** (Boolean, Optional) States if properties specified during ingestion to the source table, such as extent tags and creation time, apply to the target table. Default: false
- **managed_identity** (String, Optional) An update policy configured with a managed identity is performed on behalf of the managed identity. It must be the reserved word "system" to use the System-assigned Managed Identity of the cluster or an Object ID of a User-assigned Managed Identity.
- **enabled** (Boolean, Optional) States if update policy is enabled or disabled. Default: true
- **cluster** (Optional) `cluster` Configuration block (defined below) for the target cluster (overrides any config specified in the provider)

`cluster` Configuration block for connection details about the target ADX cluster 

*Note*: Any attributes specified here override the cluster config specified in the provider. Once a resource overrides an attribute specified in the provider, it will be stored explicitly as state for that resource and will not be possible to go back to the provider config unless explicitly unset.

- **uri** - (String, Optional) Target ADX cluster endpoint URI, starting with `https://`
- **client_id** - (String, Optional) The client ID for a service principal having admin access to this cluster/database.
- **client_secret** - (String, Optional) The client secret for a service principal having admin access to this cluster/database
- **tenant_id** - (String, Optional) Id for the tenant to which the service principal belongs

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
