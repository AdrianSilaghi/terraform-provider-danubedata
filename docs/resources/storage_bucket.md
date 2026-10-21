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

### Object Lock and Quota

```hcl
resource "danubedata_storage_bucket" "records" {
  name                               = "records"
  region                             = "fsn1"
  versioning_enabled                 = true
  object_lock_enabled                = true
  object_lock_default_retention_mode = "GOVERNANCE"
  object_lock_default_retention_days = 30
  size_limit_bytes                   = 1073741824
  tags                               = { environment = "production" }
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
  The API can change it, but cannot clear it after creation.
* `versioning_enabled` - Enable object versioning. Defaults to `false`.
* `public_access` - Allow public read access. Defaults to `false`.
* `encryption_enabled` - Enable server-side encryption. Defaults to `true`.
* `encryption_type` - Encryption type. One of `none`, `sse-s3`, `sse-kms`.
  The API defaults to `sse-s3` when encryption is enabled. Set `none` explicitly
  with `encryption_enabled = false`; when changing encryption, set both fields.
* `encryption_kms_key_id` - KMS key ID required for `sse-kms`. Sensitive. The
  API does not return it on read, so imported buckets cannot recover this value.
* `object_lock_enabled` - Enable Object Lock at creation. Requires
  `versioning_enabled = true`; changing it replaces the bucket.
* `object_lock_default_retention_mode` and
  `object_lock_default_retention_days` - Set together for an Object Lock bucket.
  Mode is `GOVERNANCE` or `COMPLIANCE`; days must be 1-36500.
* `size_limit_bytes` - Bucket quota in bytes, at least 1 GiB. The API applies
  it after creation when the bucket becomes active. Omit to remove a quota.
* `tags` - Map of up to 50 bucket tags.
* `cors_configuration` - S3 CORS configuration as JSON, for example
  `jsonencode({ CORSRules = [{ AllowedOrigins = ["https://example.com"], AllowedMethods = ["GET"] }] })`.
* `lifecycle_rules` - S3 lifecycle rules as a JSON array, supplied with
  `jsonencode(...)`. Omit it to clear rules.

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
* `monthly_cost` - Estimated monthly cost in the API billing currency.
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
  `encryption_type`, retention, quota, tags, CORS and lifecycle rules are updated in place.
- Object Lock cannot be disabled after creation. Removing configured default
  retention changes only the default for new objects, not existing locked objects.
- The platform adds a CORS rule for its console. The provider leaves that rule
  out of `cors_configuration` state and manages the user rules only.
- The provider does not manage custom bucket policy statements; the API exposes
  those through a separate policy endpoint.
- Use `minio_bucket_name` — not `name` — when addressing the bucket from an S3
  client, since the platform prefixes bucket names per team.
- For current pricing, including storage and egress allowances, see
  <https://danubedata.ro/pricing>.
- Pin the project with provider `team_id` or `DANUBEDATA_TEAM_ID`. Otherwise,
  a project-locked token uses its project and an account-wide token uses the
  account's current project. Confirm the selected project before applying.
