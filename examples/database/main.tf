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

# List available database providers
data "danubedata_database_providers" "available" {}

# Output available providers
output "database_providers" {
  description = "Available database providers"
  value = {
    for p in data.danubedata_database_providers.available.providers :
    p.name => {
      id           = p.id
      type         = p.type
      version      = p.version
      default_port = p.default_port
    }
  }
}

# Create a PostgreSQL database instance
resource "danubedata_database" "postgres" {
  name                 = "terraform-postgres"
  database_provider_id = 2 # PostgreSQL
  database_name        = "myapp"
  resource_profile     = "small"
  storage_size_gb      = 20
  memory_size_mb       = 2048
  cpu_cores            = 2
  datacenter           = "fsn1"

  timeouts {
    create = "20m"
    delete = "15m"
  }
}

# Create a MySQL database instance
resource "danubedata_database" "mysql" {
  name                 = "terraform-mysql"
  database_provider_id = 1 # MySQL
  database_name        = "production"
  resource_profile     = "small"
  storage_size_gb      = 50
  memory_size_mb       = 4096
  cpu_cores            = 2
  datacenter           = "fsn1"

  timeouts {
    create = "20m"
    delete = "15m"
  }
}

# Output database details
output "postgres_endpoint" {
  description = "PostgreSQL database endpoint"
  value       = danubedata_database.postgres.endpoint
}

output "postgres_port" {
  description = "PostgreSQL database port"
  value       = danubedata_database.postgres.port
}

output "postgres_username" {
  description = "PostgreSQL database username"
  value       = danubedata_database.postgres.username
}

output "mysql_endpoint" {
  description = "MySQL database endpoint"
  value       = danubedata_database.mysql.endpoint
}

output "mysql_monthly_cost" {
  description = "MySQL database monthly cost in dollars"
  value       = danubedata_database.mysql.monthly_cost
}
