list "axm_device_management_service" "london_mdm_servers" {
    provider = axm

    config {
        # Only return device management services whose name includes "London"
        name_contains = "London"

        # Limit the search to Apple Business Manager MDM server records
        server_type = "MDM"
    }
}
