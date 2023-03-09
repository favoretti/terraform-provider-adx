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
  allow_mv_without_rls         = true
}
```

## Argument Reference

- **name** (String, Required) Name of the function to create.
- **database_name** (String, Required) Database name in which this function should be created.
- **source_table_name** (String, Required) Name of the table being queried to produce a view
- **query** (String, Required) ADX Query which produces the desired view to materialize
- **backfill** (Boolean, Optional) Whether to create the view based on all records currently in `source_table_name` (true), or to create it "from-now-on" (false). Default is false
- **async** (Boolean, Optional) Allows materialization to happen as a background process. If set to true, failures will not be captured by Terraform and success will be assumed. Required to be true if `backfill` is set to true. Default is false
- **effective_date_time** (String, Optional) ISO8601 Date time string. If set, creation only backfills with records ingested after the datetime. `backfill` must also be set to true.
- **auto_update_schema** (Boolean, Optional) Whether to auto-update the view on source table changes. Default is false. This option is valid only for views of type `arg_max(Timestamp,*)`, `arg_min(Timestamp, *)`, `take_any(*)` (only when columns argument is *). If this option is set to true, changes to source table will be automatically reflected in the materialized view.
- **update_extents_creation_time** (Boolean, Optional) Relevant only when using `backfill`. If true, extent creation time is assigned based on datetime group-by key during the backfill process
- **allow_mv_without_rls** (Boolean, Optional) Enables `allowMaterializedViewsWithoutRowLevelSecurity` flag during policy creation
- **folder** (String, Optional) Name of the folder in which to place this entity
- **docstring** (String, Optional) Free text describing the entity to be added. This string is presented in various UX settings next to the entity names.
- **max_source_records_for_single_ingest** (Int, Optional) By default, the number of source records in each ingest operation during backfill is 2 million per node. You can change this default by setting this property to the desired number of records. (The value is the total number of records in each ingest operation.)
- **concurrency** (Int, Optional) The ingest operations, running as part of the backfill process, run concurrently. By default, concurrency is min(number_of_nodes * 2, 5).
- **cluster** (Optional) `cluster` Configuration block (defined below) for the target cluster (overrides any config specified in the provider)

`cluster` Configuration block for connection details about the target ADX cluster 

*Note*: Any attributes specified here override the cluster config specified in the provider. Once a resource overrides an attribute specified in the provider, it will be stored explicitly as state for that resource and will not be possible to go back to the provider config unless explicitly unset.

- **uri** - (String, Optional) Target ADX cluster endpoint URI, starting with `https://`
- **client_id** - (String, Optional) The client ID for a service principal having admin access to this cluster/database.
- **client_secret** - (String, Optional) The client secret for a service principal having admin access to this cluster/database
- **tenant_id** - (String, Optional) Id for the tenant to which the service principal belongs

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
