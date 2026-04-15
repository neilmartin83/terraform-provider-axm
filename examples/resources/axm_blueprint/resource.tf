resource "axm_blueprint" "example" {
  name        = "Engineering Onboarding"
  description = "Standard apps for engineering team"

  app_ids = [
    "361309726"
  ]

  configuration_ids = [
    "config-12345"
  ]

  package_ids = [
    "pkg-12345"
  ]

  device_ids = [
    "ABC123DEF456"
  ]

  user_ids = [
    "1234567890"
  ]

  user_group_ids = [
    "e0484524-fff2-4132-ad29-fe7c6258ce53"
  ]
}
