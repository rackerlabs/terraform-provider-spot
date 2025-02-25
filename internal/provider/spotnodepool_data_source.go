package provider

import (
	"context"
	"fmt"
	"strconv"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rackerlabs/terraform-provider-spot/internal/provider/datasource_spotnodepool"
	ktypes "k8s.io/apimachinery/pkg/types"
)

var (
	_ datasource.DataSource              = (*spotnodepoolDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*spotnodepoolDataSource)(nil)
)

func NewSpotnodepoolDataSource() datasource.DataSource {
	return &spotnodepoolDataSource{}
}

type spotnodepoolDataSource struct {
	ngpcClient ngpc.Client
}

func (d *spotnodepoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_spotnodepool"
}

func (d *spotnodepoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_spotnodepool.SpotnodepoolDataSourceSchema(ctx)
}

func (d *spotnodepoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.ngpcClient = spotProviderData.ngpcClient
}

func (d *spotnodepoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_spotnodepool.SpotnodepoolModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name, err := getNameFromNameOrId(data.Name.ValueString(), data.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to get name from id", err.Error())
		return
	}
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace from env", err.Error())
		return
	}
	// Read API call logic
	tflog.Info(ctx, "Getting spotnodepool", map[string]any{"name": name, "namespace": namespace})
	spotNodePool := &ngpcv1.SpotNodePool{}
	err = d.ngpcClient.Get(ctx, ktypes.NamespacedName{Name: name, Namespace: namespace}, spotNodePool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get spotnodepool", err.Error())
		return
	}
	resp.Diagnostics.Append(setSpotnodepoolDataSourceState(ctx, spotNodePool, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func setSpotnodepoolDataSourceState(ctx context.Context, spotnodepool *ngpcv1.SpotNodePool, state *datasource_spotnodepool.SpotnodepoolModel) diag.Diagnostics {
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
		state.Autoscaling = datasource_spotnodepool.NewAutoscalingValueNull()
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
		autoscalingObjVal, diagsAutoscaling := datasource_spotnodepool.AutoscalingValue{
			MinNodes: minNodes,
			MaxNodes: maxNodes,
		}.ToObjectValue(ctx)
		diags.Append(diagsAutoscaling...)
		if diags.HasError() {
			return diags
		}
		autoscalingVal, diagsAutoscaling := datasource_spotnodepool.NewAutoscalingValue(
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
	if spotnodepool.Spec.CustomLabels != nil {
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
	if spotnodepool.Spec.CustomAnnotations != nil {
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
		AttrTypes: datasource_spotnodepool.TaintsValue{}.AttributeTypes(ctx),
	}
	if len(spotnodepool.Spec.CustomTaints) > 0 {
		taintsList := make([]datasource_spotnodepool.TaintsValue, 0, len(spotnodepool.Spec.CustomTaints))
		for _, taint := range spotnodepool.Spec.CustomTaints {
			taintValue := datasource_spotnodepool.TaintsValue{
				Effect: types.StringValue(string(taint.Effect)),
				Key:    types.StringValue(taint.Key),
				Value:  types.StringValue(taint.Value),
			}
			taintsList = append(taintsList, taintValue)
		}

		taints, diags := types.ListValueFrom(ctx, taintsObjType, taintsList)
		if diags.HasError() {
			diags.Append(diags...)
			return diags
		}
		state.Taints = taints
	} else {
		state.Taints = types.ListValueMust(taintsObjType, []attr.Value{})
	}

	return diags
}
