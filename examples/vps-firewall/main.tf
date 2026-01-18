# VPS with Firewall Example
# This example creates a VPS with a firewall that allows SSH, HTTP, and HTTPS

terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.1"
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
  name           = "web-server-firewall"
  description    = "Allow SSH, HTTP, and HTTPS traffic"
  default_action = "deny"

  # Allow SSH from anywhere
  rules {
    name             = "Allow SSH"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 22
    port_range_end   = 22
    source_ips       = ["0.0.0.0/0"]
    priority         = 100
  }

  # Allow HTTP
  rules {
    name             = "Allow HTTP"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 80
    port_range_end   = 80
    source_ips       = ["0.0.0.0/0"]
    priority         = 200
  }

  # Allow HTTPS
  rules {
    name             = "Allow HTTPS"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 443
    port_range_end   = 443
    source_ips       = ["0.0.0.0/0"]
    priority         = 300
  }

  # Allow all outbound traffic
  rules {
    name       = "Allow all outbound"
    action     = "allow"
    direction  = "outbound"
    protocol   = "all"
    source_ips = ["0.0.0.0/0"]
    priority   = 1000
  }
}

# VPS instance
resource "danubedata_vps" "web" {
  name        = "web-server"
  image       = "ubuntu-22.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.main.id

  cpu_cores       = 2
  memory_size_gb  = 4
  storage_size_gb = 50
}

output "vps_public_ip" {
  value = danubedata_vps.web.public_ip
}

output "firewall_id" {
  value = danubedata_firewall.web.id
}
