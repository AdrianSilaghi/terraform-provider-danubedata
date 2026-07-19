# MySQL Database Example
# This example creates a MySQL database instance

terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.3"
    }
  }
}

provider "danubedata" {}

# Create a MySQL database instance
resource "danubedata_database" "main" {
  name             = "app-database"
  database_name    = "myapp"
  engine           = "mysql"
  version          = "8.0"
  resource_profile = "small"
  storage_size_gb  = 20
  datacenter       = "fsn1"

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
