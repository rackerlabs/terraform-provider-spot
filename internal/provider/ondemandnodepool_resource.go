package provider

import (
	"context"
	"fmt"
	"time"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"

	"github.com/rackerlabs/terraform-provider-spot/internal/provider/resource_ondemandnodepool"
)

var _ resource.Resource = (*ondemandnodepoolResource)(nil)
var _ resource.ResourceWithConfigure = (*ondemandnodepoolResource)(nil)
var _ resource.ResourceWithImportState = (*ondemandnodepoolResource)(nil)
var _ resource.ResourceWithModifyPlan = (*ondemandnodepoolResource)(nil)

func NewOndemandnodepoolResource() resource.Resource {
	return &ondemandnodepoolResource{}
}

type ondemandnodepoolResource struct {
	client *ngpc.HTTPClient
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
			ServerClass: data.ServerClass.ValueString(),
			Desired:     int(data.DesiredServerCount.ValueInt64()),
			CloudSpace:  data.CloudspaceName.ValueString(),
		},
	}

	tflog.Debug(ctx, "Creating ondemandnodepool", map[string]any{"name": onDemandNodePool.ObjectMeta.Name})
	err = r.client.Create(ctx, onDemandNodePool)
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
	err = r.client.Get(ctx, ktypes.NamespacedName{Name: name, Namespace: namespace}, ondemandnodepool)
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
	resourceVersionBytes, diags := req.Private.GetKey(ctx, keyResourceVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resourceVersion := string(resourceVersionBytes)
	ondemandnodepool := &ngpcv1.OnDemandNodePool{
		TypeMeta: metav1.TypeMeta{
			Kind:       "OnDemandNodePool",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: resourceVersion,
		},
		Spec: ngpcv1.OnDemandNodePoolSpec{
			ServerClass: plan.ServerClass.ValueString(),
			Desired:     int(plan.DesiredServerCount.ValueInt64()),
			CloudSpace:  plan.CloudspaceName.ValueString(),
		},
	}
	tflog.Debug(ctx, "Updating ondemandnodepool", map[string]any{"name": ondemandnodepool.ObjectMeta.Name})
	err = r.client.Update(ctx, ondemandnodepool)
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
	err = r.client.Delete(ctx, &ngpcv1.OnDemandNodePool{
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
		serverClasssList, err := listServerClasses(ctx, r.client)
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
	return diags
}
