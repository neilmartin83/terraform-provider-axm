---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "axm_organization_devices Data Source - terraform-provider-axm"
subcategory: ""
description: |-
  Fetches the list of devices from Apple Business or School Manager.
---

# axm_organization_devices (Data Source)

Fetches the list of devices from Apple Business or School Manager.

## Example Usage

```terraform
data "axm_organization_devices" "all" {}

output "all" {
  value = data.axm_organization_devices.all
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `devices` (Attributes List) List of organization devices. (see [below for nested schema](#nestedatt--devices))
- `id` (String) Identifier of the data source.

<a id="nestedatt--devices"></a>
### Nested Schema for `devices`

Read-Only:

- `added_to_org_date_time` (String) Date and time when device was added to organization.
- `bluetooth_mac_address` (String) Bluetooth MAC address.
- `color` (String) Device color.
- `device_capacity` (String) Device capacity.
- `device_model` (String) Device model.
- `eid` (String) EID number.
- `id` (String) Device identifier.
- `imei` (List of String) IMEI numbers.
- `meid` (List of String) MEID numbers.
- `order_date_time` (String) Order date and time.
- `order_number` (String) Order number.
- `part_number` (String) Part number.
- `product_family` (String) Product family.
- `product_type` (String) Product type.
- `purchase_source_id` (String) Purchase source identifier.
- `purchase_source_type` (String) Purchase source type.
- `serial_number` (String) Device serial number.
- `status` (String) Device status.
- `type` (String) Device type.
- `updated_date_time` (String) Last update date and time.
- `wifi_mac_address` (String) Wi-Fi MAC address.
