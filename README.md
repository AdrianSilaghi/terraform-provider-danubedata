# Terraform Provider for DanubeData

The DanubeData Terraform provider allows you to manage [DanubeData](https://danubedata.ro) infrastructure resources using Terraform.

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
      version = "~> 0.3"
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
      version = "~> 0.3"
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

  # Optional: pick a plan. CPU, memory and storage follow from it and are
  # read-only; setting them directly fails at plan time.
  resource_profile = "micro_shared"
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
| [danubedata_database_replica](docs/resources/database_replica.md) | Manage database read replicas |
| [danubedata_parameter_group](docs/resources/parameter_group.md) | Manage engine parameter groups |
| [danubedata_storage_bucket](docs/resources/storage_bucket.md) | Manage S3-compatible storage buckets |
| [danubedata_storage_access_key](docs/resources/storage_access_key.md) | Manage storage access keys |
| [danubedata_serverless](docs/resources/serverless.md) | Manage serverless containers |
| [danubedata_static_site](docs/resources/static_site.md) | Manage static sites |
| [danubedata_static_site_domain](docs/resources/static_site_domain.md) | Manage static site custom domains |
| [danubedata_vps_snapshot](docs/resources/vps_snapshot.md) | Manage VPS snapshots |
| [danubedata_database_snapshot](docs/resources/database_snapshot.md) | Manage database snapshots |
| [danubedata_cache_snapshot](docs/resources/cache_snapshot.md) | Manage cache snapshots |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| [danubedata_ssh_keys](docs/data-sources/ssh_keys.md) | List SSH keys |
| [danubedata_vps_images](docs/data-sources/vps_images.md) | List available VPS images |
| [danubedata_cache_providers](docs/data-sources/cache_providers.md) | List cache providers |
| [danubedata_database_providers](docs/data-sources/database_providers.md) | List database providers |
| [danubedata_parameter_groups](docs/data-sources/parameter_groups.md) | List parameter groups |
| [danubedata_vpss](docs/data-sources/vpss.md) | List VPS instances |
| [danubedata_databases](docs/data-sources/databases.md) | List database instances |
| [danubedata_caches](docs/data-sources/caches.md) | List cache instances |
| [danubedata_firewalls](docs/data-sources/firewalls.md) | List firewalls |
| [danubedata_serverless_containers](docs/data-sources/serverless_containers.md) | List serverless containers |
| [danubedata_static_sites](docs/data-sources/static_sites.md) | List static sites |
| [danubedata_storage_buckets](docs/data-sources/storage_buckets.md) | List storage buckets |
| [danubedata_storage_access_keys](docs/data-sources/storage_access_keys.md) | List storage access keys |
| [danubedata_vps_snapshots](docs/data-sources/vps_snapshots.md) | List VPS snapshots |
| [danubedata_cache_snapshots](docs/data-sources/cache_snapshots.md) | List cache snapshots |
| [danubedata_database_snapshots](docs/data-sources/database_snapshots.md) | List database snapshots |

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

### Documentation

The pages under `docs/` are maintained by hand — there are no `tfplugindocs`
templates and no `//go:generate` directives behind them. Edit the relevant
Markdown file directly and keep it in step with the provider schema; running a
generator over `docs/` would discard the hand-written content.

To check a schema claim against the current build:

```bash
go build -o /tmp/terraform-provider-danubedata .
terraform providers schema -json | jq .
```

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This provider is licensed under the Mozilla Public License 2.0. See [LICENSE](LICENSE) for details.
