# danubedata_databases

Lists all database instances in your account.

## Example Usage

```hcl
data "danubedata_databases" "all" {}

output "database_count" {
  value = length(data.danubedata_databases.all.instances)
}

output "database_endpoints" {
  value = {
    for db in data.danubedata_databases.all.instances : db.name => db.endpoint
  }
}
```

### Find Database by Name

```hcl
data "danubedata_databases" "all" {}

locals {
  production_db = [for db in data.danubedata_databases.all.instances : db if db.name == "production-db"][0]
}

output "production_connection" {
  value = "${local.production_db.engine}://${local.production_db.username}@${local.production_db.endpoint}:${local.production_db.port}/${local.production_db.database_name}"
}
```

### Filter by Engine

```hcl
data "danubedata_databases" "all" {}

locals {
  postgres_dbs = [for db in data.danubedata_databases.all.instances : db if db.engine == "postgresql"]
  mysql_dbs    = [for db in data.danubedata_databases.all.instances : db if db.engine == "mysql"]
}

output "postgres_count" {
  value = length(local.postgres_dbs)
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `instances` - List of database instances. Each instance contains:
  * `id` - Unique identifier for the database instance.
  * `name` - Name of the database instance.
  * `status` - Current status (`pending`, `provisioning`, `starting`, `running`, `stopping`, `stopped`, `restoring`, `updating`, `error`, `destroying`).
  * `engine` - Database engine (`mysql`, `postgresql`, `mariadb`).
  * `version` - Database version.
  * `database_name` - Name of the initial database. Null if none was created.
  * `datacenter` - Datacenter location.
  * `resource_profile` - Resource profile slug selecting CPU, memory and included storage (`micro`, `small`, `medium`, `large`).
  * `cpu_cores` - Number of CPU cores. Derived from `resource_profile`.
  * `memory_size_mb` - Memory size in MB. Derived from `resource_profile`.
  * `storage_size_gb` - Storage size in GB.
  * `endpoint` - Connection endpoint hostname. Null if not yet assigned.
  * `port` - Connection port. Null if not yet assigned.
  * `username` - Database admin username. Null if not yet assigned.
  * `monthly_cost` - Estimated monthly cost.
  * `created_at` - Timestamp when the instance was created.
