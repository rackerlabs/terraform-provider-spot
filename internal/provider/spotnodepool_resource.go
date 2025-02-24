package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/rackerlabs/terraform-provider-spot/internal/provider/resource_spotnodepool"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
)

var (
	_ resource.Resource                = (*spotnodepoolResource)(nil)
	_ resource.ResourceWithConfigure   = (*spotnodepoolResource)(nil)
	_ resource.ResourceWithImportState = (*spotnodepoolResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*spotnodepoolResource)(nil)
)

func NewSpotnodepoolResource() resource.Resource {
	return &spotnodepoolResource{}
}

type spotnodepoolResource struct {
	ngpcClient ngpc.Client
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

	spotProviderData, ok := req.ProviderData.(*SpotProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *SpotProviderData, got: %T.", req.ProviderData),
		)
		return
	}

	r.ngpcClient = spotProviderData.ngpcClient
}

func (r *spotnodepoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var serverClassVal types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root(attribServerClass), &serverClassVal)...)
	if !serverClassVal.IsNull() && !serverClassVal.IsUnknown() {
		serverClasssList, err := listServerClasses(ctx, r.ngpcClient)
		if err != nil {
			resp.Diagnostics.AddWarning("Failed to list server classes", err.Error())
		} else {
			var validServerClass bool
			for _, serverClass := range serverClasssList {
				if serverClass.Name == serverClassVal.ValueString() {
					validServerClass = true
					break
				}
			}
			if !validServerClass {
				resp.Diagnostics.AddAttributeError(path.Root(attribServerClass), "Invalid value",
					"The valid values should be read from the serverclasses data source.")
				return
			}
		}
	}
}

