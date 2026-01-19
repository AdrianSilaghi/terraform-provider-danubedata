# danubedata_storage_buckets

Lists all S3-compatible storage buckets in your account.

## Example Usage

```hcl
data "danubedata_storage_buckets" "all" {}

output "bucket_count" {
  value = length(data.danubedata_storage_buckets.all.buckets)
}

output "total_storage_bytes" {
  value = sum([for b in data.danubedata_storage_buckets.all.buckets : b.size_bytes])
}
```

### Find Bucket by Name

```hcl
data "danubedata_storage_buckets" "all" {}

locals {
  assets_bucket = [for b in data.danubedata_storage_buckets.all.buckets : b if b.name == "assets"][0]
}

output "assets_endpoint" {
  value = local.assets_bucket.endpoint_url
}

output "assets_bucket_name" {
  value = local.assets_bucket.minio_bucket_name
}
```

### Filter Public Buckets

```hcl
data "danubedata_storage_buckets" "all" {}

locals {
  public_buckets = [for b in data.danubedata_storage_buckets.all.buckets : b if b.public_access]
}

output "public_bucket_urls" {
  value = [for b in local.public_buckets : b.public_url]
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `buckets` - List of storage buckets. Each bucket contains:
  * `id` - Unique identifier for the bucket.
  * `name` - Name of the bucket.
  * `display_name` - Human-readable display name.
  * `status` - Current status.
  * `region` - Region where the bucket is located.
  * `endpoint_url` - S3-compatible endpoint URL.
  * `public_url` - Public URL (if public access enabled).
  * `minio_bucket_name` - Internal bucket name for S3 operations.
  * `public_access` - Whether public access is enabled.
  * `versioning_enabled` - Whether versioning is enabled.
  * `encryption_enabled` - Whether encryption is enabled.
  * `size_bytes` - Current size in bytes.
  * `object_count` - Number of objects in the bucket.
  * `monthly_cost` - Estimated monthly cost.
  * `created_at` - Timestamp when the bucket was created.
