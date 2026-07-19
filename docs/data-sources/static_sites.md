# danubedata_static_sites

Lists all static sites for a given team.

Unlike the other data sources, which operate on the API token owner's current
team, this one takes an explicit `team_id`.

## Example Usage

```hcl
variable "team_id" {
  description = "Numeric ID of the team whose static sites to list"
  type        = number
}

data "danubedata_static_sites" "all" {
  team_id = var.team_id
}

output "site_count" {
  value = length(data.danubedata_static_sites.all.sites)
}

output "site_urls" {
  value = {
    for site in data.danubedata_static_sites.all.sites : site.name => site.url
  }
}
```

### Find Site by Name

```hcl
data "danubedata_static_sites" "all" {
  team_id = var.team_id
}

locals {
  marketing = [for s in data.danubedata_static_sites.all.sites : s if s.name == "marketing"][0]
}

output "marketing_url" {
  value = local.marketing.url
}

output "marketing_slug" {
  value = local.marketing.slug
}
```

### Filter Active Sites

```hcl
data "danubedata_static_sites" "all" {
  team_id = var.team_id
}

locals {
  active_sites = [for s in data.danubedata_static_sites.all.sites : s if s.status == "active"]
}

output "active_site_names" {
  value = [for s in local.active_sites : s.name]
}
```

### Filter by Plan

```hcl
data "danubedata_static_sites" "all" {
  team_id = var.team_id
}

locals {
  free_sites = [for s in data.danubedata_static_sites.all.sites : s if s.plan == "free"]
  pro_sites  = [for s in data.danubedata_static_sites.all.sites : s if s.plan == "pro"]
}

output "pro_site_count" {
  value = length(local.pro_sites)
}
```

## Argument Reference

### Required

* `team_id` - Numeric ID of the team whose static sites to list. The API token
  must have access to this team.

## Attribute Reference

* `sites` - List of static sites. Each site contains:
  * `id` - Unique identifier (UUID) for the static site.
  * `name` - Name of the static site.
  * `slug` - URL-safe slug for the site.
  * `url` - Public URL for the site.
  * `plan` - Plan the site is on (`free`, `starter`, `pro`).
  * `status` - Current status (`pending`, `building`, `deploying`, `active`, `stopped`, `error`, `suspended`, `pending_review`).
  * `created_at` - Timestamp when the site was created.

## Notes

- Site IDs are UUIDs, not integers.
- The provider pages through the full result set, so `sites` contains every site for the team, not just the first page.
- This data source does not return builds, deployments or custom domains.
