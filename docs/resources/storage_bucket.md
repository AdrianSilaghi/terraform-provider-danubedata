# danubedata_storage_bucket

Manages an S3-compatible object storage bucket.

## Example Usage

### Basic Bucket

```hcl
resource "danubedata_storage_bucket" "assets" {
  name   = "my-assets"
  region = "fsn1"
}

output "bucket_endpoint" {
  value = danubedata_storage_bucket.assets.endpoint_url
}
```

### Bucket with Versioning

```hcl
resource "danubedata_storage_bucket" "backups" {
  name               = "my-backups"
  region             = "fsn1"
  versioning_enabled = true
}
```

### Public Bucket

```hcl
resource "danubedata_storage_bucket" "public" {
  name          = "public-assets"
  region        = "fsn1"
  public_access = true
}
```

### Complete Configuration

```hcl
resource "danubedata_storage_bucket" "data" {
  name               = "app-data"
  display_name       = "Application Data"
  region             = "fsn1"
  versioning_enabled = true
  public_access      = false
  encryption_enabled = true
}
```

## Argument Reference

### Required

* `name` - (Required) Name of the bucket. Must be unique and follow S3 naming conventions.
* `region` - (Required) Region for the bucket (e.g., `fsn1`).

### Optional

* `display_name` - (Optional) Human-readable display name.
* `versioning_enabled` - (Optional) Enable object versioning. Default: `false`.
* `public_access` - (Optional) Allow public access. Default: `false`.
* `encryption_enabled` - (Optional) Enable server-side encryption. Default: `false`.
* `encryption_type` - (Optional) Encryption type when encryption is enabled.

### Timeouts

* `create` - (Default `5m`) Time to wait for bucket creation.
* `update` - (Default `5m`) Time to wait for bucket updates.
* `delete` - (Default `5m`) Time to wait for bucket deletion.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The bucket ID.
* `status` - Current status.
* `endpoint_url` - S3-compatible endpoint URL.
* `public_url` - Public URL (if public access enabled).
* `minio_bucket_name` - Internal bucket name.
* `size_bytes` - Current size in bytes.
* `object_count` - Number of objects.
* `monthly_cost` - Estimated monthly cost.
* `created_at` - Creation timestamp.

## Import

Storage buckets can be imported using their ID:

```bash
terraform import danubedata_storage_bucket.example bucket-abc123
```

## Pricing

- Base: EUR 3.99/month
- Includes: 1TB storage + 1TB egress traffic
- Overage: EUR 0.01/GB for storage, EUR 0.01/GB for egress
