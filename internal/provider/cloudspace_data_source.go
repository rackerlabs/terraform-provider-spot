package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	var err error
	var id, name, namespace string
	id = data.Id.ValueString()
	if strings.Contains(id, "/") {
		tflog.Debug(ctx, "Computing name, namespace using id", map[string]any{"id": id})
		name, namespace, err = getNameAndNamespaceFromId(id)
		if err != nil {
			resp.Diagnostics.AddError("Failed to get name and namespace from id", err.Error())
			return
		}
	} else {
		// In newer approach we dont include org ns in the id because users are not aware of org ns
		name = id
		namespace = os.Getenv("RXTSPOT_ORG_NS")
		if namespace == "" {
			resp.Diagnostics.AddError("Failed to get org namespace", "RXTSPOT_ORG_NS is not set")
			return
		}
		tflog.Debug(ctx, "Using namespace from environment", map[string]any{"namespace": namespace})
	}
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
	data.CloudspaceName = types.StringValue(cloudspace.ObjectMeta.Name)
	data.Name = types.StringValue(cloudspace.ObjectMeta.Name)
	data.ApiServerEndpoint = types.StringValue(cloudspace.Status.APIServerEndpoint)
	data.Health = types.StringValue(cloudspace.Status.Health)
	data.Phase = types.StringValue(string(cloudspace.Status.Phase))
	data.Reason = types.StringValue(cloudspace.Status.Reason)
	data.HacontrolPlane = types.BoolValue(cloudspace.Spec.HAControlPlane)
	data.FirstReadyTimestamp = types.StringValue(cloudspace.Status.FirstReadyTimestamp.Format(time.RFC3339))
	if cloudspace.Spec.Webhook != "" {
		// even if we dont set string value it becomes "" by default
		// assume it as Null if it is not set
		data.PreemptionWebhook = types.StringValue(cloudspace.Spec.Webhook)
	} else {
		data.PreemptionWebhook = types.StringNull()
	}
	var diags diag.Diagnostics
	data.SpotnodepoolIds, diags = types.ListValueFrom(ctx, types.StringType, cloudspace.Spec.BidRequests)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var bidsSlice []datasource_cloudspace.BidsValue
	for _, val := range cloudspace.Status.Bids {
		var wonCount types.Int64
		if val.WonCount != nil {
			wonCount = types.Int64Value(int64(*val.WonCount))
		} else {
			wonCount = types.Int64Null()
		}
		bidObjVal, convertDiags := datasource_cloudspace.BidsValue{
			BidName:  types.StringValue(val.BidName),
			WonCount: wonCount,
		}.ToObjectValue(ctx)
		resp.Diagnostics.Append(convertDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		bidObjValuable, convertDiags := datasource_cloudspace.BidsType{}.ValueFromObject(ctx, bidObjVal)
		resp.Diagnostics.Append(convertDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		bidsSlice = append(bidsSlice, bidObjValuable.(datasource_cloudspace.BidsValue))
	}
	data.Bids, diags = types.SetValueFrom(ctx, datasource_cloudspace.BidsValue{}.Type(ctx), bidsSlice)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var allocationsSlice []datasource_cloudspace.PendingAllocationsValue
	for _, val := range cloudspace.Status.PendingAllocations {
		allocObjVal, convertDiags := datasource_cloudspace.PendingAllocationsValue{
			BidName:     types.StringValue(val.BidName),
			ServerClass: types.StringValue(val.ServerClassName),
			Count:       types.Int64Value(int64(val.Count)),
		}.ToObjectValue(ctx)
		resp.Diagnostics.Append(convertDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		allocObjValuable, convertDiags := datasource_cloudspace.PendingAllocationsType{}.ValueFromObject(ctx, allocObjVal)
		resp.Diagnostics.Append(convertDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		allocationsSlice = append(allocationsSlice, allocObjValuable.(datasource_cloudspace.PendingAllocationsValue))
	}
	data.PendingAllocations, diags = types.SetValueFrom(ctx,
		datasource_cloudspace.PendingAllocationsValue{}.Type(ctx), allocationsSlice)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

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
