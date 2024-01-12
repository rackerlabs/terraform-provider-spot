package provider

import (
	"context"
	"os"

	"terraform-provider-rxtspot/internal/provider/provider_rxtspot"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
)

var _ provider.Provider = (*rxtSpotProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &rxtSpotProvider{}
	}
}

type rxtSpotProvider struct{}

func (p *rxtSpotProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = provider_rxtspot.RxtspotProviderSchema(ctx)
}

func (p *rxtSpotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config provider_rxtspot.RxtspotModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ngpcAPIServer := os.Getenv("NGPC_APISERVER")
	if ngpcAPIServer == "" {
		ngpcAPIServer = "https://spot.rackspace.com"
	} else {
		tflog.Info(ctx, "Using provided ngpc api server", map[string]any{"ngpcAPIServer": ngpcAPIServer})
	}

	rxtSpotToken := os.Getenv("RXTSPOT_TOKEN")
	if rxtSpotToken == "" {
		rxtSpotTokenFile, found := os.LookupEnv("RXTSPOT_TOKEN_FILE")
		if !found {
			resp.Diagnostics.AddError("Missing authentication token", "Set RXTSPOT_TOKEN or RXTSPOT_TOKEN_FILE environment variable")
			return
		}
		var err error
		rxtSpotToken, err = readFileUpToNBytes(rxtSpotTokenFile, 5120)
		if err != nil {
			resp.Diagnostics.AddError("Failed to read authentication token from file", err.Error())
			return
		}
	}
	cfg := ngpc.NewConfig(ngpcAPIServer, rxtSpotToken, true)
	ngpcClient, err := ngpc.CreateClientForConfig(cfg)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create ngpc client", err.Error())
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.ResourceData = ngpcClient
	resp.DataSourceData = ngpcClient
}

func (p *rxtSpotProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rxtspot"
}

func (p *rxtSpotProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *rxtSpotProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCloudspaceResource,
		NewSpotnodepoolsResource,
	}
}
