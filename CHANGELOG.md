# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- **`danubedata_database` examples were not applyable.** The 0.3.0 schema realignment made `cpu_cores`/`memory_size_mb` read-only and `resource_profile` required, but only `examples/complete` was updated to match. `examples/database` failed with "Invalid Configuration for Read-Only Attribute" and `examples/database-mysql` additionally failed with "Missing required argument". Both now validate.
- **`docs/resources/database.md` documented a schema that no longer exists.** It listed `cpu_cores`/`memory_size_mb` as settable arguments, marked `resource_profile` optional when it is required, showed a `db-medium` profile value that has never been valid, and gave 20m create/update timeout defaults where the code uses 30m. The import example used an integer ID; database instance IDs are UUIDs.

### Changed

- `resource_profile` schema description now enumerates the valid slugs (`micro`, `small`, `medium`, `large`) with their dashboard display names (DD Puiu, DD Uzlina, DD Matita, DD Sinoe). It previously listed only "small, medium, large", omitting `micro` entirely — the slug-vs-display-name gap was a reported source of user confusion, including customers selecting a profile roughly 4x their intended spend.
- `monthly_cost` schema description corrected from "dollars" to "euros".
- `docs/resources/database.md` now documents the `database_name` charset restriction (letters, numbers, underscores; no hyphens, unlike the instance `name`), a common first-apply failure.
- `examples/database-mysql` version constraint bumped from `~> 0.1` to `~> 0.3`.

## [0.3.0] - 2026-07-15

### Fixed — API contract drift (2026-07 audit against the live API)

Unit-test fixtures were rewritten to encode the current API contract, so this class of drift now fails CI instead of passing against stale mocks.

#### Breaking decode bugs (resources were non-functional against production)
- Snapshot IDs (VPS, cache, database) are integers in API responses; the client decoded them as strings and failed on every create/list/get. All snapshot waiters also polled for status `completed` — the API's terminal status is `ready` — and now fail fast on the real failure statuses (`failed`, `create_failed`/`restore_failed`) instead of hanging until timeout.
- `danubedata_vps.ssh_key_id` is numeric in responses; decoding it as a string broke every read of an SSH-key-authenticated VPS.
- Static site and static site domain IDs are UUIDs; the client decoded them as integers, breaking all reads. Domain response fields realigned: `verification_status`/`tls_status`/`deployment_status`/`is_primary`/`dns_instructions` (nested object) replace the removed `status`/`type`/`verification_record`.
- Storage bucket `tags` is a string array, not a map; buckets with tags were unreadable.

#### Serverless — field vocabulary realigned with the current API
- `deployment_type` values are now `docker_image`/`git_repository`/`zip_upload`.
- Renamed attributes: `image_url`→`image` (+ new `image_tag`, default `latest`), `git_repository`→`repository_url`, `git_branch`→`repository_branch`, `min_instances`→`min_scale`, `max_instances`→`max_scale`. Added `source_type`, `git_auth_type`, `git_credentials` (sensitive) for git deployments.
- `repository_url` no longer forces replacement (the API supports in-place edits).

#### Firewalls — realigned with the 2025-10 firewall model
- Rule `action` is `allow`/`deny` (was `accept`/`drop`); rule ordering attribute is `order` (was `priority`); protocols `gre`/`esp` added.
- Removed `default_action` and `is_default` (no longer exist in the API; they caused "inconsistent result after apply" on every create).
- Rule changes are now sent on update (previously silently dropped).

#### Silent no-ops and perpetual diffs
- `danubedata_vps`: `cpu_cores`/`memory_size_gb`/`storage_size_gb` are Computed (derived from `resource_profile`; the API never accepted them), password changes no longer pretend to update in place, and `auth_method`/`ssh_key_id`/`custom_cloud_init`/`network_stack`/`password` now correctly force replacement. `private_ip` is read from `connection_info`. Datacenter validator matches the API (`fsn1` only).
- `danubedata_cache`: `memory_size_mb`/`cpu_cores` are Computed; name validator matches the API (2–63 chars, lowercase DNS-safe); hardcoded `resource_profile` list removed (plans are dynamic).
- `danubedata_database`: name length matches the API (2–63); hardcoded `resource_profile` list removed; datacenter `ash` removed (not offered for databases).
- `parameter_group_id` (cache and database) is preserved in state instead of being nulled on every read (the API does not echo it).
- `danubedata_static_site`: removed `team_id`/`output_directory`/`current_deployment_id` (not part of the API contract; `team_id` forced replacement on every refresh). Added `plan` (defaults to `free`).
- Cache/database engine mapping now prefers the API's `provider.type` over the display name; the caches data source no longer reports a `datacenter` the API never returns.

### Added
- `danubedata_database.storage_size_gb` is now configurable (Optional+Computed): grow-only storage resize via update, with create-then-resize when the configured value exceeds the plan minimum at create time. Shrinking is rejected by the API.

### Known upstream issues
- Firewall rule `order` and rule `name` are accepted but not persisted/echoed by the API, and rule updates are not yet applied server-side; setting them surfaces "inconsistent result after apply" until the platform fix ships.

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

[Unreleased]: https://github.com/AdrianSilaghi/terraform-provider-danubedata/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/AdrianSilaghi/terraform-provider-danubedata/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/AdrianSilaghi/terraform-provider-danubedata/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/AdrianSilaghi/terraform-provider-danubedata/releases/tag/v0.1.0
