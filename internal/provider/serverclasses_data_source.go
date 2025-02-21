package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rackerlabs/terraform-provider-spot/internal/provider/datasource_serverclasses"
)

var (
	_ datasource.DataSource              = (*serverclassesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*serverclassesDataSource)(nil)
)

func NewServerclassesDataSource() datasource.DataSource {
	return &serverclassesDataSource{}
}

type serverclassesDataSource struct {
	client ngpc.Client
}

func (d *serverclassesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serverclasses"
}

func (d *serverclassesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_serverclasses.ServerclassesDataSourceSchema(ctx)
}

func (d *serverclassesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serverclassesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data datasource_serverclasses.ServerclassesModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read API call logic
	var serverclassList ngpcv1.ServerClassList
	err := d.client.List(ctx, &serverclassList)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list server classes", err.Error())
		return
	}
	serverclasses := serverclassList.Items
	if !data.Filters.IsNull() {
		var filterValues []datasource_serverclasses.FiltersValue
		resp.Diagnostics.Append(data.Filters.ElementsAs(ctx, &filterValues, false)...)
		for _, filterValue := range filterValues {
			var values []string
			resp.Diagnostics.Append(filterValue.Values.ElementsAs(ctx, &values, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			serverclasses, err = filterServerClasses(serverclasses, filterValue.Name.ValueString(), values)
			if err != nil {
				resp.Diagnostics.AddError("Failed to filter server classes", err.Error())
				return
			}
		}
	}

	serverclassNames := make([]string, 0, len(serverclasses))
	for _, serverclass := range serverclasses {
		serverclassNames = append(serverclassNames, serverclass.Name)
	}
	serverclassListValue, diags := types.ListValueFrom(ctx, types.StringType, serverclassNames)
	resp.Diagnostics.Append(diags...)
	data.Names = serverclassListValue

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func filterServerClasses(serverclasses []ngpcv1.ServerClass, name string, values []string) ([]ngpcv1.ServerClass, error) {
	var filteredServerClasses []ngpcv1.ServerClass
	for _, serverclass := range serverclasses {
		switch name {
		case "name":
			if StrSliceContains(values, serverclass.Name) {
				filteredServerClasses = append(filteredServerClasses, serverclass)
			}
		case "availability":
			if StrSliceContains(values, serverclass.Spec.Availability) {
				filteredServerClasses = append(filteredServerClasses, serverclass)
			}
		case "category":
			if StrSliceContains(values, serverclass.Spec.Category) {
				filteredServerClasses = append(filteredServerClasses, serverclass)
			}
		case "display_name":
			if StrSliceContains(values, serverclass.Spec.DisplayName) {
				filteredServerClasses = append(filteredServerClasses, serverclass)
			}
		case "flavor_type":
			if StrSliceContains(values, serverclass.Spec.FlavorType) {
				filteredServerClasses = append(filteredServerClasses, serverclass)
			}
		case "serverclass_provider.provider_type":
			if StrSliceContains(values, serverclass.Spec.Provider.ProviderType) {
				filteredServerClasses = append(filteredServerClasses, serverclass)
			}
		case "serverclass_provider.flavor_id":
			if StrSliceContains(values, serverclass.Spec.Provider.ProviderFlavorID) {
				filteredServerClasses = append(filteredServerClasses, serverclass)
			}
		case "serverclass_provider.region":
			if StrSliceContains(values, serverclass.Spec.Region) {
				filteredServerClasses = append(filteredServerClasses, serverclass)
			}
		case "resources.cpu":
			if cpuMatchesEpressions(values, serverclass.Spec.Resources.CPU) {
				filteredServerClasses = append(filteredServerClasses, serverclass)
			}
		case "resources.memory":
			if memoryMatchesEpressions(values, serverclass.Spec.Resources.Memory) {
				filteredServerClasses = append(filteredServerClasses, serverclass)
			}
		case "status.available":
			for _, value := range values {
				if matchesExpression(value, serverclass.Status.Available) {
					filteredServerClasses = append(filteredServerClasses, serverclass)
					break
				}
			}
		case "status.reserved":
			for _, value := range values {
				if matchesExpression(value, serverclass.Status.Reserved) {
					filteredServerClasses = append(filteredServerClasses, serverclass)
					break
				}
			}
		case "status.capacity":
			for _, value := range values {
				if matchesExpression(value, serverclass.Status.Capacity) {
					filteredServerClasses = append(filteredServerClasses, serverclass)
					break
				}
			}
		case "status.last_auction":
			for _, value := range values {
				if matchesExpression(value, serverclass.Status.LastAuction) {
					filteredServerClasses = append(filteredServerClasses, serverclass)
					break
				}
			}
		default:
			return nil, fmt.Errorf("unsupported filter name %s", name)
		}
	}
	return filteredServerClasses, nil
}

func memoryMatchesEpressions(expressions []string, memory string) bool {
	memory = strings.TrimSuffix(memory, "GB")
	floatMemory, err := strconv.ParseFloat(memory, 64)
	if err != nil {
		return false
	}
	for _, expression := range expressions {
		expression = strings.TrimSuffix(strings.TrimSpace(expression), "GB")
		if matchesExpression(expression, floatMemory) {
			return true
		}
	}
	return false
}

func cpuMatchesEpressions(expressions []string, serverclassCPU string) bool {
	cpus, err := strconv.Atoi(serverclassCPU)
	if err != nil {
		return false
	}
	for _, expression := range expressions {
		if matchesExpression(strings.TrimSpace(expression), cpus) {
			return true
		}
	}
	return false
}
