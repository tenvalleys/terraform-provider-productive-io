# Manage a person in Productive.io
resource "productive_person" "example" {
  first_name = "Jane"
  last_name  = "Smith"
  email      = "jane.smith@example.com"
  title      = "Engineer"
  role_id    = 3
}

# A minimal person — only required fields
resource "productive_person" "minimal" {
  first_name = "John"
  last_name  = "Doe"
  email      = "john.doe@example.com"
}
