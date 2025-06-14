provider "axm" {
  client_id   = "BUSINESSAPI.abcdef12-3456-4789-abcd-ef1234567890"
  key_id      = "98765432-dcba-4321-9876-543210fedcba"
  private_key = file("/path/to/private_key.pem")
  scope       = "business.api"
}
