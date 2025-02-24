package provider

import (
	"context"
	"fmt"
	"time"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"

	"github.com/rackerlabs/terraform-provider-spot/internal/provider/resource_ondemandnodepool"
)

var (
	_ resource.Resource                = (*ondemandnodepoolResource)(nil)
	_ resource.ResourceWithConfigure   = (*ondemandnodepoolResource)(nil)
	_ resource.ResourceWithImportState = (*ondemandnodepoolResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*ondemandnodepoolResource)(nil)
)

func NewOndemandnodepoolResource() resource.Resource {
	return &ondemandnodepoolResource{}
}

type ondemandnodepoolResource struct {
	ngpcClient ngpc.Client
}

func (r *ondemandnodepoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ondemandnodepool"
}

func (r *ondemandnodepoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_ondemandnodepool.OndemandnodepoolResourceSchema(ctx)
}

func (r *ondemandnodepoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ondemandnodepoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_ondemandnodepool.OndemandnodepoolModel

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

	tflog.Debug(ctx, "Creating ondemandnodepool", map[string]any{"name": name, "namespace": namespace})

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
		var taintsList []resource_ondemandnodepool.TaintsValue
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

	onDemandNodePool := &ngpcv1.OnDemandNodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OnDemandNodePool",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: ngpcv1.OnDemandNodePoolSpec{
			ServerClass:       data.ServerClass.ValueString(),
			Desired:           int(data.DesiredServerCount.ValueInt64()),
			CloudSpace:        data.CloudspaceName.ValueString(),
			CustomLabels:      labels,
			CustomAnnotations: annotations,
			CustomTaints:      taints,
		},
	}

	tflog.Debug(ctx, "Creating ondemandnodepool", map[string]any{"name": onDemandNodePool.ObjectMeta.Name})
	err = r.ngpcClient.Create(ctx, onDemandNodePool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create nodepool", err.Error())
		return
	}
	tflog.Debug(ctx, "Created ondemandnodepool", map[string]any{"name": onDemandNodePool.ObjectMeta.Name})
	resp.Diagnostics.Append(setOnDemandNodePoolState(ctx, onDemandNodePool, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, keyResourceVersion, []byte(onDemandNodePool.ObjectMeta.ResourceVersion))...)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Debug(ctx, "Updated local state by getting remote api object", map[string]any{"name": onDemandNodePool.ObjectMeta.Name})
}

