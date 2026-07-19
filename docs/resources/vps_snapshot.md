# danubedata_vps_snapshot

Manages a VPS snapshot for backup and recovery.

## Example Usage

### Basic Snapshot

```hcl
resource "danubedata_vps" "server" {
  name        = "web-server"
  image       = "ubuntu-24.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.main.id
}

resource "danubedata_vps_snapshot" "backup" {
  name            = "pre-upgrade-backup"
  vps_instance_id = danubedata_vps.server.id
}
```

### Snapshot with Description

```hcl
resource "danubedata_vps_snapshot" "release" {
  name            = "v1.0-release"
  description     = "Snapshot before v1.0 release deployment"
  vps_instance_id = danubedata_vps.server.id
}
```

### Several Snapshots at Once

Every argument forces a new resource when changed, so use stable names. A name
built from `timestamp()` would change on each plan and destroy and recreate the
snapshot on every apply.

```hcl
resource "danubedata_vps_snapshot" "milestone" {
  for_each = toset(["pre-migration", "post-migration"])

  name            = each.key
  description     = "Migration checkpoint: ${each.key}"
  vps_instance_id = danubedata_vps.server.id
}
```

## Argument Reference

### Required

* `name` - Name of the snapshot. Changing this forces a new resource.
* `vps_instance_id` - ID of the VPS instance to snapshot, a UUID. Changing this
  forces a new resource.

### Optional

* `description` - Description of the snapshot. Changing this forces a new
  resource.

### Timeouts

* `create` - (Default `30m`) Time to wait for snapshot creation.
* `delete` - (Default `10m`) Time to wait for snapshot deletion.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The snapshot ID.
* `status` - Current status (`creating`, `ready`, `failed`).
* `size_gb` - Size of the snapshot in GB.
* `created_at` / `updated_at` - Timestamps.

## Import

VPS snapshots can be imported using their ID, which is a numeric identifier:

```bash
terraform import danubedata_vps_snapshot.example 4821
```

## Notes

- Snapshots capture the VPS disk state at the moment they are taken.
- This resource has no update operation. Changing `name`, `description` or
  `vps_instance_id` destroys the snapshot and takes a new one.
- Snapshots belong to their VPS: deleting the VPS deletes its snapshots. Plan
  destroys accordingly — a snapshot is not an escape hatch for a VPS you are
  about to tear down.
- Restoring from a snapshot is not exposed as a Terraform resource; use the
  dashboard or API.
