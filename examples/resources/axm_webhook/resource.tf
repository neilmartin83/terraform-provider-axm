variable "webhook_secret" {
  type      = string
  sensitive = true
}

resource "axm_webhook" "device_changes" {
  name              = "Device change notifications"
  enabled           = true
  event_types       = ["ORG_DEVICE_ATTRIBUTE_CHANGED"]
  secret_wo         = var.webhook_secret
  secret_wo_version = 1
  url               = "https://example.com/webhooks/apple"
}
