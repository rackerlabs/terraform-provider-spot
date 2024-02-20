package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rackerlabs/terraform-provider-spot/internal/provider/resource_cloudspace"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
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
	data.Id = types.StringValue(getIDFromObjectMeta(cloudspace.ObjectMeta))
	data.Region = types.StringValue(cloudspace.Spec.Region)
	data.HacontrolPlane = types.BoolValue(cloudspace.Spec.HAControlPlane)
	if cloudspace.Spec.Webhook != "" {
		// even if we dont set string value it becomes "" by default
		// assume it as Null if it is not set
		data.PreemptionWebhook = types.StringValue(cloudspace.Spec.Webhook)
	} else {
		data.PreemptionWebhook = types.StringNull()
	}
	data.CloudspaceName = types.StringValue(cloudspace.ObjectMeta.Name)
	data.ResourceVersion = types.StringValue(cloudspace.ObjectMeta.ResourceVersion)
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))
	tflog.Debug(ctx, "Updated local state", map[string]any{"data": data})

	// TODO: Use "wait_for_ready" attribute to wait for the cloudspace to be ready
	// Refer:  https://github.com/hashicorp/terraform-provider-kubernetes/blob/main/kubernetes/resource_kubernetes_deployment_v1.go#L246

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

	data.Id = types.StringValue(getIDFromObjectMeta(cloudspace.ObjectMeta))
	data.Region = types.StringValue(cloudspace.Spec.Region)
	data.HacontrolPlane = types.BoolValue(cloudspace.Spec.HAControlPlane)
	if cloudspace.Spec.Webhook != "" {
		// even if we dont set string value it becomes "" by default
		// assume it as Null if it is not set
		data.PreemptionWebhook = types.StringValue(cloudspace.Spec.Webhook)
	} else {
		data.PreemptionWebhook = types.StringNull()
	}
	data.CloudspaceName = types.StringValue(cloudspace.ObjectMeta.Name)
	data.ResourceVersion = types.StringValue(cloudspace.ObjectMeta.ResourceVersion)

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
			Name:            name,
			Namespace:       namespace,
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

	// Save updated data into Terraform state
	data.Id = types.StringValue(getIDFromObjectMeta(cloudspace.ObjectMeta))
	data.Region = types.StringValue(cloudspace.Spec.Region)
	data.HacontrolPlane = types.BoolValue(cloudspace.Spec.HAControlPlane)
	if cloudspace.Spec.Webhook != "" {
		// even if we dont set string value it becomes "" by default
		// assume it as Null if it is not set
		data.PreemptionWebhook = types.StringValue(cloudspace.Spec.Webhook)
	} else {
		data.PreemptionWebhook = types.StringNull()
	}
	data.CloudspaceName = types.StringValue(cloudspace.ObjectMeta.Name)
	data.ResourceVersion = types.StringValue(cloudspace.ObjectMeta.ResourceVersion)
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
