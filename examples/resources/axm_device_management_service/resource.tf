resource "axm_device_management_service" "example" {
  name = "Jamf Pro - Production"

  server_certificate = {
    name = "JamfPro.cer"
    data = filebase64("JamfPro.cer")
  }

  enable_mdm_disown = false

  device_ids = [
    "FAKE000ABC123",
    "FAKE111DEF456",
    "FAKE222GHI789",
  ]
}
