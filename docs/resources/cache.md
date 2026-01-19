# danubedata_cache

Manages an in-memory cache instance (Redis, Valkey, or Dragonfly).

## Example Usage

### Redis Cache

```hcl
resource "danubedata_cache" "redis" {
  name           = "my-redis"
  cache_provider = "redis"
  memory_size_mb = 512
  cpu_cores      = 1
  datacenter     = "fsn1"
  version        = "7.2"
}

output "redis_endpoint" {
  value = danubedata_cache.redis.endpoint
}
```

### Valkey Cache (Redis Fork)

```hcl
resource "danubedata_cache" "valkey" {
  name           = "my-valkey"
  cache_provider = "valkey"
  memory_size_mb = 1024
  cpu_cores      = 2
  datacenter     = "fsn1"
}
```

### Dragonfly Cache (High Performance)

```hcl
resource "danubedata_cache" "dragonfly" {
  name           = "my-dragonfly"
  cache_provider = "dragonfly"
  memory_size_mb = 2048
  cpu_cores      = 4
  datacenter     = "fsn1"
}
```

### Using Resource Profile

```hcl
resource "danubedata_cache" "standard" {
  name             = "standard-cache"
  provider         = "redis"
  resource_profile = "cache-medium"
  datacenter       = "fsn1"
}
```

## Argument Reference

### Required

* `name` - (Required) Name of the cache instance.
* `cache_provider` - (Required) Cache provider type. Valid values:
  - `redis` - Redis
  - `valkey` - Valkey (Redis fork)
  - `dragonfly` - Dragonfly (high-performance Redis alternative)
* `datacenter` - (Required) Datacenter location (e.g., `fsn1`).

### Optional

* `memory_size_mb` - (Optional) Memory size in MB.
* `cpu_cores` - (Optional) Number of CPU cores.
* `version` - (Optional) Cache version (e.g., `7.2` for Redis).
* `resource_profile` - (Optional) Predefined resource profile.

### Timeouts

* `create` - (Default `15m`) Time to wait for cache creation.
* `update` - (Default `15m`) Time to wait for cache updates.
* `delete` - (Default `15m`) Time to wait for cache deletion.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The cache instance ID.
* `status` - Current status.
* `endpoint` - Connection endpoint hostname.
* `port` - Connection port.
* `monthly_cost` - Estimated monthly cost.
* `created_at` - Creation timestamp.
* `deployed_at` - Deployment timestamp.

## Import

Cache instances can be imported using their ID:

```bash
terraform import danubedata_cache.example cache-abc123
```
