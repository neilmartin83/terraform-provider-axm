data "axm_device_management_service_serial_numbers" "example" {
  server_id = "12345678ABCD9012EFGH5678IJKL9012"
}

output "device_serial_numbers" {
  value = data.axm_device_management_service_serial_numbers.example
}
