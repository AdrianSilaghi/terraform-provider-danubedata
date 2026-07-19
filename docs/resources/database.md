# danubedata_database

Manages a managed database instance (MySQL, PostgreSQL, or MariaDB).

## Example Usage

### PostgreSQL Database

```hcl
resource "danubedata_database" "postgres" {
  name             = "analytics-db"
  database_name    = "analytics"
  engine           = "postgresql"
  version          = "18"
  resource_profile = "small"
  datacenter       = "fsn1"
}

output "postgres_endpoint" {
  value = danubedata_database.postgres.endpoint
}
```

### MySQL Database

```hcl
resource "danubedata_database" "mysql" {
  name             = "app-db"
  database_name    = "app_production"
  engine           = "mysql"
  version          = "8.4"
  resource_profile = "medium"
  datacenter       = "fsn1"
}
```

### MariaDB with Additional Storage

`storage_size_gb` defaults to the profile's included storage. Set it higher to
provision extra storage; it can be grown later but never shrunk.

```hcl
resource "danubedata_database" "mariadb" {
  name             = "legacy-db"
  database_name    = "legacy_app"
  engine           = "mariadb"
  resource_profile = "small"
  storage_size_gb  = 100
  datacenter       = "fsn1"
}
```

## Resource Profiles

`resource_profile` selects the plan, and it is the only place CPU, memory and
included storage are set. Use the **slug** in the left column — the dashboard and
pricing page show the display name, which is not a valid value here.

| Slug     | Display name | vCPU | RAM  | Included storage |
| -------- | ------------ | ---- | ---- | ---------------- |
| `micro`  | DD Puiu      | 1    | 1 GB | 10 GB            |
| `small`  | DD Uzlina    | 1    | 2 GB | 20 GB            |
| `medium` | DD Matita    | 2    | 4 GB | 50 GB            |
| `large`  | DD Sinoe     | 4    | 8 GB | 100 GB           |

For current pricing see <https://danubedata.ro/pricing>. Profiles above your
account's limit are rejected at apply time; request an increase from Account
Limits in the dashboard.

## Argument Reference

### Required

* `name` - Name of the database instance. Lowercase alphanumeric and hyphens
  only (DNS compatible), 2-63 characters. Changing this forces a new resource.
* `engine` - Database engine. One of `mysql`, `postgresql`, `mariadb`.
  Changing this forces a new resource.
* `resource_profile` - Plan slug; see [Resource Profiles](#resource-profiles).
* `datacenter` - Datacenter location. One of `fsn1`, `nbg1`, `hel1`.
  Changing this forces a new resource.

### Optional

* `database_name` - Name of the initial database to create. Must start with a
  letter and contain only letters, numbers and underscores, max 64 characters.
  **Hyphens are not permitted**, even though the instance `name` allows them.
  Changing this forces a new resource.
* `version` - Engine version, e.g. `18` for PostgreSQL or `8.4` for MySQL.
  Defaults to the current default version for the engine. Available versions
  are validated by the API; see the `danubedata_database_providers` data source
  for the current list.
* `storage_size_gb` - Storage in GB. Defaults to the profile's included
  storage. May only be increased — the API rejects shrinking.
* `parameter_group_id` - ID of a parameter group for custom engine
  configuration. Must match the instance's engine and version.
* `dns_enabled` - Whether to expose the instance publicly via DNS and a TCP
  load balancer. Defaults to `false`. Note that the API does not return live
  DNS state, so out-of-band changes are not detected until the next apply that
  explicitly re-sets this field.

### Timeouts

* `create` - (Default `30m`) Time to wait for database creation.
* `update` - (Default `30m`) Time to wait for database updates.
* `delete` - (Default `15m`) Time to wait for database deletion.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The database instance ID.
* `status` - Current status (`pending`, `provisioning`, `running`, `stopped`,
  `error`).
* `cpu_cores` - vCPU count, derived from `resource_profile`.
* `memory_size_mb` - Memory in MB, derived from `resource_profile`.
* `endpoint` - Connection endpoint hostname.
* `port` - Connection port.
* `username` - Admin username.
* `password` - Admin password. Sensitive.
* `connection_info` - Full connection URI. Sensitive.
* `monthly_cost` - Estimated monthly cost in euros.
* `monthly_cost_cents` - Estimated monthly cost in cents.
* `created_at` / `updated_at` / `deployed_at` - Timestamps.

~> **Note** `cpu_cores` and `memory_size_mb` are read-only. They are derived
from `resource_profile` and cannot be set in configuration; doing so fails at
plan time. Resize by changing `resource_profile`.

## Import

Database instances can be imported using their ID:

```bash
terraform import danubedata_database.example 9f8c2d14-3b7a-4e51-9c6d-2a1f8e0b7c33
```

## Notes

- `password` and `connection_info` are stored in state. Protect your state file
  accordingly.
- Storage can only be grown, never shrunk.
- Changing `resource_profile` resizes in place; changing `engine`, `name`,
  `database_name` or `datacenter` replaces the instance.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.
