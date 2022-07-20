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

### Argument Reference

- **table_name** (String, Required) Name of the table containing the policy to modify
- **database_name** (String, Required) Database name that the target table is in
- **query** (String, Required) The query to be run automatically when the target table is queried
- **enabled** (Boolean, Optional) Enable or disable this policy

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