func (r *ondemandnodepoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_ondemandnodepool.OndemandnodepoolModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	name := data.Name.ValueString()
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}

	tflog.Info(ctx, "Getting ondemandnodepool", map[string]any{"name": name, "namespace": namespace})
	ondemandnodepool := &ngpcv1.OnDemandNodePool{}
	err = r.ngpcClient.Get(ctx, ktypes.NamespacedName{Name: name, Namespace: namespace}, ondemandnodepool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get ondemandnodepool", err.Error())
		return
	}
	resp.Diagnostics.Append(setOnDemandNodePoolState(ctx, ondemandnodepool, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.LastUpdated = types.StringNull()
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, keyResourceVersion, []byte(ondemandnodepool.ObjectMeta.ResourceVersion))...)
	tflog.Debug(ctx, "Updating local state", map[string]any{"spec": data})
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ondemandnodepoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resource_ondemandnodepool.OndemandnodepoolModel

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
	name := plan.Name.ValueString()
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}

	// Get the latest version of the resource before updating
	// We need to get the latest version to ensure we have the most up-to-date resource version
	// This is required for Kubernetes optimistic concurrency control, even though Terraform does its own refresh
	// because other controllers may have modified the resource between our read and update
	tflog.Debug(ctx, "Getting latest version of ondemandnodepool", map[string]any{"name": name})
	latest := &ngpcv1.OnDemandNodePool{}
	err = r.ngpcClient.Get(ctx, ktypes.NamespacedName{Name: name, Namespace: namespace}, latest)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get latest version of ondemandnodepool", err.Error())
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
		var taintsList []resource_ondemandnodepool.TaintsValue
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

	ondemandnodepool := &ngpcv1.OnDemandNodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OnDemandNodePool",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: latest.ResourceVersion,
		},
		Spec: ngpcv1.OnDemandNodePoolSpec{
			ServerClass:       plan.ServerClass.ValueString(),
			Desired:           int(plan.DesiredServerCount.ValueInt64()),
			CloudSpace:        plan.CloudspaceName.ValueString(),
			CustomLabels:      labels,
			CustomAnnotations: annotations,
			CustomTaints:      taints,
		},
	}
	tflog.Debug(ctx, "Updating ondemandnodepool", map[string]any{"name": ondemandnodepool.ObjectMeta.Name})
	err = r.ngpcClient.Update(ctx, ondemandnodepool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update ondemandnodepool", err.Error())
		return
	}
	tflog.Debug(ctx, "Updated ondemandnodepool", map[string]any{"name": ondemandnodepool.ObjectMeta.Name})
	resp.Diagnostics.Append(setOnDemandNodePoolState(ctx, ondemandnodepool, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, keyResourceVersion, []byte(ondemandnodepool.ObjectMeta.ResourceVersion))...)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ondemandnodepoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_ondemandnodepool.OndemandnodepoolModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}
	tflog.Info(ctx, "Deleting ondemandnodepool", map[string]any{"name": name, "namespace": namespace})
	err = r.ngpcClient.Delete(ctx, &ngpcv1.OnDemandNodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OnDemandNodePool",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		}})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete ondemandnodepool", err.Error())
		return
	}
	// Delete API call logic, we dont need to update state on delete
	tflog.Info(ctx, "Deleted ondemandnodepool", map[string]any{"name": name, "namespace": namespace})
}

func (r *ondemandnodepoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *ondemandnodepoolResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
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

func setOnDemandNodePoolState(ctx context.Context, ondemandnodepool *ngpcv1.OnDemandNodePool, state *resource_ondemandnodepool.OndemandnodepoolModel) diag.Diagnostics {
	var diags diag.Diagnostics
	state.Name = types.StringValue(ondemandnodepool.ObjectMeta.Name)
	state.CloudspaceName = types.StringValue(ondemandnodepool.Spec.CloudSpace)
	state.ServerClass = types.StringValue(ondemandnodepool.Spec.ServerClass)
	if ondemandnodepool.Spec.Desired != 0 {
		state.DesiredServerCount = types.Int64Value(int64(ondemandnodepool.Spec.Desired))
	} else {
		state.DesiredServerCount = types.Int64Null()
	}

	state.ReservedStatus = types.StringValue(ondemandnodepool.Status.ReservedStatus)
	if ondemandnodepool.Status.ReservedCount != nil {
		state.ReservedCount = types.Int64Value(int64(*ondemandnodepool.Status.ReservedCount))
	} else {
		state.ReservedCount = types.Int64Null()
	}

	// Map labels
	if len(ondemandnodepool.Spec.CustomLabels) > 0 {
		labelsMap, diags := types.MapValueFrom(ctx, types.StringType, ondemandnodepool.Spec.CustomLabels)
		if diags.HasError() {
			diags.Append(diags...)
			return diags
		}
		state.Labels = labelsMap
	} else {
		state.Labels = types.MapNull(types.StringType)
	}

	// Map annotations
	if len(ondemandnodepool.Spec.CustomAnnotations) > 0 {
		annotationsMap, diags := types.MapValueFrom(ctx, types.StringType, ondemandnodepool.Spec.CustomAnnotations)
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
		AttrTypes: resource_ondemandnodepool.TaintsValue{}.AttributeTypes(ctx),
	}
	if len(ondemandnodepool.Spec.CustomTaints) > 0 {
		taintsList := make([]attr.Value, 0, len(ondemandnodepool.Spec.CustomTaints))
		for _, taint := range ondemandnodepool.Spec.CustomTaints {
			taintObj, diags := types.ObjectValue(
				resource_ondemandnodepool.TaintsValue{}.AttributeTypes(ctx),
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
