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
  * `id` - Unique identifier for the VPS instance.
  * `name` - Name of the VPS instance.
  * `status` - Current status (creating, running, stopped, error).
  * `image` - Operating system image.
  * `datacenter` - Datacenter location.
  * `cpu_allocation_type` - CPU allocation type (shared or dedicated).
  * `cpu_cores` - Number of CPU cores.
  * `memory_size_gb` - Memory size in GB.
  * `storage_size_gb` - Storage size in GB.
  * `public_ip` - Public IPv4 address (if assigned).
  * `private_ip` - Private IP address (if assigned).
  * `ipv6_address` - IPv6 address (if enabled).
  * `monthly_cost` - Estimated monthly cost.
  * `created_at` - Timestamp when the instance was created.
