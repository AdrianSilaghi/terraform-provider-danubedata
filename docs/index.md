# DanubeData Provider

The DanubeData provider enables you to manage infrastructure resources on [DanubeData](https://danubedata.ro) using Terraform. DanubeData offers managed cloud services including VPS instances, databases, caching, object storage, and serverless containers - all hosted in European datacenters.

## Example Usage

```hcl
terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.1"
    }
  }
}

# Configure the provider
provider "danubedata" {
  # API token can be set via DANUBEDATA_API_TOKEN environment variable
}

# Create an SSH key for VPS access
resource "danubedata_ssh_key" "main" {
  name       = "terraform-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

# Create a VPS instance
resource "danubedata_vps" "web" {
  name        = "web-server"
  image       = "ubuntu-24.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.main.id

  cpu_cores       = 2
  memory_size_gb  = 4
  storage_size_gb = 50
}

# Output the server IP
output "server_ip" {
  value = danubedata_vps.web.public_ip
}
```

## Authentication

The DanubeData provider requires an API token for authentication. You can obtain an API token from your [DanubeData account settings](https://danubedata.ro/user/api-tokens).

### Environment Variable (Recommended)

```bash
export DANUBEDATA_API_TOKEN="your-api-token"
```

Then configure the provider without credentials:

```hcl
provider "danubedata" {}
```

### Provider Configuration

You can also configure the token directly in the provider block (not recommended for production):

```hcl
provider "danubedata" {
  api_token = var.danubedata_api_token
}
```

## Schema

### Optional

- `api_token` (String, Sensitive) - API token for DanubeData authentication. Can also be set via `DANUBEDATA_API_TOKEN` environment variable.
- `base_url` (String) - Base URL for the DanubeData API. Defaults to `https://danubedata.ro/api/v1`. Can also be set via `DANUBEDATA_BASE_URL` environment variable.

## Resources

The provider supports the following resources:

### Compute
- [danubedata_vps](resources/vps.md) - Virtual Private Server instances
- [danubedata_serverless](resources/serverless.md) - Serverless containers with scale-to-zero

### Data Services
- [danubedata_database](resources/database.md) - Managed databases (MySQL, PostgreSQL, MariaDB)
- [danubedata_cache](resources/cache.md) - Managed caching (Redis, Valkey, Dragonfly)

### Storage
- [danubedata_storage_bucket](resources/storage_bucket.md) - S3-compatible object storage buckets
- [danubedata_storage_access_key](resources/storage_access_key.md) - Storage access credentials

### Security
- [danubedata_ssh_key](resources/ssh_key.md) - SSH keys for VPS authentication
- [danubedata_firewall](resources/firewall.md) - Network firewall rules

### Backup
- [danubedata_vps_snapshot](resources/vps_snapshot.md) - VPS snapshots for backup and recovery

## Data Sources

### Provider Information
- [danubedata_vps_images](data-sources/vps_images.md) - List available VPS operating system images
- [danubedata_ssh_keys](data-sources/ssh_keys.md) - List SSH keys in your account
- [danubedata_cache_providers](data-sources/cache_providers.md) - List available cache providers
- [danubedata_database_providers](data-sources/database_providers.md) - List available database providers

### Resource Listing
- [danubedata_vpss](data-sources/vpss.md) - List all VPS instances
- [danubedata_databases](data-sources/databases.md) - List all database instances
- [danubedata_caches](data-sources/caches.md) - List all cache instances
- [danubedata_firewalls](data-sources/firewalls.md) - List all firewalls
- [danubedata_serverless_containers](data-sources/serverless_containers.md) - List all serverless containers
- [danubedata_storage_buckets](data-sources/storage_buckets.md) - List all storage buckets
- [danubedata_storage_access_keys](data-sources/storage_access_keys.md) - List all storage access keys
- [danubedata_vps_snapshots](data-sources/vps_snapshots.md) - List all VPS snapshots

## Getting Started

### Prerequisites

1. A DanubeData account - [Sign up](https://danubedata.ro/register)
2. An API token from your [account settings](https://danubedata.ro/user/api-tokens)
3. Terraform 1.0 or later

### Quick Start

1. **Set your API token:**

```bash
export DANUBEDATA_API_TOKEN="your-api-token"
```

2. **Create a Terraform configuration:**

```hcl
terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.1"
    }
  }
}

provider "danubedata" {}

resource "danubedata_ssh_key" "main" {
  name       = "my-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

resource "danubedata_vps" "example" {
  name        = "my-first-vps"
  image       = "ubuntu-24.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.main.id
}
```

3. **Initialize and apply:**

```bash
terraform init
terraform plan
terraform apply
```

## Common Patterns

### Web Application Stack

```hcl
# SSH Key
resource "danubedata_ssh_key" "deploy" {
  name       = "deploy-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

# Web Server VPS
resource "danubedata_vps" "web" {
  name        = "web-server"
  image       = "ubuntu-24.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.deploy.id

  cpu_cores       = 2
  memory_size_gb  = 4
  storage_size_gb = 50
}

# Database
resource "danubedata_database" "db" {
  name            = "app-database"
  engine          = "postgresql"
  version         = "16"
  storage_size_gb = 20
  memory_size_mb  = 2048
  cpu_cores       = 2
  datacenter      = "fsn1"
}

# Cache
resource "danubedata_cache" "cache" {
  name           = "app-cache"
  cache_provider = "redis"
  version        = "7.2"
  memory_size_mb = 512
  cpu_cores      = 1
  datacenter     = "fsn1"
}

# Object Storage for assets
resource "danubedata_storage_bucket" "assets" {
  name               = "app-assets"
  region             = "fsn1"
  versioning_enabled = true
}
```

### Firewall Configuration

```hcl
resource "danubedata_firewall" "web" {
  name           = "web-firewall"
  default_action = "deny"

  rules {
    name             = "SSH"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 22
    port_range_end   = 22
    source_ips       = ["0.0.0.0/0"]
    priority         = 100
  }

  rules {
    name             = "HTTP"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 80
    port_range_end   = 80
    source_ips       = ["0.0.0.0/0"]
    priority         = 200
  }

  rules {
    name             = "HTTPS"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 443
    port_range_end   = 443
    source_ips       = ["0.0.0.0/0"]
    priority         = 300
  }
}
```

## Importing Existing Resources

All resources support importing existing infrastructure into Terraform state:

```bash
# Import a VPS
terraform import danubedata_vps.example vps-abc123

# Import a database
terraform import danubedata_database.example db-abc123

# Import a cache
terraform import danubedata_cache.example cache-abc123

# Import a storage bucket
terraform import danubedata_storage_bucket.example bucket-abc123
```

## Best Practices

### State Management

Use remote state storage for team collaboration:

```hcl
terraform {
  backend "s3" {
    bucket   = "terraform-state"
    key      = "danubedata/terraform.tfstate"
    region   = "us-east-1"
    endpoint = "https://s3.danubedata.ro"

    skip_credentials_validation = true
    skip_metadata_api_check     = true
    skip_requesting_account_id  = true
    force_path_style            = true
  }
}
```

### Sensitive Values

Use environment variables or Terraform variables for sensitive values:

```hcl
variable "db_password" {
  type      = string
  sensitive = true
}
```

### Resource Profiles

Use resource profiles for consistent sizing:

```hcl
resource "danubedata_vps" "web" {
  name             = "web-server"
  image            = "ubuntu-24.04"
  datacenter       = "fsn1"
  resource_profile = "vps-medium"  # Predefined CPU/RAM/Storage
  auth_method      = "ssh_key"
  ssh_key_id       = danubedata_ssh_key.main.id
}
```

## Support

- [Documentation](https://danubedata.ro/docs)
- [API Reference](https://danubedata.ro/docs?page=api-overview)
- [Contact Support](https://danubedata.ro/contact)
- [GitHub Issues](https://github.com/AdrianSilaghi/terraform-provider-danubedata/issues)
