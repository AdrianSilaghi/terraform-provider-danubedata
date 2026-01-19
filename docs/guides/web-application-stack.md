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
      version = "~> 0.1"
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
  name           = "${var.project_name}-web-firewall"
  description    = "Firewall for web application"
  default_action = "deny"

  # SSH access
  rules {
    name             = "SSH"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 22
    port_range_end   = 22
    source_ips       = ["0.0.0.0/0"]
    priority         = 100
  }

  # HTTP
  rules {
    name             = "HTTP"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 80
    port_range_end   = 80
    source_ips       = ["0.0.0.0/0"]
    priority         = 200
  }

  # HTTPS
  rules {
    name             = "HTTPS"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 443
    port_range_end   = 443
    source_ips       = ["0.0.0.0/0"]
    priority         = 300
  }

  # Allow all outbound
  rules {
    name       = "Outbound"
    action     = "allow"
    direction  = "outbound"
    protocol   = "all"
    source_ips = ["0.0.0.0/0"]
    priority   = 1000
  }
}

# Web Server VPS
resource "danubedata_vps" "web" {
  name        = "${var.project_name}-web"
  image       = "ubuntu-24.04"
  datacenter  = var.datacenter
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.deploy.id

  cpu_cores       = 2
  memory_size_gb  = 4
  storage_size_gb = 50

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
  name            = "${var.project_name}-db"
  database_name   = "${var.project_name}_production"
  engine          = "postgresql"
  version         = "16"
  storage_size_gb = 20
  memory_size_mb  = 2048
  cpu_cores       = 2
  datacenter      = var.datacenter
}

# Redis Cache
resource "danubedata_cache" "main" {
  name           = "${var.project_name}-cache"
  cache_provider = "redis"
  version        = "7.2"
  memory_size_mb = 512
  cpu_cores      = 1
  datacenter     = var.datacenter
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

```hcl
resource "danubedata_vps" "web" {
  # ... existing config ...

  cpu_cores       = 4    # Upgraded from 2
  memory_size_gb  = 8    # Upgraded from 4
  storage_size_gb = 100  # Upgraded from 50
}
```

### Add Database Read Replicas

For high-read workloads, contact DanubeData support to enable read replicas.

### Switch to Dragonfly for Higher Performance

```hcl
resource "danubedata_cache" "main" {
  name           = "${var.project_name}-cache"
  cache_provider = "dragonfly"  # Changed from redis
  memory_size_mb = 2048
  cpu_cores      = 4
  datacenter     = var.datacenter
}
```

## Cleanup

To destroy all resources:

```bash
terraform destroy
```
