# danubedata_cache_snapshots

Lists all cache snapshots in your account.

## Example Usage

```hcl
data "danubedata_cache_snapshots" "all" {}

output "snapshot_count" {
  value = length(data.danubedata_cache_snapshots.all.snapshots)
}

output "total_snapshot_size_mb" {
  value = sum([for s in data.danubedata_cache_snapshots.all.snapshots : s.size_mb])
}
```

### Find Snapshot by Name

```hcl
data "danubedata_cache_snapshots" "all" {}

locals {
  pre_migration = [for s in data.danubedata_cache_snapshots.all.snapshots : s if s.name == "pre-migration-backup"][0]
}

output "pre_migration_snapshot_id" {
  value = local.pre_migration.id
}
```

### Filter Snapshots by Cache Instance

```hcl
data "danubedata_cache_snapshots" "all" {}

locals {
  instance_snapshots = [
    for s in data.danubedata_cache_snapshots.all.snapshots : s
    if s.cache_instance_id == danubedata_cache.main.id
  ]
}

output "instance_snapshot_count" {
  value = length(local.instance_snapshots)
}
```

### Filter Ready Snapshots

Only snapshots in the `ready` state can be restored.

```hcl
data "danubedata_cache_snapshots" "all" {}

locals {
  ready_snapshots = [for s in data.danubedata_cache_snapshots.all.snapshots : s if s.status == "ready"]
}

output "ready_count" {
  value = length(local.ready_snapshots)
}
```

## Argument Reference

This data source has no arguments. It returns every cache snapshot in the
account; filter the result in Terraform as shown above.

## Attribute Reference

* `snapshots` - List of cache snapshots. Each snapshot contains:
  * `id` - Unique identifier for the snapshot. A numeric ID, exposed as a string.
  * `name` - Name of the snapshot.
  * `description` - Description of the snapshot.
  * `status` - Current status (`pending`, `creating`, `ready`, `failed`, `restoring`, `restore_failed`, `deleting`).
  * `cache_instance_id` - UUID of the cache instance this snapshot belongs to.
  * `size_mb` - Size of the snapshot in MB. May be fractional.
  * `created_at` - Timestamp when the snapshot was created.

~> **Note** The terminal success state is `ready`, not `completed`. `failed`
means snapshot creation did not finish; `restore_failed` means a restore
attempt from that snapshot did not finish.
