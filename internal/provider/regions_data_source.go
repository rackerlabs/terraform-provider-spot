package provider

import (
	"context"
	"fmt"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/rackerlabs/terraform-provider-spot/internal/provider/datasource_regions"
)

var (
	_ datasource.DataSource              = (*regionsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*regionsDataSource)(nil)
)

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

	spotProviderData, ok := req.ProviderData.(*SpotProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *SpotProviderData, got: %T.", req.ProviderData),
		)
		return
	}

	if spotProviderData.ngpcClient == nil {
		resp.Diagnostics.AddError(
			"Missing NGPC API client",
			"Provider configuration appears incomplete",
		)
		return
	}

	d.client = spotProviderData.ngpcClient
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
	regions := regionsList.Items

	if !data.Filters.IsNull() {
		var filtersValue []datasource_regions.FiltersValue
		diags := data.Filters.ElementsAs(ctx, &filtersValue, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, filter := range filtersValue {
			var values []string
			diags := filter.Values.ElementsAs(ctx, &values, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			regions, err = filterRegions(regions, filter.Name.ValueString(), values)
			if err != nil {
				resp.Diagnostics.AddError("Failed to filter regions", err.Error())
				return
			}
		}
	}
	regionNames := make([]string, 0, len(regions))
	for _, region := range regions {
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

func filterRegions(regions []ngpcv1.Region, name string, values []string) ([]ngpcv1.Region, error) {
	var filteredRegions []ngpcv1.Region
	for _, region := range regions {
		switch name {
		case "name":
			if StrSliceContains(values, region.Name) {
				filteredRegions = append(filteredRegions, region)
			}
		case "country":
			if StrSliceContains(values, region.Spec.Country) {
				filteredRegions = append(filteredRegions, region)
			}
		case "description":
			if StrSliceContains(values, region.Spec.Description) {
				filteredRegions = append(filteredRegions, region)
			}
		case "region_provider.region_name":
			if StrSliceContains(values, region.Spec.Provider.ProviderRegionName) {
				filteredRegions = append(filteredRegions, region)
			}
		case "region_provider.provider_type":
			if StrSliceContains(values, region.Spec.Provider.ProviderType) {
				filteredRegions = append(filteredRegions, region)
			}
		default:
			return nil, fmt.Errorf("invalid filter name: %s", name)
		}
	}
	return filteredRegions, nil
}
