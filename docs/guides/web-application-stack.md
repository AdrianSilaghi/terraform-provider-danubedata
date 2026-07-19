# Deploying a Web Application Stack

This guide demonstrates how to deploy a complete web application stack with VPS, database, cache, and object storage using the DanubeData Terraform provider.

## Architecture Overview

```
                    ┌─────────────────┐
                    │   Firewall      │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │   VPS (Web)     │
                    │   Ubuntu 24.04  │
                    └────────┬────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
┌───────▼───────┐   ┌───────▼───────┐   ┌───────▼───────┐
│  PostgreSQL   │   │    Redis      │   │ Object Storage│
│   Database    │   │    Cache      │   │   (Assets)    │
└───────────────┘   └───────────────┘   └───────────────┘
```

## Complete Configuration

Create a file named `main.tf`:

```hcl
terraform {
  required_providers {
    danubedata = {
      source  = "AdrianSilaghi/danubedata"
      version = "~> 0.3"
    }
  }
}

provider "danubedata" {}

# Variables
variable "project_name" {
  description = "Name prefix for all resources"
  default     = "myapp"
}

variable "datacenter" {
  description = "Datacenter location"
  default     = "fsn1"
}

# SSH Key for server access
resource "danubedata_ssh_key" "deploy" {
  name       = "${var.project_name}-deploy-key"
  public_key = file("~/.ssh/id_ed25519.pub")
}

# Firewall for web traffic
resource "danubedata_firewall" "web" {
  name        = "${var.project_name}-web-firewall"
  description = "Firewall for web application"

  # `rules` is a list attribute, not a repeated block. Rules are evaluated in
  # the order they appear here.
  #
  # Each rule also accepts optional `name` and `order` fields. They are left
  # out below because the API does not currently honour `order` (rules are
  # auto-numbered in submission order) and does not echo `name` back, so
  # setting either surfaces "Provider produced inconsistent result after
  # apply". Comments serve the same labelling purpose until that is fixed.
  rules = [
    # SSH access
    {
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port_range_start = 22
      port_range_end   = 22
      source_ips       = ["0.0.0.0/0"]
    },
    # HTTP
    {
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port_range_start = 80
      port_range_end   = 80
      source_ips       = ["0.0.0.0/0"]
    },
    # HTTPS
    {
      action           = "allow"
      direction        = "inbound"
      protocol         = "tcp"
      port_range_start = 443
      port_range_end   = 443
      source_ips       = ["0.0.0.0/0"]
    },
    # Allow all outbound
    {
      action     = "allow"
      direction  = "outbound"
      protocol   = "any"
      source_ips = ["0.0.0.0/0"]
    },
  ]
}

# Web Server VPS
resource "danubedata_vps" "web" {
  name        = "${var.project_name}-web"
  image       = "ubuntu-24.04"
  datacenter  = var.datacenter
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.deploy.id

  # CPU, memory and storage are set by the plan and are read-only attributes
  resource_profile = "small_shared"

  custom_cloud_init = <<-EOF
    #cloud-config
    package_update: true
    packages:
      - nginx
      - certbot
      - python3-certbot-nginx
    runcmd:
      - systemctl enable nginx
      - systemctl start nginx
  EOF
}

# PostgreSQL Database
resource "danubedata_database" "main" {
  name          = "${var.project_name}-db"
  database_name = "${var.project_name}_production" # letters, digits and underscores only
  engine        = "postgresql"
  version       = "16"
  datacenter    = var.datacenter

  resource_profile = "small"
  storage_size_gb  = 20 # optional: grows storage beyond the plan's included amount
}

# Redis Cache
resource "danubedata_cache" "main" {
  name           = "${var.project_name}-cache"
  cache_provider = "redis"
  version        = "8.0"
  datacenter     = var.datacenter

  resource_profile = "micro"
}

# Object Storage for assets
resource "danubedata_storage_bucket" "assets" {
  name               = "${var.project_name}-assets"
  region             = var.datacenter
  versioning_enabled = true
  public_access      = true
}

# Storage Access Key
resource "danubedata_storage_access_key" "app" {
  name = "${var.project_name}-app-key"
}

# Create a snapshot of the web server
resource "danubedata_vps_snapshot" "baseline" {
  name            = "${var.project_name}-baseline"
  description     = "Baseline snapshot after initial setup"
  vps_instance_id = danubedata_vps.web.id
}
```

Create `outputs.tf` for useful output values:

