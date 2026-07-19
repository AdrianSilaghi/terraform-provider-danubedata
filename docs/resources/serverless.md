# danubedata_serverless

Manages a serverless container with automatic scaling and scale-to-zero support.

## Example Usage

### Docker Image Deployment

`image` is the repository reference **without** a tag; the tag goes in
`image_tag`.

```hcl
resource "danubedata_serverless" "nginx" {
  name            = "my-nginx"
  deployment_type = "docker_image"
  image           = "nginx"
  image_tag       = "1.27"
  port            = 80
  min_scale       = 0
  max_scale       = 10
}

output "app_url" {
  value = danubedata_serverless.nginx.url
}
```

### Git Repository Deployment

```hcl
resource "danubedata_serverless" "app" {
  name              = "my-app"
  deployment_type   = "git_repository"
  repository_url    = "https://github.com/user/my-app"
  repository_branch = "main"
  source_type       = "buildpack"
  port              = 8080
  min_scale         = 1
  max_scale         = 5

  environment_variables = {
    NODE_ENV  = "production"
    LOG_LEVEL = "info"
  }
}
```

### Private Git Repository

```hcl
resource "danubedata_serverless" "private" {
  name            = "internal-api"
  deployment_type = "git_repository"
  repository_url  = "git@github.com:acme/internal-api.git"
  source_type     = "dockerfile"
  git_auth_type   = "ssh_key"
  git_credentials = var.deploy_key
  port            = 3000
}
```

### With Resource Profile

```hcl
resource "danubedata_serverless" "api" {
  name             = "api-server"
  resource_profile = "medium"
  deployment_type  = "docker_image"
  image            = "myregistry/api"
  image_tag        = "v1.0"
  port             = 3000
  min_scale        = 2
  max_scale        = 20

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
  deployment_type = "docker_image"
  image           = "myregistry/webhook"
  port            = 8080
  min_scale       = 0 # Scale to zero when idle
  max_scale       = 100
}
```

## Argument Reference

### Required

* `name` - Name of the serverless container. Changing this forces a new
  resource.
* `deployment_type` - Deployment type. One of:
  - `docker_image` - deploy a pre-built image
  - `git_repository` - build from a Git repository
  - `zip_upload` - build from an uploaded ZIP archive

  Changing this forces a new resource. See [Notes](#notes) regarding
  `zip_upload`.

### Optional

* `resource_profile` - Resource profile: `free`, `small`, `medium`, or `large`.
  Defaults to `small`. For current pricing see <https://danubedata.ro/pricing>.
* `image` - Container image reference without a tag, e.g. `nginx`. Required for
  `docker_image` deployments. Ignored for `git_repository` and `zip_upload` —
  the platform builds the image and sets this itself.
* `image_tag` - Image tag to deploy. Defaults to `latest`.
* `repository_url` - Git repository URL. Required for `git_repository`
  deployments. Can be changed after creation without replacing the container.
* `repository_branch` - Git branch to build and deploy. Defaults to `main`.
  Only applies to `git_repository` deployments.
* `source_type` - How to build the container from source: `dockerfile` or
  `buildpack`. Required for `git_repository` deployments; defaults to
  `dockerfile` for `zip_upload`.
* `git_auth_type` - Git authentication method for private repositories:
  `none`, `ssh_key`, or `access_token`. Defaults to `none`. Only applies to
  `git_repository` deployments.
* `git_credentials` - SSH private key or access token for private repository
  access. Required when `git_auth_type` is `ssh_key` or `access_token`.
  Sensitive, and never returned by the API — see [Notes](#notes).
* `port` - Port the container listens on. Between 1 and 65535. Defaults to
  `8080`.
* `min_scale` - Minimum number of instances, between 0 and 100. Defaults to
  `0` (scale to zero).
* `max_scale` - Maximum number of instances, between 1 and 100. Defaults to
  `10`.
* `environment_variables` - Map of environment variables.

### Timeouts

* `create` - (Default `15m`) Time to wait for container creation.
* `update` - (Default `15m`) Time to wait for container updates.
* `delete` - (Default `10m`) Time to wait for container deletion.

## Attribute Reference

In addition to the arguments above, the following are exported:

* `id` - The container ID.
* `status` - Current status.
* `url` - Public URL of the deployed service.
* `monthly_cost` - Current month's accrued cost so far, in the account's
  billing currency. Serverless is pay-per-use, so this accumulates from actual
  usage rather than estimating a full month.
* `created_at` / `updated_at` - Timestamps.

## Import

Serverless containers can be imported using their ID:

```bash
terraform import danubedata_serverless.example 7b3f5a92-1c4e-4a08-9d76-3e5c8f1b2a60
```

`git_credentials` is never returned by the API, so it is not populated on
import. Set it in configuration to have the provider send it on the next apply.

## Notes

- Only `name` and `deployment_type` force replacement. Everything else,
  including `resource_profile`, `image`, `repository_url` and scaling bounds,
  is updated in place.
- The provider validates `deployment_type` locally, but the per-type
  requirements (for example `image` on `docker_image`, or `repository_url` and
  `source_type` on `git_repository`) are enforced by the API and surface as
  apply-time errors.
- `zip_upload` is accepted by `deployment_type`, but the provider exposes no
  attribute for the archive itself — the ZIP must be supplied out of band. It
  cannot be driven end to end from Terraform alone.
- The provider acts on the API token owner's current team. If you belong to
  multiple teams, confirm the active team before your first apply.

## Scaling Behavior

- **min_scale = 0**: Container scales to zero after idle period (cost-effective)
- **min_scale >= 1**: Always keeps instances running (no cold starts)
- Scales up automatically based on traffic
- Scales down when traffic decreases

## Build Process (Git Deployment)

When using the `git_repository` deployment type:
1. Repository is cloned
2. `source_type` selects Dockerfile or buildpack detection
3. Container image is built
4. Image is deployed to the serverless platform
5. Automatic rebuilds on git push (via webhook)
