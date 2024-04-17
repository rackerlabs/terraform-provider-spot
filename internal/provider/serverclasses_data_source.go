package provider

import (
	"context"

	ngpcv1 "github.com/RSS-Engineering/ngpc-cp/api/v1"
	"github.com/RSS-Engineering/ngpc-cp/pkg/ngpc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/rackerlabs/terraform-provider-spot/internal/provider/datasource_serverclasses"
)

var _ datasource.DataSource = (*serverclassesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*serverclassesDataSource)(nil)

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

	// TODO: Implement filters

	serverclassNames := make([]string, 0, len(serverclassList.Items))
	for _, serverclass := range serverclassList.Items {
		serverclassNames = append(serverclassNames, serverclass.Name)
	}
	serverclassListValue, diags := types.ListValueFrom(ctx, types.StringType, serverclassNames)
	resp.Diagnostics.Append(diags...)
	data.Names = serverclassListValue

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
