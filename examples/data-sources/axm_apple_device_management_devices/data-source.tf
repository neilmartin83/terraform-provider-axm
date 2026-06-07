data "axm_apple_device_management_devices" "all" {}

output "all" {
  value = data.axm_apple_device_management_devices.all
}
