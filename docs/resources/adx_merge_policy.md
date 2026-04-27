---
page_title: "adx_merge_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the merge policy of a database, table, or materialized view in ADX.
---

# Resource `adx_merge_policy`

Manages the merge policy for a database, table, or materialized view in Azure Data Explorer. The merge policy defines if and how extents (data shards) should get merged.

A single resource handles all three scopes via the `entity_type` attribute.

See: [ADX - Merge Policy](https://learn.microsoft.com/en-us/kusto/management/merge-policy?view=azure-data-explorer)

## Example Usage

### Table Merge Policy

```terraform
resource "adx_merge_policy" "example" {
  database_name      = "test-db"
  entity_type        = "table"
  entity_name        = "my_table"
  max_range_in_hours = 48
}
```

### Database Merge Policy

```terraform
resource "adx_merge_policy" "example" {
  database_name      = "test-db"
  entity_type        = "database"
  max_range_in_hours = 48
}
```

### Materialized View Merge Policy

```terraform
resource "adx_merge_policy" "example" {
  database_name      = "test-db"
  entity_type        = "materialized_view"
  entity_name        = "my_materialized_view"
  max_range_in_hours = 336
  lookback_kind      = "HotCache"
}
```

### Custom Lookback Period

```terraform
resource "adx_merge_policy" "example" {
  database_name          = "test-db"
  entity_type            = "table"
  entity_name            = "my_table"
  lookback_kind          = "Custom"
  lookback_custom_period = "14.00:00"
}
```

## Argument Reference

- **database_name** (String, Required) Database name of the target entity.
- **entity_type** (String, Required) The scope of the merge policy. Must be one of: `database`, `table`, or `materialized_view`.
- **entity_name** (String, Optional) Name of the table or materialized view. Required when `entity_type` is `table` or `materialized_view`. Not used when `entity_type` is `database`.
- **row_count_upper_bound_for_merge** (Int, Optional) Maximum allowed row count of the merged extent. Applies to Merge operations, not Rebuild. Default: `16000000`.
- **original_size_mb_upper_bound_for_merge** (Int, Optional) Maximum allowed original size (in MBs) of the merged extent. Applies to Merge operations, not Rebuild. Default: `30000`.
- **max_extents_to_merge** (Int, Optional) Maximum allowed number of extents to be merged in a single operation. Applies to Merge operations. Default: `100`.
- **max_range_in_hours** (Int, Optional) Maximum allowed difference, in hours, between any two different extents' creation times so that they can still be merged. Default: `24`.
- **allow_rebuild** (Bool, Optional) Defines whether Rebuild operations are enabled. Default: `true`.
- **allow_merge** (Bool, Optional) Defines whether Merge operations are enabled. Default: `true`.
- **lookback_kind** (String, Optional) Defines the timespan during which extents are considered for rebuild/merge. Must be one of: `Default`, `All`, `HotCache`, or `Custom`. Default: `Default`.
- **lookback_custom_period** (String, Optional) Custom timespan period in the format `dd.hh:mm`. Required when `lookback_kind` is `Custom`.
- **cluster** (Optional) `cluster` configuration block (defined below) for the target cluster (overrides any config specified in the provider).

`cluster` Configuration block for connection details about the target ADX cluster

*Note*: Any attributes specified here override the cluster config specified in the provider. Once a resource overrides an attribute specified in the provider, it will be stored explicitly as state for that resource and will not be possible to go back to the provider config unless explicitly unset.

- **uri** - (String, Optional) Target ADX cluster endpoint URI, starting with `https://`
- **client_id** - (String, Optional) The client ID for a service principal having admin access to this cluster/database.
- **client_secret** - (String, Optional) The client secret for a service principal having admin access to this cluster/database
- **tenant_id** - (String, Optional) Id for the tenant to which the service principal belongs

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.

## Import

Merge policies can be imported using the resource ID format:

```shell
# Table scope
terraform import adx_merge_policy.example <cluster_endpoint>|<database_name>|table|<table_name>|policy|merge

# Database scope
terraform import adx_merge_policy.example <cluster_endpoint>|<database_name>|database|<database_name>|policy|merge

# Materialized view scope
terraform import adx_merge_policy.example <cluster_endpoint>|<database_name>|materialized-view|<view_name>|policy|merge
```
