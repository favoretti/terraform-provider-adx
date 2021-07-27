---
page_title: "adx_table_mapping Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages a table mapping in ADX.
---

# Resource `adx_table_mapping`

Manages a table mapping in ADX.

## Example Usage

```terraform
resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}

resource "adx_table_mapping" "test" {
  name          = "TestMapping"
  database_name = "test-db"
  table_name    = adx_table.test.name
  kind          = "Json"
  mapping {
    column   = "f1"
    path     = "$.f1"
    datatype = "string"
  }
  mapping {
    column   = "f2"
    path     = "$.f2"
    datatype = "string"
  }
}
```

### Argument Reference

- **name** (String, Required) Name of the Table mapping to create.
- **database_name** (String, Required) Database name in which this Table mapping should be created.
- **table_name** (String, Required) Table name in which this mapping should be created.
- **kind** (String, Required) Mapping kind. The only currently supported value is `Json`.
- **mapping** A `mapping` block defined below.

`mapping` Configures a mapping and supports the following:

- **column** (String, Required)
- **path** (String, Required)
- **datatype** (String, Required)
- **transform** (String, Optional)

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
