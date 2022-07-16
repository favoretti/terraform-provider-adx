---
page_title: "adx_table Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages a table in ADX.
---

# Resource `adx_table`

Manages a table in ADX.

[https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/tables](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/tables)

## Example Usage

You can pass table schema as a string:

```terraform
resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}
```

Or use HCL to construct it:

```terraform
resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"

  column {
    name = "f1"
    type = "string"
  }

  column {
    name = "f2"
    type = "string"
  }

  column {
    name = "f3"
    type = "int"
  }
}
```

Or create it from the results of a query:

```terraform
resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"

  from_query {
    query = "OtherTable | limit 0"
  }
}
```

### Argument Reference

- **name** (String, Required) Name of the Table to create.
- **database_name** (String, Required) Database name in which this Table should be created.
- **table_schema** (String, Optional) Table schema. Must contain only letters, numbers, dashes, semicolons, commas and underscores and no spaces.
- **column** (String, Optional) One or more `column` blocks defined below.
- **from_query** (String, Optional) One `from_query` blocks defined below.
- **merge_on_update** (Boolean, Optional) If true, prevent removal of columns or configuration during schema changes. Changes become additive only. See Azure docs on difference between `.alter` and `.alter-merge`. Default is false

`column` Configures a column and supports the following:

- **name** (String, Required) Column name
- **type** (String, Required) Column type

`from_query` Configures the table from the result of a query and supports the following:

See [ADX - Ingest from Query](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/data-ingestion/ingest-from-query) for behavioral details of these parameters

- **query** (String, Required) Result of this query will be used to build the target table
- **append** (Boolean, Optional) If table already contains data, add to it instead of replacing. Default is false
- **extend_schema** (Boolean, Optional) True if the command may extend the schema of the table. Default is "false". Only applied for updates
- **recreate_schema** (Boolean, Optional) True if the command may recreate the schema of the table. Default is "false". Only applied for updates. Takes precedence over recreate_schema
- **distributed** (Boolean, Optional) Indicates that the command ingests from all nodes executing the query in parallel. Default is "false"
- **force_an_update_when_value_changed** (String, Optional) A unique string. If changed the script will be applied again. Default is ""

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.

## Changing the table schema

If you alter a table, altering a column type isn't supported. Use the .alter column command instead directly against the cluster.

Please refer to this doc to understand limitations of schema changes and possible data loss scenarios:
[https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/alter-table-command](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/alter-table-command)