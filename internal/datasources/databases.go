package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DatabasesDataSource{}
var _ datasource.DataSourceWithConfigure = &DatabasesDataSource{}

type DatabasesDataSource struct {
	client *client.Client
}

type DatabasesDataSourceModel struct {
	Instances []DatabaseInstanceModel `tfsdk:"instances"`
}

type DatabaseInstanceModel struct {
	ID              types.String  `tfsdk:"id"`
	Name            types.String  `tfsdk:"name"`
	Status          types.String  `tfsdk:"status"`
	Engine          types.String  `tfsdk:"engine"`
	Version         types.String  `tfsdk:"version"`
	DatabaseName    types.String  `tfsdk:"database_name"`
	Datacenter      types.String  `tfsdk:"datacenter"`
	ResourceProfile types.String  `tfsdk:"resource_profile"`
	CPUCores        types.Int64   `tfsdk:"cpu_cores"`
	MemorySizeMB    types.Int64   `tfsdk:"memory_size_mb"`
	StorageSizeGB   types.Int64   `tfsdk:"storage_size_gb"`
	Endpoint        types.String  `tfsdk:"endpoint"`
	Port            types.Int64   `tfsdk:"port"`
	Username        types.String  `tfsdk:"username"`
	MonthlyCost     types.Float64 `tfsdk:"monthly_cost"`
	CreatedAt       types.String  `tfsdk:"created_at"`
}

func NewDatabasesDataSource() datasource.DataSource {
	return &DatabasesDataSource{}
}

func (d *DatabasesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_databases"
}

func (d *DatabasesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all database instances in your account.",
		Attributes: map[string]schema.Attribute{
			"instances": schema.ListNestedAttribute{
				Description: "List of database instances.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier for the database instance.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the database instance.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the database instance.",
							Computed:    true,
						},
						"engine": schema.StringAttribute{
							Description: "Database engine (mysql, postgresql, mariadb).",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Database version.",
							Computed:    true,
						},
						"database_name": schema.StringAttribute{
							Description: "Name of the database.",
							Computed:    true,
						},
						"datacenter": schema.StringAttribute{
							Description: "Datacenter location.",
							Computed:    true,
						},
						"resource_profile": schema.StringAttribute{
							Description: "Resource profile (predefined CPU/RAM/Storage configuration).",
							Computed:    true,
						},
						"cpu_cores": schema.Int64Attribute{
							Description: "Number of CPU cores.",
							Computed:    true,
						},
						"memory_size_mb": schema.Int64Attribute{
							Description: "Memory size in MB.",
							Computed:    true,
						},
						"storage_size_gb": schema.Int64Attribute{
							Description: "Storage size in GB.",
							Computed:    true,
						},
						"endpoint": schema.StringAttribute{
							Description: "Connection endpoint hostname.",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "Connection port.",
							Computed:    true,
						},
						"username": schema.StringAttribute{
							Description: "Database admin username.",
							Computed:    true,
						},
						"monthly_cost": schema.Float64Attribute{
							Description: "Estimated monthly cost.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the instance was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *DatabasesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *DatabasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var data DatabasesDataSourceModel

	instances, err := d.client.ListDatabases(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list database instances", err.Error())
		return
	}

	data.Instances = make([]DatabaseInstanceModel, len(instances))
	for i, inst := range instances {
		data.Instances[i] = DatabaseInstanceModel{
			ID:              types.StringValue(inst.ID),
			Name:            types.StringValue(inst.Name),
			Status:          types.StringValue(inst.Status),
			Engine:          types.StringValue(inst.Engine.Name),
			Version:         types.StringValue(inst.Version),
			Datacenter:      types.StringValue(inst.Datacenter),
			ResourceProfile: types.StringValue(inst.ResourceProfile),
			CPUCores:        types.Int64Value(int64(inst.CPUCores)),
			MemorySizeMB:    types.Int64Value(int64(inst.MemorySizeMB)),
			StorageSizeGB:   types.Int64Value(int64(inst.StorageSizeGB)),
			MonthlyCost:     types.Float64Value(inst.MonthlyCostDollars),
			CreatedAt:       types.StringValue(inst.CreatedAt),
		}
		if inst.DatabaseName != nil {
			data.Instances[i].DatabaseName = types.StringValue(*inst.DatabaseName)
		} else {
			data.Instances[i].DatabaseName = types.StringNull()
		}
		if inst.Endpoint != nil {
			data.Instances[i].Endpoint = types.StringValue(*inst.Endpoint)
		} else {
			data.Instances[i].Endpoint = types.StringNull()
		}
		if inst.Port != nil {
			data.Instances[i].Port = types.Int64Value(int64(*inst.Port))
		} else {
			data.Instances[i].Port = types.Int64Null()
		}
		if inst.Username != nil {
			data.Instances[i].Username = types.StringValue(*inst.Username)
		} else {
			data.Instances[i].Username = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
