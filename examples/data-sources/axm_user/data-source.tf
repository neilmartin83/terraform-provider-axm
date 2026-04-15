data "axm_user" "example" {
  id = "1234567890"
}

output "example_user" {
  value = data.axm_user.example
}
