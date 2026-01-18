terraform {
  required_providers {
    danubedata = {
      source = "registry.terraform.io/AdrianSilaghi/danubedata"
    }
  }
}

provider "danubedata" {
  # API token can be set via DANUBEDATA_API_TOKEN environment variable
}

# List available cache providers
data "danubedata_cache_providers" "available" {}

# Output available providers
output "cache_providers" {
  description = "Available cache providers"
  value = {
    for p in data.danubedata_cache_providers.available.providers :
    p.name => {
      id          = p.id
      type        = p.type
      version     = p.version
      default_port = p.default_port
    }
  }
}

# Create a Redis cache instance
resource "danubedata_cache" "redis" {
  name              = "terraform-redis"
  cache_provider_id = 1 # Redis
  resource_profile  = "small"
  memory_size_mb    = 512
  cpu_cores         = 1
  datacenter        = "fsn1"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

# Create a Valkey cache instance
resource "danubedata_cache" "valkey" {
  name              = "terraform-valkey"
  cache_provider_id = 2 # Valkey
  resource_profile  = "micro"
  memory_size_mb    = 256
  cpu_cores         = 1
  datacenter        = "fsn1"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

# Output cache details
output "redis_endpoint" {
  description = "Redis cache endpoint"
  value       = danubedata_cache.redis.endpoint
}

output "redis_port" {
  description = "Redis cache port"
  value       = danubedata_cache.redis.port
}

output "valkey_endpoint" {
  description = "Valkey cache endpoint"
  value       = danubedata_cache.valkey.endpoint
}
