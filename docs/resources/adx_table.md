---
page_title: "adx_table Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages a table in ADX.
---

# Resource `adx_table`

Manages a table in ADX.

## Example Usage

You can pass table schema as a string:

```terraform
resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}
```

Or use HCL to construct it:

```
resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"

  column {
    name = "f1"
    type = "string"
  }

  column {
    name = "f2"
    type = "string"
  }

  column {
    name = "f3"
    type = "int"
  }
}
```

### Argument Reference

- **name** (String, Required) Name of the Table to create. Changing this forces a new resource to be created.
- **database_name** (String, Required) Database name in which this Table should be created. Changing this forces a new resource to be created.
- **table_schema** (String, Optional) Table schema. Must contain only letters, numbers, dashes, semicolons, commas and underscores and no spaces. Changing this forces a new resource to be created.
- **column** (String, Optional) One or more `column` blocks defined below.

`column` Configures a column and supports the following:

- **name** (String, Required) Column name
- **type** (String, Required) Column type

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
