package provider

import (
	"context"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rackerlabs/terraform-provider-spot/internal/provider/datasource_serverclass"
	ktypes "k8s.io/apimachinery/pkg/types"
)

var _ datasource.DataSource = (*serverclassDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*serverclassDataSource)(nil)

func NewServerclassDataSource() datasource.DataSource {
	return &serverclassDataSource{}
}

type serverclassDataSource struct {
	client ngpc.Client
}

func (d *serverclassDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serverclass"
}

func (d *serverclassDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_serverclass.ServerclassDataSourceSchema(ctx)
}

func (d *serverclassDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ngpc.HTTPClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *ngpc.Client, got: %T. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = client
}

func (d *serverclassDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_serverclass.ServerclassModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	name := data.Name.ValueString()

	// Call the API
	var serverclass ngpcv1.ServerClass
	err := d.client.Get(ctx, ktypes.NamespacedName{Name: name}, &serverclass)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get serverclass", err.Error())
		return
	}
	data.Availability = types.StringValue(serverclass.Spec.Availability)
	data.Category = types.StringValue(serverclass.Spec.Category)
	data.DisplayName = types.StringValue(serverclass.Spec.DisplayName)
	data.FlavorType = types.StringValue(serverclass.Spec.FlavorType)
	data.Name = types.StringValue(serverclass.Name)
	onDemandPricingValue, diags := getOnDemandPricingValue(ctx, serverclass.Spec.OnDemandPricing)
	resp.Diagnostics.Append(diags...)
	data.OnDemandPricing = onDemandPricingValue
	serverclassProviderValue, diags := getRegionProviderValue(ctx, serverclass.Spec.Provider)
	resp.Diagnostics.Append(diags...)
	data.ServerclassProvider = serverclassProviderValue
	data.Region = types.StringValue(serverclass.Spec.Region)
	resources, diags := getResourcesValue(ctx, serverclass.Spec.Resources)
	resp.Diagnostics.Append(diags...)
	data.Resources = resources
	statusValue, diags := getServerclassStatusValue(ctx, serverclass.Status)
	resp.Diagnostics.Append(diags...)
	data.Status = statusValue

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func getOnDemandPricingValue(ctx context.Context, serverClassOnDemandPricing ngpcv1.ServerClassOnDemandPricing) (
	datasource_serverclass.OnDemandPricingValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	onDemandPricingObjValue, convertDiags := datasource_serverclass.OnDemandPricingValue{
		Cost:     types.StringValue(serverClassOnDemandPricing.Cost),
		Interval: types.StringValue(serverClassOnDemandPricing.Interval),
	}.ToObjectValue(ctx)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return datasource_serverclass.NewOnDemandPricingValueNull(), diags
	}
	onDemandPricingValue, convertDiags := datasource_serverclass.NewOnDemandPricingValue(
		onDemandPricingObjValue.AttributeTypes(ctx), onDemandPricingObjValue.Attributes())
	diags.Append(convertDiags...)
	return onDemandPricingValue, diags
}

func getRegionProviderValue(ctx context.Context, scProvider ngpcv1.ServerClassProvider) (
	datasource_serverclass.ServerclassProviderValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	serverclassProvider, convertDiags := datasource_serverclass.ServerclassProviderValue{
		ProviderType: types.StringValue(scProvider.ProviderType),
		FlavorId:     types.StringValue(scProvider.ProviderFlavorID),
	}.ToObjectValue(ctx)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return datasource_serverclass.NewServerclassProviderValueNull(), diags
	}
	regionProviderValue, convertDiags := datasource_serverclass.NewServerclassProviderValue(
		serverclassProvider.AttributeTypes(ctx), serverclassProvider.Attributes())
	diags.Append(convertDiags...)
	return regionProviderValue, diags
}

func getResourcesValue(ctx context.Context, resources ngpcv1.ServerResources) (
	datasource_serverclass.ResourcesValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	resourcesObjValue, convertDiags := datasource_serverclass.ResourcesValue{
		Cpu:    types.StringValue(resources.CPU),
		Memory: types.StringValue(resources.Memory),
	}.ToObjectValue(ctx)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return datasource_serverclass.NewResourcesValueNull(), diags
	}
	resourcesValue, convertDiags := datasource_serverclass.NewResourcesValue(
		resourcesObjValue.AttributeTypes(ctx), resourcesObjValue.Attributes())
	diags.Append(convertDiags...)
	return resourcesValue, diags
}

func getServerclassStatusValue(ctx context.Context, status ngpcv1.ServerClassStatus) (
	datasource_serverclass.StatusValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	spotPricingObjValue, convertDiags := datasource_serverclass.SpotPricingValue{
		MarketPricePerHour: types.StringValue(status.SpotPricing.MarketPricePerHour),
		HammerPricePerHour: types.StringValue(status.SpotPricing.HammerPricePerHour),
	}.ToObjectValue(ctx)
	diags.Append(convertDiags...)

	statusObjValue, convertDiags := datasource_serverclass.StatusValue{
		Available:   types.Int64Value(int64(status.Available)),
		Capacity:    types.Int64Value(int64(status.Capacity)),
		LastAuction: types.Int64Value(int64(status.LastAuction)),
		Reserved:    types.Int64Value(int64(status.Reserved)),
		SpotPricing: spotPricingObjValue,
	}.ToObjectValue(ctx)
	diags.Append(convertDiags...)
	if diags.HasError() {
		return datasource_serverclass.NewStatusValueNull(), diags
	}
	statusValue, convertDiags := datasource_serverclass.NewStatusValue(
		statusObjValue.AttributeTypes(ctx), statusObjValue.Attributes())
	diags.Append(convertDiags...)
	return statusValue, diags
}
