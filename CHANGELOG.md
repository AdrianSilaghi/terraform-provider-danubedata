# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-04-20

### Added

#### Resources
- `danubedata_parameter_group` - Manage cache, database, and queue parameter groups with custom parameters and locked keys.
- `danubedata_database_replica` - Manage read replicas for database instances (use `count`/`for_each` + `depends_on` to serialize creation).
- `danubedata_cache_snapshot` - Manage snapshots of cache instances.
- `danubedata_database_snapshot` - Manage snapshots of database instances.
- `danubedata_static_site` - Manage static sites (Pages). Build/deploy triggers remain a CLI/CI concern.
- `danubedata_static_site_domain` - Attach custom domains to static sites with automatic verification trigger.

#### Data Sources
- `danubedata_parameter_groups` - List parameter groups, filterable by type and provider.
- `danubedata_cache_snapshots` - List cache snapshots.
- `danubedata_database_snapshots` - List database snapshots.
- `danubedata_static_sites` - List static sites for a team.

#### Attributes
- `danubedata_cache.password` - Sensitive computed attribute exposing the cache password (from `GET /cache/{id}/connection-info`).
- `danubedata_cache.connection_info` - Sensitive computed attribute exposing the cache connection URI.
- `danubedata_cache.dns_enabled` - Optional bool for toggling public DNS via declarative state. Out-of-band DNS changes are not detected until re-applied (API limitation).
- `danubedata_database.password` - Sensitive computed attribute exposing the root password.
- `danubedata_database.connection_info` - Sensitive computed attribute exposing the database connection URI.
- `danubedata_database.dns_enabled` - Optional bool for toggling public DNS (same limitation as cache).
- `danubedata_vps.password` - Now `Optional + Computed`: always populated after provisioning from `GET /vps/{id}/password`.

### Design decisions
- **Not added to the provider** (left to CLI/CI by design):
  - Lifecycle actions: `vps start|stop|reboot|reinstall`, `cache start|stop`, `database start|stop`.
  - Snapshot restore/clone (one-shot recovery operations).
  - Serverless deploy triggers (CI/CD concern).
  - Metrics/status/usage endpoints (read-only telemetry).
  - Access key rotation (handled via `terraform taint`).

## [0.1.0] - 2025-01-19

### Added

- Initial release of the DanubeData Terraform Provider

#### Resources
- `danubedata_vps` - Manage VPS instances with support for:
  - Multiple OS images (Ubuntu, Debian, AlmaLinux, Rocky, Fedora, Alpine)
  - Custom CPU, memory, and storage allocation
  - SSH key and password authentication
  - IPv4/IPv6 dual-stack networking
  - Custom cloud-init configuration

- `danubedata_ssh_key` - Manage SSH keys for VPS authentication

- `danubedata_firewall` - Manage firewalls with:
  - Inbound and outbound rules
  - TCP, UDP, ICMP, and all protocol support
  - Port range specifications
  - Source IP filtering
  - Rule priorities

- `danubedata_cache` - Manage cache instances with support for:
  - Redis, Valkey, Dragonfly providers via `cache_provider` attribute
  - Resource profiles for simplified configuration
  - Custom memory and CPU allocation

- `danubedata_database` - Manage database instances with support for:
  - MySQL, PostgreSQL, MariaDB engines via `engine` attribute
  - Resource profiles for simplified configuration
  - Custom storage, memory, and CPU allocation

- `danubedata_storage_bucket` - Manage S3-compatible storage buckets with:
  - Versioning support
  - Public/private access control
  - Server-side encryption

- `danubedata_storage_access_key` - Manage S3 access keys

- `danubedata_serverless` - Manage serverless containers with:
  - Docker image deployment
  - Git repository deployment
  - Scale-to-zero support
  - Environment variables
  - Auto-scaling configuration

- `danubedata_vps_snapshot` - Manage VPS snapshots for backup and recovery

#### Data Sources
- `danubedata_ssh_keys` - List SSH keys
- `danubedata_vps_images` - List available VPS images
- `danubedata_cache_providers` - List cache providers
- `danubedata_database_providers` - List database providers

#### Features
- Full CRUD operations for all resources
- Import support for all resources
- Configurable timeouts for create, update, and delete operations
- Comprehensive error handling with detailed messages
- Wait for resource readiness on create/update

### Documentation
- Resource and data source documentation
- Example configurations for common use cases
- Complete infrastructure example

[Unreleased]: https://github.com/AdrianSilaghi/terraform-provider-danubedata/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/AdrianSilaghi/terraform-provider-danubedata/releases/tag/v0.1.0
