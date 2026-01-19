# danubedata_serverless_containers

Lists all serverless containers in your account.

## Example Usage

```hcl
data "danubedata_serverless_containers" "all" {}

output "container_count" {
  value = length(data.danubedata_serverless_containers.all.containers)
}

output "container_urls" {
  value = {
    for c in data.danubedata_serverless_containers.all.containers : c.name => c.url
  }
}
```

### Find Container by Name

```hcl
data "danubedata_serverless_containers" "all" {}

locals {
  api_container = [for c in data.danubedata_serverless_containers.all.containers : c if c.name == "api-server"][0]
}

output "api_url" {
  value = local.api_container.url
}

output "api_status" {
  value = local.api_container.status
}
```

### Filter by Deployment Type

```hcl
data "danubedata_serverless_containers" "all" {}

locals {
  docker_containers = [for c in data.danubedata_serverless_containers.all.containers : c if c.deployment_type == "docker"]
  git_containers    = [for c in data.danubedata_serverless_containers.all.containers : c if c.deployment_type == "git"]
}

output "docker_count" {
  value = length(local.docker_containers)
}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

* `containers` - List of serverless containers. Each container contains:
  * `id` - Unique identifier for the container.
  * `name` - Name of the container.
  * `status` - Current status (creating, building, running, error).
  * `deployment_type` - Deployment type (docker or git).
  * `image_url` - Docker image URL (for docker deployment).
  * `git_repository` - Git repository URL (for git deployment).
  * `git_branch` - Git branch (for git deployment).
  * `url` - Public HTTPS URL for the container.
  * `port` - Container port.
  * `min_instances` - Minimum number of instances (0 = scale to zero).
  * `max_instances` - Maximum number of instances.
  * `created_at` - Timestamp when the container was created.
