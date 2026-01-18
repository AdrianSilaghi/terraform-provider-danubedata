# danubedata_cache_providers

Lists available cache providers (Redis, Valkey, Dragonfly).

## Example Usage

```hcl
data "danubedata_cache_providers" "all" {}

output "cache_providers" {
  value = [for p in data.danubedata_cache_providers.all.providers : {
    id   = p.id
    name = p.name
  }]
}
```

### Get Provider by Name

```hcl
data "danubedata_cache_providers" "all" {}

locals {
  redis = [for p in data.danubedata_cache_providers.all.providers : p if p.name == "Redis"][0]
}

resource "danubedata_cache" "main" {
  name           = "my-cache"
  provider_id    = local.redis.id
  memory_size_mb = 512
  cpu_cores      = 1
  datacenter     = "fsn1"
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `providers` - List of cache providers. Each provider contains:
  * `id` - The provider ID.
  * `name` - Provider name (e.g., "Redis", "Valkey", "Dragonfly").
  * `type` - Provider type identifier.
