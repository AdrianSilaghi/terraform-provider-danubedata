# danubedata_storage_access_keys

Lists all S3 storage access keys in your account.

## Example Usage

```hcl
data "danubedata_storage_access_keys" "all" {}

output "key_count" {
  value = length(data.danubedata_storage_access_keys.all.keys)
}

output "active_keys" {
  value = [for k in data.danubedata_storage_access_keys.all.keys : k.name if k.status == "active"]
}
```

### Find Key by Name

```hcl
data "danubedata_storage_access_keys" "all" {}

locals {
  app_key = [for k in data.danubedata_storage_access_keys.all.keys : k if k.name == "app-access-key"][0]
}

output "app_access_key_id" {
  value = local.app_key.access_key_id
}
```

### Filter Active Keys

```hcl
data "danubedata_storage_access_keys" "all" {}

locals {
  active_keys  = [for k in data.danubedata_storage_access_keys.all.keys : k if k.status == "active" && !k.is_expired]
  expired_keys = [for k in data.danubedata_storage_access_keys.all.keys : k if k.is_expired]
}

output "active_count" {
  value = length(local.active_keys)
}

output "expired_count" {
  value = length(local.expired_keys)
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `keys` - List of storage access keys. Each key contains:
  * `id` - Unique identifier for the access key.
  * `name` - Name of the access key.
  * `access_key_id` - The S3 access key ID for authentication.
  * `status` - Current status (active, revoked).
  * `access_type` - Access type (full or restricted).
  * `expires_at` - Expiration timestamp (if set).
  * `last_used_at` - Timestamp when the key was last used.
  * `is_expired` - Whether the key has expired.
  * `created_at` - Timestamp when the key was created.

## Notes

- The `secret_access_key` is not included in this data source for security reasons.
- Secret access keys are only available during creation.
- To get a new secret key, create a new access key resource.