func (r *spotnodepoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_spotnodepool.SpotnodepoolModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name, err := generateRandomUUID()
	if err != nil {
		resp.Diagnostics.AddError("Failed to generate random UUID", err.Error())
		return
	}
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}

	tflog.Debug(ctx, "Creating spotnodepool", map[string]any{"name": name, "namespace": namespace})
	strBidPrice := fmt.Sprintf("%.3f", data.BidPrice.ValueFloat64())

	// Prepare custom metadata
	var labels map[string]string
	var annotations map[string]string
	var taints []corev1.Taint

	if !data.Labels.IsNull() {
		labels = make(map[string]string)
		diags := data.Labels.ElementsAs(ctx, &labels, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else {
		labels = nil
	}

	if !data.Annotations.IsNull() {
		annotations = make(map[string]string)
		diags := data.Annotations.ElementsAs(ctx, &annotations, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else {
		annotations = nil
	}

	if !data.Taints.IsNull() {
		var taintsList []resource_spotnodepool.TaintsValue
		diags := data.Taints.ElementsAs(ctx, &taintsList, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		taints = make([]corev1.Taint, 0, len(taintsList))
		for _, taint := range taintsList {
			taints = append(taints, corev1.Taint{
				Key:    taint.Key.ValueString(),
				Value:  taint.Value.ValueString(),
				Effect: corev1.TaintEffect(taint.Effect.ValueString()),
			})
		}
	} else {
		taints = nil
	}

	spotNodePool := &ngpcv1.SpotNodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SpotNodePool",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: ngpcv1.SpotNodePoolSpec{
			ServerClass:       data.ServerClass.ValueString(),
			Desired:           int(data.DesiredServerCount.ValueInt64()),
			BidPrice:          strBidPrice,
			CloudSpace:        data.CloudspaceName.ValueString(),
			CustomLabels:      labels,
			CustomAnnotations: annotations,
			CustomTaints:      taints,
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
	err = r.ngpcClient.Create(ctx, spotNodePool)
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
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, keyResourceVersion, []byte(spotNodePool.ObjectMeta.ResourceVersion))...)

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
	name, err := getNameFromNameOrId(data.Name.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name", err.Error())
		return
	}
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}

	tflog.Info(ctx, "Getting spotnodepool", map[string]any{"name": name, "namespace": namespace})
	spotNodePool := &ngpcv1.SpotNodePool{}
	err = r.ngpcClient.Get(ctx, ktypes.NamespacedName{Name: name, Namespace: namespace}, spotNodePool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get spotnodepool", err.Error())
		return
	}
	resp.Diagnostics.Append(setSpotnodepoolState(ctx, spotNodePool, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.LastUpdated = types.StringNull()
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, keyResourceVersion, []byte(spotNodePool.ObjectMeta.ResourceVersion))...)
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

	strBidPrice := fmt.Sprintf("%.3f", plan.BidPrice.ValueFloat64())
	name, err := getNameFromNameOrId(plan.Name.ValueString(), plan.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name", err.Error())
		return
	}
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}

	// Get the latest version of the resource before we update it
	// We need to get the latest version to ensure we have the most up-to-date resource version
	// This is required for Kubernetes concurrency control, even though Terraform does its own refresh
	// because other controllers may have modified the resource between our read and update
	tflog.Debug(ctx, "Getting latest version of spotnodepool", map[string]any{"name": name})
	latest := &ngpcv1.SpotNodePool{}
	err = r.ngpcClient.Get(ctx, ktypes.NamespacedName{Name: name, Namespace: namespace}, latest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get latest version of spotnodepool", err.Error())
		return
	}

	// Prepare custom metadata
	var labels map[string]string
	var annotations map[string]string
	var taints []corev1.Taint

	if !plan.Labels.IsNull() {
		labels = make(map[string]string)
		diags := plan.Labels.ElementsAs(ctx, &labels, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else {
		labels = nil
	}

	if !plan.Annotations.IsNull() {
		annotations = make(map[string]string)
		diags := plan.Annotations.ElementsAs(ctx, &annotations, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
	} else {
		annotations = nil
	}

	if !plan.Taints.IsNull() {
		var taintsList []resource_spotnodepool.TaintsValue
		diags := plan.Taints.ElementsAs(ctx, &taintsList, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		taints = make([]corev1.Taint, 0, len(taintsList))
		for _, taint := range taintsList {
			taints = append(taints, corev1.Taint{
				Key:    taint.Key.ValueString(),
				Value:  taint.Value.ValueString(),
				Effect: corev1.TaintEffect(taint.Effect.ValueString()),
			})
		}
	} else {
		taints = nil
	}

	spotNodePool := &ngpcv1.SpotNodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SpotNodePool",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: latest.ResourceVersion,
		},
		Spec: ngpcv1.SpotNodePoolSpec{
			ServerClass:       plan.ServerClass.ValueString(),
			Desired:           int(plan.DesiredServerCount.ValueInt64()),
			BidPrice:          strBidPrice,
			CloudSpace:        plan.CloudspaceName.ValueString(),
			Autoscaling:       autoscalingSpec,
			CustomLabels:      labels,
			CustomAnnotations: annotations,
			CustomTaints:      taints,
		},
	}
	tflog.Debug(ctx, "Updating spotnodepool", map[string]any{"name": spotNodePool.ObjectMeta.Name})
	err = r.ngpcClient.Update(ctx, spotNodePool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update spotnodepool", err.Error())
		return
	}
	tflog.Debug(ctx, "Updated spotnodepool", map[string]any{"name": spotNodePool.ObjectMeta.Name})
	resp.Diagnostics.Append(setSpotnodepoolState(ctx, spotNodePool, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, keyResourceVersion, []byte(spotNodePool.ObjectMeta.ResourceVersion))...)
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

	name, err := getNameFromNameOrId(data.Name.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name", err.Error())
		return
	}
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}
	tflog.Info(ctx, "Deleting spotnodepool", map[string]any{"name": name, "namespace": namespace})
	err = r.ngpcClient.Delete(ctx, &ngpcv1.SpotNodePool{
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
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
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
	state.Id = types.StringValue(spotnodepool.ObjectMeta.Name)
	state.Name = types.StringValue(spotnodepool.ObjectMeta.Name)
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
	state.BidStatus = types.StringValue(spotnodepool.Status.BidStatus)
	if spotnodepool.Status.WonCount != nil {
		state.WonCount = types.Int64Value(int64(*spotnodepool.Status.WonCount))
	} else {
		state.WonCount = types.Int64Null()
	}

	// Map labels
	if len(spotnodepool.Spec.CustomLabels) > 0 {
		labelsMap, diags := types.MapValueFrom(ctx, types.StringType, spotnodepool.Spec.CustomLabels)
		if diags.HasError() {
			diags.Append(diags...)
			return diags
		}
		state.Labels = labelsMap
	} else {
		state.Labels = types.MapNull(types.StringType)
	}

	// Map annotations
	if len(spotnodepool.Spec.CustomAnnotations) > 0 {
		annotationsMap, diags := types.MapValueFrom(ctx, types.StringType, spotnodepool.Spec.CustomAnnotations)
		if diags.HasError() {
			diags.Append(diags...)
			return diags
		}
		state.Annotations = annotationsMap
	} else {
		state.Annotations = types.MapNull(types.StringType)
	}

	// Map taints
	taintsObjType := types.ObjectType{
		AttrTypes: resource_spotnodepool.TaintsValue{}.AttributeTypes(ctx),
	}
	if len(spotnodepool.Spec.CustomTaints) > 0 {
		taintsList := make([]attr.Value, 0, len(spotnodepool.Spec.CustomTaints))
		for _, taint := range spotnodepool.Spec.CustomTaints {
			taintObj, diags := types.ObjectValue(
				resource_spotnodepool.TaintsValue{}.AttributeTypes(ctx),
				map[string]attr.Value{
					"effect": types.StringValue(string(taint.Effect)),
					"key":    types.StringValue(taint.Key),
					"value":  types.StringValue(taint.Value),
				},
			)
			if diags.HasError() {
				diags.Append(diags...)
				return diags
			}
			taintsList = append(taintsList, taintObj)
		}
		state.Taints = types.ListValueMust(taintsObjType, taintsList)
	} else {
		state.Taints = types.ListNull(taintsObjType)
	}

	return diags
}
