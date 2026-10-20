# danubedata_static_site

Manages a static site.

This resource manages the site container only, not its content. Deployments are
triggered out of band — from the `danube` CLI or from CI/CD — so a site created
by Terraform exists and has a URL before anything has been published to it.

## Example Usage

### Basic Site

```hcl
resource "danubedata_static_site" "marketing" {
  name = "marketing-site"
}

output "site_url" {
  value = danubedata_static_site.marketing.url
}
```

### Site with a Custom Domain

```hcl
resource "danubedata_static_site" "marketing" {
  name = "marketing-site"
}

resource "danubedata_static_site_domain" "www" {
  static_site_id = danubedata_static_site.marketing.id
  domain         = "www.example.com"
}
```

## Argument Reference

### Required

* `name` - Name of the static site. Changing this forces a new resource.

### Optional

* `plan` - Deprecated and ignored by the API. Retained for compatibility with
  existing configurations; the project's Static Sites plan covers every site.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The static site ID.
* `slug` - URL slug for the site.
* `url` - Default URL of the deployed site.
* `status` - Current status of the site.
* `created_at` / `updated_at` - Timestamps.

## Import

Static sites can be imported using their ID:

```bash
terraform import danubedata_static_site.example 7c3e5a91-42bd-4f08-9a17-5d8b0c6e4f22
```

## Notes

- The API exposes no update operation for static sites. Changing `name` replaces
  the site and its attached domains. Manage the project's Static Sites plan
  outside this resource.
- Creating the resource does not publish content. Deploy with the `danube` CLI
  or from CI/CD after the site exists.
- Custom domains are managed separately, with `danubedata_static_site_domain`.
- Pin the project with provider `team_id` or `DANUBEDATA_TEAM_ID`. Otherwise,
  a project-locked token uses its project and an account-wide token uses the
  account's current project. Confirm the selected project before applying.
