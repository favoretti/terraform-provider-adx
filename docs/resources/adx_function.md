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

### Argument Reference

- **name** (String, Required) Name of the function to create.
- **database_name** (String, Required) Database name in which this function should be created.
- **body** (String, Required) Function body enclosed in curly braces {}
- **parameters** (String, Optional) Function parameters enclosed in parenthesis (myLimit:long)

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
