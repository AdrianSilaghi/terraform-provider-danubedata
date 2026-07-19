# danubedata_cache

Manages an in-memory cache instance (Redis, Valkey, or Dragonfly).

## Example Usage

### Redis Cache

```hcl
resource "danubedata_cache" "redis" {
  name             = "my-redis"
  cache_provider   = "redis"
  resource_profile = "small"
  datacenter       = "fsn1"
  version          = "7.4"
}

output "redis_endpoint" {
  value = danubedata_cache.redis.endpoint
}
```

### Valkey Cache (Redis Fork)

```hcl
resource "danubedata_cache" "valkey" {
  name             = "my-valkey"
  cache_provider   = "valkey"
  resource_profile = "medium"
  datacenter       = "fsn1"
}
```

### Dragonfly Cache (High Performance)

```hcl
resource "danubedata_cache" "dragonfly" {
  name             = "my-dragonfly"
  cache_provider   = "dragonfly"
  resource_profile = "large"
  datacenter       = "fsn1"
}
```

### Publicly Reachable Cache

```hcl
resource "danubedata_cache" "public" {
  name             = "edge-cache"
  cache_provider   = "valkey"
  resource_profile = "micro"
  datacenter       = "fsn1"
  dns_enabled      = true
}
```

## Resource Profiles

`resource_profile` selects the plan, and it is the only place CPU and memory are
set. Use the **slug** in the left column — the dashboard and pricing page show
the display name, which is not a valid value here.

The same four slugs are available for every provider. Memory and storage are
identical across providers; only Dragonfly scales vCPU with the plan.

| Slug     | Display name | RAM     | Storage | vCPU (Redis / Valkey) | vCPU (Dragonfly) |
| -------- | ------------ | ------- | ------- | --------------------- | ---------------- |
| `micro`  | DD Puiu      | 0.25 GB | 2 GB    | 1                     | 1                |
| `small`  | DD Rosu      | 1 GB    | 4 GB    | 1                     | 2                |
| `medium` | DD Dranov    | 3 GB    | 16 GB   | 1                     | 4                |
| `large`  | DD Razim     | 6 GB    | 32 GB   | 1                     | 8                |

For current pricing see <https://danubedata.ro/pricing>. Profiles above your
account's limit are rejected at apply time; request an increase from Account
Limits in the dashboard.

## Argument Reference

### Required

* `name` - Name of the cache instance. Lowercase alphanumeric and hyphens only
  (DNS compatible), 2-63 characters.
* `cache_provider` - Cache provider type. One of:
  - `redis` - Redis
  - `valkey` - Valkey (Redis fork)
  - `dragonfly` - Dragonfly (high-performance Redis alternative)

  Changing this forces a new resource.
* `resource_profile` - Plan slug; see [Resource Profiles](#resource-profiles).
* `datacenter` - Datacenter location. One of `fsn1`, `nbg1`, `hel1`, `ash`.
  Changing this forces a new resource.

### Optional

* `version` - Version of the cache software, e.g. `7.4` for Redis. Defaults to
  the current default version for the provider. Versions are validated by the
  API; as of this release the offered versions are:

  | Provider    | Versions                    |
  | ----------- | --------------------------- |
  | `redis`     | `8.4`, `8.0`, `7.4`, `7.2`  |
  | `valkey`    | `9.1`, `9.0`, `8.1`, `8.0`, `7.2` |
  | `dragonfly` | `1.24`, `1.23`              |

  Only applied at creation — see [Notes](#notes).
* `parameter_group_id` - ID of a parameter group for custom configuration.
* `dns_enabled` - Whether to expose the instance publicly via DNS and a TCP
  load balancer. Defaults to `false`. Note that the API does not return live
  DNS state, so out-of-band changes are not detected until the next apply that
  explicitly re-sets this field.

### Timeouts

* `create` - (Default `30m`) Time to wait for cache creation.
* `update` - (Default `30m`) Time to wait for cache updates.
* `delete` - (Default `15m`) Time to wait for cache deletion.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The cache instance ID.
* `status` - Current status (`pending`, `provisioning`, `running`, `stopped`,
  `error`).
* `cpu_cores` - vCPU count, derived from `resource_profile`.
* `memory_size_mb` - Memory in MB, derived from `resource_profile`.
* `endpoint` - Connection endpoint hostname.
* `port` - Connection port.
* `password` - Cache password. Sensitive.
* `connection_info` - Full connection URI, e.g. `redis://host:6379`. Sensitive.
* `monthly_cost` - Estimated monthly cost in euros.
* `monthly_cost_cents` - Estimated monthly cost in cents.
* `created_at` / `updated_at` / `deployed_at` - Timestamps.

~> **Note** `cpu_cores` and `memory_size_mb` are read-only. They are derived
from `resource_profile` and cannot be set in configuration; doing so fails at
plan time. Resize by changing `resource_profile`.

## Import

Cache instances can be imported using their ID:

```bash
terraform import danubedata_cache.example 4c1e9b70-8a52-4d6f-b3c1-6e2f9d0a5b48
```

## Notes

- `password` and `connection_info` are stored in state. Protect your state file
  accordingly.
- Changing `resource_profile` resizes in place; changing `cache_provider` or
  `datacenter` replaces the instance. `name` is updated in place.
- `version` is sent only when the instance is created. The update request
  carries `name`, `resource_profile` and `parameter_group_id` only, so editing
  `version` afterwards does not upgrade a running instance.
- Updates are applied asynchronously: the API returns before the change is
  deployed, so the instance briefly reports `pending` and returns to `running`
  once reconciled.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.
