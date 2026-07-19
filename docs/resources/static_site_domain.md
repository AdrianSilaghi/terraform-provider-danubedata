# danubedata_static_site_domain

Manages a custom domain attached to a static site.

Attaching a domain does not make it live. The resource is created in `pending`
verification status; you then add the DNS record described in `dns_instructions`
and trigger verification out of band — see [Verification](#verification).

## Example Usage

### Attach a Domain

```hcl
resource "danubedata_static_site" "marketing" {
  name = "marketing-site"
}

resource "danubedata_static_site_domain" "www" {
  static_site_id = danubedata_static_site.marketing.id
  domain         = "www.example.com"
}

output "dns_record" {
  value = danubedata_static_site_domain.www.dns_instructions
}
```

### Several Domains on One Site

```hcl
resource "danubedata_static_site_domain" "aliases" {
  for_each = toset(["www.example.com", "example.com"])

  static_site_id = danubedata_static_site.marketing.id
  domain         = each.key
}
```

### Surfacing the Record to Add

```hcl
output "verification_record" {
  value = {
    type  = danubedata_static_site_domain.www.dns_instructions.record_type
    name  = danubedata_static_site_domain.www.dns_instructions.record_name
    value = danubedata_static_site_domain.www.dns_instructions.record_value
  }
}
```

## Verification

1. Apply the configuration. The domain is attached with
   `verification_status = "pending"`.
2. Read `dns_instructions` and create the record it describes at your DNS
   provider. The record type is whatever `dns_instructions.record_type`
   reports — do not hardcode an assumption about it.
3. Once the record has propagated, trigger verification:

   ```bash
   danube pages domains verify www.example.com
   ```

4. Run `terraform plan` (or `terraform refresh`) to pull the updated
   `verification_status`, `tls_status` and `deployment_status` into state.

Terraform does not wait for or poll any of these steps: a successful `apply`
means the domain was attached, not that it is serving traffic.

## Argument Reference

### Required

* `static_site_id` - ID of the parent static site, a UUID. Changing this forces
  a new resource.
* `domain` - The custom domain, e.g. `www.example.com`. Changing this forces a
  new resource.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - Composite identifier, `{static_site_id}:{domain_id}`.
* `domain_id` - ID of the domain attachment, a UUID.
* `verification_status` - DNS ownership verification status (`pending`,
  `verifying`, `verified`, `failed`).
* `tls_status` - TLS certificate provisioning status (`pending`,
  `provisioning`, `active`, `failed`).
* `deployment_status` - Status of routing the domain to the site's active
  deployment (`pending`, `deploying`, `active`, `failed`).
* `is_primary` - Whether this is the primary domain for the site.
* `dns_instructions` - The DNS record to add for ownership verification, an
  object with:
  * `record_type` - DNS record type.
  * `record_name` - DNS record name.
  * `record_value` - DNS record value.
  * `instructions` - Human-readable instructions for configuring the record.
* `created_at` - Timestamp.

## Import

Domain attachments are imported using `{static_site_id}:{domain}` — the domain
**name**, not `domain_id`:

```bash
terraform import danubedata_static_site_domain.www 7c3e5a91-42bd-4f08-9a17-5d8b0c6e4f22:www.example.com
```

Note the asymmetry: the `id` attribute written to state is
`{static_site_id}:{domain_id}`, but the import address takes the domain name.
The provider rewrites `id` to its canonical form once the first read resolves
the domain's UUID.

## Notes

- This resource has no update operation. Both arguments force replacement, so
  correcting a typo in `domain` detaches the old domain and attaches the new
  one — verification starts over.
- Verification and TLS issuance happen asynchronously and are never awaited by
  Terraform. The three status attributes reflect whatever the API reported at
  the last refresh.
- `is_primary` is read-only. Promoting a domain to primary is not exposed
  through this resource.
- There is no `updated_at` attribute on this resource, and no `timeouts` block.
- Destroying the parent `danubedata_static_site` takes its domains with it.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.
