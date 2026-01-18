# danubedata_database_providers

Lists available database providers (MySQL, PostgreSQL, MariaDB).

## Example Usage

```hcl
data "danubedata_database_providers" "all" {}

output "database_providers" {
  value = [for p in data.danubedata_database_providers.all.providers : {
    id   = p.id
    name = p.name
  }]
}
```

### Get Provider by Name

```hcl
data "danubedata_database_providers" "all" {}

locals {
  postgres = [for p in data.danubedata_database_providers.all.providers : p if p.name == "PostgreSQL"][0]
}

resource "danubedata_database" "main" {
  name            = "my-database"
  provider_id     = local.postgres.id
  storage_size_gb = 20
  memory_size_mb  = 2048
  cpu_cores       = 2
  datacenter      = "fsn1"
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `providers` - List of database providers. Each provider contains:
  * `id` - The provider ID.
  * `name` - Provider name (e.g., "MySQL", "PostgreSQL", "MariaDB").
  * `type` - Provider type identifier.
