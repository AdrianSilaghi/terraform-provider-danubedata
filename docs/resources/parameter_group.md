# danubedata_parameter_group

Manages a parameter group — a reusable, named set of engine configuration
parameters that cache, database and queue instances can attach to.

## Example Usage

### Redis Cache Parameter Group

```hcl
resource "danubedata_parameter_group" "redis_tuned" {
  name          = "redis-high-connections"
  type          = "cache"
  provider_type = "redis"
  family        = "redis7.x"
  description   = "Raised connection limit with LRU eviction"

  parameters = {
    maxclients       = "10000"
    maxmemory-policy = "allkeys-lru"
    timeout          = "300"
  }
}

resource "danubedata_cache" "sessions" {
  name               = "sessions"
  cache_provider     = "redis"
  resource_profile   = "small"
  datacenter         = "fsn1"
  parameter_group_id = danubedata_parameter_group.redis_tuned.id
}
```

### PostgreSQL Group with Locked Parameters

Keys listed in `locked_parameters` cannot be overridden by the individual
instances that attach to the group — useful for settings you want to hold
constant across a fleet.

```hcl
resource "danubedata_parameter_group" "postgres_baseline" {
  name          = "postgres-baseline"
  type          = "database"
  provider_type = "postgresql"
  description   = "Company baseline for PostgreSQL instances"

  parameters = {
    max_connections            = "200"
    work_mem                   = "8MB"
    log_min_duration_statement = "500"
  }

  locked_parameters = [
    "max_connections",
  ]
}

resource "danubedata_database" "analytics" {
  name               = "analytics-db"
  engine             = "postgresql"
  resource_profile   = "medium"
  datacenter         = "fsn1"
  parameter_group_id = danubedata_parameter_group.postgres_baseline.id
}
```

### Default Group for a Provider

```hcl
resource "danubedata_parameter_group" "valkey_default" {
  name          = "valkey-team-default"
  type          = "cache"
  provider_type = "valkey"
  is_default    = true

  parameters = {
    maxmemory-policy = "volatile-lru"
  }
}
```

## Argument Reference

### Required

* `name` - Name of the parameter group, max 255 characters. Updated in place.
* `type` - Parameter group type. One of `cache`, `database`, `queue`. Changing
  this forces a new resource.
* `provider_type` - The engine the group targets, e.g. `redis`, `valkey`,
  `dragonfly`, `mysql`, `postgresql`, `mariadb`. Free-form string, max 50
  characters — the API does not restrict it to a fixed list, so a typo is
  accepted at apply time and simply never matches an instance. Changing this
  forces a new resource.
* `parameters` - Key/value map of engine parameters. **Values must be strings**;
  express numbers and booleans as strings (`"10000"`, `"true"`). Keys are
  passed through to the engine as-is, so consult the engine's own configuration
  reference for valid names.

### Optional

* `description` - Free-text description of the group.
* `family` - Family label, e.g. `redis7.x`. See [Notes](#notes) — this is only
  sent when the group is created.
* `locked_parameters` - List of parameter keys that instances using this group
  cannot override.
* `is_default` - Whether this group is the default for its provider type.
  Defaults to `false`.
* `is_active` - Whether the group is active. Defaults to `true`. See
  [Notes](#notes) — this is only honoured on update, not at creation.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The parameter group ID.
* `is_system` - Whether this is a system-managed group. System groups cannot be
  modified or deleted.
* `created_at` / `updated_at` - Timestamps.

## Import

Parameter groups can be imported using their ID, which is a numeric identifier:

```bash
terraform import danubedata_parameter_group.example 42
```

## Notes

- `is_active` is not part of the create payload — the API forces new groups to
  active. Setting `is_active = false` in a configuration that has never been
  applied does not take effect at creation; it is only applied by a subsequent
  update.
- `family` is sent only when the group is created. The update request does not
  carry it and the attribute does not force replacement, so editing `family`
  on an existing group has no effect on the API.
- System groups (`is_system = true`) are rejected by the API for both update and
  delete. Do not import one into Terraform — every apply that touches it will
  fail. Clone it in the dashboard and manage the copy instead.
- The API refuses to delete a group that is still referenced by any cache or
  database instance. When instances point at the group via
  `parameter_group_id`, Terraform already orders the destroy correctly;
  detaching a group by hand elsewhere can leave a destroy blocked.
- `type` and `provider_type` force replacement; `name`, `description`,
  `parameters`, `locked_parameters`, `is_default` and `is_active` are updated in
  place.
- Attaching a group to an instance is done from the instance side, through
  `parameter_group_id` on `danubedata_cache` or `danubedata_database`.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.
