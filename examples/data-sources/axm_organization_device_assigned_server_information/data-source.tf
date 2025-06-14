data "axm_organization_device_assigned_server_information" "example" {
  device_id = "GX7N12345XYZ"
}

output "assigned_server" {
  value = data.axm_organization_device_assigned_server_information.example
}
