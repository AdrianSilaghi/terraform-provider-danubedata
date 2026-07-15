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

var _ datasource.DataSource = &CacheSnapshotsDataSource{}
var _ datasource.DataSourceWithConfigure = &CacheSnapshotsDataSource{}

type CacheSnapshotsDataSource struct {
	client *client.Client
}

type CacheSnapshotsDataSourceModel struct {
	Snapshots []CacheSnapshotModel `tfsdk:"snapshots"`
}

type CacheSnapshotModel struct {
	ID              types.String  `tfsdk:"id"`
	Name            types.String  `tfsdk:"name"`
	Description     types.String  `tfsdk:"description"`
	Status          types.String  `tfsdk:"status"`
	CacheInstanceID types.String  `tfsdk:"cache_instance_id"`
	SizeMB          types.Float64 `tfsdk:"size_mb"`
	CreatedAt       types.String  `tfsdk:"created_at"`
}

func NewCacheSnapshotsDataSource() datasource.DataSource {
	return &CacheSnapshotsDataSource{}
}

func (d *CacheSnapshotsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cache_snapshots"
}

func (d *CacheSnapshotsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all cache snapshots in your account.",
		Attributes: map[string]schema.Attribute{
			"snapshots": schema.ListNestedAttribute{
				Description: "List of cache snapshots.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                schema.StringAttribute{Computed: true},
						"name":              schema.StringAttribute{Computed: true},
						"description":       schema.StringAttribute{Computed: true},
						"status":            schema.StringAttribute{Computed: true},
						"cache_instance_id": schema.StringAttribute{Computed: true},
						"size_mb":           schema.Float64Attribute{Computed: true},
						"created_at":        schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *CacheSnapshotsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CacheSnapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured Client", "Expected configured client.")
		return
	}

	var data CacheSnapshotsDataSourceModel

	snapshots, err := d.client.ListCacheSnapshots(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list cache snapshots", err.Error())
		return
	}

	data.Snapshots = make([]CacheSnapshotModel, len(snapshots))
	for i, s := range snapshots {
		data.Snapshots[i] = CacheSnapshotModel{
			ID:              types.StringValue(strconv.FormatInt(s.ID, 10)),
			Name:            types.StringValue(s.Name),
			Description:     types.StringValue(s.Description),
			Status:          types.StringValue(s.Status),
			CacheInstanceID: types.StringValue(s.CacheInstanceID),
			SizeMB:          types.Float64Value(s.SizeMB),
			CreatedAt:       types.StringValue(s.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
