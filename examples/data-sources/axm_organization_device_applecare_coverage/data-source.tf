data "axm_organization_device_applecare_coverage" "example" {
  id = "FAKESERIAL12345"
}

output "applecare_coverage" {
  value = data.axm_organization_device_applecare_coverage.example
}
