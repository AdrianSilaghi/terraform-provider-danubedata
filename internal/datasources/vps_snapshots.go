package datasources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VpsSnapshotsDataSource{}
var _ datasource.DataSourceWithConfigure = &VpsSnapshotsDataSource{}

type VpsSnapshotsDataSource struct {
	client *client.Client
}

type VpsSnapshotsDataSourceModel struct {
	Snapshots []VpsSnapshotModel `tfsdk:"snapshots"`
}

type VpsSnapshotModel struct {
	ID            types.String  `tfsdk:"id"`
	Name          types.String  `tfsdk:"name"`
	Description   types.String  `tfsdk:"description"`
	Status        types.String  `tfsdk:"status"`
	VpsInstanceID types.String  `tfsdk:"vps_instance_id"`
	SizeGB        types.Float64 `tfsdk:"size_gb"`
	SizeBytes     types.Int64   `tfsdk:"size_bytes"`
	SizeBasis     types.String  `tfsdk:"size_basis"`
	ExpiresAt     types.String  `tfsdk:"expires_at"`
	CreatedAt     types.String  `tfsdk:"created_at"`
}

func NewVpsSnapshotsDataSource() datasource.DataSource {
	return &VpsSnapshotsDataSource{}
}

func (d *VpsSnapshotsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_snapshots"
}

func (d *VpsSnapshotsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all VPS snapshots in your account.",
		Attributes: map[string]schema.Attribute{
			"snapshots": schema.ListNestedAttribute{
				Description: "List of VPS snapshots.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier for the snapshot.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the snapshot.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the snapshot.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the snapshot (creating, ready, failed).",
							Computed:    true,
						},
						"vps_instance_id": schema.StringAttribute{
							Description: "ID of the VPS instance this snapshot belongs to.",
							Computed:    true,
						},
						"size_gb": schema.Float64Attribute{
							Description: "Size of the snapshot in GB.",
							Computed:    true,
						},
						"size_bytes": schema.Int64Attribute{Description: "Snapshot size in bytes when measured.", Computed: true},
						"size_basis": schema.StringAttribute{Description: "How the size was measured.", Computed: true},
						"expires_at": schema.StringAttribute{Description: "Automatic deletion time, when applicable.", Computed: true},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the snapshot was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *VpsSnapshotsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VpsSnapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var data VpsSnapshotsDataSourceModel

	snapshots, err := d.client.ListVpsSnapshots(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list VPS snapshots", err.Error())
		return
	}

	data.Snapshots = make([]VpsSnapshotModel, len(snapshots))
	for i, s := range snapshots {
		data.Snapshots[i] = VpsSnapshotModel{
			ID:            types.StringValue(strconv.FormatInt(s.ID, 10)),
			Name:          types.StringValue(s.Name),
			Description:   types.StringValue(s.Description),
			Status:        types.StringValue(s.Status),
			VpsInstanceID: types.StringValue(s.VpsInstanceID),
			CreatedAt:     types.StringValue(s.CreatedAt),
		}
		if s.SizeKnown {
			data.Snapshots[i].SizeGB = types.Float64Value(s.SizeGB)
		} else {
			data.Snapshots[i].SizeGB = types.Float64Null()
		}
		if s.Description == "" {
			data.Snapshots[i].Description = types.StringNull()
		}
		if s.SizeBytes != nil {
			data.Snapshots[i].SizeBytes = types.Int64Value(*s.SizeBytes)
		} else {
			data.Snapshots[i].SizeBytes = types.Int64Null()
		}
		if s.SizeBasis != nil {
			data.Snapshots[i].SizeBasis = types.StringValue(*s.SizeBasis)
		} else {
			data.Snapshots[i].SizeBasis = types.StringNull()
		}
		if s.ExpiresAt != nil {
			data.Snapshots[i].ExpiresAt = types.StringValue(*s.ExpiresAt)
		} else {
			data.Snapshots[i].ExpiresAt = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
