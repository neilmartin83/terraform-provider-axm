data "axm_audit_events" "example" {
  start_timestamp = "2026-03-01T00:00:00Z"
  end_timestamp   = "2026-03-02T23:59:59Z"
}

output "example_audit_events" {
  value = data.axm_audit_events.example
}
