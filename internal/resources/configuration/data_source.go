// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configuration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

var _ datasource.DataSource = &ConfigurationDataSource{}

// NewConfigurationDataSource returns a new data source for a single configuration.
func NewConfigurationDataSource() datasource.DataSource {
	return &ConfigurationDataSource{}
}

// ConfigurationDataSource defines the data source implementation.
type ConfigurationDataSource struct {
	client *client.Client
}

// ConfigurationDataSourceModel describes the data source data model.
type ConfigurationDataSourceModel struct {
	ID                     types.String   `tfsdk:"id"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
	Type                   types.String   `tfsdk:"type"`
	Name                   types.String   `tfsdk:"name"`
	ConfigurationType      types.String   `tfsdk:"configuration_type"`
	ConfiguredForPlatforms []types.String `tfsdk:"configured_for_platforms"`
	ConfigurationProfile   types.String   `tfsdk:"configuration_profile"`
	Filename               types.String   `tfsdk:"filename"`
	CreatedDateTime        types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime        types.String   `tfsdk:"updated_date_time"`
}

func (d *ConfigurationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configuration"
}

func (d *ConfigurationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific configuration from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The configuration ID.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The resource type.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The configuration name.",
			},
			"configuration_type": schema.StringAttribute{
				Computed:    true,
				Description: "The configuration type (e.g. AIR_DROP, AUTHENTICATION_SCREEN_LOCK, CUSTOM_SETTING).",
			},
			"configured_for_platforms": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Platforms the configuration applies to.",
			},
			"configuration_profile": schema.StringAttribute{
				Computed:    true,
				Description: "The configuration profile payload. Only present for CUSTOM_SETTING configurations.",
			},
			"filename": schema.StringAttribute{
				Computed:    true,
				Description: "The configuration profile filename. Only present for CUSTOM_SETTING configurations.",
			},
			"created_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The created date and time.",
			},
			"updated_date_time": schema.StringAttribute{
				Computed:    true,
				Description: "The updated date and time.",
			},
		},
	}
}

func (d *ConfigurationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_configuration data source") {
		return
	}
	d.client = c
}

func (d *ConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConfigurationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readCtx, cancel, timeoutDiags := common.ResolveReadTimeout(ctx, data.Timeouts, common.DefaultReadTimeout)
	resp.Diagnostics.Append(timeoutDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	defer cancel()

	cfg, err := d.client.GetConfiguration(readCtx, data.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read configuration", err.Error())
		return
	}

	data.Type = types.StringValue(cfg.Type)
	data.Name = types.StringValue(cfg.Attributes.Name)
	data.ConfigurationType = types.StringValue(cfg.Attributes.Type)
	data.ConfiguredForPlatforms = common.StringsToTypesStrings(cfg.Attributes.ConfiguredForPlatforms)
	data.CreatedDateTime = types.StringValue(cfg.Attributes.CreatedDateTime)
	data.UpdatedDateTime = types.StringValue(cfg.Attributes.UpdatedDateTime)

	if cfg.Attributes.CustomSettingsValues != nil {
		data.ConfigurationProfile = types.StringValue(cfg.Attributes.CustomSettingsValues.ConfigurationProfile)
		data.Filename = types.StringValue(cfg.Attributes.CustomSettingsValues.Filename)
	}

	tflog.Debug(ctx, "Read configuration", map[string]any{
		"configuration_id": data.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
