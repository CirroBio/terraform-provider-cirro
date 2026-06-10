resource "cirro_agent" "example" {
  name          = "On-Prem HPC Cluster"
  agent_role_arn = "arn:aws:iam::123456789012:role/CirroAgentRole"

  tags = {
    cluster = "hpc-01"
    site    = "us-east"
  }

  environment_configuration = {
    QUEUE          = "default"
    MAX_VCPUS      = "256"
  }
}
