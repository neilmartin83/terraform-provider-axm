// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package app

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

var _ datasource.DataSource = &AppDataSource{}

// NewAppDataSource returns a new data source for a single app.
func NewAppDataSource() datasource.DataSource {
	return &AppDataSource{}
}

// AppDataSource defines the data source implementation.
type AppDataSource struct {
	client *client.Client
}

// AppDataSourceModel describes the data source data model.
type AppDataSourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
	Type        types.String   `tfsdk:"type"`
	Name        types.String   `tfsdk:"name"`
	BundleID    types.String   `tfsdk:"bundle_id"`
	WebsiteURL  types.String   `tfsdk:"website_url"`
	Version     types.String   `tfsdk:"version"`
	SupportedOS []types.String `tfsdk:"supported_os"`
	IsCustomApp types.Bool     `tfsdk:"is_custom_app"`
	AppStoreURL types.String   `tfsdk:"app_store_url"`
}

func (d *AppDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (d *AppDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific app from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The app ID.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The resource type.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The app name.",
			},
			"bundle_id": schema.StringAttribute{
				Computed:    true,
				Description: "The app bundle identifier.",
			},
			"website_url": schema.StringAttribute{
				Computed:    true,
				Description: "The app website URL.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The app version.",
			},
			"supported_os": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Supported operating systems.",
			},
			"is_custom_app": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the app is custom.",
			},
			"app_store_url": schema.StringAttribute{
				Computed:    true,
				Description: "The App Store URL.",
			},
		},
	}
}

func (d *AppDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_app data source") {
		return
	}
	d.client = c
}

func (d *AppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AppDataSourceModel

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

	app, err := d.client.GetApp(readCtx, data.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read app", err.Error())
		return
	}

	data.Type = types.StringValue(app.Type)
	data.Name = types.StringValue(app.Attributes.Name)
	data.BundleID = types.StringValue(app.Attributes.BundleID)
	data.WebsiteURL = types.StringPointerValue(common.StringPointerOrNil(app.Attributes.WebsiteURL))
	data.Version = types.StringPointerValue(common.StringPointerOrNil(app.Attributes.Version))
	data.SupportedOS = common.StringsToTypesStrings(app.Attributes.SupportedOS)
	data.IsCustomApp = types.BoolValue(app.Attributes.IsCustomApp)
	data.AppStoreURL = types.StringPointerValue(common.StringPointerOrNil(app.Attributes.AppStoreURL))

	tflog.Debug(ctx, "Read app", map[string]any{
		"app_id": data.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
