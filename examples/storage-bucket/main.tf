# S3 Storage Bucket Example
# This example creates an S3-compatible storage bucket with an access key

terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.1"
    }
  }
}

provider "danubedata" {}

# Create a storage bucket
resource "danubedata_storage_bucket" "assets" {
  name               = "my-app-assets"
  region             = "fsn1"
  versioning_enabled = true
  public_access      = false
}

# Create an access key for the bucket
resource "danubedata_storage_access_key" "app" {
  name = "app-storage-key"
}

output "bucket_endpoint" {
  description = "S3 endpoint URL"
  value       = danubedata_storage_bucket.assets.endpoint_url
}

output "bucket_name" {
  description = "Bucket name for S3 operations"
  value       = danubedata_storage_bucket.assets.minio_bucket_name
}

output "access_key_id" {
  description = "S3 access key ID"
  value       = danubedata_storage_access_key.app.access_key_id
}

output "secret_access_key" {
  description = "S3 secret access key"
  value       = danubedata_storage_access_key.app.secret_access_key
  sensitive   = true
}

# Example: Configure AWS CLI
# aws configure set aws_access_key_id <access_key_id>
# aws configure set aws_secret_access_key <secret_access_key>
# aws --endpoint-url <bucket_endpoint> s3 ls s3://<bucket_name>/
