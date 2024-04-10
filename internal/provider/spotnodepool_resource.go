package provider

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/rackerlabs/terraform-provider-spot/internal/provider/resource_spotnodepool"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/google/uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
)

var _ resource.Resource = (*spotnodepoolResource)(nil)
var _ resource.ResourceWithConfigure = (*spotnodepoolResource)(nil)
var _ resource.ResourceWithImportState = (*spotnodepoolResource)(nil)

// TODO: Implement serverclass validation using ConfigValidator
// var _ resource.ResourceWithConfigValidators = (*spotnodepoolResource)(nil)

func NewSpotnodepoolResource() resource.Resource {
	return &spotnodepoolResource{}
}

type spotnodepoolResource struct {
	client *ngpc.HTTPClient
}

func (r *spotnodepoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spotnodepool"
}

func (r *spotnodepoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_spotnodepool.SpotnodepoolResourceSchema(ctx)
}

func (r *spotnodepoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *spotnodepoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_spotnodepool.SpotnodepoolModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API call logic
	namespace := os.Getenv("RXTSPOT_ORG_NS")
	tflog.Debug(ctx, "Using namespace from environment variable", map[string]any{"namespace": namespace})
	strBidPrice := fmt.Sprintf("%.3f", data.BidPrice.ValueFloat64())
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
			BidPrice:    strBidPrice,
			CloudSpace:  data.CloudspaceName.ValueString(),
		},
	}
	if !data.Autoscaling.IsNull() {
		spotNodePool.Spec.Autoscaling = ngpcv1.AutoscalingSpec{
			Enabled:  true,
			MinNodes: int(data.Autoscaling.MinNodes.ValueInt64()),
			MaxNodes: int(data.Autoscaling.MaxNodes.ValueInt64()),
		}
	}
	tflog.Debug(ctx, "Creating spotnodepool", map[string]any{"name": spotNodePool.ObjectMeta.Name})
	err := r.client.Create(ctx, spotNodePool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create nodepool", err.Error())
		return
	}
	tflog.Debug(ctx, "Created spotnodepool", map[string]any{"name": spotNodePool.ObjectMeta.Name})
	resp.Diagnostics.Append(setSpotnodepoolState(ctx, spotNodePool, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Debug(ctx, "Updated local state by getting remote api object", map[string]any{"name": spotNodePool.ObjectMeta.Name})
}

func (r *spotnodepoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_spotnodepool.SpotnodepoolModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	name, namespace, err := getNameAndNamespaceFromId(data.Id.ValueString())
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
	resp.Diagnostics.Append(setSpotnodepoolState(ctx, spotNodePool, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.LastUpdated = types.StringNull()
	tflog.Debug(ctx, "Updating local state", map[string]any{"spec": data})
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *spotnodepoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resource_spotnodepool.SpotnodepoolModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// TODO: Find the difference between state and plan and update only the changed fields using patch
	autoscalingSpec, diags := convertAutoscalingValueToSpec(plan.Autoscaling)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	var serverClassList ngpcv1.ServerClassList
	err := r.client.List(ctx, &serverClassList)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list serverclasses", err.Error())
		return
	}
	var serverClassExists bool
	for _, serverClass := range serverClassList.Items {
		if serverClass.Name == plan.ServerClass.ValueString() {
			serverClassExists = true
			break
		}
	}
	if !serverClassExists {
		var serverClassNames []string
		for _, serverClass := range serverClassList.Items {
			//TODO: Filter serverclasses based on region in cloudspace
			serverClassNames = append(serverClassNames, serverClass.Name)
		}
		resp.Diagnostics.AddError("ServerClass does not exist", fmt.Sprintf("Available serverclasses: %v", serverClassNames))
		return
	}
	strBidPrice := fmt.Sprintf("%.3f", plan.BidPrice.ValueFloat64())
	name, namespace, err := getNameAndNamespaceFromId(plan.Id.ValueString())
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
			ResourceVersion: state.ResourceVersion.ValueString(),
		},
		Spec: ngpcv1.SpotNodePoolSpec{
			ServerClass: plan.ServerClass.ValueString(),
			Desired:     int(plan.DesiredServerCount.ValueInt64()),
			BidPrice:    strBidPrice,
			CloudSpace:  plan.CloudspaceName.ValueString(),
			Autoscaling: autoscalingSpec,
		},
	}
	tflog.Debug(ctx, "Updating spotnodepool", map[string]any{"name": spotNodePool.ObjectMeta.Name})
	err = r.client.Update(ctx, spotNodePool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update spotnodepool", err.Error())
		return
	}
	tflog.Debug(ctx, "Updated spotnodepool", map[string]any{"name": spotNodePool.ObjectMeta.Name})
	resp.Diagnostics.Append(setSpotnodepoolState(ctx, spotNodePool, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *spotnodepoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_spotnodepool.SpotnodepoolModel
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

func (r *spotnodepoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	namespace := os.Getenv("RXTSPOT_ORG_NS")
	if namespace == "" {
		resp.Diagnostics.AddError("Failed to get org namespace", "RXTSPOT_ORG_NS is not set")
		return
	}
	req.ID = fmt.Sprintf("%s/%s", namespace, req.ID)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// convertAutoscalingValueToSpec converts the autoscaling spec from terraform type to k8s type
func convertAutoscalingValueToSpec(
	autoscalingValue resource_spotnodepool.AutoscalingValue) (ngpcv1.AutoscalingSpec, diag.Diagnostics) {
	var diags diag.Diagnostics
	if autoscalingValue.IsNull() {
		// User removed block from the tf spec, hence disable autoscaling
		return ngpcv1.AutoscalingSpec{
			Enabled:  false,
			MinNodes: 0, // equivalent to null in tf because of omitempty
			MaxNodes: 0,
		}, diags
	}

	return ngpcv1.AutoscalingSpec{
		Enabled:  true,
		MinNodes: int(autoscalingValue.MinNodes.ValueInt64()),
		MaxNodes: int(autoscalingValue.MaxNodes.ValueInt64()),
	}, diags
}

func setSpotnodepoolState(ctx context.Context, spotnodepool *ngpcv1.SpotNodePool, state *resource_spotnodepool.SpotnodepoolModel) diag.Diagnostics {
	var diags diag.Diagnostics
	state.Id = types.StringValue(getIDFromObjectMeta(spotnodepool.ObjectMeta))
	state.CloudspaceName = types.StringValue(spotnodepool.Spec.CloudSpace)
	state.ServerClass = types.StringValue(spotnodepool.Spec.ServerClass)
	if spotnodepool.Spec.Desired != 0 {
		state.DesiredServerCount = types.Int64Value(int64(spotnodepool.Spec.Desired))
	} else {
		state.DesiredServerCount = types.Int64Null()
	}
	floatBidPrice, err := strconv.ParseFloat(spotnodepool.Spec.BidPrice, 64)
	if err != nil {
		diags.AddError("Failed to parse bid price returned from remote service", err.Error())
		return diags
	}
	autoscalingSpec := spotnodepool.Spec.Autoscaling
	if !autoscalingSpec.Enabled {
		state.Autoscaling = resource_spotnodepool.NewAutoscalingValueNull()
	} else {
		var minNodes, maxNodes basetypes.Int64Value
		if autoscalingSpec.MinNodes == 0 {
			minNodes = basetypes.NewInt64Null()
		} else {
			minNodes = types.Int64Value(int64(autoscalingSpec.MinNodes))
		}
		if autoscalingSpec.MaxNodes == 0 {
			maxNodes = basetypes.NewInt64Null()
		} else {
			maxNodes = types.Int64Value(int64(autoscalingSpec.MaxNodes))
		}
		autoscalingObjVal, diagsAutoscaling := resource_spotnodepool.AutoscalingValue{
			MinNodes: minNodes,
			MaxNodes: maxNodes,
		}.ToObjectValue(ctx)
		diags.Append(diagsAutoscaling...)
		if diags.HasError() {
			return diags
		}
		autoscalingVal, diagsAutoscaling := resource_spotnodepool.NewAutoscalingValue(
			autoscalingObjVal.AttributeTypes(ctx),
			autoscalingObjVal.Attributes(),
		)
		diags.Append(diagsAutoscaling...)
		if diags.HasError() {
			return diags
		}
		state.Autoscaling = autoscalingVal
	}

	state.BidPrice = types.Float64Value(floatBidPrice)
	state.ResourceVersion = types.StringValue(spotnodepool.ObjectMeta.ResourceVersion)
	state.BidStatus = types.StringValue(spotnodepool.Status.BidStatus)
	if spotnodepool.Status.WonCount != nil {
		state.WonCount = types.Int64Value(int64(*spotnodepool.Status.WonCount))
	} else {
		state.WonCount = types.Int64Null()
	}
	return diags
}
