# Look up the MDM servers by name to get their IDs.
data "axm_device_management_services" "all" {}

locals {
  jamf_pro_id  = one([for s in data.axm_device_management_services.all.device_management_services : s.id if s.name == "Jamf Pro - Production"])
  jamf_edu_id  = one([for s in data.axm_device_management_services.all.device_management_services : s.id if s.name == "Jamf School"])
}

resource "axm_default_device_assignment" "org" {
  apple_tv         = local.jamf_edu_id
  apple_vision_pro = local.jamf_pro_id
  ipad             = local.jamf_edu_id
  iphone           = local.jamf_pro_id
  ipod             = local.jamf_edu_id
  mac              = local.jamf_pro_id
}
