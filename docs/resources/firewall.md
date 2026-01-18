# danubedata_firewall

Manages a firewall with rules for network traffic control.

## Example Usage

### Basic Firewall

```hcl
resource "danubedata_firewall" "web" {
  name           = "web-firewall"
  description    = "Allow HTTP/HTTPS and SSH"
  default_action = "deny"

  rules {
    name             = "Allow SSH"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 22
    port_range_end   = 22
    source_ips       = ["0.0.0.0/0"]
    priority         = 100
  }

  rules {
    name             = "Allow HTTP"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 80
    port_range_end   = 80
    source_ips       = ["0.0.0.0/0"]
    priority         = 200
  }

  rules {
    name             = "Allow HTTPS"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 443
    port_range_end   = 443
    source_ips       = ["0.0.0.0/0"]
    priority         = 300
  }
}
```

### Firewall with IP Restrictions

```hcl
resource "danubedata_firewall" "admin" {
  name           = "admin-firewall"
  description    = "Restricted admin access"
  default_action = "deny"

  rules {
    name             = "Allow SSH from office"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 22
    port_range_end   = 22
    source_ips       = ["203.0.113.0/24", "198.51.100.0/24"]
    priority         = 100
  }

  rules {
    name             = "Allow all outbound"
    action           = "allow"
    direction        = "outbound"
    protocol         = "all"
    source_ips       = ["0.0.0.0/0"]
    priority         = 1000
  }
}
```

## Argument Reference

### Required

* `name` - (Required) Name of the firewall.

### Optional

* `description` - (Optional) Description of the firewall.
* `default_action` - (Optional) Default action for unmatched traffic: `allow` or `deny`. Default: `deny`.
* `is_default` - (Optional) Whether this is the default firewall. Default: `false`.
* `rules` - (Optional) List of firewall rules. See [Rules](#rules) below.

### Rules

Each rule supports:

* `name` - (Optional) Name of the rule.
* `action` - (Required) Action: `allow` or `deny`.
* `direction` - (Required) Direction: `inbound` or `outbound`.
* `protocol` - (Required) Protocol: `tcp`, `udp`, `icmp`, or `all`.
* `port_range_start` - (Optional) Start of port range.
* `port_range_end` - (Optional) End of port range.
* `source_ips` - (Optional) List of source IP addresses/CIDRs.
* `priority` - (Optional) Rule priority (lower = higher priority).

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The firewall ID.
* `status` - Current status of the firewall.
* `created_at` - Creation timestamp.

## Import

Firewalls can be imported using their ID:

```bash
terraform import danubedata_firewall.example fw-abc123
```
