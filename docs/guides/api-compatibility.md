# API compatibility

This update covers the provider's 15 existing resources and 16 data sources.
It was checked against the published [API schema](https://danubedata.ro/docs/api.json)
on 2026-10-20, danube-cli v1.8.0, and the API serializers and request validation
where generated schema metadata was incomplete.

## Upgrading to 0.4.0

Set the provider constraint to `version = "~> 0.4"`, add the required fields
listed below, and refresh the provider installation:

```shell
terraform init -upgrade
terraform plan
```

Review the plan before applying, particularly changes that replace a resource.
Protect or back up your Terraform state before upgrading.

## Configuration changes

- Every `danubedata_vps` configuration must choose `cpu_platform = "amd"` or
  `cpu_platform = "intel"`. The API requires a choice and has no default.
- Every `danubedata_serverless` configuration must supply its DNS-compatible
  `slug`. Changing a slug replaces the container.
- The serverless create API currently discards Git credentials. The provider
  refuses authenticated `git_auth_type` values and configured `git_credentials`
  on creation; public Git deployments and imports of existing private
  deployments remain available. Credential changes require replacement because
  the update API does not rotate or clear them.
  ZIP archive creation is rejected during planning; use the Danube CLI to
  upload an archive.
- `danubedata_static_site_domain.dns_instructions` is the instruction string
  returned by the API. It is not an object containing DNS record attributes.
- A static site's legacy `plan` setting is deprecated. Static site capacity is
  provided by the project's plan; selecting a per-site plan has no effect.

## Project selection

Use a project-locked API token or set provider `team_id` / environment variable
`DANUBEDATA_TEAM_ID` to pin the intended project. Explicit provider configuration
takes precedence over the environment variable. Account-wide tokens without a
project selection follow the account's current project. SSH keys are user-owned.

## Readback limits

Some API endpoints accept settings that their resource responses omit. The
provider retains these configured values across refresh rather than reporting
them removed. Terraform cannot detect changes made outside Terraform to a field
the API does not return; individual resource pages identify these fields.

The cache and database provider data sources expose compatibility metadata,
rather than a live engine/version catalog. The public API currently exposes
plans but no complete engine-version discovery endpoint. Verify engine versions
against the current product offering before applying a version-pinned configuration.

The API encodes associative PHP arrays as objects when populated and sometimes
as `[]` when empty. The response fixtures use source-confirmed object shapes for
parameters, environment variables and tags, and typed objects for firewall
sources where the generated OpenAPI item schema is underspecified.

## Verification

The local tests include real Terraform plan/apply/refresh/import/destroy cycles
against loopback HTTP servers, with fake credentials and current API response
shapes. Standalone examples are also checked with `terraform validate` against
the built provider.

These checks do not provision live infrastructure. Live acceptance tests still
require an explicitly selected test project and an API token with the necessary
abilities. A local test result does not verify a release, registry installation,
or production provisioning.
