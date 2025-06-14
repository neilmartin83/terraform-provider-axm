data "axm_organization_devices" "all" {}

output "all" {
  value = data.axm_organization_devices.all
}
