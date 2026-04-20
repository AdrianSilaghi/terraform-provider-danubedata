package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DatabaseSnapshotsDataSource{}
var _ datasource.DataSourceWithConfigure = &DatabaseSnapshotsDataSource{}

type DatabaseSnapshotsDataSource struct {
	client *client.Client
}

type DatabaseSnapshotsDataSourceModel struct {
	Snapshots []DatabaseSnapshotModel `tfsdk:"snapshots"`
}

type DatabaseSnapshotModel struct {
	ID                 types.String  `tfsdk:"id"`
	Name               types.String  `tfsdk:"name"`
	Description        types.String  `tfsdk:"description"`
	Status             types.String  `tfsdk:"status"`
	DatabaseInstanceID types.String  `tfsdk:"database_instance_id"`
	SizeGB             types.Float64 `tfsdk:"size_gb"`
	CreatedAt          types.String  `tfsdk:"created_at"`
}

func NewDatabaseSnapshotsDataSource() datasource.DataSource {
	return &DatabaseSnapshotsDataSource{}
}

func (d *DatabaseSnapshotsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_snapshots"
}

func (d *DatabaseSnapshotsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all database snapshots in your account.",
		Attributes: map[string]schema.Attribute{
			"snapshots": schema.ListNestedAttribute{
				Description: "List of database snapshots.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                   schema.StringAttribute{Computed: true},
						"name":                 schema.StringAttribute{Computed: true},
						"description":          schema.StringAttribute{Computed: true},
						"status":               schema.StringAttribute{Computed: true},
						"database_instance_id": schema.StringAttribute{Computed: true},
						"size_gb":              schema.Float64Attribute{Computed: true},
						"created_at":           schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *DatabaseSnapshotsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabaseSnapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured Client", "Expected configured client.")
		return
	}

	var data DatabaseSnapshotsDataSourceModel

	snapshots, err := d.client.ListDatabaseSnapshots(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list database snapshots", err.Error())
		return
	}

	data.Snapshots = make([]DatabaseSnapshotModel, len(snapshots))
	for i, s := range snapshots {
		data.Snapshots[i] = DatabaseSnapshotModel{
			ID:                 types.StringValue(s.ID),
			Name:               types.StringValue(s.Name),
			Description:        types.StringValue(s.Description),
			Status:             types.StringValue(s.Status),
			DatabaseInstanceID: types.StringValue(s.DatabaseInstanceID),
			SizeGB:             types.Float64Value(s.SizeGB),
			CreatedAt:          types.StringValue(s.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
