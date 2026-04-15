data "axm_user_groups" "example" {}

output "example_user_groups" {
  value = data.axm_user_groups.example
}
