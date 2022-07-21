---
page_title: "adx_materialized_view Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages a materialized view in ADX.
---

# Resource `adx_materialized_view`

Manages a materialized view in ADX.

See: [ADX - Materialized Views](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/materialized-views/materialized-view-overview)

## Example Usage

```terraform
resource "adx_materialized_view" "test" {
  name                         = "test_mv"
  database_name                = "test-db"
  source_table_name            = adx_table.test.name
  query                        = format("%s | extend hi=true | summarize count(), dcount(f1) by f2",adx_table.test.name)
}
```

### Argument Reference

- **name** (String, Required) Name of the function to create.
- **database_name** (String, Required) Database name in which this function should be created.
- **source_table_name** (String, Required) Name of the table being queried to produce a view
- **query** (String, Required) ADX Query which produces the desired view to materialize
- **backfill** (Boolean, Optional) Whether to create the view based on all records currently in `source_table_name` (true), or to create it "from-now-on" (false). Default is false
- **async** (Boolean, Optional) Allows materialization to happen as a background process. If set to true, failures will not be captured by Terraform and success will be assumed. Required to be true if `backfill` is set to true. Default is false
- **effective_date_time** (String, Optional) ISO8601 Date time string. If set, creation only backfills with records ingested after the datetime. `backfill` must also be set to true. 
- **auto_update_schema** (Boolean, Optional) Whether to auto-update the view on source table changes. Default is false. This option is valid only for views of type `arg_max(Timestamp,*)`, `arg_min(Timestamp, *)`, `take_any(*)` (only when columns argument is *). If this option is set to true, changes to source table will be automatically reflected in the materialized view.
- **update_extents_creation_time** (Boolean, Optional) Relevant only when using `backfill`. If true, extent creation time is assigned based on datetime group-by key during the backfill process

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
