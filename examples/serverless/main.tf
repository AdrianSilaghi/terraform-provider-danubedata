# Serverless Container Example
# This example creates a serverless container from a Docker image

terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.1"
    }
  }
}

provider "danubedata" {}

# Simple nginx container
resource "danubedata_serverless" "web" {
  name            = "my-web-app"
  deployment_type = "docker"
  image_url       = "nginx:latest"
  port            = 80
  min_instances   = 0  # Scale to zero when idle
  max_instances   = 10

  environment_variables = {
    NGINX_HOST = "localhost"
    NGINX_PORT = "80"
  }

  timeouts {
    create = "10m"
    delete = "5m"
  }
}

output "app_url" {
  description = "Public URL of the serverless container"
  value       = danubedata_serverless.web.url
}

output "app_status" {
  description = "Current status of the container"
  value       = danubedata_serverless.web.status
}
