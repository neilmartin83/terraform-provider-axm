data "axm_packages" "example" {}

output "example_packages" {
  value = data.axm_packages.example
}
