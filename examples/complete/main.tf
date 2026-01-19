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

# Variables
variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "root_password" {
  description = "Root password for the VPS"
  type        = string
  sensitive   = true
}

# VPS Instance
resource "danubedata_vps" "web" {
  name             = "${var.environment}-web-server"
  resource_profile = "small_shared"
  image            = "ubuntu-24.04"
  datacenter       = "fsn1"
  network_stack    = "dual_stack"
  auth_method      = "password"
  password         = var.root_password

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

# Redis Cache for sessions
resource "danubedata_cache" "sessions" {
  name             = "${var.environment}-sessions"
  cache_provider   = "redis"
  resource_profile = "small"
  memory_size_mb   = 512
  cpu_cores        = 1
  datacenter       = "fsn1"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

# PostgreSQL Database
resource "danubedata_database" "main" {
  name             = "${var.environment}-database"
  engine           = "postgresql"
  database_name    = "app"
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

# Storage Bucket for uploads
resource "danubedata_storage_bucket" "uploads" {
  name               = "${var.environment}-uploads"
  display_name       = "User Uploads"
  region             = "fsn1"
  versioning_enabled = true
  public_access      = false
  encryption_enabled = true

  timeouts {
    create = "5m"
    delete = "5m"
  }
}

# Storage Access Key for the application
resource "danubedata_storage_access_key" "app" {
  name = "${var.environment}-app-storage-key"
}

# Outputs
output "vps_public_ip" {
  description = "VPS public IP address"
  value       = danubedata_vps.web.public_ip
}

output "vps_ipv6_address" {
  description = "VPS IPv6 address"
  value       = danubedata_vps.web.ipv6_address
}

output "cache_endpoint" {
  description = "Redis cache endpoint"
  value       = danubedata_cache.sessions.endpoint
}

output "database_endpoint" {
  description = "PostgreSQL database endpoint"
  value       = danubedata_database.main.endpoint
}

output "database_port" {
  description = "PostgreSQL database port"
  value       = danubedata_database.main.port
}

output "storage_endpoint" {
  description = "S3 storage endpoint"
  value       = danubedata_storage_bucket.uploads.endpoint_url
}

output "storage_bucket_name" {
  description = "S3 bucket name"
  value       = danubedata_storage_bucket.uploads.minio_bucket_name
}

output "storage_access_key_id" {
  description = "S3 access key ID"
  value       = danubedata_storage_access_key.app.access_key_id
}

output "storage_secret_access_key" {
  description = "S3 secret access key"
  value       = danubedata_storage_access_key.app.secret_access_key
  sensitive   = true
}

output "monthly_costs" {
  description = "Monthly costs breakdown"
  value = {
    vps      = danubedata_vps.web.monthly_cost
    cache    = danubedata_cache.sessions.monthly_cost
    database = danubedata_database.main.monthly_cost
    storage  = danubedata_storage_bucket.uploads.monthly_cost
  }
}
