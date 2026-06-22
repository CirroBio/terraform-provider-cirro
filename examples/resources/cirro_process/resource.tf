# Example 1: NEXTFLOW process with pipeline code and custom settings
resource "cirro_process" "nextflow_pipeline" {
  id          = "my-nextflow-pipeline"
  name        = "My Nextflow Pipeline"
  description = "Runs quality control and alignment on sequencing data."
  executor    = "NEXTFLOW"

  # Projects that can run this process
  linked_project_ids = ["proj-abc123"]
  is_tenant_wide    = false

  # Processes that feed into this one (leave empty if none)
  parent_process_ids = []

  # Processes that can run after this one (leave empty if none)
  child_process_ids = []

  # Location of the Nextflow workflow code on GitHub
  pipeline_code {
    repository_path = "my-org/my-nextflow-workflow"
    version         = "main"
    repository_type = "GITHUB_PUBLIC"
    entry_point     = "main.nf"
  }

  # Location of the Cirro process definition (UI form, file manifest, etc.)
  custom_settings {
    repository      = "my-org/my-nextflow-workflow"
    branch          = "main"
    folder          = ".cirro"
    repository_type = "GITHUB_PUBLIC"
  }

  category          = "Quality Control"
  documentation_url = "https://github.com/my-org/my-nextflow-workflow"
}
