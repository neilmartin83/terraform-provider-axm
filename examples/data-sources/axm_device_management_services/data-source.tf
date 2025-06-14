data "axm_device_management_services" "all" {}

output "all_mdm_servers" {
  value = data.axm_device_management_services.all
}
