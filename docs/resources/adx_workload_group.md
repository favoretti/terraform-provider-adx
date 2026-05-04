---
page_title: "adx_workload_group Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages a workload group in ADX.
---

# Resource `adx_workload_group`

Manages a workload group in ADX (Azure Data Explorer). Workload groups allow you to group together sets of management commands and queries based on shared characteristics, and apply policies to control per-request limits and request rate limits.

See: [ADX - Workload Groups](https://learn.microsoft.com/en-us/kusto/management/workload-groups)

## Example Usage

### Basic workload group with request rate limits

```terraform
resource "adx_workload_group" "adhoc_queries" {
  database_name = "my-database"
  name          = "Ad-hoc queries"

  request_rate_limit_policies = jsonencode([
    {
      IsEnabled = true
      Scope     = "WorkloadGroup"
      LimitKind = "ConcurrentRequests"
      Properties = {
        MaxConcurrentRequests = 100
      }
    },
    {
      IsEnabled = true
      Scope     = "Principal"
      LimitKind = "ConcurrentRequests"
      Properties = {
        MaxConcurrentRequests = 25
      }
    }
  ])
}
```

### Workload group with request limits and rate limits

```terraform
resource "adx_workload_group" "reports" {
  database_name = "my-database"
  name          = "Reports"

  request_limits_policy = jsonencode({
    DataScope = {
      IsRelaxable = true
      Value       = "HotCache"
    }
    MaxMemoryPerQueryPerNode = {
      IsRelaxable = false
      Value       = 6442450944
    }
    MaxExecutionTime = {
      IsRelaxable = true
      Value       = "00:04:00"
    }
    MaxResultRecords = {
      IsRelaxable = true
      Value       = 500000
    }
    MaxResultBytes = {
      IsRelaxable = true
      Value       = 67108864
    }
  })

  request_rate_limit_policies = jsonencode([
    {
      IsEnabled = true
      Scope     = "WorkloadGroup"
      LimitKind = "ConcurrentRequests"
      Properties = {
        MaxConcurrentRequests = 50
      }
    }
  ])

  query_consistency_policy = jsonencode({
    QueryConsistency = {
      IsRelaxable = true
      Value       = "Weak"
    }
  })
}
```

## Argument Reference

- **name** (String, Required) Name of the workload group. Up to 16 custom workload groups can be defined per cluster.
- **database_name** (String, Required) Database name used as context for the management command. Workload groups are cluster-level resources.
- **request_limits_policy** (String, Optional) JSON representation of the request limits policy. Controls per-request resource limits such as memory, execution time, and result size.
- **request_rate_limit_policies** (String, Optional) JSON representation of the request rate limit policies array. Controls how many concurrent requests are allowed per workload group or principal.
- **request_rate_limits_enforcement_policy** (String, Optional) JSON representation of the request rate limits enforcement policy. Controls at which level (Cluster, Database, QueryHead) rate limits are enforced.
- **request_queuing_policy** (String, Optional) JSON representation of the request queuing policy. When enabled, requests are queued instead of rejected when rate limits are exceeded.
- **query_consistency_policy** (String, Optional) JSON representation of the query consistency policy. Controls the consistency mode (Strong/Weak) for queries in this workload group.
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

## Import

Workload groups can be imported using the resource ID format:

```shell
terraform import adx_workload_group.example <cluster_endpoint>|<database_name>|workload_group|<workload_group_name>
```
