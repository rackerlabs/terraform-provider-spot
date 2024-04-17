package provider

import (
	"context"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rackerlabs/terraform-provider-spot/internal/provider/datasource_regions"
)

var _ datasource.DataSource = (*regionsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*regionsDataSource)(nil)

func NewRegionsDataSource() datasource.DataSource {
	return &regionsDataSource{}
}

type regionsDataSource struct {
	client ngpc.Client
}

func (d *regionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *regionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_regions.RegionsDataSourceSchema(ctx)
}

func (d *regionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ngpc.HTTPClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			"Expected *ngpc.HTTPClient, got: %T. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = client
}

func (d *regionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_regions.RegionsModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	tflog.Debug(ctx, "Listing regions")
	regionsList := &ngpcv1.RegionList{}
	err := d.client.List(ctx, regionsList)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get regions", err.Error())
		return
	}

	// TODO: Implement filters

	regionNames := make([]string, 0, len(regionsList.Items))
	for _, region := range regionsList.Items {
		regionNames = append(regionNames, region.Name)
	}

	regionNamesVal, diags := types.ListValueFrom(ctx, types.StringType, regionNames)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Names = regionNamesVal
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
