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
  docker_containers = [for c in data.danubedata_serverless_containers.all.containers : c if c.deployment_type == "docker_image"]
  git_containers    = [for c in data.danubedata_serverless_containers.all.containers : c if c.deployment_type == "git_repository"]
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
  * `status` - Current status (`pending`, `building`, `deploying`, `running`, `stopped`, `starting`, `stopping`, `provisioning`, `error`, `degraded`).
  * `deployment_type` - Deployment type (`docker_image`, `git_repository`, `zip_upload`).
  * `image` - Container image reference. Null unless `deployment_type` is `docker_image`.
  * `repository_url` - Git repository URL. Null unless `deployment_type` is `git_repository`.
  * `repository_branch` - Git branch, for `git_repository` deployments.
  * `url` - Public HTTPS URL for the container.
  * `port` - Container port.
  * `min_scale` - Minimum number of instances (0 = scale to zero).
  * `max_scale` - Maximum number of instances.
  * `created_at` - Timestamp when the container was created.
