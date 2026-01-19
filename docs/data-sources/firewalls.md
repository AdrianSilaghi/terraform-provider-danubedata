# danubedata_firewalls

Lists all firewalls in your account.

## Example Usage

```hcl
data "danubedata_firewalls" "all" {}

output "firewall_count" {
  value = length(data.danubedata_firewalls.all.firewalls)
}

output "firewall_names" {
  value = [for fw in data.danubedata_firewalls.all.firewalls : fw.name]
}
```

### Find Firewall by Name

```hcl
data "danubedata_firewalls" "all" {}

locals {
  web_firewall = [for fw in data.danubedata_firewalls.all.firewalls : fw if fw.name == "web-firewall"][0]
}

output "web_firewall_id" {
  value = local.web_firewall.id
}

output "web_firewall_rules" {
  value = local.web_firewall.rules_count
}
```

### Find Default Firewall

```hcl
data "danubedata_firewalls" "all" {}

locals {
  default_firewall = [for fw in data.danubedata_firewalls.all.firewalls : fw if fw.is_default][0]
}

output "default_firewall_id" {
  value = local.default_firewall.id
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `firewalls` - List of firewalls. Each firewall contains:
  * `id` - Unique identifier for the firewall.
  * `name` - Name of the firewall.
  * `description` - Description of the firewall.
  * `status` - Current status.
  * `default_action` - Default action for unmatched traffic (allow or deny).
  * `is_default` - Whether this is the default firewall.
  * `rules_count` - Number of rules in the firewall.
  * `created_at` - Timestamp when the firewall was created.
