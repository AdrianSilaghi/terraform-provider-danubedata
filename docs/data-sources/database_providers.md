# danubedata_database_providers

Lists available database providers (MySQL, PostgreSQL, MariaDB).

This list is compiled into the provider rather than fetched from the API, so it
does not reflect per-account availability.

## Example Usage

```hcl
data "danubedata_database_providers" "all" {}

output "database_providers" {
  value = [for p in data.danubedata_database_providers.all.providers : {
    type = p.type
    name = p.name
  }]
}
```

### Get Provider by Name

The `danubedata_database` resource selects its engine with the `type` value, not
the numeric `id`. CPU and memory come from `resource_profile` and cannot be set
directly.

```hcl
data "danubedata_database_providers" "all" {}

locals {
  postgres = [for p in data.danubedata_database_providers.all.providers : p if p.name == "PostgreSQL"][0]
}

resource "danubedata_database" "main" {
  name             = "my-database"
  engine           = local.postgres.type
  resource_profile = "small"
  storage_size_gb  = 20
  datacenter       = "fsn1"
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `providers` - List of database providers. Each provider contains:
  * `id` - The provider ID (a number). Not accepted by any resource argument; use `type` instead.
  * `name` - Provider name (`MySQL`, `PostgreSQL`, `MariaDB`).
  * `type` - Provider type identifier (`mysql`, `postgresql`, `mariadb`). This is what the `danubedata_database` resource's `engine` argument expects.
  * `description` - Provider description.
  * `version` - Default version.
  * `default_port` - Default port number.
