data "axm_user_group" "example" {
  id = "UG123456"
}

output "example_user_group" {
  value = data.axm_user_group.example
}
