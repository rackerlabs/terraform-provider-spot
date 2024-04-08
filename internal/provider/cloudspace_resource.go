package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rackerlabs/terraform-provider-spot/internal/provider/resource_cloudspace"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
)

var _ resource.Resource = (*cloudspaceResource)(nil)
var _ resource.ResourceWithConfigure = (*cloudspaceResource)(nil)
var _ resource.ResourceWithImportState = (*cloudspaceResource)(nil)

// TODO: Implement ResourceWithConfigValidators for region validation
// var _ resource.ResourceWithConfigValidators = (*cloudspaceResource)(nil)

func NewCloudspaceResource() resource.Resource {
	return &cloudspaceResource{}
}

type cloudspaceResource struct {
	client ngpc.Client
}

func (r *cloudspaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudspace"
}

func (r *cloudspaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_cloudspace.CloudspaceResourceSchema(ctx)
}

func (r *cloudspaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cloudspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_cloudspace.CloudspaceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespace := os.Getenv("RXTSPOT_ORG_NS")
	if namespace == "" {
		resp.Diagnostics.AddError("Failed to get org namespace", "RXTSPOT_ORG_NS is not set")
		return
	}
	tflog.Debug(ctx, "Using namespace from environment", map[string]any{"namespace": namespace})

	regionsList := ngpcv1.RegionList{}
	err := r.client.List(ctx, &regionsList)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get regions", err.Error())
		return
	}
	var validRegion bool
	for _, region := range regionsList.Items {
		if region.Name == data.Region.ValueString() {
			validRegion = true
			break
		}
	}
	if !validRegion {
		regionNames := make([]string, len(regionsList.Items))
		for i, region := range regionsList.Items {
			regionNames[i] = region.Name
		}
		resp.Diagnostics.AddAttributeError(path.Root("region"), "Invalid region", fmt.Sprintf("Allowed values are: %v", regionNames))
		return
	}

	cloudspace := &ngpcv1.CloudSpace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudSpace",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.CloudspaceName.ValueString(),
			Namespace: namespace,
		},
		Spec: ngpcv1.CloudSpaceSpec{
			Region:         data.Region.ValueString(),
			Cloud:          "default",
			HAControlPlane: data.HacontrolPlane.ValueBool(),
			Webhook:        data.PreemptionWebhook.ValueString(),
		},
	}
	tflog.Info(ctx, "Creating cloudspace", map[string]any{"name": cloudspace.ObjectMeta.Name})
	err = r.client.Create(ctx, cloudspace)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cloudspace", err.Error())
		return
	}
	tflog.Info(ctx, "Created cloudspace", map[string]any{"name": cloudspace.ObjectMeta.Name})
	diags := steCloudspaceState(ctx, cloudspace, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	// TODO: Use "wait_until_ready" attribute to wait for the cloudspace to be ready
	// Refer:  https://github.com/hashicorp/terraform-provider-kubernetes/blob/main/kubernetes/resource_kubernetes_deployment_v1.go#L246

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Debug(ctx, "Updated local state")
}

func (r *cloudspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_cloudspace.CloudspaceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
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
	err = r.client.Get(ctx, ktypes.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, cloudspace)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get cloudspace", err.Error())
		return
	}
	diags := steCloudspaceState(ctx, cloudspace, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.LastUpdated = types.StringNull()

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, curData resource_cloudspace.CloudspaceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &curData)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 'curData' is state of the resource on remote(current state) and 'data' is planned state(new state) of the resource
	tflog.Debug(ctx, "Computing name, namespace using resource id", map[string]any{"id": data.Id.ValueString()})
	name, namespace, err := getNameAndNamespaceFromId(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name and namespace from id", err.Error())
		return
	}
	tflog.Debug(ctx, "Name, namespace using resource id", map[string]any{"name": name, "namespace": namespace})

	// TODO: Find the difference between state and plan and update only the changed fields using patch
	cloudspace := &ngpcv1.CloudSpace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudSpace",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			// TODO: Save resource version in the private state not in the public state
			ResourceVersion: curData.ResourceVersion.ValueString(),
		},
		Spec: ngpcv1.CloudSpaceSpec{
			Region:         data.Region.ValueString(),
			Cloud:          "default",
			HAControlPlane: data.HacontrolPlane.ValueBool(),
			Webhook:        data.PreemptionWebhook.ValueString(),
		},
	}
	tflog.Debug(ctx, "Updating cloudspace", map[string]any{"name": name})
	err = r.client.Update(ctx, cloudspace)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cloudspace", err.Error())
		return
	}
	tflog.Info(ctx, "Updated cloudspace", map[string]any{"name": data.CloudspaceName.ValueString()})
	diags := steCloudspaceState(ctx, cloudspace, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_cloudspace.CloudspaceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete API call logic
	tflog.Debug(ctx, "Computing name, namespace using resource id", map[string]any{"id": data.Id.ValueString()})
	name, namespace, err := getNameAndNamespaceFromId(data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name and namespace from id", err.Error())
		return
	}
	tflog.Debug(ctx, "Name, namespace using resource id", map[string]any{"name": name, "namespace": namespace})

	tflog.Debug(ctx, "Deleting cloudspace", map[string]any{"name": name})
	err = r.client.Delete(ctx, &ngpcv1.CloudSpace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudSpace",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete cloudspace", err.Error())
		return
	}
	tflog.Info(ctx, "Deleted cloudspace", map[string]any{"name": data.CloudspaceName.ValueString()})
}

