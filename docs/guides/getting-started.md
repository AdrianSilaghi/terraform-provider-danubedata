# Getting Started with DanubeData Terraform Provider

This guide walks you through setting up and using the DanubeData Terraform provider to manage your cloud infrastructure.

## Prerequisites

Before you begin, ensure you have:

1. **Terraform** 1.0 or later installed ([Download](https://www.terraform.io/downloads))
2. **A DanubeData account** ([Sign up](https://danubedata.ro/register))
3. **An API token** from your [account settings](https://danubedata.ro/user/api-tokens)

## Step 1: Set Up Authentication

The recommended way to authenticate is using an environment variable:

```bash
export DANUBEDATA_API_TOKEN="your-api-token-here"
```

Add this to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) for persistence.

## Step 2: Create Your First Configuration

Create a new directory for your Terraform project:

```bash
mkdir my-danubedata-project
cd my-danubedata-project
```

Create a file named `main.tf`:

```hcl
terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.1"
    }
  }
}

provider "danubedata" {}

# Create an SSH key for server access
resource "danubedata_ssh_key" "main" {
  name       = "my-terraform-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

# Create a VPS instance
resource "danubedata_vps" "web" {
  name        = "my-first-server"
  image       = "ubuntu-24.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.main.id

  cpu_cores       = 2
  memory_size_gb  = 4
  storage_size_gb = 50
}

# Output useful information
output "server_ip" {
  value       = danubedata_vps.web.public_ip
  description = "Public IP address of the server"
}

output "ssh_command" {
  value       = "ssh root@${danubedata_vps.web.public_ip}"
  description = "SSH command to connect to the server"
}
```

## Step 3: Initialize Terraform

Initialize the working directory:

```bash
terraform init
```

This downloads the DanubeData provider and prepares your project.

## Step 4: Plan Your Infrastructure

Preview the changes Terraform will make:

```bash
terraform plan
```

Review the output to ensure it matches your expectations.

## Step 5: Apply the Configuration

Create your infrastructure:

```bash
terraform apply
```

Type `yes` when prompted to confirm.

## Step 6: Connect to Your Server

After the apply completes, use the output SSH command to connect:

```bash
ssh root@<server_ip>
```

## Step 7: Clean Up (Optional)

When you're done, destroy the resources:

```bash
terraform destroy
```

Type `yes` to confirm deletion.

## Next Steps

Now that you've created your first resource, explore more capabilities:

- **[Web Application Stack](web-application-stack.md)** - Deploy a complete web app with database and cache
- **[CI/CD Integration](ci-cd-integration.md)** - Automate deployments with GitHub Actions
- **[Multi-Environment Setup](multi-environment.md)** - Manage dev, staging, and production

## Troubleshooting

### "Missing API Token" Error

Ensure your API token is set:

```bash
echo $DANUBEDATA_API_TOKEN
```

If empty, set it:

```bash
export DANUBEDATA_API_TOKEN="your-token"
```

### "Resource Not Found" on Import

When importing existing resources, use the resource ID from the DanubeData dashboard:

```bash
terraform import danubedata_vps.web vps-abc123
```

### Timeout Errors

For large resources or slow operations, increase timeouts:

```hcl
resource "danubedata_vps" "large" {
  # ... configuration ...

  timeouts {
    create = "30m"
    update = "30m"
    delete = "20m"
  }
}
```
