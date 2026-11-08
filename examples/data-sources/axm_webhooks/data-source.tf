data "axm_webhooks" "all" {}

output "webhooks" {
  value = data.axm_webhooks.all.webhooks
}
