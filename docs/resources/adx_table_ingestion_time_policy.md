---
page_title: "adx_table_ingestion_time_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the ingestion time policy of a table in ADX.
---

# Resource `adx_table_ingestion_time_policy`

Turns on or turns off a table's ingestion time policy. This policy adds a hidden datetime column in the table, called $IngestionTime. Whenever new data is ingested, the time of ingestion is recorded in the hidden column.

See: [ADX - ingestion time Policy](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/management/ingestiontimepolicy)

## Example Usage

```terraform

resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}

resource "adx_table_ingestion_time_policy" "test" {
  database_name      = "test-db"
  table_name         = adx_table.test.name
  enabled            = true
}

```

## Argument Reference

- **table_name** (String, Required) Name of the table containing the policy to modify
- **database_name** (String, Required) Database name that the target table is in
- **enabled** (Bool, Required) Determines whether to turn on or turn off the policy. true turns on the policy. false turns off the policy.


## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
