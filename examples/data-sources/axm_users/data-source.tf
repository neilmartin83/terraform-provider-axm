data "axm_users" "example" {}

output "example_users" {
  value = data.axm_users.example
}
