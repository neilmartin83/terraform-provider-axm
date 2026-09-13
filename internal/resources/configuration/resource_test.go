// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/neilmartin83/terraform-provider-axm/internal/provider"
	"github.com/neilmartin83/terraform-provider-axm/internal/resources/configuration"
)

func TestConfigurationResourceMetadata(t *testing.T) {
	r := configuration.NewConfigurationResource()
	resp := tfresource.MetadataResponse{}
	r.Metadata(context.Background(), tfresource.MetadataRequest{ProviderTypeName: "axm"}, &resp)

	if resp.TypeName != "axm_configuration" {
		t.Errorf("expected TypeName %q, got %q", "axm_configuration", resp.TypeName)
	}
}

func TestConfigurationResourceSchema(t *testing.T) {
	r := configuration.NewConfigurationResource()
	resp := tfresource.SchemaResponse{}
	r.Schema(context.Background(), tfresource.SchemaRequest{}, &resp)

	if resp.Schema.Description == "" {
		t.Error("expected non-empty schema Description")
	}

	tests := []struct {
		name     string
		required bool
		optional bool
		computed bool
	}{
		{"id", false, false, true},
		{"name", true, false, false},
		{"type", false, false, true},
		{"configured_for_platforms", false, true, true},
		{"configuration_profile", false, true, true},
		{"filename", false, true, true},
		{"created_date_time", false, false, true},
		{"updated_date_time", false, false, true},
		{"timeouts", false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attr, ok := resp.Schema.Attributes[tt.name]
			if !ok {
				t.Fatalf("attribute %q not found in schema", tt.name)
			}
			if attr.IsRequired() != tt.required {
				t.Errorf("expected Required=%v, got %v", tt.required, attr.IsRequired())
			}
			if attr.IsOptional() != tt.optional {
				t.Errorf("expected Optional=%v, got %v", tt.optional, attr.IsOptional())
			}
			if attr.IsComputed() != tt.computed {
				t.Errorf("expected Computed=%v, got %v", tt.computed, attr.IsComputed())
			}
		})
	}

	configuredAttr, ok := resp.Schema.Attributes["configured_for_platforms"].(resourceschema.SetAttribute)
	if !ok {
		t.Fatal("configured_for_platforms is not a SetAttribute")
	}
	if configuredAttr.ElementType != types.StringType {
		t.Errorf("expected configured_for_platforms ElementType to be StringType")
	}

	configProfileAttr, ok := resp.Schema.Attributes["configuration_profile"].(resourceschema.StringAttribute)
	if !ok {
		t.Fatal("configuration_profile is not a StringAttribute")
	}
	if configProfileAttr.Sensitive {
		t.Error("expected configuration_profile to not be Sensitive")
	}
}

func TestConfigurationResourceIdentitySchema(t *testing.T) {
	r := configuration.NewConfigurationResource()

	ri, ok := r.(tfresource.ResourceWithIdentity)
	if !ok {
		t.Fatal("resource does not implement ResourceWithIdentity")
	}

	resp := tfresource.IdentitySchemaResponse{}
	ri.IdentitySchema(context.Background(), tfresource.IdentitySchemaRequest{}, &resp)

	idAttr, ok := resp.IdentitySchema.Attributes["id"]
	if !ok {
		t.Fatal("identity schema missing 'id' attribute")
	}

	idIdentityAttr, ok := idAttr.(identityschema.StringAttribute)
	if !ok {
		t.Fatal("identity 'id' attribute is not a StringAttribute")
	}
	if !idIdentityAttr.RequiredForImport {
		t.Error("expected identity 'id' to have RequiredForImport=true")
	}
}

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"axm": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}
	for _, envVar := range []string{"AXM_CLIENT_ID", "AXM_KEY_ID", "AXM_PRIVATE_KEY", "AXM_SCOPE"} {
		if os.Getenv(envVar) == "" {
			t.Skipf("%s must be set for acceptance tests", envVar)
		}
	}
}

// accTestProfileXML returns a valid configuration profile plist containing a
// Wi-Fi payload. Apple's API rejects profiles with an empty PayloadContent
// array (validation: PayloadContent must have at least one member).
func accTestProfileXML(displayName, identifier, uuid string) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>AutoJoin</key>
			<true/>
			<key>CaptiveBypass</key>
			<false/>
			<key>DisableAssociationMACRandomization</key>
			<false/>
			<key>EncryptionType</key>
			<string>Any</string>
			<key>HIDDEN_NETWORK</key>
			<false/>
			<key>IsHotspot</key>
			<false/>
			<key>PayloadDescription</key>
			<string>Configures Wi-Fi settings</string>
			<key>PayloadDisplayName</key>
			<string>Wi-Fi</string>
			<key>PayloadIdentifier</key>
			<string>com.apple.wifi.managed.%s</string>
			<key>PayloadType</key>
			<string>com.apple.wifi.managed</string>
			<key>PayloadUUID</key>
			<string>%s</string>
			<key>PayloadVersion</key>
			<integer>1</integer>
			<key>ProxyType</key>
			<string>None</string>
			<key>SSID_STR</key>
			<string>Test Network</string>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>%s</string>
	<key>PayloadIdentifier</key>
	<string>%s</string>
	<key>PayloadType</key>
	<string>Configuration</string>
	<key>PayloadUUID</key>
	<string>%s</string>
	<key>PayloadVersion</key>
	<integer>1</integer>
</dict>
</plist>`, uuid, uuid, displayName, identifier, uuid)
}

func TestAccConfigurationResource_basic(t *testing.T) {
	testAccPreCheck(t)
	name := "tf-acc-test-config-basic"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "axm_configuration" "test" {
						name                    = %q
						configured_for_platforms = ["PLATFORM_MACOS"]
						configuration_profile    = <<-EOT
%s
EOT
						filename                = "test-configuration.mobileconfig"
					}
				`, name, accTestProfileXML("Test Config", "com.test.profile", "00000000-0000-0000-0000-000000000000")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("axm_configuration.test", "id"),
					resource.TestCheckResourceAttr("axm_configuration.test", "name", name),
					resource.TestCheckResourceAttrSet("axm_configuration.test", "type"),
					resource.TestCheckResourceAttr("axm_configuration.test", "filename", "test-configuration.mobileconfig"),
				),
			},
			{
				Config: fmt.Sprintf(`
					resource "axm_configuration" "test" {
						name                    = %q
						configured_for_platforms = ["PLATFORM_MACOS"]
						configuration_profile    = <<-EOT
%s
EOT
						filename                = "updated-configuration.mobileconfig"
					}
				`, name+"-updated", accTestProfileXML("Updated Config", "com.test.profile.updated", "00000000-0000-0000-0000-000000000001")),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("axm_configuration.test", "name", name+"-updated"),
					resource.TestCheckResourceAttr("axm_configuration.test", "filename", "updated-configuration.mobileconfig"),
				),
			},
		},
	})
}

func TestAccConfigurationResource_import(t *testing.T) {
	testAccPreCheck(t)
	name := "tf-acc-test-config-import"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
					resource "axm_configuration" "test" {
						name                    = %q
						configured_for_platforms = ["PLATFORM_MACOS"]
						configuration_profile    = <<-EOT
%s
EOT
						filename                = "import-configuration.mobileconfig"
					}
				`, name, accTestProfileXML("Import Config", "com.test.profile.import", "00000000-0000-0000-0000-000000000002")),
			},
			{
				ResourceName:            "axm_configuration.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts", "configuration_profile", "filename"},
			},
		},
	})
}
