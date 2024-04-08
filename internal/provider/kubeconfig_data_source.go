package provider

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"os"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rackerlabs/terraform-provider-spot/internal/provider/datasource_kubeconfig"
	ktypes "k8s.io/apimachinery/pkg/types"
)

//go:embed kubeconfig.yaml.tmpl
var kubeconfigTemplate string

var _ datasource.DataSource = (*kubeconfigDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*kubeconfigDataSource)(nil)

func NewKubeconfigDataSource() datasource.DataSource {
	return &kubeconfigDataSource{}
}

type kubeconfigDataSource struct {
	client ngpc.Client
}

func (d *kubeconfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubeconfig"
}

func (d *kubeconfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_kubeconfig.KubeconfigDataSourceSchema(ctx)
}

func (d *kubeconfigDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ngpc.HTTPClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ngpc.HTTPClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *kubeconfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_kubeconfig.KubeconfigModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	token := os.Getenv("RXTSPOT_TOKEN")
	if token == "" {
		resp.Diagnostics.AddError("Missing authentication token", "Set RXTSPOT_TOKEN environment variable")
		return
	}
	tflog.Debug(ctx, "Computing name, namespace using resource id", map[string]any{"id": data.Id.ValueString()})
	name, namespace, err := getNameAndNamespaceFromId(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name and namespace from id", err.Error())
		return
	}
	tflog.Debug(ctx, "Name, namespace using resource id", map[string]any{"name": name, "namespace": namespace})
	cloudspace := &ngpcv1.CloudSpace{}
	err = d.client.Get(ctx, ktypes.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, cloudspace)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get cloudspace", err.Error())
		return
	}
	auth0ClientApps, err := d.client.Organizer().GetAuth0Clients(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get auth0 client apps", err.Error())
		return
	}
	kubeconfigVars := KubeconfigVars{
		User:                  "ngpc-user",
		Token:                 token,
		Host:                  fmt.Sprintf("https://%s/", cloudspace.Status.APIServerEndpoint),
		ClusterName:           cloudspace.Name,
		InsecureSkipTLSVerify: true, // TODO: false on production
	}
	for _, auth0Client := range auth0ClientApps {
		if auth0Client.Name == nil || auth0Client.ClientID == nil || auth0Client.Domain == nil {
			continue
		}
		if *auth0Client.Name == "NGPC CLI" {
			kubeconfigVars.OidcClientID = *auth0Client.ClientID
			kubeconfigVars.OidcIssuerURL = fmt.Sprintf("https://%s/", *auth0Client.Domain)
		}
	}
	if kubeconfigVars.OidcClientID == "" || kubeconfigVars.OidcIssuerURL == "" {
		resp.Diagnostics.AddError("Failed to get oidc client id or issuer url", "Please check if NGPC CLI client is created in Auth0")
		return
	}
	kubeconfigVars.OrgID = os.Getenv("RXTSPOT_ORG_ID")
	if kubeconfigVars.OrgID == "" {
		resp.Diagnostics.AddError("Missing organization id", "Set RXTSPOT_ORG_ID environment variable")
		return
	}
	orgName, err := FindOrgName(ctx, d.client, token, kubeconfigVars.OrgID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get organization name", err.Error())
		return
	}
	kubeconfigVars.OrgName = orgName

	kubeconfigBlob, err := generateKubeconfig(kubeconfigVars, kubeconfigTemplate)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create kubeconfig", err.Error())
		return
	}
	data.Raw = types.StringValue(kubeconfigBlob)

	tokenKubecfg, diags := datasource_kubeconfig.KubeconfigsValue{
		Cluster:  types.StringValue(kubeconfigVars.ClusterName),
		Exec:     types.ObjectNull(datasource_kubeconfig.ExecValue{}.AttributeTypes(ctx)),
		Host:     types.StringValue(kubeconfigVars.Host),
		Insecure: types.BoolValue(kubeconfigVars.InsecureSkipTLSVerify),
		Name:     types.StringValue(fmt.Sprintf("%s-%s", kubeconfigVars.OrgName, cloudspace.Name)),
		Token:    types.StringValue(token),
		Username: types.StringValue(kubeconfigVars.User),
	}.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	execArgs := []string{
		"oidc-login",
		"get-token",
		fmt.Sprintf("--oidc-issuer-url=%s", kubeconfigVars.OidcIssuerURL),
		fmt.Sprintf("--oidc-client-id=%s", kubeconfigVars.OidcClientID),
		"--oidc-extra-scope=openid",
		"--oidc-extra-scope=profile",
		"--oidc-extra-scope=email",
		fmt.Sprintf("--oidc-auth-request-extra-params=organization=%s", kubeconfigVars.OrgID),
		fmt.Sprintf("--token-cache-dir=~/.kube/cache/oidc-login/%s", kubeconfigVars.OrgID),
	}
	execArgsListVal, diags := types.ListValueFrom(ctx, types.StringType, execArgs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	execObjVal, diags := datasource_kubeconfig.ExecValue{
		ApiVersion: types.StringValue("client.authentication.k8s.io/v1beta1"),
		Command:    types.StringValue("kubectl"),
		Args:       execArgsListVal,
		Env:        types.MapNull(types.StringType),
	}.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	oidcKubecfg, diags := datasource_kubeconfig.KubeconfigsValue{
		Cluster:  types.StringValue(cloudspace.Name),
		Host:     types.StringValue(cloudspace.Status.APIServerEndpoint),
		Insecure: types.BoolValue(kubeconfigVars.InsecureSkipTLSVerify),
		Name:     types.StringValue(fmt.Sprintf("%s-%s-oidc", kubeconfigVars.OrgName, cloudspace.Name)),
		Token:    types.StringNull(),
		Username: types.StringValue("oidc"),
		Exec:     execObjVal,
	}.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kubeCfgVal, diags := types.ListValueFrom(ctx, datasource_kubeconfig.KubeconfigsValue{}.Type(ctx), []basetypes.ObjectValue{tokenKubecfg, oidcKubecfg})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Kubeconfigs = kubeCfgVal
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type KubeconfigVars struct {
	OrgID                 string
	OrgName               string
	User                  string
	Token                 string
	Host                  string
	ClusterName           string
	InsecureSkipTLSVerify bool
	OidcIssuerURL         string
	OidcClientID          string
}

func generateKubeconfig(kubeconfigVars KubeconfigVars, templatedStr string) (string, error) {
	var tpl bytes.Buffer
	t := template.Must(template.New("kubeconfig").Parse(templatedStr))
	err := t.Execute(&tpl, kubeconfigVars)
	if err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}
	return tpl.String(), nil
}
