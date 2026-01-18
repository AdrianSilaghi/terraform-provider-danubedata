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

# Create a storage bucket
resource "danubedata_storage_bucket" "assets" {
  name               = "terraform-assets"
  display_name       = "Application Assets"
  region             = "fsn1"
  versioning_enabled = true
  public_access      = false
  encryption_enabled = true
  encryption_type    = "sse-s3"

  timeouts {
    create = "5m"
    delete = "5m"
  }
}

# Create a public storage bucket for static files
resource "danubedata_storage_bucket" "static" {
  name               = "terraform-static"
  display_name       = "Static Website Files"
  region             = "fsn1"
  versioning_enabled = false
  public_access      = true
  encryption_enabled = true

  timeouts {
    create = "5m"
    delete = "5m"
  }
}

# Create a storage access key for application use
resource "danubedata_storage_access_key" "app" {
  name = "terraform-app-key"
}

# Create a storage access key with expiration
resource "danubedata_storage_access_key" "temp" {
  name       = "terraform-temp-key"
  expires_at = "2025-12-31T23:59:59Z"
}

# Output bucket details
output "assets_bucket_name" {
  description = "Assets bucket internal name"
  value       = danubedata_storage_bucket.assets.minio_bucket_name
}

output "assets_endpoint" {
  description = "Assets bucket S3 endpoint"
  value       = danubedata_storage_bucket.assets.endpoint_url
}

output "static_bucket_name" {
  description = "Static files bucket internal name"
  value       = danubedata_storage_bucket.static.minio_bucket_name
}

# Output access key credentials
output "app_access_key_id" {
  description = "Application access key ID"
  value       = danubedata_storage_access_key.app.access_key_id
}

output "app_secret_access_key" {
  description = "Application secret access key"
  value       = danubedata_storage_access_key.app.secret_access_key
  sensitive   = true
}

output "temp_access_key_id" {
  description = "Temporary access key ID"
  value       = danubedata_storage_access_key.temp.access_key_id
}
