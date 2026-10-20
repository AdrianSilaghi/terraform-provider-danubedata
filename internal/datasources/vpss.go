package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VpssDataSource{}
var _ datasource.DataSourceWithConfigure = &VpssDataSource{}

type VpssDataSource struct {
	client *client.Client
}

type VpssDataSourceModel struct {
	Instances []VpsInstanceModel `tfsdk:"instances"`
}

type VpsInstanceModel struct {
	ID                      types.String  `tfsdk:"id"`
	Name                    types.String  `tfsdk:"name"`
	Status                  types.String  `tfsdk:"status"`
	StatusLabel             types.String  `tfsdk:"status_label"`
	VNCAccessURL            types.String  `tfsdk:"vnc_access_url"`
	TeamID                  types.Int64   `tfsdk:"team_id"`
	UserID                  types.Int64   `tfsdk:"user_id"`
	Image                   types.String  `tfsdk:"image"`
	Datacenter              types.String  `tfsdk:"datacenter"`
	ResourceProfile         types.String  `tfsdk:"resource_profile"`
	CPUAllocationType       types.String  `tfsdk:"cpu_allocation_type"`
	CPUPlatform             types.String  `tfsdk:"cpu_platform"`
	NetworkTier             types.String  `tfsdk:"network_tier"`
	PrimaryIPv4FloatingIPID types.String  `tfsdk:"primary_ipv4_floating_ip_id"`
	PrimaryIPv6FloatingIPID types.String  `tfsdk:"primary_ipv6_floating_ip_id"`
	CPUCores                types.Int64   `tfsdk:"cpu_cores"`
	MemorySizeGB            types.Int64   `tfsdk:"memory_size_gb"`
	StorageSizeGB           types.Int64   `tfsdk:"storage_size_gb"`
	PublicIP                types.String  `tfsdk:"public_ip"`
	PrivateIP               types.String  `tfsdk:"private_ip"`
	IPv6Address             types.String  `tfsdk:"ipv6_address"`
	MonthlyCost             types.Float64 `tfsdk:"monthly_cost"`
	CreatedAt               types.String  `tfsdk:"created_at"`
}

func NewVpssDataSource() datasource.DataSource {
	return &VpssDataSource{}
}

func (d *VpssDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpss"
}

func (d *VpssDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all VPS instances in your account.",
		Attributes: map[string]schema.Attribute{
			"instances": schema.ListNestedAttribute{
				Description: "List of VPS instances.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier for the VPS instance.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the VPS instance.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the VPS instance.",
							Computed:    true,
						},
						"status_label":   schema.StringAttribute{Description: "Human-readable VPS status.", Computed: true},
						"vnc_access_url": schema.StringAttribute{Description: "VNC access URL when available.", Computed: true},
						"team_id":        schema.Int64Attribute{Description: "Owning project ID.", Computed: true},
						"user_id":        schema.Int64Attribute{Description: "Owning user ID.", Computed: true},
						"image": schema.StringAttribute{
							Description: "Operating system image.",
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
						"cpu_allocation_type": schema.StringAttribute{
							Description: "CPU allocation type (shared or dedicated).",
							Computed:    true,
						},
						"cpu_platform":                schema.StringAttribute{Description: "Processor platform.", Computed: true},
						"network_tier":                schema.StringAttribute{Description: "Network tier.", Computed: true},
						"primary_ipv4_floating_ip_id": schema.StringAttribute{Description: "Reserved primary Floating IPv4 ID.", Computed: true},
						"primary_ipv6_floating_ip_id": schema.StringAttribute{Description: "Reserved primary Floating IPv6 ID.", Computed: true},
						"cpu_cores": schema.Int64Attribute{
							Description: "Number of CPU cores.",
							Computed:    true,
						},
						"memory_size_gb": schema.Int64Attribute{
							Description: "Memory size in GB.",
							Computed:    true,
						},
						"storage_size_gb": schema.Int64Attribute{
							Description: "Storage size in GB.",
							Computed:    true,
						},
						"public_ip": schema.StringAttribute{
							Description: "Public IPv4 address.",
							Computed:    true,
						},
						"private_ip": schema.StringAttribute{
							Description: "Private IP address.",
							Computed:    true,
						},
						"ipv6_address": schema.StringAttribute{
							Description: "IPv6 address.",
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

func (d *VpssDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VpssDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var data VpssDataSourceModel

	instances, err := d.client.ListVps(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list VPS instances", err.Error())
		return
	}

	data.Instances = make([]VpsInstanceModel, len(instances))
	for i, inst := range instances {
		data.Instances[i] = VpsInstanceModel{
			ID:                types.StringValue(inst.ID),
			Name:              types.StringValue(inst.Name),
			Status:            types.StringValue(inst.Status),
			StatusLabel:       types.StringValue(inst.StatusLabel),
			TeamID:            types.Int64Value(int64(inst.TeamID)),
			UserID:            types.Int64Value(int64(inst.UserID)),
			Image:             types.StringValue(inst.Image),
			Datacenter:        types.StringValue(inst.Datacenter),
			ResourceProfile:   types.StringValue(inst.ResourceProfile),
			CPUAllocationType: types.StringValue(inst.CPUAllocationType),
			CPUPlatform:       types.StringValue(inst.CPUPlatform),
			NetworkTier:       types.StringValue(inst.NetworkTier),
			CPUCores:          types.Int64Value(int64(inst.CPUCores)),
			MemorySizeGB:      types.Int64Value(int64(inst.MemorySizeGB)),
			StorageSizeGB:     types.Int64Value(int64(inst.StorageSizeGB)),
			MonthlyCost:       types.Float64Value(inst.MonthlyCost),
			CreatedAt:         types.StringValue(inst.CreatedAt),
		}
		if inst.VNCAccessURL != nil {
			data.Instances[i].VNCAccessURL = types.StringValue(*inst.VNCAccessURL)
		} else {
			data.Instances[i].VNCAccessURL = types.StringNull()
		}
		if inst.CPUPlatform == "" {
			data.Instances[i].CPUPlatform = types.StringNull()
		}
		if inst.PrimaryIPv4FloatingIPID != nil {
			data.Instances[i].PrimaryIPv4FloatingIPID = types.StringValue(*inst.PrimaryIPv4FloatingIPID)
		} else {
			data.Instances[i].PrimaryIPv4FloatingIPID = types.StringNull()
		}
		if inst.PrimaryIPv6FloatingIPID != nil {
			data.Instances[i].PrimaryIPv6FloatingIPID = types.StringValue(*inst.PrimaryIPv6FloatingIPID)
		} else {
			data.Instances[i].PrimaryIPv6FloatingIPID = types.StringNull()
		}
		if inst.PublicIP != nil {
			data.Instances[i].PublicIP = types.StringValue(*inst.PublicIP)
		} else {
			data.Instances[i].PublicIP = types.StringNull()
		}
		if inst.PrivateIP != nil {
			data.Instances[i].PrivateIP = types.StringValue(*inst.PrivateIP)
		} else {
			data.Instances[i].PrivateIP = types.StringNull()
		}
		if inst.IPv6Address != nil {
			data.Instances[i].IPv6Address = types.StringValue(*inst.IPv6Address)
		} else {
			data.Instances[i].IPv6Address = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
