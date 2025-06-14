data "axm_organization_device" "example" {
  id = "GX7N12345XYZ"
}

output "example_device" {
  value = data.axm_organization_device.example
}
