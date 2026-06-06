resource "axm_device_management_service" "example" {
  name = "Jamf Pro - Production"

  server_certificate = {
    name = "PublicKey.pem"
    data = filebase64("${path.module}/PublicKey.pem")
  }

  allow_release = false

  device_ids = [
    "FAKE000ABC123",
    "FAKE111DEF456",
    "FAKE222GHI789",
  ]
}
