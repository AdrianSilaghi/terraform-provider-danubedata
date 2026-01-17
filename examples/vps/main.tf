terraform {
  required_providers {
    danubedata = {
      source = "registry.terraform.io/AdrianSilaghi/danubedata"
    }
  }
}

provider "danubedata" {
  # API token can be set via DANUBEDATA_API_TOKEN environment variable
  # api_token = var.danubedata_token

  # Base URL defaults to https://danubedata.ro/api/v1
  # base_url = "https://danubedata.ro/api/v1"
}

# List available VPS images
data "danubedata_vps_images" "available" {}

# Output available images
output "available_images" {
  description = "List of available VPS image IDs"
  value       = data.danubedata_vps_images.available.images[*].id
}

# Create a VPS instance
resource "danubedata_vps" "example" {
  name             = "terraform-test"
  resource_profile = "nano_shared"
  image            = "ubuntu-24.04"
  datacenter       = "fsn1"
  network_stack    = "dual_stack"
  auth_method      = "password"
  password         = var.root_password

  timeouts {
    create = "15m"
    delete = "10m"
  }
}

# Output VPS details
output "vps_id" {
  description = "VPS instance ID"
  value       = danubedata_vps.example.id
}

output "vps_public_ip" {
  description = "VPS public IPv4 address"
  value       = danubedata_vps.example.public_ip
}

output "vps_ipv6_address" {
  description = "VPS IPv6 address"
  value       = danubedata_vps.example.ipv6_address
}

output "vps_status" {
  description = "VPS current status"
  value       = danubedata_vps.example.status
}

output "vps_monthly_cost" {
  description = "VPS monthly cost in dollars"
  value       = danubedata_vps.example.monthly_cost
}

variable "root_password" {
  description = "Root password for the VPS (at least 12 characters)"
  type        = string
  sensitive   = true
}
