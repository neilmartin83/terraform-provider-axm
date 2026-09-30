# Initial creation adopts existing assignments without moving any devices.
# Keep other resources' device_ids unset for these servers.
resource "axm_device_assignment_policy" "example" {
  assignments = {
    EXAMPLESERIAL1 = "11111111111111111111111111111111"
    EXAMPLESERIAL2 = "UNASSIGNED"
  }
  authoritative_server_ids = ["11111111111111111111111111111111"]
  fallback_server_id       = "22222222222222222222222222222222"
  server_names = {
    "11111111111111111111111111111111" = "Example MDM"
    "22222222222222222222222222222222" = "Default MDM"
  }
  adoption_only = true
  max_changes   = 1
}
