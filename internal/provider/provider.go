package axm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// providerModel is the Terraform schema model (internal)
type providerModel struct {
	TeamID     types.String `tfsdk:"team_id"`
	ClientID   types.String `tfsdk:"client_id"`
	KeyID      types.String `tfsdk:"key_id"`
	PrivateKey types.String `tfsdk:"private_key"`
}

type axmProvider struct{}

func (p *axmProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "axm"
}

func (p *axmProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "Apple Business Manager Team ID (starts with BUSINESSAPI.)",
			},
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "Client ID (same as Team ID for AxM)",
			},
			"key_id": schema.StringAttribute{
				Required:    true,
				Description: "Key ID for the .p8 private key",
			},
			"private_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Contents of the .p8 private key",
			},
		},
	}
}

func (p *axmProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := NewClient(
		config.TeamID.ValueString(),
		config.ClientID.ValueString(),
		config.KeyID.ValueString(),
		config.PrivateKey.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("AXM Client Init Failed", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *axmProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *axmProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New() provider.Provider {
	return &axmProvider{}
}
