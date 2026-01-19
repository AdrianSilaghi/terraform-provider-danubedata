package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CachesDataSource{}
var _ datasource.DataSourceWithConfigure = &CachesDataSource{}

type CachesDataSource struct {
	client *client.Client
}

type CachesDataSourceModel struct {
	Instances []CacheInstanceModel `tfsdk:"instances"`
}

type CacheInstanceModel struct {
	ID              types.String  `tfsdk:"id"`
	Name            types.String  `tfsdk:"name"`
	Status          types.String  `tfsdk:"status"`
	CacheProvider   types.String  `tfsdk:"cache_provider"`
	Version         types.String  `tfsdk:"version"`
	Datacenter      types.String  `tfsdk:"datacenter"`
	ResourceProfile types.String  `tfsdk:"resource_profile"`
	CPUCores        types.Int64   `tfsdk:"cpu_cores"`
	MemorySizeMB    types.Int64   `tfsdk:"memory_size_mb"`
	Endpoint        types.String  `tfsdk:"endpoint"`
	Port            types.Int64   `tfsdk:"port"`
	MonthlyCost     types.Float64 `tfsdk:"monthly_cost"`
	CreatedAt       types.String  `tfsdk:"created_at"`
}

func NewCachesDataSource() datasource.DataSource {
	return &CachesDataSource{}
}

func (d *CachesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_caches"
}

func (d *CachesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all cache instances in your account.",
		Attributes: map[string]schema.Attribute{
			"instances": schema.ListNestedAttribute{
				Description: "List of cache instances.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier for the cache instance.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the cache instance.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the cache instance.",
							Computed:    true,
						},
						"cache_provider": schema.StringAttribute{
							Description: "Cache provider (redis, valkey, dragonfly).",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Cache version.",
							Computed:    true,
						},
						"datacenter": schema.StringAttribute{
							Description: "Datacenter location.",
							Computed:    true,
						},
						"resource_profile": schema.StringAttribute{
							Description: "Resource profile (predefined CPU/RAM configuration).",
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
						"endpoint": schema.StringAttribute{
							Description: "Connection endpoint hostname.",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "Connection port.",
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

func (d *CachesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CachesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var data CachesDataSourceModel

	instances, err := d.client.ListCaches(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list cache instances", err.Error())
		return
	}

	data.Instances = make([]CacheInstanceModel, len(instances))
	for i, inst := range instances {
		data.Instances[i] = CacheInstanceModel{
			ID:              types.StringValue(inst.ID),
			Name:            types.StringValue(inst.Name),
			Status:          types.StringValue(inst.Status),
			CacheProvider:   types.StringValue(inst.Provider.Name),
			Version:         types.StringValue(inst.Version),
			Datacenter:      types.StringValue(inst.Datacenter),
			ResourceProfile: types.StringValue(inst.ResourceProfile),
			CPUCores:        types.Int64Value(int64(inst.CPUCores)),
			MemorySizeMB:    types.Int64Value(int64(inst.MemorySizeMB)),
			MonthlyCost:     types.Float64Value(inst.MonthlyCostDollars),
			CreatedAt:       types.StringValue(inst.CreatedAt),
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
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
