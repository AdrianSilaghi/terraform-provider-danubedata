# Basic VPS Example
# This example creates a simple VPS instance with SSH key authentication

terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.3"
    }
  }
}

provider "danubedata" {}

# Create an SSH key for authentication
resource "danubedata_ssh_key" "main" {
  name       = "my-ssh-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

# Create a VPS instance
resource "danubedata_vps" "web" {
  name        = "web-server"
  image       = "ubuntu-22.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.main.id

  # Optional: pick a resource profile (defaults to "nano_shared" if omitted).
  # cpu_cores, memory_size_gb, and storage_size_gb are read-only outputs
  # derived from the profile, not inputs.
  # resource_profile = "micro_shared"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

output "vps_id" {
  description = "The VPS instance ID"
  value       = danubedata_vps.web.id
}

output "vps_public_ip" {
  description = "The public IP address of the VPS"
  value       = danubedata_vps.web.public_ip
}

output "ssh_command" {
  description = "SSH command to connect to the VPS"
  value       = "ssh ubuntu@${danubedata_vps.web.public_ip}"
}
