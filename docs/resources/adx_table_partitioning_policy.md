---
page_title: "adx_table_partitioning_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the partitioning policy of a table in ADX.
---

# Resource `adx_table_partitioning_policy`

Manages partitioning policy for a table in ADX.

See: [ADX - Partitioning Policy](https://docs.microsoft.com/en-us/azure/data-explorer/kusto/management/partitioningpolicy)

## Example Usage

```terraform

resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}

resource "adx_table_partitioning_policy" "test" {
  database_name       = "test-db"
  table_name          = adx_table.test.name
  effective_date_time = "2022-07-19T13:56:45Z"

  partition_key {
    column_name = "f1"
    kind        = "Hash"

    hash_properties {
      function                  = "XxHash64"
      max_partition_count       = 64
      seed                      = 2
      partition_assignment_mode = "Uniform"
    }
  }
}

```

### Argument Reference

- **table_name** (String, Required) Name of the table containing the policy to modify
- **database_name** (String, Required) Database name that the target table is in
- **effective_date_time** (String, Required) ISO8601 Timestamp after which this policy will apply
- **partition_key** (Optional) One to two `partition_key` blocks defined below
- **cluster** (Optional) `cluster` Configuration block (defined below) for the target cluster (overrides any config specified in the provider)

`partition_key` Configures a single partition key (ADX allows up to two) and supports the following:

- **column_name** (String, Required) Column name on which to partition
- **kind** (String, Required) Partition strategy (ADX Supports 'Hash' and 'UniformRange')
- **hash_properties** (Optional) `hash_properties` Block defined below (Required if kind is 'Hash')
- **uniform_range_properties** (Optional) `uniform_range_properties` Block defined below (Required if kind is 'UniformRange')

`hash_properties` Configuration block for 'Hash' partition strategy

- **function** (String, Required) A timespan scalar constant that indicates the size of each datetime partition (Ex: '1.00:00:00' for 1 day)
- **max_partition_count** (Int, Required) Maximum number of partitions to create using this key
- **seed** (Integer, Optional) Positive integer used for randomizing the hash value (Default: 1)
- **partition_assignment_mode** (String, Optional) The mode used for assigning partitions to nodes in the cluster ('Default' or 'Uniform') (Default is 'Default')

`uniform_range_properties` Configuration block for 'UniformRange' partition strategy

- **range_size** (String, Required) A timespan scalar constant that indicates the size of each datetime partition (Ex: '1.00:00:00' for 1 day)
- **reference** (String, Required) ISO8601 String that indicates a fixed point in time, according to which datetime partitions are aligned (Ex: '1970-01-01 00:00:00')
- **override_creation_time** (Boolean, Optional) Indicates whether or not the result extent's minimum and maximum creation times should be overridden by the range of the values in the partition key. (Default is false)

There can be up to one column on which 'Hash' partitioning is defined

`cluster` Configuration block for connection details about the target ADX cluster 

*Note*: Any attributes specified here override the cluster config specified in the provider. Once a resource overrides an attribute specified in the provider, it will be stored explicitly as state for that resource and will not be possible to go back to the provider config.

- **cluster_uri** - (String, Optional) Target ADX cluster endpoint URI, starting with `https://`
- **client_id** - (String, Optional) The client ID for a service principal having admin access to this cluster/database. 
- **client_secret** - (String, Optional) The client secret for a service principal having admin access to this cluster/database
- **tenant_id** - (String, Optional) Id for the tenant to which the service principal belongs

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
