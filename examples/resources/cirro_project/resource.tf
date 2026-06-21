# Cirro-hosted project (minimal)
resource "cirro_project" "hosted" {
  name               = "My Research Project"
  description        = "Genomics analysis project for the Smith lab."
  billing_account_id = cirro_billing_account.example.id

  # Up to 10 contacts. At least one required.
  contacts = [
    {
      name         = "Jane Smith"
      organization = "Acme Corp"
      email        = "jane@example.com"
      phone        = "+1-555-0100"
    }
  ]

  settings = {
    # --- Budget (required) ---
    # Spend limit for the period. Must be > 0.
    budget_amount = 500
    # Period the budget resets: MONTHLY, QUARTERLY, or ANNUALLY.
    budget_period = "MONTHLY"

    # --- Storage (optional) ---
    # How long (days) datasets are retained before expiry. Default: 7.
    retention_policy_days = 30
    # How long (days) temporary/scratch storage persists. Default: 14.
    temporary_storage_lifetime_days = 14

    # --- Features (optional) ---
    # Enable automated S3 backup for project data. Default: false.
    enable_backup = true
    # Enable SFTP access to project storage. Default: false.
    enable_sftp = false

    # --- Compute quotas (optional) ---
    max_spot_vcpu           = 256
    max_gpu_vcpu            = 0
    max_fpga_vcpu           = 0
    max_workspaces_vcpu     = 96
    max_workspaces_gpu_vcpu = 0
    max_workspaces_per_user = 10
    max_shared_filesystems  = 5
    enable_advanced_gpu_config    = false

    # --- Access controls (optional) ---
    enable_custom_workspace_roles = true
    is_discoverable               = false
    is_shareable                  = false
    is_ai_enabled                 = false
  }

  # --- Governance (optional) ---
  # Attach data classification IDs to enforce compliance requirements.
  # classification_ids = [cirro_classification.phi.id]

  # --- Tags (optional) ---
  # Free-text labels shown in the Cirro UI.
  # tags = ["genomics", "smith-lab"]
}

# BYOA (Bring Your Own Account) project
resource "cirro_project" "byoa" {
  name               = "BYOA Research Project"
  description        = "Project running in our own AWS account."
  billing_account_id = cirro_billing_account.example.id

  # Up to 10 contacts. At least one required.
  contacts = [
    {
      name         = "Jane Smith"
      organization = "Acme Corp"
      email        = "jane@example.com"
      phone        = "+1-555-0100"
    }
  ]

  # --- Cloud account (required for BYOA) ---
  # account_type and account_id cannot be changed after creation.
  account = {
    account_type = "BYOA"
    account_id   = "123456789012"
    account_name = "my-aws-account"
    region_name  = "us-east-1"
  }

  settings = {
    # --- Budget (required) ---
    budget_amount = 500
    budget_period = "ANNUALLY"

    # --- Storage (optional) ---
    retention_policy_days           = 30
    temporary_storage_lifetime_days = 14

    # --- Features (optional) ---
    enable_backup = true
    enable_sftp   = false

    # --- Encryption (optional) ---
    kms_arn = "arn:aws:kms:us-east-1:123456789012:key/mrk-abc123"

    # --- Network (required for BYOA) ---
    vpc_id            = "vpc-0abc123def456"
    batch_subnets     = ["subnet-aaa", "subnet-bbb"]
    workspace_subnets = ["subnet-ccc", "subnet-ddd"]

    # --- Compute quotas (optional) ---
    max_spot_vcpu           = 256
    max_gpu_vcpu            = 0
    max_fpga_vcpu           = 0
    max_workspaces_vcpu     = 96
    max_workspaces_gpu_vcpu = 0
    max_workspaces_per_user = 10
    max_shared_filesystems  = 5
    enable_advanced_gpu_config    = false

    # --- Access controls (optional) ---
    enable_custom_workspace_roles = true
    is_discoverable               = false
    is_shareable                  = false
    is_ai_enabled                 = false
  }

  # --- Governance (optional) ---
  # Attach data classification IDs to enforce compliance requirements.
  # classification_ids = [cirro_classification.phi.id]

  # --- Tags (optional) ---
  # Free-text labels shown in the Cirro UI.
  # tags = ["genomics", "smith-lab"]
}
