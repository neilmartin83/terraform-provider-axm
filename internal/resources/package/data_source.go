// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package packageinfo

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

var _ datasource.DataSource = &PackageDataSource{}

// NewPackageDataSource returns a new data source for a single package.
func NewPackageDataSource() datasource.DataSource {
	return &PackageDataSource{}
}

// PackageDataSource defines the data source implementation.
type PackageDataSource struct {
	client *client.Client
}

// PackageDataSourceModel describes the data source data model.
type PackageDataSourceModel struct {
	ID              types.String   `tfsdk:"id"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
	Type            types.String   `tfsdk:"type"`
	Name            types.String   `tfsdk:"name"`
	URL             types.String   `tfsdk:"url"`
	Hash            types.String   `tfsdk:"hash"`
	BundleIDs       []types.String `tfsdk:"bundle_ids"`
	Description     types.String   `tfsdk:"description"`
	Version         types.String   `tfsdk:"version"`
	CreatedDateTime types.String   `tfsdk:"created_date_time"`
	UpdatedDateTime types.String   `tfsdk:"updated_date_time"`
}

func (d *PackageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_package"
}

func (d *PackageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a specific package from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The package ID.",
			},
			"timeouts": timeouts.Attributes(ctx),
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The resource type.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The package name.",
			},
			"url": schema.StringAttribute{
				Computed:    true,
				Description: "The package URL.",
			},
			"hash": schema.StringAttribute{
				Computed:    true,
				Description: "The package hash.",
			},
			"bundle_ids": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Bundle IDs in the package.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The package description.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "The package version.",
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

func (d *PackageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_package data source") {
		return
	}
	d.client = c
}

func (d *PackageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PackageDataSourceModel

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

	pkg, err := d.client.GetPackage(readCtx, data.ID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read package", err.Error())
		return
	}

	data.Type = types.StringValue(pkg.Type)
	data.Name = types.StringValue(pkg.Attributes.Name)
	data.URL = types.StringValue(pkg.Attributes.URL)
	data.Hash = types.StringValue(pkg.Attributes.Hash)
	data.BundleIDs = common.StringsToTypesStrings(pkg.Attributes.BundleIDs)
	data.Description = types.StringPointerValue(common.StringPointerOrNil(pkg.Attributes.Description))
	data.Version = types.StringPointerValue(common.StringPointerOrNil(pkg.Attributes.Version))
	data.CreatedDateTime = types.StringPointerValue(common.StringPointerOrNil(pkg.Attributes.CreatedDateTime))
	data.UpdatedDateTime = types.StringPointerValue(common.StringPointerOrNil(pkg.Attributes.UpdatedDateTime))

	tflog.Debug(ctx, "Read package", map[string]any{
		"package_id": data.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
