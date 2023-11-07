---
page_title: "adx_external_table Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages an update policy for a table in ADX.
---

# Resource `adx_table_continuous_export`

Manages external tables in ADX.

See: [ADX - External Table ](https://learn.microsoft.com/en-us/azure/data-explorer/kusto/management/external-tables-azurestorage-azuredatalake)

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

- **name** (String, Required) An external table name that adheres to the entity names rules. An external table can't have the same name as a regular table in the same database.
- **database_name** (String, Required) Database name within ADX that the target external table is in
- **schema** (String, Required) The external data schema is a comma-separated list of one or more column names and data types, where each item follows the format: ColumnName : ColumnType. If the schema is unknown, use infer_storage_schema to infer the schema based on external file contents.
- **data_format** (String, Required) The data format, which can be any of the ingestion formats. We recommend using the Parquet format for external tables to improve query and export performance, unless you use JSON paths mapping. When using an external table for export scenario, you're limited to the following formats: CSV, TSV, JSON and Parquet.
- **storage_connection_string** (String, Required) One or more comma-separated paths to Azure Blob Storage blob containers, Azure Data Lake Gen 2 file systems or Azure Data Lake Gen 1 containers, including credentials. The external table storage type is determined by the provided connection strings. See storage connection strings.
- **partitions** (String, Optional) A comma-separated list of columns by which the external table is partitioned. Partition column can exist in the data file itself, or as part of the file path. See partitions formatting to learn how this value should look.
- **path_format** (String, Optional) An external data folder URI path format to use with partitions. See path format.
- **folder** (String, Optional) Table's folder
- **doc_string** (String, Optional) String documenting the table
- **compressed** (Bool, Optional) If set, indicates whether the files are compressed as .gz files (used in export scenario only)
- **include_headers** (String, Optional) For delimited text formats (CSV, TSV, ...), indicates whether files contain a header. Possible values are: All (all files contain a header), FirstFile (first file in a folder contains a header), None (no files contain a header). Default: All
- **name_prefix** (String, Optional) If set, indicates the prefix of the files. On write operations, all files will be written with this prefix. On read operations, only files with this prefix are read.
- **file_extension** (String, Optional) If set, indicates file extensions of the files. On write, files names will end with this suffix. On read, only files with this file extension will be read.
- **encoding** (String, Optional) Indicates how the text is encoded: UTF8NoBOM (default) or UTF8BOM.
- **sample_uris** (Bool, Optional) If set, the command result provides several examples of simulated external data files URI as they're expected by the external table definition. This option helps validate whether the Partitions and PathFormat parameters are defined properly.
- **files_preview** (Bool, Optional) If set, one of the command result tables contains a preview of .show external table artifacts command. Like sampleUri, the option helps validate the Partitions and PathFormat parameters of external table definition.
- **validate_not_empty** (Bool, Optional) If set, the connection strings are validated for having content in them. The command will fail if the specified URI location doesn't exist, or if there are insufficient permissions to access it.
- **dry_run** (Bool, Optional) 	If set, the external table definition isn't persisted. This option is useful for validating the external table definition, especially in conjunction with the filesPreview or sampleUris parameter. 
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
