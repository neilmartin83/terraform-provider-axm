// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package apps

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

var _ datasource.DataSource = &AppsDataSource{}

// NewAppsDataSource returns a new data source for all apps.
func NewAppsDataSource() datasource.DataSource {
	return &AppsDataSource{}
}

// AppsDataSource defines the data source implementation.
type AppsDataSource struct {
	client *client.Client
}

// AppsDataSourceModel describes the data source data model.
type AppsDataSourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
	Apps     []AppModel     `tfsdk:"apps"`
}

// AppModel describes an app.
type AppModel struct {
	ID          types.String   `tfsdk:"id"`
	Type        types.String   `tfsdk:"type"`
	Name        types.String   `tfsdk:"name"`
	BundleID    types.String   `tfsdk:"bundle_id"`
	WebsiteURL  types.String   `tfsdk:"website_url"`
	Version     types.String   `tfsdk:"version"`
	SupportedOS []types.String `tfsdk:"supported_os"`
	IsCustomApp types.Bool     `tfsdk:"is_custom_app"`
	AppStoreURL types.String   `tfsdk:"app_store_url"`
}

func (d *AppsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_apps"
}

func (d *AppsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of apps from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"apps": schema.ListNestedAttribute{
				Description: "List of apps.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The app ID.",
						},
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
				},
			},
		},
	}
}

func (d *AppsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_apps data source") {
		return
	}
	d.client = c
}

func (d *AppsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AppsDataSourceModel

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

	apps, err := d.client.GetApps(readCtx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read apps", err.Error())
		return
	}

	data.Apps = make([]AppModel, 0, len(apps))
	for _, app := range apps {
		data.Apps = append(data.Apps, AppModel{
			ID:          types.StringValue(app.ID),
			Type:        types.StringValue(app.Type),
			Name:        types.StringValue(app.Attributes.Name),
			BundleID:    types.StringValue(app.Attributes.BundleID),
			WebsiteURL:  types.StringPointerValue(common.StringPointerOrNil(app.Attributes.WebsiteURL)),
			Version:     types.StringPointerValue(common.StringPointerOrNil(app.Attributes.Version)),
			SupportedOS: common.StringsToTypesStrings(app.Attributes.SupportedOS),
			IsCustomApp: types.BoolValue(app.Attributes.IsCustomApp),
			AppStoreURL: types.StringPointerValue(common.StringPointerOrNil(app.Attributes.AppStoreURL)),
		})
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read apps", map[string]any{
		"app_count": len(data.Apps),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
