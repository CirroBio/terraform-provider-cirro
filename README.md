# Terraform Provider for Cirro

Manage [Cirro](https://cirro.bio) resources via infrastructure as code.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.7
- [Go](https://golang.org/doc/install) >= 1.22 (to build from source)

## Using the Provider

```hcl
terraform {
  required_providers {
    cirro = {
      source  = "cirro-bio/cirro"
      version = "~> 0.1"
    }
  }
}

provider "cirro" {
  base_url = "https://app.cirro.bio"
}
```

### Authentication

The provider uses OAuth2 client credentials. Set credentials via environment variables (recommended):

```sh
export CIRRO_BASE_URL="https://app.cirro.bio"
export CIRRO_CLIENT_ID="your-client-id"
export CIRRO_CLIENT_SECRET="your-client-secret"
```

The provider exchanges credentials for a Bearer token at `{base_url}/auth/token` and refreshes it automatically before expiry.

---

## Resources

### `cirro_billing_account`

Manages a billing account. Billing accounts are referenced by projects to track spend.

**[→ Example](examples/resources/cirro_billing_account/resource.tf)**

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Billing account name |
| `customer_type` | string | yes | `INTERNAL`, `CONSORTIUM`, or `EXTERNAL` |
| `billing_method` | string | yes | `BUDGET_NUMBER`, `PURCHASE_ORDER`, or `CREDIT` |
| `primary_budget_number` | string | yes | Budget number or reference code |
| `owner` | string | yes | Username of the account owner |
| `shared_with` | list(string) | yes | Usernames the account is shared with |
| `contacts` | list(object) | yes | At least one contact with `name`, `organization`, `email`, `phone` |

---

### `cirro_project`

Manages a Cirro project. Projects are the top-level container for datasets, analyses, and team access.

> **Note:** Cirro does not expose a project deletion API. Destroying this resource removes it from Terraform state but leaves the project in Cirro.

**[→ Example](examples/resources/cirro_project/resource.tf)**

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Project name (3–100 characters) |
| `description` | string | yes | Project description |
| `billing_account_id` | string | yes | ID of the billing account to charge |
| `contacts` | list(object) | yes | 1–10 contacts with `name`, `organization`, `email`, `phone` |
| `settings` | object | yes | See settings sub-arguments below |
| `account` | object | no | Cloud account config (required for BYOA projects) |
| `classification_ids` | list(string) | no | Governance classification IDs to attach |
| `tags` | list(string) | no | Free-text labels shown in the Cirro UI |

Computed attributes: `id`, `status`, `status_message`, `organization`, `created_by`, `created_at`, `updated_at`, `deployed_at`.

**`settings` sub-arguments**

| Argument | Type | Default | Description |
|---|---|---|---|
| `budget_amount` | number | — | Spend limit for the period (must be > 0) |
| `budget_period` | string | — | `MONTHLY`, `QUARTERLY`, or `ANNUALLY` |
| `retention_policy_days` | number | `7` | Days before datasets expire |
| `temporary_storage_lifetime_days` | number | `14` | Days before temporary storage is cleared |
| `enable_backup` | bool | `false` | Enable automated S3 backup |
| `enable_sftp` | bool | `false` | Enable SFTP access to project storage |
| `service_connections` | list(string) | — | Service connection IDs to attach |
| `kms_arn` | string | — | Customer-managed KMS key ARN for encryption |
| `vpc_id` | string | — | VPC ID for BYOA projects (format: `vpc-*`) |
| `batch_subnets` | list(string) | — | Subnet IDs for batch compute (BYOA) |
| `workspace_subnets` | list(string) | — | Subnet IDs for workspaces (BYOA) |
| `max_spot_vcpu` | number | — | Maximum spot vCPU quota |
| `max_fpga_vcpu` | number | — | Maximum FPGA vCPU quota |
| `max_gpu_vcpu` | number | — | Maximum GPU vCPU quota |
| `enable_dragen` | bool | — | Enable DRAGEN compute environment |
| `dragen_ami` | string | — | AMI for DRAGEN instances |
| `max_workspaces_vcpu` | number | — | Maximum vCPU quota for workspaces |
| `max_workspaces_gpu_vcpu` | number | — | Maximum GPU vCPU quota for workspaces |
| `max_workspaces_per_user` | number | — | Maximum concurrent workspaces per user |
| `enable_advanced_gpu_config` | bool | — | Enable advanced GPU configuration |
| `enable_custom_workspace_roles` | bool | — | Enable custom workspace roles |
| `max_shared_filesystems` | number | — | Maximum number of shared filesystems |
| `is_discoverable` | bool | — | Allow other users to discover the project |
| `is_shareable` | bool | — | Allow datasets to be shared outside the project |
| `is_ai_enabled` | bool | — | Enable AI features |

Computed settings: `has_pipelines_enabled`, `has_workspaces_enabled`, `has_shared_filesystems_enabled`.

**`account` sub-arguments**

| Argument | Type | Required | Description |
|---|---|---|---|
| `account_type` | string | yes | `HOSTED` or `BYOA` — **cannot be changed after creation** |
| `account_id` | string | no | AWS account ID (12-digit) — **cannot be changed after creation** |
| `account_name` | string | no | Human-readable account name |
| `region_name` | string | no | AWS region (e.g. `us-east-1`) |

Import: `terraform import cirro_project.example {project_id}`

---

### `cirro_project_member`

Manages a user's role within a project. Destroying this resource sets the user's role to `NONE`, removing their access.

**[→ Example](examples/resources/cirro_project_member/resource.tf)**

| Argument | Type | Required | Description |
|---|---|---|---|
| `project_id` | string | yes | Project ID — forces replacement if changed |
| `username` | string | yes | Cirro username — forces replacement if changed |
| `role` | string | yes | `OWNER`, `ADMIN`, `CONTRIBUTOR`, or `COLLABORATOR` |
| `suppress_notification` | bool | no | Suppress the email sent to the user (default: `false`) |

Import: `terraform import cirro_project_member.example {project_id}/{username}`

---

### `cirro_user`

Invites a user to Cirro and manages their profile.

> **Note:** Cirro does not expose a user deletion API. Destroying this resource removes it from Terraform state but the user account remains in Cirro.

**[→ Example](examples/resources/cirro_user/resource.tf)**

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Full name (3–70 characters) |
| `email` | string | yes | Email address — forces replacement if changed |
| `organization` | string | yes | Organization name (2–40 characters) |
| `phone` | string | no | Phone number |
| `department` | string | no | Department |
| `job_title` | string | no | Job title |
| `global_roles` | list(string) | no | System-wide roles. Allowed values: `administrators`, `sys-admins`, `pipeline-developers`, `app-developers` |

The `username` attribute is computed — it is assigned by Cirro after the invitation is accepted.

Import: `terraform import cirro_user.example {username}`

---

### `cirro_agent`

Registers a Cirro compute agent. The agent software must be installed separately on your compute infrastructure; it will populate the registration fields (`status`, `registration_hostname`, etc.) once it checks in.

**[→ Example](examples/resources/cirro_agent/resource.tf)**

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Display name for the agent |
| `agent_role_arn` | string | yes | ARN of the AWS IAM role the agent assumes |
| `tags` | map(string) | no | Key-value labels shown to users selecting this agent |
| `environment_configuration` | map(string) | no | Environment variables passed to the agent |

Computed attributes: `status`, `registration_hostname`, `registration_os`, `registration_agent_version`, `created_by`, `created_at`, `updated_at`.

Import: `terraform import cirro_agent.example {agent_id}`

---

### `cirro_classification`

Manages a data governance classification. Classifications are applied to projects to signal and enforce compliance requirements.

**[→ Example](examples/resources/cirro_classification/resource.tf)**

| Argument | Type | Required | Description |
|---|---|---|---|
| `name` | string | yes | Classification name (max 100 characters) |
| `description` | string | yes | What this classification means |
| `requirement_ids` | list(string) | no | IDs of governance requirements to attach |

Computed attributes: `created_by`, `created_at`, `updated_at`.

Import: `terraform import cirro_classification.example {classification_id}`

---

### `cirro_process`

Manages a custom Cirro process (pipeline or ingest data type). Processes define the workflow code and UI form that project members run to analyze data.

> **Note:** Destroying this resource archives the process in Cirro (it is not deleted permanently).

**[→ Example](examples/resources/cirro_process/resource.tf)**

| Argument | Type | Required | Description |
|---|---|---|---|
| `id` | string | yes | Unique process ID (4–80 chars, lowercase, numbers, dashes, underscores). **Cannot be changed after creation.** |
| `name` | string | yes | Friendly display name (4–80 characters) |
| `description` | string | yes | What the process does (4–500 characters) |
| `executor` | string | yes | Execution engine: `INGEST`, `NEXTFLOW`, `CROMWELL`, or `OMICS_READY2RUN` |
| `linked_project_ids` | list(string) | yes | IDs of projects that can run this process |
| `parent_process_ids` | list(string) | yes | IDs of processes whose output feeds into this one (empty list if none) |
| `child_process_ids` | list(string) | yes | IDs of processes that can run after this one (empty list if none) |
| `pipeline_code` | object | no | Location of workflow code (required for NEXTFLOW/CROMWELL/OMICS_READY2RUN, not used for INGEST) |
| `custom_settings` | object | no | Location of the Cirro process definition in a GitHub repo |
| `data_type` | string | no | Name of the data type produced by this process |
| `category` | string | no | UI category label (e.g. `Microbial Analysis`) |
| `documentation_url` | string | no | Link to process documentation |
| `file_requirements_message` | string | no | Instructions shown when uploading files (INGEST processes) |
| `is_tenant_wide` | bool | no | Share across the entire tenant (default: `false`) |
| `allow_multiple_sources` | bool | no | Accept multiple dataset sources (default: `false`) |
| `uses_sample_sheet` | bool | no | Use the Cirro-provided sample sheet (default: `false`) |

**`pipeline_code` sub-arguments**

| Argument | Type | Required | Description |
|---|---|---|---|
| `repository_path` | string | yes | GitHub repository containing the workflow code (`org/repo`) |
| `version` | string | yes | Branch, tag, or commit hash |
| `repository_type` | string | yes | `NONE`, `AWS`, `GITHUB_PUBLIC`, or `GITHUB_PRIVATE` |
| `entry_point` | string | yes | Main script to execute (e.g. `main.nf`) |
| `executor_version` | string | no | Version of the executor runtime |

**`custom_settings` sub-arguments**

| Argument | Type | Required | Description |
|---|---|---|---|
| `repository` | string | yes | GitHub repository containing the process definition (`org/repo`) |
| `branch` | string | no | Branch, tag, or commit hash (default: `main`) |
| `folder` | string | no | Folder within the repo (default: `.cirro`) |
| `repository_type` | string | no | `NONE`, `AWS`, `GITHUB_PUBLIC`, or `GITHUB_PRIVATE` |

Computed attributes: `owner`, `is_archived`, `created_at`, `updated_at`, `custom_settings.last_sync`, `custom_settings.sync_status`, `custom_settings.commit_hash`.

Import: `terraform import cirro_process.example {process_id}`

---

## Data Sources

### `data.cirro_project`

Looks up a project by ID.

```hcl
data "cirro_project" "example" {
  id = "your-project-id"
}

output "project_name" {
  value = data.cirro_project.example.name
}
```

### `data.cirro_user`

Looks up a user by username.

```hcl
data "cirro_user" "example" {
  username = "jsmith"
}
```

---

## Local Development

### Build the binary

```sh
make build
```

This produces a `terraform-provider-cirro` binary in the current directory. To cross-compile for a specific platform:

```sh
GOOS=linux GOARCH=amd64 go build -o terraform-provider-cirro_linux_amd64 .
```

### Install into Go bin (alternative)

```sh
go install .
```

Installs the binary to `$GOPATH/bin` (typically `~/go/bin`).

### Configure Terraform to use the local binary

Add a dev override to `~/.terraformrc` (Linux/Mac) or `%APPDATA%\terraform.rc` (Windows):

```hcl
provider_installation {
  dev_overrides {
    "cirro-bio/cirro" = "/path/to/go/bin"
  }
  direct {}
}
```

With the override active, skip `terraform init` and run `terraform plan` / `terraform apply` directly.

### Running tests

```sh
make test       # unit tests
make testacc    # acceptance tests (requires CIRRO_BASE_URL, CIRRO_CLIENT_ID, CIRRO_CLIENT_SECRET)
```

### Generating docs

```sh
make generate
```
