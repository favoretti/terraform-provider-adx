---
page_title: "adx_table_caching_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the caching policy of a table in ADX.
---

# Resource `adx_table_caching_policy`

Manages caching policy for a table in ADX.

See: [ADX - caching Policy](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/cachepolicy)

## Example Usage

```terraform

resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}

resource "adx_table_caching_policy" "test" {
  database_name      = "test-db"
  table_name         = adx_table.test.name
  data_hot_span      = "30d"
}

```

### Argument Reference

- **table_name** (String, Required) Name of the table containing the policy to modify
- **database_name** (String, Required) Database name that the target table is in
- **data_hot_span** (String, Required) Timespan to store rows in SSD hot cache (Example: 30d for 30 days)

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
