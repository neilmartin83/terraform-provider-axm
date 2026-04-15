// Copyright Neil Martin 2026
// SPDX-License-Identifier: MPL-2.0

package packages

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

var _ datasource.DataSource = &PackagesDataSource{}

// NewPackagesDataSource returns a new data source for all packages.
func NewPackagesDataSource() datasource.DataSource {
	return &PackagesDataSource{}
}

// PackagesDataSource defines the data source implementation.
type PackagesDataSource struct {
	client *client.Client
}

// PackagesDataSourceModel describes the data source data model.
type PackagesDataSourceModel struct {
	ID       types.String   `tfsdk:"id"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
	Packages []PackageModel `tfsdk:"packages"`
}

// PackageModel describes a package.
type PackageModel struct {
	ID              types.String   `tfsdk:"id"`
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

func (d *PackagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_packages"
}

func (d *PackagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches the list of packages from Apple Business Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"timeouts": timeouts.Attributes(ctx),
			"packages": schema.ListNestedAttribute{
				Description: "List of packages.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The package ID.",
						},
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
				},
			},
		},
	}
}

func (d *PackagesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	c, diags := common.ConfigureClient(req.ProviderData, "Data Source")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !common.RequireBusinessScope(c, &resp.Diagnostics, "axm_packages data source") {
		return
	}
	d.client = c
}

func (d *PackagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PackagesDataSourceModel

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

	packages, err := d.client.GetPackages(readCtx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read packages", err.Error())
		return
	}

	data.Packages = make([]PackageModel, 0, len(packages))
	for _, pkg := range packages {
		data.Packages = append(data.Packages, PackageModel{
			ID:              types.StringValue(pkg.ID),
			Type:            types.StringValue(pkg.Type),
			Name:            types.StringValue(pkg.Attributes.Name),
			URL:             types.StringValue(pkg.Attributes.URL),
			Hash:            types.StringValue(pkg.Attributes.Hash),
			BundleIDs:       common.StringsToTypesStrings(pkg.Attributes.BundleIDs),
			Description:     types.StringPointerValue(common.StringPointerOrNil(pkg.Attributes.Description)),
			Version:         types.StringPointerValue(common.StringPointerOrNil(pkg.Attributes.Version)),
			CreatedDateTime: types.StringPointerValue(common.StringPointerOrNil(pkg.Attributes.CreatedDateTime)),
			UpdatedDateTime: types.StringPointerValue(common.StringPointerOrNil(pkg.Attributes.UpdatedDateTime)),
		})
	}

	data.ID = types.StringValue(time.Now().UTC().String())

	tflog.Debug(ctx, "Read packages", map[string]any{
		"package_count": len(data.Packages),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
