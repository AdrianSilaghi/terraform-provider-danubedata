# Terraform Provider for DanubeData

The DanubeData Terraform provider allows you to manage [DanubeData](https://danubedata.com) infrastructure resources using Terraform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- A DanubeData account and API token

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.1"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/AdrianSilaghi/terraform-provider-danubedata.git
cd terraform-provider-danubedata
make install
```

## Authentication

The provider requires an API token for authentication. You can provide it in several ways:

### Environment Variable (Recommended)

```bash
export DANUBEDATA_API_TOKEN="your-api-token"
```

### Provider Configuration

```hcl
provider "danubedata" {
  api_token = "your-api-token"  # Not recommended for production
}
```

## Quick Start

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

# Create an SSH key
resource "danubedata_ssh_key" "example" {
  name       = "my-ssh-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

# Create a VPS instance
resource "danubedata_vps" "example" {
  name        = "my-server"
  image       = "ubuntu-22.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.example.id

  # Optional: specify resources
  cpu_cores      = 2
  memory_size_gb = 4
  storage_size_gb = 50
}

output "server_ip" {
  value = danubedata_vps.example.public_ip
}
```

## Resources

| Resource | Description |
|----------|-------------|
| [danubedata_vps](docs/resources/vps.md) | Manage VPS instances |
| [danubedata_ssh_key](docs/resources/ssh_key.md) | Manage SSH keys |
| [danubedata_firewall](docs/resources/firewall.md) | Manage firewalls with rules |
| [danubedata_cache](docs/resources/cache.md) | Manage Redis/Valkey/Dragonfly cache instances |
| [danubedata_database](docs/resources/database.md) | Manage MySQL/PostgreSQL/MariaDB databases |
| [danubedata_storage_bucket](docs/resources/storage_bucket.md) | Manage S3-compatible storage buckets |
| [danubedata_storage_access_key](docs/resources/storage_access_key.md) | Manage storage access keys |
| [danubedata_serverless](docs/resources/serverless.md) | Manage serverless containers |
| [danubedata_vps_snapshot](docs/resources/vps_snapshot.md) | Manage VPS snapshots |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| [danubedata_ssh_keys](docs/data-sources/ssh_keys.md) | List SSH keys |
| [danubedata_vps_images](docs/data-sources/vps_images.md) | List available VPS images |
| [danubedata_cache_providers](docs/data-sources/cache_providers.md) | List cache providers |
| [danubedata_database_providers](docs/data-sources/database_providers.md) | List database providers |

## Examples

See the [examples](examples/) directory for complete configuration examples:

- [VPS with SSH Key](examples/vps-basic/)
- [VPS with Firewall](examples/vps-firewall/)
- [Redis Cache](examples/cache-redis/)
- [MySQL Database](examples/database-mysql/)
- [S3 Storage Bucket](examples/storage-bucket/)
- [Serverless Container](examples/serverless/)
- [Complete Infrastructure](examples/complete/)

## Development

### Prerequisites

- Go 1.21+
- Terraform 1.0+
- Make

### Building

```bash
# Build the provider
make build

# Install to local Terraform plugins directory
make install

# Run tests
make test

# Run acceptance tests (requires API token)
make testacc
```

### Running Acceptance Tests

Acceptance tests create real resources and may incur charges. Use a test account.

```bash
export DANUBEDATA_API_TOKEN="your-test-token"
make testacc
```

### Generating Documentation

```bash
make docs
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This provider is licensed under the Mozilla Public License 2.0. See [LICENSE](LICENSE) for details.
