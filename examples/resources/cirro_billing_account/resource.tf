resource "cirro_billing_account" "example" {
  name                 = "Smith Lab"
  customer_type        = "INTERNAL"
  billing_method       = "BUDGET_NUMBER"
  primary_budget_number = "BDG-2024-001"
  owner                = "jsmith"
  shared_with          = ["jdoe", "alee"]

  contacts = [
    {
      name         = "Jane Smith"
      organization = "Acme Corp"
      email        = "jane@example.com"
      phone        = "+1-555-0100"
    }
  ]
}
