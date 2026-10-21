# danubedata_cache_snapshot

Manages a cache instance snapshot for backup and recovery.

## Example Usage

### Basic Snapshot

```hcl
resource "danubedata_cache" "sessions" {
  name             = "sessions"
  cache_provider   = "redis"
  resource_profile = "small"
  datacenter       = "fsn1"
}

resource "danubedata_cache_snapshot" "backup" {
  name              = "pre-upgrade-backup"
  cache_instance_id = danubedata_cache.sessions.id
}
```

### Snapshot with Description

```hcl
resource "danubedata_cache_snapshot" "release" {
  name              = "v1.0-release"
  description       = "Snapshot before v1.0 release deployment"
  cache_instance_id = danubedata_cache.sessions.id
}
```

### Several Snapshots at Once

Every argument forces a new resource when changed, so use stable names. A name
built from `timestamp()` would change on each plan and destroy and recreate the
snapshot on every apply.

```hcl
resource "danubedata_cache_snapshot" "milestone" {
  for_each = toset(["pre-migration", "post-migration"])

  name              = each.key
  description       = "Migration checkpoint: ${each.key}"
  cache_instance_id = danubedata_cache.sessions.id
}
```

## Argument Reference

### Required

* `name` - Name of the snapshot. Changing this forces a new resource.
* `cache_instance_id` - ID of the cache instance to snapshot, a UUID. Changing
  this forces a new resource.

### Optional

* `description` - Description of the snapshot. Changing this forces a new
  resource.

### Timeouts

* `create` - (Default `30m`) Time to wait for snapshot creation.
* `delete` - (Default `10m`) Time to wait for snapshot deletion.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The snapshot ID.
* `status` - Current status. `creating` while in progress, `ready` when
  complete.
* `size_mb` - Size of the snapshot in MB.
* `size_bytes` - Exact size in bytes, or null while unknown.
* `size_basis` - How size was measured (`stored`, `uploaded`, or `captured`), or null while unknown.
* `created_at` / `updated_at` - Timestamps.

## Import

Cache snapshots can be imported using their ID, which is a numeric identifier:

```bash
terraform import danubedata_cache_snapshot.example 3907
```

## Notes

- Create waits for the snapshot to reach `ready`. It fails fast if the snapshot
  reports `failed` or `restore_failed`.
- This resource has no update operation. Changing `name`, `description` or
  `cache_instance_id` destroys the snapshot and takes a new one.
- Snapshots belong to their instance: deleting the cache instance deletes its
  snapshots. Plan destroys accordingly — a snapshot is not an escape hatch for
  an instance you are about to tear down.
- Restoring from a snapshot is not exposed as a Terraform resource; use the
  dashboard or API.
- Pin the project with provider `team_id` or `DANUBEDATA_TEAM_ID`. Otherwise,
  a project-locked token uses its project and an account-wide token uses the
  account's current project. Confirm the selected project before applying.
