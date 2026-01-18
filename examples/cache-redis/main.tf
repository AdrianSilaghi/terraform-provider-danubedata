# Redis Cache Example
# This example creates a Redis cache instance

terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.1"
    }
  }
}

provider "danubedata" {}

# Look up the Redis provider
data "danubedata_cache_providers" "all" {}

locals {
  redis_provider = [for p in data.danubedata_cache_providers.all.providers : p if p.name == "Redis"][0]
}

# Create a Redis cache instance
resource "danubedata_cache" "main" {
  name           = "app-cache"
  provider_id    = local.redis_provider.id
  memory_size_mb = 512
  cpu_cores      = 1
  datacenter     = "fsn1"
  version        = "7.2"

  timeouts {
    create = "10m"
    delete = "10m"
  }
}

output "cache_endpoint" {
  description = "Redis connection endpoint"
  value       = danubedata_cache.main.endpoint
}

output "cache_port" {
  description = "Redis connection port"
  value       = danubedata_cache.main.port
}

output "redis_url" {
  description = "Full Redis connection URL"
  value       = "redis://${danubedata_cache.main.endpoint}:${danubedata_cache.main.port}"
}
