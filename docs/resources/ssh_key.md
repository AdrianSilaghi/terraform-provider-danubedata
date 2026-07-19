# danubedata_ssh_key

Manages an SSH key for VPS authentication.

## Example Usage

```hcl
resource "danubedata_ssh_key" "main" {
  name       = "my-laptop"
  public_key = file("~/.ssh/id_ed25519.pub")
}
```

### Using with VPS

```hcl
resource "danubedata_ssh_key" "deploy" {
  name       = "deploy-key"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... deploy@example.com"
}

resource "danubedata_vps" "server" {
  name        = "web-server"
  image       = "ubuntu-24.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.deploy.id
}
```

## Argument Reference

### Required

* `name` - A descriptive name for the SSH key.
* `public_key` - The SSH public key in OpenSSH format, e.g. `ssh-rsa AAAA...`
  or `ssh-ed25519 AAAA...`.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The SSH key ID.
* `fingerprint` - The SHA256 fingerprint of the SSH key.
* `created_at` / `updated_at` - Timestamps.

## Import

SSH keys can be imported using their ID, which is a numeric identifier:

```bash
terraform import danubedata_ssh_key.example 137
```

## Notes

- Only the public key is sent to the API. Keep the matching private key out of
  Terraform configuration and state.
- `ssh_key_id` on `danubedata_vps` is create-only: pointing an existing VPS at
  a different key replaces the instance. To rotate keys on a running VPS,
  change `authorized_keys` on the machine itself.
