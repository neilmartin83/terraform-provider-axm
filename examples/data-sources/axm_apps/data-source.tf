data "axm_apps" "example" {}

output "example_apps" {
  value = data.axm_apps.example
}
