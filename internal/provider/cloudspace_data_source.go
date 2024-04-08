package provider

import (
	"context"
	"fmt"
	"os"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rackerlabs/terraform-provider-spot/internal/provider/datasource_cloudspace"
	ktypes "k8s.io/apimachinery/pkg/types"
)

var _ datasource.DataSource = (*cloudspaceDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*cloudspaceDataSource)(nil)

func NewCloudspaceDataSource() datasource.DataSource {
	return &cloudspaceDataSource{}
}

type cloudspaceDataSource struct {
	client ngpc.Client
}

func (d *cloudspaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudspace"
}

func (d *cloudspaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_cloudspace.CloudspaceDataSourceSchema(ctx)
}

func (d *cloudspaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cloudspaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_cloudspace.CloudspaceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
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

	data.Id = types.StringValue(getIDFromObjectMeta(cloudspace.ObjectMeta))
	data.Region = types.StringValue(cloudspace.Spec.Region)
	data.Name = types.StringValue(cloudspace.ObjectMeta.Name)
	data.ApiServerEndpoint = types.StringValue(cloudspace.Status.APIServerEndpoint)
	data.Health = types.StringValue(cloudspace.Status.Health)
	data.Phase = types.StringValue(string(cloudspace.Status.Phase))
	data.Reason = types.StringValue(cloudspace.Status.Reason)

	token := os.Getenv("RXTSPOT_TOKEN")
	if token == "" {
		resp.Diagnostics.AddError("Missing authentication token", "Set RXTSPOT_TOKEN environment variable")
		return
	}
	kubeconfigVars := KubeconfigVars{
		OrgName:               "rxtspot",
		User:                  "ngpc-user",
		Token:                 token,
		Host:                  fmt.Sprintf("https://%s/", cloudspace.Status.APIServerEndpoint),
		ClusterName:           cloudspace.Name,
		InsecureSkipTLSVerify: true, // TODO: false on production
	}
	data.Token = types.StringValue(kubeconfigVars.Token)
	data.User = types.StringValue(kubeconfigVars.User)
	kubeconfigBlob, err := generateKubeconfig(kubeconfigVars, kubeconfigTemplateTokenBased)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create kubeconfig", err.Error())
		return
	}
	data.Kubeconfig = types.StringValue(kubeconfigBlob)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// TODO: Remove this const because this is deprecated
const kubeconfigTemplateTokenBased = `apiVersion: v1
clusters:
  - cluster:
      insecure-skip-tls-verify: {{.InsecureSkipTLSVerify}}
      server: >-
        {{.Host}}
    name: {{.ClusterName}}
contexts:
  - context:
      cluster: {{.ClusterName}}
      namespace: default
      user: {{.User}}
    name: {{.OrgName}}-{{.ClusterName}}
current-context: {{.OrgName}}-{{.ClusterName}}
kind: Config
preferences: {}
users:
  - name: {{.User}}
    user:
      token: >-
        {{.Token}}
`
