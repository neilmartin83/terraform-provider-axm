data "axm_webhook" "example" {
  id = "webhook-12345"
}

output "webhook" {
  value = data.axm_webhook.example
}
