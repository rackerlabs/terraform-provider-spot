package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/coreos/go-oidc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/oauth2"

	"github.com/rackerlabs/terraform-provider-spot/internal/provider/provider_spot"
)

var _ provider.Provider = (*spotProvider)(nil)

// SpotProviderData is wrapper over all the dependencies
// needed by Resources and DataSources
type SpotProviderData struct {
	ngpcClient      ngpc.Client
	organizerClient *ngpc.OrganizerClient
}

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

	// Below "ngpcCfg" & "organizerClient" is used create a unauthenticated
	// ngpc client to query the organizer for Auth0 client list.
	ngpcCfg := ngpc.NewConfig(ngpcAPIServer, "", p.Version == "dev")
	organizerClient := ngpc.NewOrganizerClient(ngpcCfg)
	// get the refresh token from the user input
	auth0ClientApps, err := organizerClient.GetAuth0Clients(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get auth0 client apps", err.Error())
		return
	}
	if organizerClient == nil {
		resp.Diagnostics.AddError("Failed to create organizer client", "organizerClient is nil")
		return
	}

	// get the auth0 client list from organizer
	var auth0ClientURL, auth0ClientId string
	for _, auth0Client := range auth0ClientApps {
		if auth0Client.Name == nil || auth0Client.ClientID == nil || auth0Client.Domain == nil {
			continue
		}
		if *auth0Client.Name == Auth0AppName {
			auth0ClientURL = fmt.Sprintf("https://%s/", *auth0Client.Domain)
			if auth0Client.ClientID != nil {
				auth0ClientId = *auth0Client.ClientID
			}
		}
	}
	if auth0ClientId == "" || auth0ClientURL == "" {
		resp.Diagnostics.AddError("Failed to get auth0 client details", "auth0 clientId (or) clientURL is empty")
		return
	}

	// Create an OIDC provider
	oidcProvider, err := oidc.NewProvider(ctx, auth0ClientURL)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create OIDC provider", err.Error())
		return
	}
	// Configure the OAuth2 client
	oauth2Config := &oauth2.Config{
		ClientID: auth0ClientId,
		Endpoint: oidcProvider.Endpoint(),
		Scopes:   []string{"openid", "profile", "email", "offline_access"},
	}

	// use the client address to get the access token
	// and set it in the strRxtSpotToken var
	if !tokenStringVal.IsNull() && !tokenStringVal.IsUnknown() {
		strRxtSpotToken, err = GetAccessToken(ctx, oauth2Config, tokenStringVal.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("error getting the access token", err.Error())
			return
		}
	} else {
		rxtRefreshToken := os.Getenv("RXTSPOT_TOKEN")
		if rxtRefreshToken == "" {
			rxtSpotTokenFile, found := os.LookupEnv("RXTSPOT_TOKEN_FILE")
			if !found {
				resp.Diagnostics.AddError("Missing authentication token", "Set RXTSPOT_TOKEN or RXTSPOT_TOKEN_FILE environment variable")
				return
			}
			tflog.Debug(ctx, "Reading authentication token from file", map[string]any{"rxtSpotTokenFile": rxtSpotTokenFile})
			var err error
			rxtRefreshToken, err = readFileUpToNBytes(rxtSpotTokenFile, 5120)
			if err != nil {
				resp.Diagnostics.AddError("Failed to read authentication token from file", err.Error())
				return
			}
		}

		strRxtSpotToken, err = GetAccessToken(ctx, oauth2Config, rxtRefreshToken)
		if err != nil {
			resp.Diagnostics.AddError("error getting the access token", err.Error())
			return
		}
	}
	// Setting token in environment variable for other workflows like kubeconfig generation
	// TODO: Use SpotProviderData to store all these variables
	err = os.Setenv("RXTSPOT_TOKEN", strRxtSpotToken)
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
	if ngpcClient == nil {
		resp.Diagnostics.AddError("Failed to create ngpc client", "ngpcClient is nil")
		return
	}
	// spotProviderData contains all dependencies needed by Resources and DataSources,
	// including API clients and global provider state
	spotProviderData := &SpotProviderData{
		ngpcClient:      ngpcClient,
		organizerClient: organizerClient,
	}
	resp.ResourceData = spotProviderData
	resp.DataSourceData = spotProviderData
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
		NewOndemandnodepoolDataSource,
	}
}

func (p *spotProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCloudspaceResource,
		NewSpotnodepoolResource,
		NewOndemandnodepoolResource,
	}
}

func GetAccessToken(ctx context.Context, config *oauth2.Config, refreshToken string) (string, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	tokenSource := config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return "", err
	}

	accessToken, ok := newToken.Extra("id_token").(string)
	if !ok {
		return "", fmt.Errorf("id token not found")
	}

	return accessToken, nil
}
