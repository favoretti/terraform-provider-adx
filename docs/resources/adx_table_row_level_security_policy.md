---
page_title: "adx_table_row_level_security_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the row level security policy of a table in ADX.
---

# Resource `adx_table_row_level_security_policy`

Manages the row level security policy of a table in ADX.

See: [ADX - Row Level Security Policy](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/rowlevelsecuritypolicy)

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
  body          = format("{%s | limit 10}", adx_table.test.name)
}

resource "adx_table_row_level_security_policy" "test" {
  database_name = "test-db"
  table_name    = adx_table.test.name
  query         = adx_function.test.name
  enabled       = true
}

```

## Argument Reference

- **table_name** (String, Required) Name of the table containing the policy to modify
- **database_name** (String, Required) Database name that the target table is in
- **query** (String, Required) The query to be run automatically when the target table is queried
- **enabled** (Boolean, Optional) Enable or disable this policy
- **allow_mv_without_rls** (Boolean, Optional) Enables the allowMaterializedViewsWithoutRowLevelSecurity flag during policy creation
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
