# danubedata_storage_access_key

Manages an S3-compatible storage access key for bucket authentication.

## Example Usage

### Basic Access Key

```hcl
resource "danubedata_storage_access_key" "main" {
  name = "app-access-key"
}

output "access_key_id" {
  value = danubedata_storage_access_key.main.access_key_id
}

output "secret_access_key" {
  value     = danubedata_storage_access_key.main.secret_access_key
  sensitive = true
}
```

### Key with an Expiry

```hcl
resource "danubedata_storage_access_key" "temporary" {
  name       = "ci-access-key"
  expires_at = "2027-01-01T00:00:00Z"
}
```

### Key Limited to One Bucket

```hcl
resource "danubedata_storage_access_key" "reader" {
  name  = "backup-reader"
  scope = "buckets"
  bucket_permissions = {
    (danubedata_storage_bucket.backups.id) = "read"
  }
}
```

Set `scope = "none"` for a key that receives access only through bucket policies.
The provider checks the created scope and permissions before storing the secret;
if the API returns broader access, it attempts to revoke the new key and fails.

### Using with AWS Provider for S3 Operations

```hcl
resource "danubedata_storage_bucket" "data" {
  name   = "my-data"
  region = "fsn1"
}

resource "danubedata_storage_access_key" "s3" {
  name = "s3-access"
}

provider "aws" {
  alias  = "danubedata"
  region = "us-east-1" # Required but ignored

  access_key = danubedata_storage_access_key.s3.access_key_id
  secret_key = danubedata_storage_access_key.s3.secret_access_key

  endpoints {
    s3 = danubedata_storage_bucket.data.endpoint_url
  }

  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  s3_use_path_style           = true
}

resource "aws_s3_object" "example" {
  provider = aws.danubedata
  bucket   = danubedata_storage_bucket.data.minio_bucket_name
  key      = "example.txt"
  content  = "Hello, World!"
}
```

## Argument Reference

### Required

* `name` - Name of the access key, 1-32 characters.

### Optional

* `expires_at` - Expiration date for the access key, in ISO 8601 format.
  Omit for a key that does not expire.
* `scope` - `team` (default, all team buckets), `buckets`, or `none`. Changing it replaces the key.
* `bucket_permissions` - Map of bucket IDs to `read`, `readwrite`, or `full`.
  Required with `scope = "buckets"` (1-50 entries) and empty with other scopes.
  Changing permissions replaces the key.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The access key resource ID.
* `access_key_id` - The S3 access key ID for authentication.
* `secret_access_key` - The S3 secret access key. Sensitive, and only available
  after creation.
* `status` - Current status of the key (`active`, `revoked`).
* `arn` - S3 principal ARN, when available.
* `created_at` / `updated_at` - Timestamps.

~> **Note** `id` and `access_key_id` are different values. `id` identifies the
resource in the DanubeData API and is what `terraform import` expects;
`access_key_id` is the credential you give to an S3 client.

## Import

Storage access keys can be imported using their ID:

```bash
terraform import danubedata_storage_access_key.example 8e6b0c47-2f19-4d3a-b75c-1a0d9e3f6482
```

**Note:** The `secret_access_key` is only returned during creation and cannot be
retrieved later. Importing an existing key emits a warning and leaves the secret
absent from state; to get a usable secret you must create a replacement key.

## Notes

- `secret_access_key` is stored in state. Protect your state file accordingly.
- This resource does not declare `timeouts`; access keys are created and
  deleted synchronously.
- The API returns the secret only at creation. Refresh preserves that value in
  existing Terraform state; import cannot recover it.
- A `team` key reaches every bucket in the active team. For limited access use
  `scope = "buckets"` and explicit permissions or `scope = "none"` with a bucket policy.
- Pin the project with provider `team_id` or `DANUBEDATA_TEAM_ID`. Otherwise,
  a project-locked token uses its project and an account-wide token uses the
  account's current project. Confirm the selected project before applying.
