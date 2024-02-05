package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/rackerlabs/terraform-provider-spot/internal/provider/provider_spot"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
)

var _ provider.Provider = (*spotProvider)(nil)

// New creates Provider with given version
// Version is not connected to any framework functionality currently, but may be in the future.
// Terraform uses the version from the GH release tag only. Hence value set here doesnt matter.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &spotProvider{
			Version: version,
		}
	}
}

type spotProvider struct {
	Version string
}

func (p *spotProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = provider_spot.SpotProviderSchema(ctx)
}

func (p *spotProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config provider_spot.SpotModel
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
		tflog.Debug(ctx, "Reading authentication token from file", map[string]any{"rxtSpotTokenFile": rxtSpotTokenFile})
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

func (p *spotProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "spot"
	resp.Version = p.Version
}

func (p *spotProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *spotProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCloudspaceResource,
		NewSpotnodepoolsResource,
	}
}
