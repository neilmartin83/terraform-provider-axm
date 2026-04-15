resource "axm_configuration" "example" {
  name = "Wi-Fi Configuration"

  configured_for_platforms = [
    "PLATFORM_IOS",
    "PLATFORM_MACOS"
  ]

  configuration_profile = <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
  <dict>
    <key>PayloadContent</key>
    <array>
      <dict>
        <key>PayloadType</key>
        <string>com.apple.wifi.managed</string>
      </dict>
    </array>
  </dict>
</plist>
EOF

  filename = "WiFi.mobileconfig"
}
