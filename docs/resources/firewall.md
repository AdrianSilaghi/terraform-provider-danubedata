# danubedata_firewall

Manages a firewall with rules for network traffic control.

## Example Usage

### Basic Firewall

`rules` is a list attribute, so it is assigned with `=` and square brackets —
not repeated `rules { }` blocks.

Rules are evaluated in the order they appear in the list. `name` and `order`
are omitted below deliberately — see
[Rule `order` and `name`](#rule-order-and-name).

```hcl
resource "danubedata_firewall" "web" {
  name        = "web-firewall"
  description = "Allow HTTP/HTTPS and SSH"

  rules = [
    {
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port_range_start = 22
      port_range_end   = 22
      source_ips       = ["0.0.0.0/0"]
    },
    {
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port_range_start = 80
      port_range_end   = 80
      source_ips       = ["0.0.0.0/0"]
    },
    {
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port_range_start = 443
      port_range_end   = 443
      source_ips       = ["0.0.0.0/0"]
    },
  ]
}
```

### Firewall with IP Restrictions

```hcl
resource "danubedata_firewall" "admin" {
  name        = "admin-firewall"
  description = "Restricted admin access"

  rules = [
    {
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port_range_start = 22
      port_range_end   = 22
      source_ips       = ["203.0.113.0/24", "198.51.100.0/24"]
    },
    {
      action     = "allow"
      direction  = "outbound"
      protocol   = "any"
      source_ips = ["0.0.0.0/0"]
    },
  ]
}
```

### Firewall with No Rules

`rules` may be omitted entirely, which creates the firewall with an empty rule
set for you to populate later.

```hcl
resource "danubedata_firewall" "placeholder" {
  name        = "staging-firewall"
  description = "Rules managed out of band for now"
}
```

## Argument Reference

### Required

* `name` - Name of the firewall.

### Optional

* `description` - Description of the firewall.
* `rules` - List of firewall rules. See [Rules](#rules) below.

### Rules

Each element of `rules` supports:

* `action` - (Required) Action to take. One of `allow`, `deny`.
* `direction` - (Required) Direction. One of `inbound`, `outbound`.
* `protocol` - (Required) Protocol. One of `tcp`, `udp`, `icmp`, `any`, `gre`,
  `esp`.
* `name` - (Optional) Name/description of the rule. **Not yet honoured by the
  API** — see below.
* `port_range_start` - (Optional) Start of port range (1-65535).
* `port_range_end` - (Optional) End of port range (1-65535).
* `source_ips` - (Optional) List of source IP addresses or CIDR blocks.
* `order` - (Optional) Rule evaluation order; lower numbers are evaluated
  first. **Not yet honoured by the API** — see below.
* `id` - (Read-only) Rule ID, assigned by the API.

### Rule `order` and `name`

~> **Caveat** `order` and `name` are part of the intended rule contract and are
accepted by the provider schema, but the API does not yet honour either one.
Setting them on a rule causes the apply to fail with **"Provider produced
inconsistent result after apply"**. Omit both for a clean apply until the
platform fix lands.

Specifically:

* A submitted `order` is discarded. Rules are auto-numbered in the sequence they
  appear in the `rules` list, and `order` is never returned by the API.
* A submitted `name` is stored, but under a different field than the one the API
  reads back, so it always returns as null.

Neither gap costs you control over evaluation order: because rules are numbered
in submission sequence, **the order of elements in the `rules` list is the
effective evaluation order**. Write rules most-specific first, and reorder the
list to reorder evaluation.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The firewall ID.
* `status` - Current status (`draft`, `active`, `deploying`).
* `created_at` / `updated_at` - Timestamps.

## Import

Firewalls can be imported using their ID, which is a UUID:

```bash
terraform import danubedata_firewall.example 4d1e7b90-6c25-4a38-b1f7-8e93c05a2d64
```

## Notes

- Rules are replaced wholesale on update: the provider sends the full `rules`
  list on every change, so removing an element from configuration removes the
  rule.
- `order` and `name` on a rule are not yet honoured server-side and will fail
  the apply; see [Rule `order` and `name`](#rule-order-and-name).
- Each rule's `id` is assigned by the API and is read-only; do not set it in
  configuration.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.
