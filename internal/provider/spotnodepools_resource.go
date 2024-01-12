package provider

import (
	"context"
	"fmt"
	"terraform-provider-rxtspot/internal/provider/resource_spotnodepools"
	"time"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
)

var _ resource.Resource = (*spotnodepoolsResource)(nil)
var _ resource.ResourceWithConfigure = (*spotnodepoolsResource)(nil)
var _ resource.ResourceWithImportState = (*spotnodepoolsResource)(nil)

func NewSpotnodepoolsResource() resource.Resource {
	return &spotnodepoolsResource{}
}

type spotnodepoolsResource struct {
	client *ngpc.HTTPClient
}

func (r *spotnodepoolsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spotnodepools"
}

func (r *spotnodepoolsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_spotnodepools.SpotnodepoolsResourceSchema(ctx)
}

func (r *spotnodepoolsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *spotnodepoolsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_spotnodepools.SpotnodepoolsModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	// TODO: OrgName should be read from parent cloudspace resource or provider config
	orgName := data.Organization.ValueString()
	tflog.Debug(ctx, "Getting namespace associated with organization", map[string]any{"name": orgName})
	namespace, err := findNamespaceForOrganization(ctx, r.client, orgName)
	if err != nil {
		resp.Diagnostics.AddError("Failed to find namespace for organization", err.Error())
		return
	}
	tflog.Debug(ctx, "Got namespace associated with organization", map[string]any{"namespace": namespace})

	// Creating spotnodepool with same cloudspace name, they will be linked by cloudspace name
	spotNodePool := &ngpcv1.SpotNodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SpotNodePool",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      uuid.New().String(),
			Namespace: namespace,
		},
		Spec: ngpcv1.SpotNodePoolSpec{
			ServerClass: data.ServerClass.ValueString(),
			Desired:     int(data.DesiredServerCount.ValueInt64()),
			BidPrice:    data.BidPrice.ValueString(),
			CloudSpace:  data.CloudspaceName.ValueString(),
		},
	}
	tflog.Debug(ctx, "Creating spotnodepool", map[string]any{"name": spotNodePool.ObjectMeta.Name})
	tflog.Trace(ctx, "Creating spotnodepool", map[string]any{"req": spotNodePool})
	err = r.client.Create(ctx, spotNodePool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create nodepool", err.Error())
		return
	}
	tflog.Debug(ctx, "Created nodepool", map[string]any{"name": spotNodePool.ObjectMeta.Name})
	data.Id = types.StringValue(getIDFromObjectMeta(spotNodePool.ObjectMeta))
	data.CloudspaceName = types.StringValue(spotNodePool.Spec.CloudSpace)
	data.ServerClass = types.StringValue(spotNodePool.Spec.ServerClass)
	data.DesiredServerCount = types.Int64Value(int64(spotNodePool.Spec.Desired))
	data.BidPrice = types.StringValue(spotNodePool.Spec.BidPrice)
	data.ResourceVersion = types.StringValue(spotNodePool.ObjectMeta.ResourceVersion)
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	// TODO: Use "wait_for_ready" attribute to wait for the spotNodePool to be ready or win bids
	// Ref:  https://github.com/hashicorp/terraform-provider-kubernetes/blob/main/kubernetes/resource_kubernetes_deployment_v1.go#L246

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Debug(ctx, "Updated local state by getting remote api object", map[string]any{"name": spotNodePool.ObjectMeta.Name})
}

func (r *spotnodepoolsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_spotnodepools.SpotnodepoolsModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	name, namespace, err := getNameAndNamespaceFromId(data.Id.ValueString()) // TODO: Handle error
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name and namespace from id", err.Error())
		return
	}
	tflog.Info(ctx, "Getting spotnodepool", map[string]any{"name": name, "namespace": namespace})
	spotNodePool := &ngpcv1.SpotNodePool{}
	err = r.client.Get(ctx, ktypes.NamespacedName{Name: name, Namespace: namespace}, spotNodePool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get spotnodepool", err.Error())
		return
	}
	tflog.Trace(ctx, "Got spotNodePool", map[string]any{"spotNodePoolOnRemote": spotNodePool})
	data.Id = types.StringValue(getIDFromObjectMeta(spotNodePool.ObjectMeta))
	data.CloudspaceName = types.StringValue(spotNodePool.Spec.CloudSpace)
	data.ServerClass = types.StringValue(spotNodePool.Spec.ServerClass)
	data.DesiredServerCount = types.Int64Value(int64(spotNodePool.Spec.Desired))
	data.BidPrice = types.StringValue(spotNodePool.Spec.BidPrice)
	// TODO: Should we update the resource version here? because the resource version
	// can only change if the resource is updated outside of terraform
	data.ResourceVersion = types.StringValue(spotNodePool.ObjectMeta.ResourceVersion)
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *spotnodepoolsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, curData resource_spotnodepools.SpotnodepoolsModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &curData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// TODO: Find the difference between curData and data and update only the changed fields using patch
	// Update API call logic
	name, namespace, err := getNameAndNamespaceFromId(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name and namespace from id", err.Error())
		return
	}
	spotNodePool := &ngpcv1.SpotNodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SpotNodePool",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: curData.ResourceVersion.ValueString(),
		},
		Spec: ngpcv1.SpotNodePoolSpec{
			ServerClass: data.ServerClass.ValueString(),
			Desired:     int(data.DesiredServerCount.ValueInt64()),
			BidPrice:    data.BidPrice.ValueString(),
			CloudSpace:  data.CloudspaceName.ValueString(),
		},
	}
	err = r.client.Update(ctx, spotNodePool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update spotnodepool", err.Error())
		return
	}
	tflog.Debug(ctx, "Updated spotnodepool", map[string]any{"name": spotNodePool.ObjectMeta.Name})
	// r.client.Update(ctx, spotNodePool) updates the object in place, so we can use the same object to update state
	data.Id = types.StringValue(getIDFromObjectMeta(spotNodePool.ObjectMeta))
	data.CloudspaceName = types.StringValue(spotNodePool.Spec.CloudSpace)
	data.ServerClass = types.StringValue(spotNodePool.Spec.ServerClass)
	data.DesiredServerCount = types.Int64Value(int64(spotNodePool.Spec.Desired))
	data.BidPrice = types.StringValue(spotNodePool.Spec.BidPrice)
	data.ResourceVersion = types.StringValue(spotNodePool.ObjectMeta.ResourceVersion)
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *spotnodepoolsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_spotnodepools.SpotnodepoolsModel
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name, namespace, err := getNameAndNamespaceFromId(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name and namespace from id", err.Error())
		return
	}
	tflog.Info(ctx, "Deleting spotnodepool", map[string]any{"name": name, "namespace": namespace})
	err = r.client.Delete(ctx, &ngpcv1.SpotNodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SpotNodePool",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		}})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete spotnodepool", err.Error())
		return
	}
	// Delete API call logic, we dont need to update state on delete
	tflog.Info(ctx, "Deleted spotnodepool", map[string]any{"name": name, "namespace": namespace})
}

func (r *spotnodepoolsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
