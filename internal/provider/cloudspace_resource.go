package provider

import (
	"context"
	"fmt"
	"terraform-provider-rackspacespot/internal/provider/resource_cloudspace"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
)

var _ resource.Resource = (*cloudspaceResource)(nil)

func NewCloudspaceResource() resource.Resource {
	return &cloudspaceResource{}
}

type cloudspaceResource struct {
	client ngpc.Client
}

// type cloudspaceResourceModel struct {
// 	Id types.String `tfsdk:"id"`
// }

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

	// Create API call logic
	cloudspaceCR := &ngpcv1.CloudSpace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudSpace",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: data.Metadata.Name.ValueString(),
			// TODO: Get namespace from org name; add org tf resource
			Namespace: data.Metadata.Namespace.ValueString(),
		},
		Spec: ngpcv1.CloudSpaceSpec{
			Region:         data.Spec.Region.ValueString(),
			Cloud:          data.Spec.Cloud.ValueString(),
			HAControlPlane: data.Spec.HacontrolPlane.ValueBool(),
			Webhook:        data.Spec.Webhook.ValueString(),
		},
	}
	err := r.client.Create(ctx, cloudspaceCR)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cloudspace", err.Error())
		return
	}
	tflog.Info(ctx, "Created cloudspace", map[string]any{"status": cloudspaceCR.Status})
	data.Spec.Cloud = types.StringValue(cloudspaceCR.Spec.Cloud)
	data.Spec.Region = types.StringValue(cloudspaceCR.Spec.Region)
	data.Spec.HacontrolPlane = types.BoolValue(cloudspaceCR.Spec.HAControlPlane)
	data.Spec.Webhook = types.StringValue(cloudspaceCR.Spec.Webhook)

	// Example data value setting
	// data.Id = types.StringValue("example-id")

	// Save data into Terraform state
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

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	cloudspaceCR := &ngpcv1.CloudSpace{}
	err := r.client.Get(ctx,
		ktypes.NamespacedName{
			Name:      data.Metadata.Name.ValueString(),
			Namespace: data.Metadata.Namespace.ValueString()},
		cloudspaceCR,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}
	data.Spec.Cloud = types.StringValue(cloudspaceCR.Spec.Cloud)
	data.Spec.Region = types.StringValue(cloudspaceCR.Spec.Region)
	data.Spec.HacontrolPlane = types.BoolValue(cloudspaceCR.Spec.HAControlPlane)
	data.Spec.Webhook = types.StringValue(cloudspaceCR.Spec.Webhook)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_cloudspace.CloudspaceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Update API call logic
	cloudspaceCR := &ngpcv1.CloudSpace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudSpace",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: data.Metadata.Name.ValueString(),
			// TODO: Get namespace from org name; add org tf resource
			Namespace: data.Metadata.Namespace.ValueString(),
		},
		Spec: ngpcv1.CloudSpaceSpec{
			Region:         data.Spec.Region.ValueString(),
			Cloud:          data.Spec.Cloud.ValueString(),
			HAControlPlane: data.Spec.HacontrolPlane.ValueBool(),
			Webhook:        data.Spec.Webhook.ValueString(),
		},
	}
	err := r.client.Update(ctx, cloudspaceCR)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cloudspace", err.Error())
		return
	}
	data.Spec.Cloud = types.StringValue(cloudspaceCR.Spec.Cloud)
	data.Spec.Region = types.StringValue(cloudspaceCR.Spec.Region)
	data.Spec.HacontrolPlane = types.BoolValue(cloudspaceCR.Spec.HAControlPlane)
	data.Spec.Webhook = types.StringValue(cloudspaceCR.Spec.Webhook)

	// Save updated data into Terraform state
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
	err := r.client.Delete(ctx, &ngpcv1.CloudSpace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudSpace",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      data.Metadata.Name.ValueString(),
			Namespace: data.Metadata.Namespace.ValueString(),
		},
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete cloudspace", err.Error())
		return
	}
}
