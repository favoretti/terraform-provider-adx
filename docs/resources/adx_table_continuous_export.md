---
page_title: "adx_table_continuous_export Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages an update policy for a table in ADX.
---

# Resource `adx_table_continuous_export`

Manages an continuous export for a table in ADX.

See: [ADX - Continuous Export ](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/management/data-export/continuous-data-export)

## Example Usage

```terraform

resource "adx_table" "test_table" {
  name            = "my_test_table"
  database_name   = "my_db"
  table_schema    = "name:string,color:string,number:int"
  merge_on_update = false
}


resource "adx_external_table" "test_external_table" {
    database_name               = "my_db"
    name                        = "et_my_test_external_table"
    data_format                 = "csv"
    storage_connection_string   = "your_storage_SAS_url"
    schema                      = "name:string,color:string,number:int"

    depends_on = [ adx_table.test_table ]
}


resource "adx_table_continuous_export" "test_cont_export" {
  database_name         = "my_db"
  name                  = "ce_my_test_cont_export"
  external_table_name   = "et_my_test_external_table"
  query                  = "my_test_table"
  interval_between_runs = "6m"

  depends_on = [ adx_external_table.test_external_table ]
}


```

## Argument Reference

- **name** (String, Required) The name of the continuous export. Must be unique within the database.
- **database_name** (String, Required) Database name within ADX that the target continuous export is in
- **external_table_name** (String, Required) The name of the external table export target.
- **query** (String, Required) The query to export.
- **interval_between_runs** (String, Optional) The time span between continuous export executions. Must be greater than 1 minute. Default: 10h (10:00:00)
- **forced_latency** (String, Optional) An optional period of time to limit the query to records that were ingested only prior to this period (relative to current time). This property is useful if, for example, the query performs some aggregations/joins and you would like to make sure all relevant records have already been ingested before running the export.
- **size_limit** (Int, Optional) The size limit in bytes of a single storage artifact being written (prior to compression). Allowed range is 100 MB (default) to 1 GB.
- **distributed** (Bool, Optional) Disable/enable distributed export. Setting to false is equivalent to single distribution hint. Default is true.
- **parquet_row_group_size** (Int, Optional) Relevant only when data format is Parquet. Controls the row group size in the exported files. Default row group size is 100,000 records.
- **use_native_parquet_writer** (Bool, Optional) Use the new export implementation when exporting to Parquet, this implementation is a more performant, resource light export mechanism. Note that an exported 'datetime' column is currently unsupported by Synapse SQL 'COPY'. Default is false.
- **managed_identity** (String, Optional) The managed identity on behalf of which the continuous export job will run. The managed identity can be an object ID, or the system reserved word. For more information, see Use a managed identity to run a continuous export job.
- **is_disabled** (Bool, Optional) Disable/enable the continuous export. Default is false.
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
