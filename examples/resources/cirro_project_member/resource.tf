resource "cirro_project_member" "example" {
  project_id = cirro_project.example.id
  username   = "jsmith"
  role       = "CONTRIBUTOR"
}
