# danubedata_database_snapshot

Manages a database instance snapshot for backup and recovery.

## Example Usage

### Basic Snapshot

```hcl
resource "danubedata_database" "primary" {
  name             = "app-db"
  database_name    = "app_production"
  engine           = "postgresql"
  resource_profile = "medium"
  datacenter       = "fsn1"
}

resource "danubedata_database_snapshot" "backup" {
  name                 = "pre-migration-backup"
  database_instance_id = danubedata_database.primary.id
}
```

### Snapshot with Description

```hcl
resource "danubedata_database_snapshot" "release" {
  name                 = "v1.0-release"
  description          = "Snapshot before v1.0 schema migration"
  database_instance_id = danubedata_database.primary.id
}
```

### Several Snapshots at Once

Every argument forces a new resource when changed, so use stable names. A name
built from `timestamp()` would change on each plan and destroy and recreate the
snapshot on every apply.

```hcl
resource "danubedata_database_snapshot" "milestone" {
  for_each = toset(["pre-migration", "post-migration"])

  name                 = each.key
  description          = "Migration checkpoint: ${each.key}"
  database_instance_id = danubedata_database.primary.id
}
```

## Argument Reference

### Required

* `name` - Name of the snapshot. Changing this forces a new resource.
* `database_instance_id` - ID of the database instance to snapshot, a UUID.
  Changing this forces a new resource.

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
* `size_gb` - Size of the snapshot in GB.
* `created_at` / `updated_at` - Timestamps.

## Import

Database snapshots can be imported using their ID, which is a numeric
identifier:

```bash
terraform import danubedata_database_snapshot.example 5142
```

## Notes

- Create waits for the snapshot to reach `ready`. It fails fast if the snapshot
  reports `failed` or `restore_failed`.
- This resource has no update operation. Changing `name`, `description` or
  `database_instance_id` destroys the snapshot and takes a new one.
- Snapshots belong to their instance: deleting the database instance deletes its
  snapshots. Plan destroys accordingly — a snapshot is not an escape hatch for
  an instance you are about to tear down.
- Restoring from a snapshot is not exposed as a Terraform resource; use the
  dashboard or API.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.
