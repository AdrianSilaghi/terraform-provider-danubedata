# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-01-17

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
  - Redis
  - Valkey (Redis fork)
  - Dragonfly (high-performance cache)
  - Custom memory and CPU allocation

- `danubedata_database` - Manage database instances with support for:
  - MySQL (8.0-9.1)
  - PostgreSQL (15-17)
  - MariaDB (10.11-11.6)
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
