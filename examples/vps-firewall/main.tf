# VPS with Firewall Example
# This example creates a VPS with a firewall that allows SSH, HTTP, and HTTPS

terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.3"
    }
  }
}

provider "danubedata" {}

# SSH key for authentication
resource "danubedata_ssh_key" "main" {
  name       = "web-server-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

# Firewall for web server
resource "danubedata_firewall" "web" {
  name        = "web-server-firewall"
  description = "Allow SSH, HTTP, and HTTPS traffic"

  # `rules` is a list attribute, not a repeated block.
  #
  # NOTE: the API does not yet honour `order` (rules are auto-numbered in the
  # sequence they are submitted) and does not echo `name` back. Setting either
  # currently surfaces "Provider produced inconsistent result after apply".
  # They are shown here because they are the intended contract; omit them if
  # you need a clean apply until the platform fix ships.
  rules = [
    {
      name             = "Allow SSH"
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port_range_start = 22
      port_range_end   = 22
      source_ips       = ["0.0.0.0/0"]
      order            = 100
    },
    {
      name             = "Allow HTTP"
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port_range_start = 80
      port_range_end   = 80
      source_ips       = ["0.0.0.0/0"]
      order            = 200
    },
    {
      name             = "Allow HTTPS"
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port_range_start = 443
      port_range_end   = 443
      source_ips       = ["0.0.0.0/0"]
      order            = 300
    },
    {
      name       = "Allow all outbound"
      action     = "allow"
      direction  = "outbound"
      protocol   = "any"
      source_ips = ["0.0.0.0/0"]
      order      = 1000
    },
  ]
}

# VPS instance
resource "danubedata_vps" "web" {
  name        = "web-server"
  image       = "ubuntu-22.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.main.id

  # CPU, memory and storage are derived from the plan and are read-only.
  resource_profile = "micro_shared"
}

output "vps_public_ip" {
  value = danubedata_vps.web.public_ip
}

output "firewall_id" {
  value = danubedata_firewall.web.id
}
