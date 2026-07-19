# danubedata_vps_images

Lists available VPS operating system images.

## Example Usage

```hcl
data "danubedata_vps_images" "all" {}

output "available_images" {
  value = [for img in data.danubedata_vps_images.all.images : img.id]
}
```

### Filter Ubuntu Images

```hcl
data "danubedata_vps_images" "all" {}

locals {
  ubuntu_images = [for img in data.danubedata_vps_images.all.images : img if img.distro == "ubuntu"]
}

output "ubuntu_images" {
  value = [for img in local.ubuntu_images : img.id]
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
  image       = local.latest_ubuntu.id
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.main.id
}
```

Prefer `id` over `image` for the VPS `image` argument. Both are accepted, but
`image` is a fully-pinned registry reference that changes whenever the image is
rebuilt, which produces a spurious diff on the next plan.

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `images` - List of available images. Each image contains:
  * `id` - Short image identifier, e.g. `ubuntu-24.04`. This is the stable value to pass to the `danubedata_vps` resource's `image` argument.
  * `image` - Full image reference, e.g. a pinned container registry path. Also accepted by `danubedata_vps`, but it changes on every image rebuild.
  * `label` - Human-readable label.
  * `description` - Image description.
  * `distro` - Distribution name (`ubuntu`, `debian`, `alma`, `rocky`, `fedora`, `alpine`).
  * `version` - Distribution version.
  * `family` - OS family (`debian`, `redhat`, `fedora`, `alpine`). Null if the image does not declare one.
  * `default_user` - Default SSH user for this image.
