# danubedata_vps_snapshot

Manages a VPS snapshot for backup and recovery.

## Example Usage

### Basic Snapshot

```hcl
resource "danubedata_vps" "server" {
  name        = "web-server"
  image       = "ubuntu-22.04"
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

### Multiple Snapshots

```hcl
resource "danubedata_vps_snapshot" "daily" {
  name            = "daily-${formatdate("YYYY-MM-DD", timestamp())}"
  description     = "Daily automated snapshot"
  vps_instance_id = danubedata_vps.server.id

  lifecycle {
    create_before_destroy = true
  }
}
```

## Argument Reference

### Required

* `name` - (Required) Name of the snapshot.
* `vps_instance_id` - (Required) ID of the VPS instance to snapshot.

### Optional

* `description` - (Optional) Description of the snapshot.

### Timeouts

* `create` - (Default `15m`) Time to wait for snapshot creation.
* `delete` - (Default `5m`) Time to wait for snapshot deletion.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The snapshot ID.
* `status` - Current status (`creating`, `ready`, `error`).
* `size_gb` - Size of the snapshot in GB.
* `created_at` - Creation timestamp.

## Import

VPS snapshots can be imported using their ID:

```bash
terraform import danubedata_vps_snapshot.example snap-abc123
```

## Notes

- Snapshots capture the entire VPS disk state
- Creating a snapshot may briefly impact VPS performance
- Snapshots are stored separately from the VPS and persist after VPS deletion
- Use snapshots for point-in-time recovery or cloning VPS instances
