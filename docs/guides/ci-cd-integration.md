# CI/CD Integration with DanubeData Terraform Provider

This guide covers integrating the DanubeData Terraform provider with popular CI/CD platforms for automated infrastructure deployments.

## GitHub Actions

### Basic Workflow

Create `.github/workflows/terraform.yml`:

```yaml
name: Terraform

on:
  push:
    branches: [main]
    paths:
      - 'infrastructure/**'
  pull_request:
    branches: [main]
    paths:
      - 'infrastructure/**'

env:
  TF_VERSION: '1.6.0'
  WORKING_DIR: 'infrastructure'

jobs:
  terraform:
    name: Terraform
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: ${{ env.WORKING_DIR }}

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: ${{ env.TF_VERSION }}

      - name: Terraform Format Check
        run: terraform fmt -check -recursive

      - name: Terraform Init
        run: terraform init
        env:
          DANUBEDATA_API_TOKEN: ${{ secrets.DANUBEDATA_API_TOKEN }}

      - name: Terraform Validate
        run: terraform validate

      - name: Terraform Plan
        id: plan
        run: terraform plan -no-color -out=tfplan
        env:
          DANUBEDATA_API_TOKEN: ${{ secrets.DANUBEDATA_API_TOKEN }}
        continue-on-error: true

      - name: Comment Plan on PR
        if: github.event_name == 'pull_request'
        uses: actions/github-script@v7
        with:
          script: |
            const output = `#### Terraform Plan 📖

            \`\`\`
            ${{ steps.plan.outputs.stdout }}
            \`\`\`

            *Pushed by: @${{ github.actor }}*`;

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: output
            })

      - name: Terraform Apply
        if: github.ref == 'refs/heads/main' && github.event_name == 'push'
        run: terraform apply -auto-approve tfplan
        env:
          DANUBEDATA_API_TOKEN: ${{ secrets.DANUBEDATA_API_TOKEN }}
```

### Multi-Environment Workflow

For managing multiple environments (dev, staging, production):

```yaml
name: Terraform Multi-Environment

on:
  push:
    branches:
      - main
      - develop
  pull_request:
    branches: [main, develop]

jobs:
  determine-environment:
    runs-on: ubuntu-latest
    outputs:
      environment: ${{ steps.set-env.outputs.environment }}
    steps:
      - id: set-env
        run: |
          if [[ "${{ github.ref }}" == "refs/heads/main" ]]; then
            echo "environment=production" >> $GITHUB_OUTPUT
          elif [[ "${{ github.ref }}" == "refs/heads/develop" ]]; then
            echo "environment=staging" >> $GITHUB_OUTPUT
          else
            echo "environment=development" >> $GITHUB_OUTPUT
          fi

  terraform:
    needs: determine-environment
    runs-on: ubuntu-latest
    environment: ${{ needs.determine-environment.outputs.environment }}

    steps:
      - uses: actions/checkout@v4

      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: '1.6.0'

      - name: Terraform Init
        run: |
          terraform init \
            -backend-config="key=${{ needs.determine-environment.outputs.environment }}/terraform.tfstate"
        env:
          DANUBEDATA_API_TOKEN: ${{ secrets.DANUBEDATA_API_TOKEN }}

      - name: Select Workspace
        run: terraform workspace select ${{ needs.determine-environment.outputs.environment }} || terraform workspace new ${{ needs.determine-environment.outputs.environment }}

      - name: Terraform Plan
        run: terraform plan -var-file="environments/${{ needs.determine-environment.outputs.environment }}.tfvars" -out=tfplan
        env:
          DANUBEDATA_API_TOKEN: ${{ secrets.DANUBEDATA_API_TOKEN }}

      - name: Terraform Apply
        if: github.event_name == 'push'
        run: terraform apply -auto-approve tfplan
        env:
          DANUBEDATA_API_TOKEN: ${{ secrets.DANUBEDATA_API_TOKEN }}
```

## GitLab CI

Create `.gitlab-ci.yml`:

```yaml
image:
  name: hashicorp/terraform:1.6
  entrypoint: [""]

variables:
  TF_ROOT: ${CI_PROJECT_DIR}/infrastructure

cache:
  key: terraform-cache
  paths:
    - ${TF_ROOT}/.terraform

stages:
  - validate
  - plan
  - apply

