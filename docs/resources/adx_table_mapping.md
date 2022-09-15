---
page_title: "adx_table_mapping Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages a table mapping in ADX.
---

# Resource `adx_table_mapping`

Manages a table mapping in ADX.

[ADX - Data Mappings](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/mappings)

## Example Usage

```terraform
resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}

resource "adx_table_mapping" "test" {
  name          = "TestMapping"
  database_name = "test-db"
  table_name    = adx_table.test.name
  kind          = "Json"
  mapping {
    column   = "f1"
    path     = "$.f1"
    datatype = "string"
  }
  mapping {
    column   = "f2"
    path     = "$.f2"
    datatype = "string"
  }
}
```

## Argument Reference

- **name** (String, Required) Name of the Table mapping to create.
- **database_name** (String, Required) Database name in which this Table mapping should be created.
- **table_name** (String, Required) Table name in which this mapping should be created.
- **kind** (String, Required) Mapping kind. (json, csv, orc, avro, parquet, w3clogfile)
- **mapping** A `mapping` block defined below.
- **cluster** (Optional) `cluster` Configuration block (defined below) for the target cluster (overrides any config specified in the provider)

`mapping` Configures a mapping and supports the following:

- **column** (String, Required)
- **path** (String, Optional)
- **ordinal** (String, Optional)
- **field** (String, Optional)
- **constvalue** (String, Optional)
- **datatype** (String, Optional)
- **transform** (String, Optional)

`cluster` Configuration block for connection details about the target ADX cluster

*Note*: Any attributes specified here override the cluster config specified in the provider. Once a resource overrides an attribute specified in the provider, it will be stored explicitly as state for that resource and will not be possible to go back to the provider config unless explicitly unset.

- **uri** - (String, Optional) Target ADX cluster endpoint URI, starting with `https://`
- **client_id** - (String, Optional) The client ID for a service principal having admin access to this cluster/database. 
- **client_secret** - (String, Optional) The client secret for a service principal having admin access to this cluster/database
- **tenant_id** - (String, Optional) Id for the tenant to which the service principal belongs

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
