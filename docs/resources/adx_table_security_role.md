---
page_title: "adx_table_security_role Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages a security role assignment on a table in ADX.
---

# Resource `adx_table_security_role`

Manages a security role assignment (principal) on a table in ADX.

See: [ADX - Manage table security roles](https://learn.microsoft.com/en-us/kusto/management/manage-table-security-roles)

## Example Usage

### Grant a user the admin role on a table

```terraform
resource "adx_table" "test" {
  name          = "StormEvents"
  database_name = "test-db"
  table_schema  = "StartTime:datetime,EndTime:datetime,EventType:string"
}

resource "adx_table_security_role" "admin" {
  database_name = "test-db"
  table_name    = adx_table.test.name
  role          = "admins"
  principal_fqn = "aaduser=user@example.com"
  notes         = "Granted via Terraform"
}
```

### Grant an application the ingestor role on a table

```terraform
resource "adx_table_security_role" "ingestor" {
  database_name = "test-db"
  table_name    = adx_table.test.name
  role          = "ingestors"
  principal_fqn = "aadapp=4c7e82bd-6adb-46c3-b413-fdd44834c69b;contoso.com"
}
```

### Grant multiple users the same role using `for_each`

Each resource instance manages a single principal. Use `for_each` to assign the same role to multiple principals:

```terraform
resource "adx_table_security_role" "ingestors" {
  for_each = toset([
    "aaduser=user1@example.com",
    "aaduser=user2@example.com",
    "aadapp=4c7e82bd-6adb-46c3-b413-fdd44834c69b;contoso.com",
  ])

  database_name = "test-db"
  table_name    = adx_table.test.name
  role          = "ingestors"
  principal_fqn = each.value
}
```

## Argument Reference

- **table_name** (String, Required) Name of the table on which to manage the security role.
- **database_name** (String, Required) Database name that the target table is in.
- **role** (String, Required) The security role to assign. Must be one of `admins` or `ingestors`.
- **principal_fqn** (String, Required) Fully qualified name of the principal, e.g. `aaduser=user@example.com`, `aadapp=<app-id>;<tenant>`, or `aadgroup=group@example.com`.
- **notes** (String, Optional) Free text describing the role assignment. Displayed when using the `.show` command.
- **cluster** (Optional) `cluster` Configuration block (defined below) for the target cluster (overrides any config specified in the provider).

`cluster` Configuration block for connection details about the target ADX cluster.

*Note*: Any attributes specified here override the cluster config specified in the provider. Once a resource overrides an attribute specified in the provider, it will be stored explicitly as state for that resource and will not be possible to go back to the provider config unless explicitly unset.

- **uri** - (String, Optional) Target ADX cluster endpoint URI, starting with `https://`
- **client_id** - (String, Optional) The client ID for a service principal having admin access to this cluster/database.
- **client_secret** - (String, Optional) The client secret for a service principal having admin access to this cluster/database.
- **tenant_id** - (String, Optional) Id for the tenant to which the service principal belongs.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

- **id** - The ID of this resource.
- **principal_type** - The type of the principal (e.g., `AAD User`, `AAD App`, `AAD Group`).
- **principal_display_name** - The display name of the principal.

## Import

Table security roles can be imported using the resource ID format:

```
terraform import adx_table_security_role.example "<cluster_uri>|<database_name>|table|<table_name>|security_role|<role>|<principal_fqn>"
```

Example:

```
terraform import adx_table_security_role.admin "mycluster.eastus.kusto.windows.net|mydb|table|StormEvents|security_role|admins|aaduser=user@example.com"
```
