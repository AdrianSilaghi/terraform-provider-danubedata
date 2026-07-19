# danubedata_vps

Manages a VPS (Virtual Private Server) instance.

## Example Usage

### Basic VPS with SSH Key

```hcl
resource "danubedata_ssh_key" "main" {
  name       = "my-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

resource "danubedata_vps" "web" {
  name             = "web-server"
  image            = "ubuntu-24.04"
  datacenter       = "fsn1"
  resource_profile = "nano_shared"
  auth_method      = "ssh_key"
  ssh_key_id       = danubedata_ssh_key.main.id
}

output "web_ip" {
  value = danubedata_vps.web.public_ip
}
```

### VPS on a Dedicated-CPU Profile

vCPU, memory and disk all come from `resource_profile` and cannot be set
individually.

```hcl
resource "danubedata_vps" "app" {
  name                = "app-server"
  image               = "debian-12"
  datacenter          = "fsn1"
  resource_profile    = "small"
  cpu_allocation_type = "dedicated"
  auth_method         = "ssh_key"
  ssh_key_id          = danubedata_ssh_key.main.id
}
```

### VPS with Password Authentication

```hcl
resource "danubedata_vps" "dev" {
  name        = "dev-server"
  image       = "ubuntu-24.04"
  datacenter  = "fsn1"
  auth_method = "password"
  password    = var.server_password # at least 12 characters
}
```

### IPv6-only VPS

```hcl
resource "danubedata_vps" "edge" {
  name          = "edge-node"
  image         = "ubuntu-24.04"
  datacenter    = "fsn1"
  network_stack = "ipv6_only"
  auth_method   = "ssh_key"
  ssh_key_id    = danubedata_ssh_key.main.id
}
```

## Resource Profiles

`resource_profile` selects the plan, and it is the only place vCPU, memory and
disk are set. Use the **slug** — the dashboard and pricing page show a display
name (e.g. "DD Litcov"), which is not a valid value here.

| Shared vCPU     | Dedicated vCPU |
| --------------- | -------------- |
| `pico_shared`   | `nano`         |
| `nano_shared`   | `micro`        |
| `micro_shared`  | `small`        |
| `small_shared`  | `medium`       |
| `medium_shared` | `large`        |
| `large_shared`  | `xlarge`       |

Defaults to `nano_shared`. For current specs and pricing see
<https://danubedata.ro/pricing>. Not every profile is available to every
account — profiles above your account's limit, and plans restricted to
particular teams, are rejected at apply time.

## Argument Reference

### Required

* `name` - Name of the VPS instance. Lowercase alphanumeric and hyphens only,
  starting and ending with an alphanumeric character, 1-255 characters.
  Changing this forces a new resource.
* `image` - Operating system image, e.g. `ubuntu-24.04` or `debian-12`. Use the
  `danubedata_vps_images` data source for the current list. Changing this
  forces a new resource.
* `datacenter` - Datacenter location. Only `fsn1` is accepted. Changing this
  forces a new resource.
* `auth_method` - Authentication method. One of `ssh_key`, `password`. Changing
  this forces a new resource.

### Optional

* `resource_profile` - Plan slug; see [Resource Profiles](#resource-profiles).
  Defaults to `nano_shared`. Changing this resizes the instance in place.
* `cpu_allocation_type` - CPU allocation type. One of `shared`, `dedicated`.
  Defaults to `shared`. Changing this is applied in place.
* `ssh_key_id` - ID of the SSH key to install. Required when `auth_method` is
  `ssh_key`. Changing this forces a new resource.
* `password` - Root password. Required when `auth_method` is `password`, and
  must be at least 12 characters. When `auth_method` is `ssh_key`, leave it
  unset — the API generates a password after provisioning and the provider
  reads it back into state. Changing a value you configured forces a new
  resource. Sensitive.
* `network_stack` - Network stack. One of `ipv4_only`, `ipv6_only`,
  `dual_stack`. Defaults to `dual_stack`. Changing a value you configured
  forces a new resource.
* `custom_cloud_init` - Custom cloud-init configuration script, max 10000
  characters. Changing this forces a new resource.

### Timeouts

* `create` - (Default `30m`) Time to wait for VPS creation.
* `update` - (Default `30m`) Time to wait for VPS updates.
* `delete` - (Default `15m`) Time to wait for VPS deletion.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The VPS instance ID.
* `status` - Current status (`pending`, `provisioning`, `running`, `stopped`,
  `error`).
* `cpu_cores` - vCPU count, derived from `resource_profile`.
* `memory_size_gb` - Memory in GB, derived from `resource_profile`.
* `storage_size_gb` - Disk size in GB, derived from `resource_profile`.
* `public_ip` - Public IPv4 address.
* `private_ip` - Private IP address.
* `ipv6_address` - IPv6 address.
* `monthly_cost` - Estimated monthly cost in euros.
* `monthly_cost_cents` - Estimated monthly cost in cents.
* `created_at` / `updated_at` / `deployed_at` - Timestamps.

~> **Note** `cpu_cores`, `memory_size_gb` and `storage_size_gb` are read-only.
They are derived from `resource_profile` and cannot be set in configuration;
doing so fails at plan time. Resize by changing `resource_profile`.

## Import

VPS instances can be imported using their ID, which is a UUID:

```bash
terraform import danubedata_vps.example 9f8c2d14-3b7a-4e51-9c6d-2a1f8e0b7c33
```

## Notes

- `password` is stored in state, whether you set it or the API generated it.
  Protect your state file accordingly.
- Only `resource_profile` and `cpu_allocation_type` are updated in place. Every
  other argument replaces the instance when changed, which destroys the disk.
- Deleting a VPS also deletes its snapshots. See `danubedata_vps_snapshot`.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.
