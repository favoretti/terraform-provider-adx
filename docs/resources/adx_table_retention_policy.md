---
page_title: "adx_table_retention_policy Resource - terraform-provider-adx"
subcategory: ""
description: |-
  Manages the retention policy of a table in ADX.
---

# Resource `adx_table_retention_policy`

Manages retention policy for a table in ADX.

See: [ADX - retention Policy](https://learn.microsoft.com/en-us/kusto/management/retention-policy)

## Example Usage

```terraform

resource "adx_table" "test" {
  name          = "Test1"
  database_name = "test-db"
  table_schema  = "f1:string,f2:string,f3:int"
}

resource "adx_table_retention_policy" "test" {
  database_name      = "test-db"
  table_name         = adx_table.test.name
  soft_delete_period = "500m"
  recoverability     = false
}

```

## Argument Reference

- **table_name** (String, Required) Name of the table containing the policy to modify
- **database_name** (String, Required) Database name that the target table is in
- **soft_delete_period** (String, Required) Time span for which it's guaranteed that the data is kept available to query, in the format `<amount><unit>` where unit is one of `d`, `h`, `m`, `s` (for example `500m` or `30d`), up to a maximum of `364635d` (999 years). The period is measured starting from the time the data was ingested (see note in ADX docs about this being imprecise). A zero period such as `0s` is also valid and means no data is kept at all; ADX then requires every update policy referencing this table as its source to be transactional
- **recoverability** (Boolean, Required) Data recoverability (true/false) after the data was soft-deleted. Must be `false` when `soft_delete_period` is `0s`, since ADX does not commit the ingested data in that case
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
