# danubedata_serverless

Manages a serverless container with automatic scaling and scale-to-zero support.

## Example Usage

### Docker Image Deployment

```hcl
resource "danubedata_serverless" "nginx" {
  name            = "my-nginx"
  deployment_type = "docker"
  image_url       = "nginx:latest"
  port            = 80
  min_instances   = 0
  max_instances   = 10
}

output "app_url" {
  value = danubedata_serverless.nginx.url
}
```

### Git Repository Deployment

```hcl
resource "danubedata_serverless" "app" {
  name            = "my-app"
  deployment_type = "git"
  git_repository  = "https://github.com/user/my-app"
  git_branch      = "main"
  port            = 8080
  min_instances   = 1
  max_instances   = 5

  environment_variables = {
    NODE_ENV = "production"
    LOG_LEVEL = "info"
  }
}
```

### With Resource Profile

```hcl
resource "danubedata_serverless" "api" {
  name             = "api-server"
  resource_profile = "medium"
  deployment_type  = "docker"
  image_url        = "myregistry/api:v1.0"
  port             = 3000
  min_instances    = 2
  max_instances    = 20

  environment_variables = {
    DATABASE_URL = "postgres://..."
    REDIS_URL    = "redis://..."
  }
}
```

### Scale to Zero Configuration

```hcl
resource "danubedata_serverless" "webhook" {
  name            = "webhook-handler"
  deployment_type = "docker"
  image_url       = "myregistry/webhook:latest"
  port            = 8080
  min_instances   = 0  # Scale to zero when idle
  max_instances   = 100
}
```

## Argument Reference

### Required

* `name` - (Required) Name of the serverless container.
* `deployment_type` - (Required) Deployment type: `docker` or `git`.
* `port` - (Required) Container port to expose.

### Optional

* `resource_profile` - (Optional) Resource profile: `small`, `medium`, or `large`.
* `image_url` - (Optional) Docker image URL. Required for `docker` deployment type.
* `git_repository` - (Optional) Git repository URL. Required for `git` deployment type.
* `git_branch` - (Optional) Git branch to deploy. Default: `main`.
* `min_instances` - (Optional) Minimum number of instances. Default: `0` (scale to zero).
* `max_instances` - (Optional) Maximum number of instances. Default: `10`.
* `environment_variables` - (Optional) Map of environment variables.

### Timeouts

* `create` - (Default `15m`) Time to wait for container creation.
* `update` - (Default `15m`) Time to wait for container updates.
* `delete` - (Default `10m`) Time to wait for container deletion.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The container ID.
* `status` - Current status.
* `url` - Public HTTPS URL for the container.
* `current_month_cost_cents` - Current month's cost in cents.
* `created_at` - Creation timestamp.

## Import

Serverless containers can be imported using their ID:

```bash
terraform import danubedata_serverless.example srv-abc123
```

## Scaling Behavior

- **min_instances = 0**: Container scales to zero after idle period (cost-effective)
- **min_instances >= 1**: Always keeps instances running (no cold starts)
- Scales up automatically based on traffic
- Scales down when traffic decreases

## Build Process (Git Deployment)

When using `git` deployment type:
1. Repository is cloned
2. Buildpack detection or Dockerfile is used
3. Container image is built
4. Image is deployed to serverless platform
5. Automatic rebuilds on git push (via webhook)
