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
      port             = 22
      source_ips       = ["0.0.0.0/0"]
    },
    {
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port             = 80
      source_ips       = ["0.0.0.0/0"]
    },
    {
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port             = 443
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
      port             = 22
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
* `name` - (Optional) Name/description of the rule.
* `port` - (Optional) Single destination port (1-65535).
* `port_range_start` - (Optional) Start of port range (1-65535).
* `port_range_end` - (Optional) End of port range (1-65535).
* `source_ips` - (Optional) List of source IP addresses or CIDR blocks.
* `source_instance_ids` - (Optional) Source resources as `{ id, type }` pairs. Omitted restrictions on an identifiable existing rule are preserved; set `[]` to remove them explicitly.
* `order` - (Optional, computed) Rule evaluation order. The API assigns it from list order.
* `id` - (Read-only) Rule identity returned by the API. Platform-locked rules remain on the server and are excluded from the Terraform-managed `rules` list.

### Rule `order` and `name`

The API stores `name` and assigns `order` based on list order. Place the most specific rules first. A metadata-only firewall update leaves rules untouched; changing `rules` sends the full desired rule list.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The firewall ID.
* `status` - Current status (`draft`, `active`, `applying`, `error`).
* `is_default` / `is_system` - API default marker and platform-managed flag. A system firewall cannot be detached or deleted.
* `default_action` / `team_id` - Default action and owning project ID.
* `created_at` / `updated_at` - Timestamps.

## Import

Firewalls can be imported using their ID, which is a UUID:

```bash
terraform import danubedata_firewall.example 4d1e7b90-6c25-4a38-b1f7-8e93c05a2d64
```

## Notes

- Changing `rules` sends the full desired list. Removing an element removes that rule. Name or description changes leave the rule set untouched.
- Each rule's `id` is assigned by the API and is read-only; do not set it in
  configuration.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.
