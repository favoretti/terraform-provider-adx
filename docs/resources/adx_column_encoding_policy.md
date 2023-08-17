---
page_title: "adx_column_encoding_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the encoding policy of a column in ADX.
---

# Resource `adx_column_encoding_policy`

Manages encoding policy for a column in ADX.

See: [ADX - encoding Policy](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/management/encoding-policy)

## Example Usage

```terraform

resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:dynamic,f2:string,f3:int"
}

resource "adx_column_encoding_policy" "test" {
  database_name         = "test-db"
  entity_identifier     = "Test1.f1"
  encoding_policy_type  = "BigObject"
}

```

## Argument Reference

- **database_name** (String, Required) Database name that the target column is in
- **entity_identifier** (String, Required) The identifier for the column.
- **encoding_policy_type** (String, Optional) The type of the encoding policy to apply to the specified column. If you omit the type, the existing encoding policy profile is cleared reset to the default value.


## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
