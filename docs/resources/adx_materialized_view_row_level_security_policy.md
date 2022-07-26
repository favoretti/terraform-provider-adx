---
page_title: "adx_materialized_view_row_level_security_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the row level security policy of a materialized view in ADX.
---

# Resource `adx_materialized_view_row_level_security_policy`

Manages the row level security policy of a materialized view in ADX.

See: [ADX - Materialized View Policies](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/materialized-views/

## Example Usage

```terraform

resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}

resource "adx_materialized_view" "test" {
  name               = "test_mv"
  database_name      = "test-db"
  source_table_name  = adx_table.test.name
  query              = format("%s | extend hi=true | summarize count(), dcount(f1) by f2",adx_table.test.name)
}

resource "adx_function" "test" {
  database_name = "test-db"
  name          = "test_function"
  body          = format("{%s | limit 10}", adx_materialized_view.test.name)
}

resource "adx_materialized_view_row_level_security_policy" "test" {
  database_name = "test-db"
  view_name     = adx_materialized_view.test.name
  query         = adx_function.test.name
  enabled       = true
}

```

## Argument Reference

- **view_name** (String, Required) Name of the materialized view containing the policy to modify
- **database_name** (String, Required) Database name that the target materialized view is in
- **query** (String, Required) The query to be run automatically when the target materialized view is queried
- **enabled** (Boolean, Optional) Enable or disable this policy
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
