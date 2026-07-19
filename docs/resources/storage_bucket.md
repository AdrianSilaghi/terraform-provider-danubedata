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
  encryption_type    = "sse-s3"
}
```

## Argument Reference

### Required

* `name` - Name of the bucket. Must follow S3 bucket naming rules: 3-63
  characters, lowercase alphanumeric and hyphens, starting and ending with an
  alphanumeric character. Changing this forces a new resource.
* `region` - Region for the bucket. Currently `fsn1` is the only accepted
  value. Changing this forces a new resource.

### Optional

* `display_name` - Human-readable display name, max 255 characters.
* `versioning_enabled` - Enable object versioning. Defaults to `false`.
* `public_access` - Allow public read access. Defaults to `false`.
* `encryption_enabled` - Enable server-side encryption. Defaults to `true`.
* `encryption_type` - Encryption type. One of `none`, `sse-s3`, `sse-kms`.
  Defaults to `sse-s3`.

### Timeouts

* `create` - (Default `10m`) Time to wait for bucket creation.
* `update` - (Default `10m`) Time to wait for bucket updates.
* `delete` - (Default `10m`) Time to wait for bucket deletion.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The bucket ID.
* `status` - Current status (`pending`, `active`, `error`, `destroying`).
* `endpoint_url` - S3 endpoint URL for accessing the bucket.
* `minio_bucket_name` - Internal bucket name, including the team prefix. This
  is the name to pass to S3 clients.
* `size_bytes` - Current size of the bucket in bytes.
* `object_count` - Number of objects in the bucket.
* `monthly_cost` - Estimated monthly cost in euros.
* `monthly_cost_cents` - Estimated monthly cost in cents.
* `created_at` / `updated_at` - Timestamps.

## Import

Storage buckets can be imported using their ID:

```bash
terraform import danubedata_storage_bucket.example 2d9a7f31-6c08-4b5e-8a13-9f4e2c7d0b56
```

## Notes

- `encryption_enabled` defaults to `true`. Set it explicitly to `false` if you
  want an unencrypted bucket.
- Changing `name` or `region` replaces the bucket. `display_name`,
  `versioning_enabled`, `public_access`, `encryption_enabled` and
  `encryption_type` are updated in place.
- Use `minio_bucket_name` — not `name` — when addressing the bucket from an S3
  client, since the platform prefixes bucket names per team.
- For current pricing, including storage and egress allowances, see
  <https://danubedata.ro/pricing>.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.
