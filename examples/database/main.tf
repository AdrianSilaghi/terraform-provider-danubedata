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

# Create a PostgreSQL database instance
resource "danubedata_database" "postgres" {
  name             = "terraform-postgres"
  engine           = "postgresql"
  database_name    = "myapp"
  resource_profile = "small"
  storage_size_gb  = 20
  memory_size_mb   = 2048
  cpu_cores        = 2
  datacenter       = "fsn1"

  timeouts {
    create = "20m"
    delete = "15m"
  }
}

# Create a MySQL database instance
resource "danubedata_database" "mysql" {
  name             = "terraform-mysql"
  engine           = "mysql"
  database_name    = "production"
  resource_profile = "small"
  storage_size_gb  = 50
  memory_size_mb   = 4096
  cpu_cores        = 2
  datacenter       = "fsn1"

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
