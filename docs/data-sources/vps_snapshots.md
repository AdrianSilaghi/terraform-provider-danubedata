# danubedata_vps_snapshots

Lists all VPS snapshots in your account.

## Example Usage

```hcl
data "danubedata_vps_snapshots" "all" {}

output "snapshot_count" {
  value = length(data.danubedata_vps_snapshots.all.snapshots)
}

output "total_snapshot_size_gb" {
  value = sum([for s in data.danubedata_vps_snapshots.all.snapshots : s.size_gb])
}
```

### Find Snapshot by Name

```hcl
data "danubedata_vps_snapshots" "all" {}

locals {
  pre_upgrade = [for s in data.danubedata_vps_snapshots.all.snapshots : s if s.name == "pre-upgrade-backup"][0]
}

output "pre_upgrade_snapshot_id" {
  value = local.pre_upgrade.id
}
```

### Filter Snapshots by VPS Instance

```hcl
data "danubedata_vps_snapshots" "all" {}

variable "vps_id" {
  default = "vps-abc123"
}

locals {
  vps_snapshots = [for s in data.danubedata_vps_snapshots.all.snapshots : s if s.vps_instance_id == var.vps_id]
}

output "vps_snapshot_count" {
  value = length(local.vps_snapshots)
}
```

### Filter Ready Snapshots

```hcl
data "danubedata_vps_snapshots" "all" {}

locals {
  ready_snapshots = [for s in data.danubedata_vps_snapshots.all.snapshots : s if s.status == "ready"]
}

output "ready_count" {
  value = length(local.ready_snapshots)
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `snapshots` - List of VPS snapshots. Each snapshot contains:
  * `id` - Unique identifier for the snapshot.
  * `name` - Name of the snapshot.
  * `description` - Description of the snapshot.
  * `status` - Current status (creating, ready, error).
  * `vps_instance_id` - ID of the VPS instance this snapshot belongs to.
  * `size_gb` - Size of the snapshot in GB.
  * `created_at` - Timestamp when the snapshot was created.