```hcl
output "web_server_ip" {
  value       = danubedata_vps.web.public_ip
  description = "Public IP of the web server"
}

output "ssh_command" {
  value       = "ssh root@${danubedata_vps.web.public_ip}"
  description = "SSH command to connect"
}

output "database_endpoint" {
  value       = danubedata_database.main.endpoint
  description = "Database connection endpoint"
}

output "database_port" {
  value       = danubedata_database.main.port
  description = "Database port"
}

output "database_connection_string" {
  value       = "postgresql://${danubedata_database.main.username}@${danubedata_database.main.endpoint}:${danubedata_database.main.port}/${var.project_name}_production"
  description = "PostgreSQL connection string (add password)"
  sensitive   = true
}

output "cache_endpoint" {
  value       = danubedata_cache.main.endpoint
  description = "Redis cache endpoint"
}

output "cache_port" {
  value       = danubedata_cache.main.port
  description = "Redis port"
}

output "cache_url" {
  value       = "redis://${danubedata_cache.main.endpoint}:${danubedata_cache.main.port}"
  description = "Redis connection URL"
}

output "storage_endpoint" {
  value       = danubedata_storage_bucket.assets.endpoint_url
  description = "S3 endpoint URL"
}

output "storage_bucket" {
  value       = danubedata_storage_bucket.assets.minio_bucket_name
  description = "S3 bucket name"
}

output "storage_access_key_id" {
  value       = danubedata_storage_access_key.app.access_key_id
  description = "S3 access key ID"
}

output "storage_secret_key" {
  value       = danubedata_storage_access_key.app.secret_access_key
  description = "S3 secret access key"
  sensitive   = true
}

output "monthly_cost_estimate" {
  value = format("€%.2f/month",
    danubedata_vps.web.monthly_cost +
    danubedata_database.main.monthly_cost +
    danubedata_cache.main.monthly_cost +
    danubedata_storage_bucket.assets.monthly_cost
  )
  description = "Estimated monthly cost"
}
```

## Deployment

1. **Initialize Terraform:**

```bash
terraform init
```

2. **Preview changes:**

```bash
terraform plan
```

3. **Apply configuration:**

```bash
terraform apply
```

4. **Get connection details:**

```bash
terraform output
```

## Connecting Your Application

### Database Connection (Node.js example)

```javascript
const { Pool } = require('pg');

const pool = new Pool({
  host: process.env.DB_HOST,     // From terraform output database_endpoint
  port: process.env.DB_PORT,     // From terraform output database_port
  database: 'myapp_production',
  user: process.env.DB_USER,     // From terraform output database_username
  password: process.env.DB_PASS, // From DanubeData dashboard
  ssl: { rejectUnauthorized: false }
});
```

### Redis Connection

```javascript
const Redis = require('ioredis');

const redis = new Redis({
  host: process.env.REDIS_HOST,  // From terraform output cache_endpoint
  port: process.env.REDIS_PORT   // From terraform output cache_port
});
```

### S3 Storage Connection

```javascript
const { S3Client, PutObjectCommand } = require('@aws-sdk/client-s3');

const s3 = new S3Client({
  endpoint: process.env.S3_ENDPOINT,        // From terraform output storage_endpoint
  region: 'us-east-1',                       // Required but ignored
  credentials: {
    accessKeyId: process.env.S3_ACCESS_KEY,  // From terraform output storage_access_key_id
    secretAccessKey: process.env.S3_SECRET   // From terraform output storage_secret_key
  },
  forcePathStyle: true
});
```

## Scaling Up

### Upgrade VPS Resources

Resize by moving to a larger plan — CPU, memory and storage cannot be set
individually:

```hcl
resource "danubedata_vps" "web" {
  # ... existing config ...

  resource_profile = "medium_shared" # upgraded from small_shared
}
```

The same applies to the database and cache instances, whose profiles are
`micro`, `small`, `medium` and `large`.

### Add Database Read Replicas

For high-read workloads, attach a replica to the instance:

```hcl
resource "danubedata_database_replica" "read" {
  database_instance_id = danubedata_database.main.id
}
```

The replica's `endpoint`, `replication_status` and `seconds_behind_master` are
exported once it is ready.

### Switch to Dragonfly for Higher Performance

```hcl
resource "danubedata_cache" "main" {
  name           = "${var.project_name}-cache"
  cache_provider = "dragonfly" # Changed from redis
  datacenter     = var.datacenter

  resource_profile = "medium"
}
```

## Cleanup

To destroy all resources:

```bash
terraform destroy
```
