---
page_title: "adx_function Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages a user defined function in ADX.
---

# Resource `adx_function`

Manages a user defined function in ADX.

See: [ADX - Create Stored Function](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/create-function)

## Example Usage

```terraform
resource "adx_function" "test" {
  database_name = "test-db"
  name          = "my_function"
  body          = "{table1 | limit myLimit}"
  parameters    = "(myLimit:long)"
}
```

## Argument Reference

- **name** (String, Required) Name of the function to create.
- **database_name** (String, Required) Database name in which this function should be created.
- **body** (String, Required) Function body enclosed in curly braces {}
- **parameters** (String, Optional) Function parameters enclosed in parenthesis (myLimit:long)
- **cluster** (Optional) `cluster` Configuration block (defined below) for the target cluster (overrides any config specified in the provider)

`cluster` Configuration block for connection details about the target ADX cluster 

*Note*: Any attributes specified here override the cluster config specified in the provider. Once a resource overrides an attribute specified in the provider, it will be stored explicitly as state for that resource and will not be possible to go back to the provider config unless explicitly unset.

- **cluster_uri** - (String, Optional) Target ADX cluster endpoint URI, starting with `https://`
- **client_id** - (String, Optional) The client ID for a service principal having admin access to this cluster/database.
- **client_secret** - (String, Optional) The client secret for a service principal having admin access to this cluster/database
- **tenant_id** - (String, Optional) Id for the tenant to which the service principal belongs

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
