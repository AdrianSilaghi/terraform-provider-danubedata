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
  region = "us-east-1"  # Required but ignored

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

* `name` - (Required) Name of the access key.

### Optional

* `expires_at` - (Optional) Expiration timestamp in RFC3339 format.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The access key ID (internal).
* `access_key_id` - The S3 access key ID for authentication.
* `secret_access_key` - The S3 secret access key for authentication. Only available after creation.
* `status` - Current status of the key.
* `created_at` - Creation timestamp.

## Import

Storage access keys can be imported using their ID:

```bash
terraform import danubedata_storage_access_key.example key-abc123
```

**Note:** The `secret_access_key` is only returned during creation and cannot be retrieved later. If you import an existing key, you will need to rotate it to get a new secret.
