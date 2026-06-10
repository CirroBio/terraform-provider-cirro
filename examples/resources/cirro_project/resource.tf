resource "cirro_project" "example" {
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
    # --- Budget ---
    # Spend limit for the period. Must be > 0.
    budget_amount = 500
    # Period the budget resets: MONTHLY, QUARTERLY, or ANNUALLY.
    budget_period = "MONTHLY"

    # --- Storage ---
    # How long (days) datasets are retained before expiry. Default: 7.
    retention_policy_days = 30
    # How long (days) temporary/scratch storage persists. Default: 14.
    temporary_storage_lifetime_days = 14

    # --- Features ---
    # Enable automated S3 backup for project data. Default: false.
    enable_backup = true
    # Enable SFTP access to project storage. Default: false.
    enable_sftp = false

    # --- Integrations (optional) ---
    # IDs of service connections (e.g. notification or data services) to attach.
    service_connections = []

    # --- Encryption (optional) ---
    # Customer-managed KMS key ARN. Leave unset to use the Cirro-managed key.
    # kms_arn = "arn:aws:kms:us-east-1:123456789012:key/mrk-abc123"

    # --- Network (optional) ---
    # VPC to deploy project resources into (BYOA projects only).
    # vpc_id = "vpc-0abc123def456"
  }

  # --- Cloud account (optional) ---
  # Required for BYOA (Bring Your Own Account) projects.
  # Omit for Cirro-hosted projects.
  # account = {
  #   account_type = "BYOA"   # HOSTED or BYOA
  #   account_id   = "123456789012"
  #   account_name = "my-aws-account"
  #   region_name  = "us-east-1"
  # }

  # --- Governance (optional) ---
  # Attach data classification IDs to enforce compliance requirements.
  # classification_ids = [cirro_classification.phi.id]

  # --- Tags (optional) ---
  # Free-text labels shown in the Cirro UI.
  # tags = ["genomics", "smith-lab"]
}
