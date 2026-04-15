data "axm_package" "example" {
  id = "pkg-12345"
}

output "example_package" {
  value = data.axm_package.example
}
