package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/rackerlabs/terraform-provider-spot/internal/provider/resource_cloudspace"

	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
)

var (
	_ resource.Resource                = (*cloudspaceResource)(nil)
	_ resource.ResourceWithConfigure   = (*cloudspaceResource)(nil)
	_ resource.ResourceWithImportState = (*cloudspaceResource)(nil)
	_ resource.ResourceWithModifyPlan  = (*cloudspaceResource)(nil)
)

func NewCloudspaceResource() resource.Resource {
	return &cloudspaceResource{}
}

type cloudspaceResource struct {
	ngpcClient ngpc.Client
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

func (r *cloudspaceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var regionVal types.String
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root(attribRegion), &regionVal)...)
	if !regionVal.IsNull() && !regionVal.IsUnknown() {
		regionsList, err := listRegions(ctx, r.ngpcClient)
		if err != nil {
			resp.Diagnostics.AddWarning("Failed to validate region", err.Error())
		} else {
			var validRegion bool
			regionNames := make([]string, len(regionsList))
			for i, region := range regionsList {
				if region.Name == regionVal.ValueString() {
					validRegion = true
				}
				regionNames[i] = region.Name
			}
			if !validRegion {
				resp.Diagnostics.AddAttributeError(path.Root(attribRegion), "Invalid region",
					fmt.Sprintf("Allowed values are: %v", regionNames))
				return
			}
		}
	}
}

func (r *cloudspaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_cloudspace.CloudspaceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.CloudspaceName.ValueString()
	if name == "" {
		name = data.Name.ValueString()
	}

	name, err := getNameFromNameOrId(name, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name", err.Error())
		return
	}
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}
	tflog.Debug(ctx, "Creating cloudspace", map[string]any{"name": name, "namespace": namespace})

	cloudspace := &ngpcv1.CloudSpace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudSpace",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: ngpcv1.CloudSpaceSpec{
			Region:            data.Region.ValueString(),
			Cloud:             "default",
			HAControlPlane:    data.HacontrolPlane.ValueBool(),
			Webhook:           data.PreemptionWebhook.ValueString(),
			DeploymentType:    data.DeploymentType.ValueString(),
			KubernetesVersion: data.KubernetesVersion.ValueString(),
			CNI:               data.Cni.ValueString(),
		},
	}
	tflog.Info(ctx, "Creating cloudspace", map[string]any{"name": cloudspace.ObjectMeta.Name})
	err = r.ngpcClient.Create(ctx, cloudspace)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cloudspace", err.Error())
		return
	}
	tflog.Info(ctx, "Created cloudspace", map[string]any{"name": cloudspace.ObjectMeta.Name})
	diags := setCloudspaceState(ctx, cloudspace, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, keyResourceVersion, []byte(cloudspace.ObjectMeta.ResourceVersion))...)
	data.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Debug(ctx, "Updated local state")

	if data.WaitUntilReady.ValueBool() {
		tflog.Info(ctx, "Waiting for cloudspace to be ready")
		// If you dont find the Timeouts attribute in the data, run make generate-code
		createTimeout, diags := data.Timeouts.Create(ctx, DefaultCloudSpaceCreateTimeout)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		maxRetries := uint64(createTimeout/DefaultRefreshInterval) + 1
		backoffStrategy := backoff.WithMaxRetries(backoff.NewConstantBackOff(DefaultRefreshInterval), maxRetries)
		err := backoff.Retry(waitForCloudSpaceControlPlaneReady(ctx, r.ngpcClient, name, namespace), backoffStrategy)
		if err != nil {
			resp.Diagnostics.AddWarning("Failed to wait for cloudspace to be ready", err.Error())
			return
		}
	}
}

