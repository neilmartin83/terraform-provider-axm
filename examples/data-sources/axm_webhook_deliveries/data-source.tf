data "axm_webhook_deliveries" "example" {
  webhook_id = "webhook-12345"
}

output "webhook_deliveries" {
  value = data.axm_webhook_deliveries.example.deliveries
}
