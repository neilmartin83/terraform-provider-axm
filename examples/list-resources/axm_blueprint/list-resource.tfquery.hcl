list "axm_blueprint" "engineering_blueprints" {
    provider = axm

    config {
        # Match Blueprints that contain "Engineering" in the name
        name_contains = "Engineering"
    }
}
