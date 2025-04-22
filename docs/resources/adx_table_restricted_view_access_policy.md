---
page_title: "adx_table_restricted_view_access_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the restricted_view_access policy of a table in ADX.
---

# Resource `adx_table_restricted_view_access_policy`

Manages restricted view access policy for a table in ADX.

See: [ADX - restricted view access Policy](https://learn.microsoft.com/en-us/kusto/management/restricted-view-access-policy)

## Example Usage

```terraform

resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}


resource "adx_table_restricted_view_access_policy" "test" {
  database_name      = "test-db"
  table_name         = adx_table.test.name
  enabled            = true
}

```

## Argument Reference

- **table_name** (String, Required) Name of the table containing the policy to modify
- **database_name** (String, Required) Database name that the target table is in
- **enabled** (boolean or String, Required) Enables or disables policy
- **follower_database** (Bool, Optional) True if the target table is from an attached/follower database
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
