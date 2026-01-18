# MySQL Database Example
# This example creates a MySQL database instance

terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.1"
    }
  }
}

provider "danubedata" {}

# Look up the MySQL provider
data "danubedata_database_providers" "all" {}

locals {
  mysql_provider = [for p in data.danubedata_database_providers.all.providers : p if p.name == "MySQL"][0]
}

# Create a MySQL database instance
resource "danubedata_database" "main" {
  name            = "app-database"
  database_name   = "myapp"
  provider_id     = local.mysql_provider.id
  version         = "8.0"
  storage_size_gb = 20
  memory_size_mb  = 2048
  cpu_cores       = 2
  datacenter      = "fsn1"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

output "database_endpoint" {
  description = "MySQL connection endpoint"
  value       = danubedata_database.main.endpoint
}

output "database_port" {
  description = "MySQL connection port"
  value       = danubedata_database.main.port
}

output "database_name" {
  description = "Database name"
  value       = "myapp"
}

output "connection_string" {
  description = "MySQL connection string (password not included)"
  value       = "mysql://${danubedata_database.main.username}@${danubedata_database.main.endpoint}:${danubedata_database.main.port}/myapp"
}