func (r *cloudspaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_cloudspace.CloudspaceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name := data.CloudspaceName.ValueString()
	if name == "" {
		name = data.Name.ValueString()
	}
	name, err := getNameFromNameOrId(name, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name", err.Error())
		return
	}
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}

	// Read API call logic
	tflog.Debug(ctx, "Reading cloudspace", map[string]any{"name": name, "namespace": namespace})
	cloudspace := &ngpcv1.CloudSpace{}
	err = r.ngpcClient.Get(ctx, ktypes.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, cloudspace)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get cloudspace", err.Error())
		return
	}
	diags := setCloudspaceState(ctx, cloudspace, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, keyResourceVersion, []byte(cloudspace.ObjectMeta.ResourceVersion))...)
	data.LastUpdated = types.StringNull()

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *cloudspaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resource_cloudspace.CloudspaceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// 'state' is state of the resource on remote(current state) and 'data' is planned state(new state) of the resource

	name := state.CloudspaceName.ValueString()
	if name == "" {
		name = state.Name.ValueString()
	}
	name, err := getNameFromNameOrId(name, state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name", err.Error())
		return
	}
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}

	if plan.DeploymentType.ValueString() != state.DeploymentType.ValueString() {
		resp.Diagnostics.AddError("Update to the deployment_type is not allowed", fmt.Sprintf("%s to %s is not allowed", state.DeploymentType.ValueString(), plan.DeploymentType.ValueString()))
		return
	}

	resourceVersionBytes, diags := req.Private.GetKey(ctx, keyResourceVersion)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resourceVersion := string(resourceVersionBytes)
	// TODO: Find the difference between state and plan and update only the changed fields using patch
	cloudspace := &ngpcv1.CloudSpace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudSpace",
			APIVersion: "ngpc.rxt.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       namespace,
			ResourceVersion: resourceVersion,
		},
		Spec: ngpcv1.CloudSpaceSpec{
			Region:         plan.Region.ValueString(),
			Cloud:          "default",
			HAControlPlane: plan.HacontrolPlane.ValueBool(),
			Webhook:        plan.PreemptionWebhook.ValueString(),
			DeploymentType: plan.DeploymentType.ValueString(),
		},
	}
	tflog.Debug(ctx, "Updating cloudspace", map[string]any{"name": name, "namespace": namespace})
	err = r.ngpcClient.Update(ctx, cloudspace)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cloudspace", err.Error())
		return
	}
	tflog.Info(ctx, "Updated cloudspace", map[string]any{"name": name})
	diags = setCloudspaceState(ctx, cloudspace, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, keyResourceVersion, []byte(cloudspace.ObjectMeta.ResourceVersion))...)
	state.LastUpdated = types.StringValue(time.Now().Format(time.RFC3339))
	state.WaitUntilReady = plan.WaitUntilReady
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cloudspaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_cloudspace.CloudspaceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.CloudspaceName.ValueString()
	if name == "" {
		name = data.Name.ValueString()
	}
	name, err := getNameFromNameOrId(name, data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name", err.Error())
		return
	}
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace", err.Error())
		return
	}

	// Delete API call logic
	tflog.Debug(ctx, "Deleting cloudspace", map[string]any{"name": name, "namespace": namespace})
	err = r.ngpcClient.Delete(ctx, &ngpcv1.CloudSpace{
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
	tflog.Info(ctx, "Deleted cloudspace", map[string]any{"name": name})
}

func (r *cloudspaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func setCloudspaceState(ctx context.Context, cloudspace *ngpcv1.CloudSpace, state *resource_cloudspace.CloudspaceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	state.Id = types.StringValue(cloudspace.ObjectMeta.Name)
	state.Name = types.StringValue(cloudspace.ObjectMeta.Name)
	state.CloudspaceName = types.StringValue(cloudspace.ObjectMeta.Name)
	state.Region = types.StringValue(cloudspace.Spec.Region)
	state.HacontrolPlane = types.BoolValue(cloudspace.Spec.HAControlPlane)
	state.DeploymentType = types.StringValue(cloudspace.Spec.DeploymentType)
	state.KubernetesVersion = types.StringValue(cloudspace.Spec.KubernetesVersion)
	state.Cni = types.StringValue(cloudspace.Spec.CNI)
	if cloudspace.Spec.Webhook != "" {
		// even if we dont set string value it becomes "" by default
		// assume it as Null if it is not set
		state.PreemptionWebhook = types.StringValue(cloudspace.Spec.Webhook)
	} else {
		state.PreemptionWebhook = types.StringNull()
	}
	state.FirstReadyTimestamp = types.StringValue(cloudspace.Status.FirstReadyTimestamp.UTC().Format(time.RFC3339))

	// Always set SpotnodepoolIds to a known value since it may not be available after initial cloudspace creation
	// This prevents Terraform from seeing an "unknown" value after apply
	if len(cloudspace.Spec.BidRequests) > 0 {
		state.SpotnodepoolIds, diags = types.ListValueFrom(ctx, types.StringType, cloudspace.Spec.BidRequests)
	} else {
		state.SpotnodepoolIds = types.ListValueMust(types.StringType, []attr.Value{})
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

// This function returns retry function that waits for cloudspace to be ready with some resilience
// TODO: Is there non-polling based approach?
func waitForCloudSpaceControlPlaneReady(ctx context.Context, ngpcClient ngpc.Client, name string, namespace string) backoff.Operation {
	consecutiveErrorCount := 0
	maxConsecutiveErrors := 3
	startTime := time.Now()
	initialGracePeriod := 2 * time.Minute

	return func() error {
		tflog.Debug(ctx, "Checking cloudspace readiness status", map[string]any{
			"name":      name,
			"namespace": namespace,
			"age":       time.Since(startTime).String(),
		})

		cloudspace := &ngpcv1.CloudSpace{}
		err := ngpcClient.Get(ctx, ktypes.NamespacedName{
			Name:      name,
			Namespace: namespace,
		}, cloudspace)
		if err != nil {
			// Only return permanent error if we can't reach the API
			return backoff.Permanent(fmt.Errorf("failed to get cloudspace: %w", err))
		}

		// If APIServerEndpoint is available, the cloudspace is ready
		if len(cloudspace.Status.APIServerEndpoint) > 0 {
			tflog.Debug(ctx, "Cloudspace control plane is ready", map[string]any{"name": name})
			return nil
		}

		tflog.Debug(ctx, "Cloudspace status", map[string]any{
			"name":              name,
			"phase":             cloudspace.Status.Phase,
			"reason":            cloudspace.Status.Reason,
			"age":               time.Since(startTime).String(),
			"consecutiveErrors": consecutiveErrorCount,
		})

		inGracePeriod := time.Since(startTime) < initialGracePeriod

		// Handle states with resilience
		switch cloudspace.Status.Phase {
		case ngpcv1.CloudSpacePhaseDeleting:
			return backoff.Permanent(fmt.Errorf("cloudspace %s is being deleted", name))

		case ngpcv1.CloudSpacePhaseError:
			if inGracePeriod {
				// During grace period, log but continue waiting
				tflog.Info(ctx, "Ignoring error state during grace period", map[string]any{
					"name":          name,
					"reason":        cloudspace.Status.Reason,
					"timeRemaining": (initialGracePeriod - time.Since(startTime)).String(),
				})
				return fmt.Errorf("cloudspace %s is in Error phase during grace period: %s",
					name, cloudspace.Status.Reason)
			}

			consecutiveErrorCount++
			if consecutiveErrorCount >= maxConsecutiveErrors {
				// If too many consecutive errors, give up
				return backoff.Permanent(fmt.Errorf("cloudspace %s is persistently in Error phase: %s",
					name, cloudspace.Status.Reason))
			}

			tflog.Info(ctx, "Cloudspace in error state but continuing to wait", map[string]any{
				"name":                 name,
				"reason":               cloudspace.Status.Reason,
				"consecutiveErrors":    consecutiveErrorCount,
				"maxConsecutiveErrors": maxConsecutiveErrors,
			})
			return fmt.Errorf("cloudspace %s is in Error phase (%d/%d consecutive errors): %s",
				name, consecutiveErrorCount, maxConsecutiveErrors, cloudspace.Status.Reason)

		case ngpcv1.CloudSpacePhaseProvisioning, ngpcv1.CloudSpacePhaseUpgrading, ngpcv1.CloudSpacePhaseNoWinningBids:
			// Known, reset error count and continue waiting
			consecutiveErrorCount = 0
			tflog.Info(ctx, "Cloudspace is in a transitional state", map[string]any{
				"name":  name,
				"phase": cloudspace.Status.Phase,
			})
			return fmt.Errorf("cloudspace %s is in %s phase", name, cloudspace.Status.Phase)

		default:
			// Unknown state, no increment to error counter
			tflog.Info(ctx, "Cloudspace is not ready yet", map[string]any{
				"name":  name,
				"phase": cloudspace.Status.Phase,
			})
			return fmt.Errorf("cloudspace %s is not ready yet (phase: %s)", name, cloudspace.Status.Phase)
		}
	}
}