func (r *cloudspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	namespace := os.Getenv("RXTSPOT_ORG_NS")
	if namespace == "" {
		resp.Diagnostics.AddError("Failed to get org namespace", "RXTSPOT_ORG_NS is not set")
		return
	}
	req.ID = fmt.Sprintf("%s/%s", namespace, req.ID)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func steCloudspaceState(ctx context.Context, cloudspace *ngpcv1.CloudSpace, state *resource_cloudspace.CloudspaceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	state.Id = types.StringValue(getIDFromObjectMeta(cloudspace.ObjectMeta))
	state.Region = types.StringValue(cloudspace.Spec.Region)
	state.HacontrolPlane = types.BoolValue(cloudspace.Spec.HAControlPlane)
	if cloudspace.Spec.Webhook != "" {
		// even if we dont set string value it becomes "" by default
		// assume it as Null if it is not set
		state.PreemptionWebhook = types.StringValue(cloudspace.Spec.Webhook)
	} else {
		state.PreemptionWebhook = types.StringNull()
	}
	state.CloudspaceName = types.StringValue(cloudspace.ObjectMeta.Name)
	state.ResourceVersion = types.StringValue(cloudspace.ObjectMeta.ResourceVersion)
	state.FirstReadyTimestamp = types.StringValue(cloudspace.Status.FirstReadyTimestamp.Format(time.RFC3339))
	state.SpotnodepoolIds, diags = types.ListValueFrom(ctx, types.StringType, cloudspace.Spec.BidRequests)
	diags.Append(diags...)
	if diags.HasError() {
		return diags
	}

	var bidsSlice []resource_cloudspace.BidsValue
	for _, val := range cloudspace.Status.Bids {
		var wonCount types.Int64
		if val.WonCount != nil {
			wonCount = types.Int64Value(int64(*val.WonCount))
		} else {
			wonCount = types.Int64Null()
		}
		bidObjVal, convertDiags := resource_cloudspace.BidsValue{
			BidName:  types.StringValue(val.BidName),
			WonCount: wonCount,
		}.ToObjectValue(ctx)
		diags.Append(convertDiags...)
		if diags.HasError() {
			return diags
		}
		bidObjValuable, convertDiags := resource_cloudspace.BidsType{}.ValueFromObject(ctx, bidObjVal)
		diags.Append(convertDiags...)
		if diags.HasError() {
			return diags
		}
		bidsSlice = append(bidsSlice, bidObjValuable.(resource_cloudspace.BidsValue))
	}
	state.Bids, diags = types.SetValueFrom(ctx, resource_cloudspace.BidsValue{}.Type(ctx), bidsSlice)
	diags.Append(diags...)
	if diags.HasError() {
		return diags
	}
	var allocationsSlice []resource_cloudspace.PendingAllocationsValue
	for _, val := range cloudspace.Status.PendingAllocations {
		allocObjVal, convertDiags := resource_cloudspace.PendingAllocationsValue{
			BidName:     types.StringValue(val.BidName),
			ServerClass: types.StringValue(val.ServerClassName),
			Count:       types.Int64Value(int64(val.Count)),
		}.ToObjectValue(ctx)
		diags.Append(convertDiags...)
		if diags.HasError() {
			return diags
		}
		allocObjValuable, convertDiags := resource_cloudspace.PendingAllocationsType{}.ValueFromObject(ctx, allocObjVal)
		diags.Append(convertDiags...)
		if diags.HasError() {
			return diags
		}
		allocationsSlice = append(allocationsSlice, allocObjValuable.(resource_cloudspace.PendingAllocationsValue))
	}
	var convertDiags diag.Diagnostics
	state.PendingAllocations, convertDiags = types.SetValueFrom(ctx,
		resource_cloudspace.PendingAllocationsValue{}.Type(ctx), allocationsSlice)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return diags
	}
	return diags
}
