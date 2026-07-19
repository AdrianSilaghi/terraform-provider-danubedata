# danubedata_parameter_groups

Lists parameter groups available for cache, database, or queue instances.

Results include both your team's own groups and the built-in system groups.

## Example Usage

```hcl
data "danubedata_parameter_groups" "all" {}

output "parameter_group_names" {
  value = [for g in data.danubedata_parameter_groups.all.groups : g.name]
}
```

### Filter by Type and Provider

Both filters are applied by the API. Combine them to narrow the list to the
groups that are valid for a particular engine.

```hcl
data "danubedata_parameter_groups" "postgres" {
  type          = "database"
  provider_type = "postgresql"
}

output "postgres_group_names" {
  value = [for g in data.danubedata_parameter_groups.postgres.groups : g.name]
}
```

### Attach a Group to a Database

```hcl
data "danubedata_parameter_groups" "postgres" {
  type          = "database"
  provider_type = "postgresql"
}

locals {
  tuned = [for g in data.danubedata_parameter_groups.postgres.groups : g if g.name == "postgres-tuned"][0]
}

resource "danubedata_database" "analytics" {
  name               = "analytics-db"
  engine             = "postgresql"
  resource_profile   = "small"
  datacenter         = "fsn1"
  parameter_group_id = local.tuned.id
}
```

### Find the Default Group

```hcl
data "danubedata_parameter_groups" "cache" {
  type          = "cache"
  provider_type = "redis"
}

locals {
  default_group = [for g in data.danubedata_parameter_groups.cache.groups : g if g.is_default][0]
}

output "default_group_id" {
  value = local.default_group.id
}
```

## Argument Reference

### Optional

* `type` - Filter by parameter group type. One of `cache`, `database`, `queue`.
* `provider_type` - Filter by provider type, e.g. `redis`, `mysql`, `postgresql`.

Omit both to list every group available to your team.

## Attribute Reference

* `groups` - List of parameter groups matching the filters. Each group contains:
  * `id` - The parameter group ID. A numeric ID, exposed as a string; pass it to the `parameter_group_id` argument on the `danubedata_cache` or `danubedata_database` resource.
  * `name` - Name of the parameter group.
  * `type` - Group type (`cache`, `database`, `queue`).
  * `provider_type` - Provider the group applies to, e.g. `redis`, `mysql`, `postgresql`.
  * `family` - Engine family the group targets. Null if not set.
  * `description` - Description of the group. Null if not set.
  * `is_default` - Whether this is the default group for its provider.
  * `is_active` - Whether the group is active.
  * `is_system` - Whether this is a built-in system group rather than one your team created.
  * `created_at` - Timestamp when the group was created. Null if not set.

~> **Note** This data source does not return the parameter values themselves, nor
the list of locked parameters. A group may lock individual parameters so that an
instance cannot override them; to read or manage parameter contents, use the
`danubedata_parameter_group` resource.

## Notes

- System groups (`is_system = true`) are shared across all teams and cannot be modified.
- The provider pages through the full result set, so `groups` contains every matching group, not just the first page.
- A parameter group must match the instance's engine and version to be accepted at apply time.
