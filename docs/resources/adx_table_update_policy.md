---
page_title: "adx_table_update_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages an update policy for a table in ADX.
---

# Resource `adx_table_update_policy`

Manages an update policy for a table in ADX.

See: [ADX - Update Policy](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/updatepolicy)

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
  body          = format("{%s | extend timestamp=ingestion_time()}", adx_table.test.name)
}

resource "adx_table" "test_update" {
  name          = "test_update"
  database_name = "test-db"

  table_schema = "f1:string,f2:string,f3:int"
}

resource "adx_table_update_policy" "test_update" {
  database_name = "test-db"
  table_name    = adx_table.test_update.name
  query         = adx_function.test.name
  source_table  = adx_table.test.name
  transactional = true
}

```

### Argument Reference

- **table_name** (String, Required) Name of the table containing the policy to modify
- **database_name** (String, Required) Database name that the target table is in
source_table
- **query** (String, Required) A query used to produce data for the update
- **transactional** (Boolean, Required) States if the update policy is transactional or not, default is false). If transactional and the update policy fails, the source table is not updated.
- **propagate_ingestion_properties** (Boolean, Optional) States if properties specified during ingestion to the source table, such as extent tags and creation time, apply to the target table. Default: false
- **enabled** (Boolean, Optional) States if update policy is enabled or disabled. Default: true

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
