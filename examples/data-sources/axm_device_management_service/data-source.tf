data "axm_device_management_service" "example" {
  id = "12345678ABCD9012EFGH5678IJKL9012"
}

output "mdm_server" {
  value = data.axm_device_management_service.example
}
