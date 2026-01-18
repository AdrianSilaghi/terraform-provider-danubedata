# danubedata_ssh_keys

Lists all SSH keys in your account.

## Example Usage

```hcl
data "danubedata_ssh_keys" "all" {}

output "ssh_key_ids" {
  value = [for key in data.danubedata_ssh_keys.all.keys : key.id]
}

output "ssh_key_names" {
  value = [for key in data.danubedata_ssh_keys.all.keys : key.name]
}
```

### Find Key by Name

```hcl
data "danubedata_ssh_keys" "all" {}

locals {
  deploy_key = [for key in data.danubedata_ssh_keys.all.keys : key if key.name == "deploy-key"][0]
}

resource "danubedata_vps" "server" {
  name        = "web-server"
  image       = "ubuntu-22.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = local.deploy_key.id
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `keys` - List of SSH keys. Each key contains:
  * `id` - The SSH key ID.
  * `name` - Name of the key.
  * `fingerprint` - SSH key fingerprint.
  * `public_key` - The public key content.
  * `created_at` - Creation timestamp.
