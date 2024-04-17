package provider

import (
	"context"
	"fmt"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rackerlabs/terraform-provider-spot/internal/provider/datasource_region"
	ktypes "k8s.io/apimachinery/pkg/types"
)

var _ datasource.DataSource = (*regionDataSource)(nil)

func NewRegionDataSource() datasource.DataSource {
	return &regionDataSource{}
}

type regionDataSource struct {
	client ngpc.Client
}

func (d *regionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_region"
}

func (d *regionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_region.RegionDataSourceSchema(ctx)
}

func (d *regionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *regionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_region.RegionModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

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
	region := &ngpcv1.Region{}
	err = d.client.Get(ctx, ktypes.NamespacedName{Name: name, Namespace: namespace}, region)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get region", err.Error())
		return
	}
	data.Country = types.StringValue(region.Spec.Country)
	data.Description = types.StringValue(region.Spec.Description)
	data.Name = types.StringValue(region.ObjectMeta.Name)
	providerObjVal, diags := datasource_region.RegionProviderValue{
		RegionName:   types.StringValue(region.Spec.Provider.ProviderRegionName),
		ProviderType: types.StringValue(region.Spec.Provider.ProviderType),
	}.ToObjectValue(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	providerValue, diags := datasource_region.NewRegionProviderValue(providerObjVal.AttributeTypes(ctx), providerObjVal.Attributes())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.RegionProvider = providerValue
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
