# danubedata_cache_providers

Lists available cache providers (Redis, Valkey, Dragonfly).

This list is compiled into the provider rather than fetched from the API, so it
does not reflect per-account availability.

## Example Usage

```hcl
data "danubedata_cache_providers" "all" {}

output "cache_providers" {
  value = [for p in data.danubedata_cache_providers.all.providers : {
    type = p.type
    name = p.name
  }]
}
```

### Get Provider by Name

The `danubedata_cache` resource selects its engine with the `type` value, not the
numeric `id`.

```hcl
data "danubedata_cache_providers" "all" {}

locals {
  redis = [for p in data.danubedata_cache_providers.all.providers : p if p.name == "Redis"][0]
}

resource "danubedata_cache" "main" {
  name             = "my-cache"
  cache_provider   = local.redis.type
  resource_profile = "small"
  datacenter       = "fsn1"
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `providers` - List of cache providers. Each provider contains:
  * `id` - The provider ID (a number). Not accepted by any resource argument; use `type` instead.
  * `name` - Provider name (`Redis`, `Valkey`, `Dragonfly`).
  * `type` - Provider type identifier (`redis`, `valkey`, `dragonfly`). This is what the `danubedata_cache` resource's `cache_provider` argument expects.
  * `description` - Provider description.
  * `version` - Default version.
  * `default_port` - Default port number.
