package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
)

var _ provider.Provider = (*rackspacespotProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &rackspacespotProvider{}
	}
}

type rackspacespotProvider struct {
	NGPCApiserver types.String `tfsdk:"ngpc_apiserver"`
}

func (p *rackspacespotProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"ngpc_apiserver": schema.StringAttribute{
				Description: "The address of the NGPC API server",
				Required:    true,
			},
		},
	}
}

func (p *rackspacespotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config rackspacespotProvider
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if config.NGPCApiserver.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("ngpc_apiserver"),
			"Unknown ngpc_apiserver",
			"The ngpc_apiserver is unknown",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}
	ngpcApiserver := os.Getenv("NGPC_APISERVER")

	if !config.NGPCApiserver.IsNull() {
		ngpcApiserver = config.NGPCApiserver.ValueString()
	}

	if ngpcApiserver == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("ngpc_apiserver"),
			"Missing ngpc_apiserver",
			"Missing ngpc_apiserver.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	ngpcAuthToken := os.Getenv("NGPC_AUTH_TOKEN")
	if ngpcAuthToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("ngpc_auth_token"),
			"Missing ngpc_auth_token",
			"Missing ngpc_auth_token.",
		)
	}

	cfg := ngpc.NewConfig(ngpcApiserver, ngpcAuthToken, true)
	ngpcClient, err := ngpc.CreateClientForConfig(cfg)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create ngpc client", err.Error())
	}
	resp.ResourceData = ngpcClient
}

func (p *rackspacespotProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rackspacespot"
}

func (p *rackspacespotProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *rackspacespotProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCloudspaceResource,
	}
}
