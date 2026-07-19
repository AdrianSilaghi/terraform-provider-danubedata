# danubedata_vpss

Lists all VPS instances in your account.

## Example Usage

```hcl
data "danubedata_vpss" "all" {}

output "vps_count" {
  value = length(data.danubedata_vpss.all.instances)
}

output "vps_names" {
  value = [for vps in data.danubedata_vpss.all.instances : vps.name]
}
```

### Find VPS by Name

```hcl
data "danubedata_vpss" "all" {}

locals {
  web_server = [for vps in data.danubedata_vpss.all.instances : vps if vps.name == "web-server"][0]
}

output "web_server_ip" {
  value = local.web_server.public_ip
}
```

### Filter Running Instances

```hcl
data "danubedata_vpss" "all" {}

locals {
  running_instances = [for vps in data.danubedata_vpss.all.instances : vps if vps.status == "running"]
}

output "running_count" {
  value = length(local.running_instances)
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `instances` - List of VPS instances. Each instance contains:
  * `id` - Unique identifier (UUID) for the VPS instance.
  * `name` - Name of the VPS instance.
  * `status` - Current status (`pending`, `provisioning`, `starting`, `running`, `stopping`, `stopped`, `rebooting`, `restoring`, `reinstalling`, `error`, `destroying`, `recreating`).
  * `image` - Operating system image.
  * `datacenter` - Datacenter location.
  * `resource_profile` - Resource profile slug (e.g. `nano_shared`, `micro_shared`) selecting CPU, memory and storage.
  * `cpu_allocation_type` - CPU allocation type (`shared` or `dedicated`).
  * `cpu_cores` - Number of CPU cores. Derived from `resource_profile`.
  * `memory_size_gb` - Memory size in GB. Derived from `resource_profile`.
  * `storage_size_gb` - Storage size in GB. Derived from `resource_profile`.
  * `public_ip` - Public IPv4 address. Null if not assigned.
  * `private_ip` - Private IP address. Null if not assigned.
  * `ipv6_address` - IPv6 address. Null unless the instance's network stack includes IPv6.
  * `monthly_cost` - Estimated monthly cost.
  * `created_at` - Timestamp when the instance was created.
