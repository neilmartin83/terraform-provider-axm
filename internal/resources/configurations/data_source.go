// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package configurations

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/datasource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/neilmartin83/terraform-provider-axm/internal/client"
	"github.com/neilmartin83/terraform-provider-axm/internal/common"
)

var _ datasource.DataSource = &ConfigurationsDataSource{}

// NewConfigurationsDataSource returns a new data source for all configurations.
func NewConfigurationsDataSource() datasource.DataSource {
	return &ConfigurationsDataSource{}
}

// ConfigurationsDataSource defines the data source implementation.
type ConfigurationsDataSource struct {
	client *client.Client
}

// ConfigurationsDataSourceModel describes the data source data model.
type ConfigurationsDataSourceModel struct {
	ID             types.String       `tfsdk:"id"`
	Timeouts       timeouts.Value     `tfsdk:"timeouts"`
	Configurations []ConfigurationModel `tfsdk:"configurations"`
}

// ConfigurationModel describes a configuration in the list.
type ConfigurationModel struct {
	ID                     types.String   `tfsdk:"id"`
	Type                   types.String   `tfsdk:"type"`
	Name                   types.String   `tfsdk:"name"`
	ConfigurationType      types.String   `tfsdk:"configuration_type"`
	ConfiguredForPlatforms []types.String `tfsdk:"configured_for_platforms"`
	CreatedDateTime        types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime        types.String   `tfsdk:"updated_date_time"`
}

func (d *ConfigurationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_configurations"
}

func (d *ConfigurationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of configurations from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"configurations": schema.ListNestedAttribute{
				Description: "List of configurations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The configuration ID.",
						},
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
						"created_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The created date and time.",
						},
						"updated_date_time": schema.StringAttribute{
							Computed:    true,
							Description: "The updated date and time.",
						},
					},
				},
			},
		},
	}
}

func (d *ConfigurationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_configurations data source") {
		return
	}
	d.client = c
}

func (d *ConfigurationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConfigurationsDataSourceModel

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

	configs, err := d.client.GetConfigurations(readCtx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read configurations", err.Error())
		return
	}

	data.Configurations = make([]ConfigurationModel, 0, len(configs))
	for _, cfg := range configs {
		data.Configurations = append(data.Configurations, ConfigurationModel{
			ID:                     types.StringValue(cfg.ID),
			Type:                   types.StringValue(cfg.Type),
			Name:                   types.StringValue(cfg.Attributes.Name),
			ConfigurationType:      types.StringValue(cfg.Attributes.Type),
			ConfiguredForPlatforms: common.StringsToTypesStrings(cfg.Attributes.ConfiguredForPlatforms),
			CreatedDateTime:        types.StringValue(cfg.Attributes.CreatedDateTime),
			UpdatedDateTime:        types.StringValue(cfg.Attributes.UpdatedDateTime),
		})
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read configurations", map[string]any{
		"configuration_count": len(data.Configurations),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
