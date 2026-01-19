# danubedata_caches

Lists all cache instances in your account.

## Example Usage

```hcl
data "danubedata_caches" "all" {}

output "cache_count" {
  value = length(data.danubedata_caches.all.instances)
}

output "cache_endpoints" {
  value = {
    for cache in data.danubedata_caches.all.instances : cache.name => "${cache.endpoint}:${cache.port}"
  }
}
```

### Find Cache by Name

```hcl
data "danubedata_caches" "all" {}

locals {
  main_cache = [for c in data.danubedata_caches.all.instances : c if c.name == "main-cache"][0]
}

output "redis_url" {
  value = "redis://${local.main_cache.endpoint}:${local.main_cache.port}"
}
```

### Filter by Provider

```hcl
data "danubedata_caches" "all" {}

locals {
  redis_caches     = [for c in data.danubedata_caches.all.instances : c if c.cache_provider == "Redis"]
  dragonfly_caches = [for c in data.danubedata_caches.all.instances : c if c.cache_provider == "Dragonfly"]
}

output "redis_count" {
  value = length(local.redis_caches)
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `instances` - List of cache instances. Each instance contains:
  * `id` - Unique identifier for the cache instance.
  * `name` - Name of the cache instance.
  * `status` - Current status (creating, running, stopped, error).
  * `cache_provider` - Cache provider (Redis, Valkey, Dragonfly).
  * `version` - Cache version.
  * `datacenter` - Datacenter location.
  * `cpu_cores` - Number of CPU cores.
  * `memory_size_mb` - Memory size in MB.
  * `endpoint` - Connection endpoint hostname.
  * `port` - Connection port.
  * `monthly_cost` - Estimated monthly cost.
  * `created_at` - Timestamp when the instance was created.
