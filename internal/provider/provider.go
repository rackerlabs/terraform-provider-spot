package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

	var strRxtSpotToken string
	var tokenStringVal basetypes.StringValue
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("token"), &tokenStringVal)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !tokenStringVal.IsNull() && !tokenStringVal.IsUnknown() {
		strRxtSpotToken = tokenStringVal.ValueString()
	} else {
		strRxtSpotToken = os.Getenv("RXTSPOT_TOKEN")
		if strRxtSpotToken == "" {
			rxtSpotTokenFile, found := os.LookupEnv("RXTSPOT_TOKEN_FILE")
			if !found {
				resp.Diagnostics.AddError("Missing authentication token", "Set RXTSPOT_TOKEN or RXTSPOT_TOKEN_FILE environment variable")
				return
			}
			tflog.Debug(ctx, "Reading authentication token from file", map[string]any{"rxtSpotTokenFile": rxtSpotTokenFile})
			var err error
			strRxtSpotToken, err = readFileUpToNBytes(rxtSpotTokenFile, 5120)
			if err != nil {
				resp.Diagnostics.AddError("Failed to read authentication token from file", err.Error())
				return
			}
		}
	}
	// Setting token in environment variable for other workflows like kubeconfig generation
	err := os.Setenv("RXTSPOT_TOKEN", strRxtSpotToken)
	if err != nil {
		resp.Diagnostics.AddError("Failed to set RXTSPOT_TOKEN in environment variable", err.Error())
		return
	}
	rxtSpotToken := NewRxtSpotToken(strRxtSpotToken)
	if err := rxtSpotToken.Parse(); err != nil {
		resp.Diagnostics.AddError("Failed to parse token", err.Error())
		return
	}

	expired, err := rxtSpotToken.IsExpired()
	if err != nil {
		resp.Diagnostics.AddError("Failed to check if token is expired", err.Error())
		return
	}
	if expired {
		resp.Diagnostics.AddError("Token is expired", "Please use a valid token")
		return
	}

	if !rxtSpotToken.IsEmailVerified() {
		resp.Diagnostics.AddError("Email is not verified", "Please verify your email to use Spot services")
		return
	}

	isValidSignature, err := rxtSpotToken.IsValidSignature()
	if err != nil {
		resp.Diagnostics.AddError("Failed to check if token has valid signature", err.Error())
		return
	}
	if !isValidSignature {
		resp.Diagnostics.AddError("Token has invalid signature", "Please use a valid token")
		return
	}
	orgID, err := rxtSpotToken.GetOrgID()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get org_id from authentication token", err.Error())
		return
	}
	if err = os.Setenv("RXTSPOT_ORG_ID", orgID); err != nil {
		resp.Diagnostics.AddError("Failed to set org_id in environment variable RXTSPOT_ORG_ID", err.Error())
		return
	}
	orgNamespace := findNamespaceFromID(orgID)
	tflog.Debug(ctx, "Setting org_id in environment variable RXTSPOT_ORG_NS", map[string]any{"org_id": orgID, "orgNamespace": orgNamespace})
	if err = os.Setenv("RXTSPOT_ORG_NS", orgNamespace); err != nil {
		resp.Diagnostics.AddError("Failed to set org_id in environment variable RXTSPOT_ORG_NS", err.Error())
		return
	}

	tflog.Info(ctx, "Token verified successfully", map[string]any{"org_id": orgID, "orgNamespace": orgNamespace})
	tflog.Debug(ctx, "Creating ngpc client", map[string]any{"ngpcAPIServer": ngpcAPIServer})
	cfg := ngpc.NewConfig(ngpcAPIServer, strRxtSpotToken, p.Version == "dev")
	ngpcClient, err := ngpc.CreateClientForConfig(cfg)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create ngpc client", err.Error())
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
	return []func() datasource.DataSource{
		NewCloudspaceDataSource,
		NewKubeconfigDataSource,
		NewSpotnodepoolDataSource,
		NewRegionDataSource,
		NewRegionsDataSource,
		NewServerclassDataSource,
		NewServerclassesDataSource,
	}
}

func (p *spotProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCloudspaceResource,
		NewSpotnodepoolResource,
	}
}
