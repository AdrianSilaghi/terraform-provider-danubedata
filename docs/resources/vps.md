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
  name        = "web-server"
  image       = "ubuntu-22.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.main.id
}
```

### VPS with Custom Resources

```hcl
resource "danubedata_vps" "app" {
  name              = "app-server"
  image             = "debian-12"
  datacenter        = "fsn1"
  auth_method       = "ssh_key"
  ssh_key_id        = danubedata_ssh_key.main.id

  cpu_allocation_type = "dedicated"
  cpu_cores           = 4
  memory_size_gb      = 8
  storage_size_gb     = 100
}
```

### VPS with Password Authentication

```hcl
resource "danubedata_vps" "dev" {
  name        = "dev-server"
  image       = "ubuntu-22.04"
  datacenter  = "fsn1"
  auth_method = "password"
  password    = var.server_password
}
```

### VPS with Resource Profile

```hcl
resource "danubedata_vps" "standard" {
  name             = "standard-server"
  image            = "ubuntu-22.04"
  datacenter       = "fsn1"
  resource_profile = "vps-medium"
  auth_method      = "ssh_key"
  ssh_key_id       = danubedata_ssh_key.main.id
}
```

## Argument Reference

### Required

* `name` - (Required) Name of the VPS instance.
* `image` - (Required) Operating system image. Use the `danubedata_vps_images` data source to list available images.
* `datacenter` - (Required) Datacenter location (e.g., `fsn1`).
* `auth_method` - (Required) Authentication method: `ssh_key` or `password`.

### Optional

* `ssh_key_id` - (Optional) ID of the SSH key. Required when `auth_method` is `ssh_key`.
* `password` - (Optional, Sensitive) Root password. Required when `auth_method` is `password`.
* `resource_profile` - (Optional) Predefined resource profile (e.g., `vps-small`, `vps-medium`, `vps-large`).
* `cpu_allocation_type` - (Optional) CPU allocation type: `shared` (default) or `dedicated`.
* `cpu_cores` - (Optional) Number of CPU cores.
* `memory_size_gb` - (Optional) Memory size in GB.
* `storage_size_gb` - (Optional) Storage size in GB.
* `network_stack` - (Optional) Network stack: `ipv4`, `ipv6`, or `dual` (default).
* `custom_cloud_init` - (Optional) Custom cloud-init configuration.

### Timeouts

* `create` - (Default `20m`) Time to wait for VPS creation.
* `update` - (Default `20m`) Time to wait for VPS updates.
* `delete` - (Default `20m`) Time to wait for VPS deletion.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The VPS instance ID.
* `status` - Current status of the VPS.
* `public_ip` - Public IPv4 address.
* `private_ip` - Private IP address.
* `ipv6_address` - IPv6 address (if enabled).
* `monthly_cost` - Estimated monthly cost.
* `created_at` - Creation timestamp.
* `deployed_at` - Deployment timestamp.

## Import

VPS instances can be imported using their ID:

```bash
terraform import danubedata_vps.example vps-abc123
```
