package provider

import (
	"context"
	"fmt"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	ktypes "k8s.io/apimachinery/pkg/types"

	"github.com/rackerlabs/terraform-provider-spot/internal/provider/datasource_ondemandnodepool"
)

var (
	_ datasource.DataSource              = (*ondemandnodepoolDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*ondemandnodepoolDataSource)(nil)
)

func NewOndemandnodepoolDataSource() datasource.DataSource {
	return &ondemandnodepoolDataSource{}
}

type ondemandnodepoolDataSource struct {
	ngpcClient ngpc.Client
}

func (d *ondemandnodepoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ondemandnodepool"
}

func (d *ondemandnodepoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ondemandnodepoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ondemandnodepoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_ondemandnodepool.OndemandnodepoolModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	namespace, err := getNamespaceFromEnv()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get namespace from env", err.Error())
		return
	}
	// Read API call logic
	tflog.Info(ctx, "Getting ondemandnodepool", map[string]any{"name": name, "namespace": namespace})
	onDemandNodePool := &ngpcv1.OnDemandNodePool{}
	err = d.ngpcClient.Get(ctx, ktypes.NamespacedName{Name: name, Namespace: namespace}, onDemandNodePool)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get ondemandnodepool", err.Error())
		return
	}
	resp.Diagnostics.Append(setOnDemandNodePoolDataSourceState(ctx, onDemandNodePool, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func setOnDemandNodePoolDataSourceState(ctx context.Context, ondemandnodepool *ngpcv1.OnDemandNodePool, state *datasource_ondemandnodepool.OndemandnodepoolModel) diag.Diagnostics {
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
	if ondemandnodepool.Spec.CustomLabels != nil {
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
	if ondemandnodepool.Spec.CustomAnnotations != nil {
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
		AttrTypes: datasource_ondemandnodepool.TaintsValue{}.AttributeTypes(ctx),
	}
	if len(ondemandnodepool.Spec.CustomTaints) > 0 {
		taintsList := make([]datasource_ondemandnodepool.TaintsValue, 0, len(ondemandnodepool.Spec.CustomTaints))
		for _, taint := range ondemandnodepool.Spec.CustomTaints {
			taintValue := datasource_ondemandnodepool.TaintsValue{
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
