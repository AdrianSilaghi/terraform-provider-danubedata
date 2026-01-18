# danubedata_database

Manages a managed database instance (MySQL, PostgreSQL, or MariaDB).

## Example Usage

### MySQL Database

```hcl
resource "danubedata_database" "mysql" {
  name               = "my-mysql"
  database_name      = "app_production"
  provider_id        = 1  # MySQL
  version            = "8.0"
  storage_size_gb    = 20
  memory_size_mb     = 2048
  cpu_cores          = 2
  datacenter         = "fsn1"
}

output "mysql_endpoint" {
  value = danubedata_database.mysql.endpoint
}

output "mysql_port" {
  value = danubedata_database.mysql.port
}
```

### PostgreSQL Database

```hcl
resource "danubedata_database" "postgres" {
  name               = "my-postgres"
  database_name      = "analytics"
  provider_id        = 2  # PostgreSQL
  version            = "16"
  storage_size_gb    = 50
  memory_size_mb     = 4096
  cpu_cores          = 4
  datacenter         = "fsn1"
}
```

### MariaDB Database

```hcl
resource "danubedata_database" "mariadb" {
  name               = "my-mariadb"
  database_name      = "legacy_app"
  provider_id        = 3  # MariaDB
  storage_size_gb    = 20
  memory_size_mb     = 2048
  cpu_cores          = 2
  datacenter         = "fsn1"
}
```

### Using Resource Profile

```hcl
resource "danubedata_database" "standard" {
  name             = "standard-db"
  provider_id      = 1
  resource_profile = "db-medium"
  datacenter       = "fsn1"
}
```

## Argument Reference

### Required

* `name` - (Required) Name of the database instance.
* `provider_id` - (Required) Database provider ID:
  - `1` - MySQL
  - `2` - PostgreSQL
  - `3` - MariaDB
* `datacenter` - (Required) Datacenter location (e.g., `fsn1`).

### Optional

* `database_name` - (Optional) Name of the initial database to create.
* `storage_size_gb` - (Optional) Storage size in GB.
* `memory_size_mb` - (Optional) Memory size in MB.
* `cpu_cores` - (Optional) Number of CPU cores.
* `version` - (Optional) Database version (e.g., `8.0` for MySQL, `16` for PostgreSQL).
* `resource_profile` - (Optional) Predefined resource profile.

### Timeouts

* `create` - (Default `20m`) Time to wait for database creation.
* `update` - (Default `20m`) Time to wait for database updates.
* `delete` - (Default `15m`) Time to wait for database deletion.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The database instance ID.
* `status` - Current status.
* `endpoint` - Connection endpoint hostname.
* `port` - Connection port.
* `username` - Database admin username.
* `monthly_cost` - Estimated monthly cost.
* `created_at` - Creation timestamp.
* `deployed_at` - Deployment timestamp.

## Import

Database instances can be imported using their ID:

```bash
terraform import danubedata_database.example db-abc123
```

## Notes

- Database credentials are managed separately and can be retrieved via the API.
- Storage size can only be increased, not decreased.
- Version upgrades may cause brief downtime.
