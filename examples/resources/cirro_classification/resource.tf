resource "cirro_classification" "example" {
  name        = "PHI"
  description = "Protected Health Information — requires HIPAA controls."

  # Optionally link to governance requirements
  # requirement_ids = ["req-abc123"]
}
