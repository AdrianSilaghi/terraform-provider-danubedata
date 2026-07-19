# danubedata_database_snapshots

Lists all database snapshots in your account.

## Example Usage

```hcl
data "danubedata_database_snapshots" "all" {}

output "snapshot_count" {
  value = length(data.danubedata_database_snapshots.all.snapshots)
}

output "total_snapshot_size_gb" {
  value = sum([for s in data.danubedata_database_snapshots.all.snapshots : s.size_gb])
}
```

### Find Snapshot by Name

```hcl
data "danubedata_database_snapshots" "all" {}

locals {
  pre_upgrade = [for s in data.danubedata_database_snapshots.all.snapshots : s if s.name == "pre-upgrade-backup"][0]
}

output "pre_upgrade_snapshot_id" {
  value = local.pre_upgrade.id
}
```

### Filter Snapshots by Database Instance

```hcl
data "danubedata_database_snapshots" "all" {}

locals {
  instance_snapshots = [
    for s in data.danubedata_database_snapshots.all.snapshots : s
    if s.database_instance_id == danubedata_database.postgres.id
  ]
}

output "instance_snapshot_count" {
  value = length(local.instance_snapshots)
}
```

### Filter Ready Snapshots

Only snapshots in the `ready` state can be restored.

```hcl
data "danubedata_database_snapshots" "all" {}

locals {
  ready_snapshots = [for s in data.danubedata_database_snapshots.all.snapshots : s if s.status == "ready"]
}

output "ready_count" {
  value = length(local.ready_snapshots)
}
```

## Argument Reference

This data source has no arguments. It returns every database snapshot in the
account; filter the result in Terraform as shown above.

## Attribute Reference

* `snapshots` - List of database snapshots. Each snapshot contains:
  * `id` - Unique identifier for the snapshot. A numeric ID, exposed as a string.
  * `name` - Name of the snapshot.
  * `description` - Description of the snapshot.
  * `status` - Current status (`pending`, `creating`, `ready`, `failed`, `restoring`, `restore_failed`, `deleting`).
  * `database_instance_id` - UUID of the database instance this snapshot belongs to.
  * `size_gb` - Size of the snapshot in GB. May be fractional.
  * `created_at` - Timestamp when the snapshot was created.

~> **Note** The terminal success state is `ready`, not `completed`. `failed`
means snapshot creation did not finish; `restore_failed` means a restore
attempt from that snapshot did not finish.
