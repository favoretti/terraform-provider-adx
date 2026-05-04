---
page_title: "adx_cluster_request_classification_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the cluster-level request classification policy in ADX.
---

# Resource `adx_cluster_request_classification_policy`

Manages the cluster-level request classification policy in ADX (Azure Data Explorer). The request classification policy assigns incoming requests to workload groups based on request characteristics using a user-defined KQL function.

See: [ADX - Request Classification Policy](https://learn.microsoft.com/en-us/kusto/management/request-classification-policy)

## Example Usage

### Simple classification to a single workload group

```terraform
resource "adx_cluster_request_classification_policy" "example" {
  database_name = "my-database"
  is_enabled    = true

  classification_function = <<-EOT
    iff(request_properties.current_application == "Kusto.Explorer" and request_properties.request_type == "Query",
        "Ad-hoc queries",
        "default")
  EOT
}
```

### Classification to multiple workload groups

```terraform
resource "adx_cluster_request_classification_policy" "example" {
  database_name = "my-database"
  is_enabled    = true

  classification_function = <<-EOT
    case(
      current_principal_is_member_of('aadgroup=admins@contoso.com'), "Premium",
      request_properties.current_database == "Analytics" and request_properties.current_principal has 'aadapp=', "Automated",
      request_properties.current_application == "Kusto.Explorer" and request_properties.request_type == "Query", "Ad-hoc queries",
      request_properties.current_application == "KustoQueryRunner", "Scheduled",
      "default")
  EOT
}
```

## Argument Reference

- **database_name** (String, Required) Database name used as context for the management command. The policy is cluster-level.
- **is_enabled** (Bool, Required) Whether the request classification policy is enabled.
- **classification_function** (String, Required) The body of the KQL function used for classifying requests into workload groups. The function must return a string (the workload group name) and has access to `request_properties` with fields like `current_database`, `current_application`, `current_principal`, `request_type`, `request_description`, etc.
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

The cluster request classification policy can be imported using the resource ID format:

```shell
terraform import adx_cluster_request_classification_policy.example <cluster_endpoint>|<database_name>|cluster|request_classification
```

## Notes

- Only one request classification policy can be defined per cluster.
- The classification function is evaluated for every incoming request — keep it lightweight.
- The function must not reference any other entity (database, table, or function).
- If the function returns an empty string, "default", or a non-existent workload group name, the request is assigned to the `default` workload group.
