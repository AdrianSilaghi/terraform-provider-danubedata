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
  image       = "ubuntu-22.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.deploy.id
}
```

## Argument Reference

### Required

* `name` - (Required) Name of the SSH key.
* `public_key` - (Required) The public key content in OpenSSH format.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The SSH key ID.
* `fingerprint` - The SSH key fingerprint.
* `created_at` - Creation timestamp.

## Import

SSH keys can be imported using their ID:

```bash
terraform import danubedata_ssh_key.example key-abc123
```
