data "axm_apple_device_management_device" "example" {
  id = "DEE58089D88646E78395FE80BEAAFF94"
}

output "example_device" {
  value = data.axm_apple_device_management_device.example
}