before_script:
  - cd ${TF_ROOT}

validate:
  stage: validate
  script:
    - terraform init -backend=false
    - terraform fmt -check -recursive
    - terraform validate

plan:
  stage: plan
  script:
    - terraform init
    - terraform plan -out=plan.tfplan
  artifacts:
    paths:
      - ${TF_ROOT}/plan.tfplan
    expire_in: 1 week
  rules:
    - if: $CI_PIPELINE_SOURCE == "merge_request_event"
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH

apply:
  stage: apply
  script:
    - terraform init
    - terraform apply -auto-approve plan.tfplan
  dependencies:
    - plan
  rules:
    - if: $CI_COMMIT_BRANCH == $CI_DEFAULT_BRANCH
      when: manual
  environment:
    name: production
```

## Terraform Cloud

### Configuration

Create `backend.tf`:

```hcl
terraform {
  cloud {
    organization = "your-organization"

    workspaces {
      name = "danubedata-production"
    }
  }
}
```

### Workspace Variables

In Terraform Cloud, set these workspace variables:

| Variable | Category | Sensitive |
|----------|----------|-----------|
| `DANUBEDATA_API_TOKEN` | Environment | Yes |

### VCS Integration

1. Connect your GitHub/GitLab repository
2. Configure working directory if not root
3. Enable auto-apply for main branch (optional)

## Jenkins

Create `Jenkinsfile`:

```groovy
pipeline {
    agent {
        docker {
            image 'hashicorp/terraform:1.6'
            args '-u root'
        }
    }

    environment {
        DANUBEDATA_API_TOKEN = credentials('danubedata-api-token')
        TF_IN_AUTOMATION = 'true'
    }

    stages {
        stage('Init') {
            steps {
                dir('infrastructure') {
                    sh 'terraform init'
                }
            }
        }

        stage('Validate') {
            steps {
                dir('infrastructure') {
                    sh 'terraform fmt -check'
                    sh 'terraform validate'
                }
            }
        }

        stage('Plan') {
            steps {
                dir('infrastructure') {
                    sh 'terraform plan -out=tfplan'
                }
            }
        }

        stage('Approve') {
            when {
                branch 'main'
            }
            steps {
                input message: 'Apply Terraform changes?', ok: 'Apply'
            }
        }

        stage('Apply') {
            when {
                branch 'main'
            }
            steps {
                dir('infrastructure') {
                    sh 'terraform apply -auto-approve tfplan'
                }
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
```

## Best Practices

### 1. Use Remote State

Store state remotely for team collaboration:

```hcl
terraform {
  backend "s3" {
    bucket   = "terraform-state"
    key      = "danubedata/terraform.tfstate"
    region   = "us-east-1"
    endpoint = "https://s3.danubedata.ro"

    skip_credentials_validation = true
    skip_metadata_api_check     = true
    skip_requesting_account_id  = true
    force_path_style            = true
  }
}
```

### 2. Lock State

Enable state locking to prevent concurrent modifications:

```hcl
terraform {
  backend "s3" {
    # ... other config ...

    dynamodb_table = "terraform-locks"  # If using DynamoDB for locking
  }
}
```

### 3. Use Workspaces or Directories

Separate environments using workspaces:

```bash
terraform workspace new production
terraform workspace new staging
terraform workspace new development
```

Or separate directories:

```
infrastructure/
├── modules/
│   ├── vps/
│   ├── database/
│   └── cache/
├── environments/
│   ├── production/
│   ├── staging/
│   └── development/
└── shared/
```

### 4. Secure Secrets

- Never commit API tokens to version control
- Use CI/CD secret management (GitHub Secrets, GitLab CI Variables, etc.)
- Rotate tokens periodically
- Use least-privilege tokens when possible

### 5. Plan Before Apply

Always run `terraform plan` and review changes before applying:

```yaml
- name: Terraform Plan
  run: terraform plan -out=tfplan

- name: Terraform Apply
  run: terraform apply tfplan  # Uses the saved plan
```

### 6. Use Automated Formatting

Enforce consistent formatting in CI:

```yaml
- name: Format Check
  run: terraform fmt -check -recursive -diff
```

### 7. Validate Configuration

Always validate before planning:

```yaml
- name: Validate
  run: terraform validate
```
