# danubedata_vps_images

Lists available VPS operating system images.

## Example Usage

```hcl
data "danubedata_vps_images" "all" {}

output "available_images" {
  value = [for img in data.danubedata_vps_images.all.images : img.image]
}
```

### Filter Ubuntu Images

```hcl
data "danubedata_vps_images" "all" {}

locals {
  ubuntu_images = [for img in data.danubedata_vps_images.all.images : img if img.distro == "ubuntu"]
}

output "ubuntu_images" {
  value = [for img in local.ubuntu_images : img.image]
}
```

### Get Latest Ubuntu LTS

```hcl
data "danubedata_vps_images" "all" {}

locals {
  ubuntu_lts = [for img in data.danubedata_vps_images.all.images : img if img.distro == "ubuntu" && can(regex("LTS", img.label))]
  latest_ubuntu = local.ubuntu_lts[length(local.ubuntu_lts) - 1]
}

resource "danubedata_vps" "server" {
  name        = "web-server"
  image       = local.latest_ubuntu.image
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.main.id
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `images` - List of available images. Each image contains:
  * `id` - The image ID.
  * `image` - Image identifier used when creating VPS.
  * `label` - Human-readable label.
  * `description` - Image description.
  * `distro` - Distribution name (e.g., `ubuntu`, `debian`, `almalinux`).
  * `version` - Distribution version.
  * `family` - Image family (if applicable).
  * `default_user` - Default SSH user for this image.
