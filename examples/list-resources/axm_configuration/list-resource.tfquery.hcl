list "axm_configuration" "wifi_configurations" {
    provider = axm

    config {
        # Match Configurations that contain "Wi-Fi" in the name
        name_contains = "Wi-Fi"
    }
}
