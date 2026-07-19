# danubedata_database_replica

Manages a single read replica of a managed database instance.

Each resource represents one replica. The API assigns the replica's index
server-side, so replicas must be created one at a time — see
[Multiple Replicas](#multiple-replicas).

## Example Usage

### Single Replica

```hcl
resource "danubedata_database" "primary" {
  name             = "app-db"
  database_name    = "app_production"
  engine           = "postgresql"
  resource_profile = "medium"
  datacenter       = "fsn1"
}

resource "danubedata_database_replica" "read" {
  database_instance_id = danubedata_database.primary.id
}

output "replica_endpoint" {
  value = danubedata_database_replica.read.endpoint
}
```

### Multiple Replicas

The API derives a new replica's index from the highest index that currently
exists, so two replica creations running at the same time compute the same index
and collide. Creation must be serialized.

Terraform cannot express "each element of this resource depends on the previous
one" — `depends_on` may not reference the resource it is declared in. For more
than one replica, use separate resource blocks chained with `depends_on`:

```hcl
resource "danubedata_database_replica" "read_1" {
  database_instance_id = danubedata_database.primary.id
}

resource "danubedata_database_replica" "read_2" {
  database_instance_id = danubedata_database.primary.id

  depends_on = [danubedata_database_replica.read_1]
}
```

`count` and `for_each` keep the configuration shorter, but Terraform will create
the instances in parallel. If you use them, apply with `-parallelism=1` so the
replicas are still created one at a time:

```hcl
resource "danubedata_database_replica" "read" {
  count = 3

  database_instance_id = danubedata_database.primary.id
}
```

```bash
terraform apply -parallelism=1
```

## Argument Reference

### Required

* `database_instance_id` - ID of the parent database instance, a UUID. Changing
  this forces a new resource.

### Timeouts

* `create` - (Default `30m`) Time to wait for the replica to become ready.
* `delete` - (Default `10m`) Time to wait for replica deletion.

## Attribute Reference

In addition to the argument above, the following are exported:

* `id` - Composite identifier, `{database_instance_id}:{replica_index}`.
* `replica_index` - 1-based index of this replica within the parent instance,
  assigned by the API.
* `name` - Name of the replica node.
* `node_id` - Internal node identifier.
* `endpoint` - Connection endpoint for the replica. May be null before the
  replica is fully provisioned.
* `status` - Current status of the replica.
* `ready` - Whether the replica is ready to serve reads.
* `replication_status` - Replication status (`healthy`, `lagging`, `broken`).
  May be null.
* `seconds_behind_master` - Replication lag in seconds. May be null.
* `is_replication_healthy` - Whether replication is healthy.

## Import

Replicas can be imported using the composite ID, `{database_instance_id}:{replica_index}`:

```bash
terraform import danubedata_database_replica.read_1 9f8c2d14-3b7a-4e51-9c6d-2a1f8e0b7c33:1
```

## Notes

- The parent instance must be `running`. The API rejects adding or removing a
  replica while the instance is in any other state, so a replica added in the
  same apply as its parent depends on the parent's create waiter having
  completed — which the reference to `danubedata_database.primary.id` already
  guarantees.
- Replica indexes are assigned as "highest existing index + 1" and are never
  renumbered. Deleting a replica leaves a gap that is not reused: destroying
  index 1 while index 2 exists means the next replica created is index 3.
- This resource has no update operation. `database_instance_id` forces
  replacement; every other attribute is read-only.
- Each replica is a separate billable node in addition to the primary.
- Create waits for the replica to report ready, and fails fast if it enters
  `error` or `failed`.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.
