package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &StaticSitesDataSource{}
var _ datasource.DataSourceWithConfigure = &StaticSitesDataSource{}

type StaticSitesDataSource struct {
	client *client.Client
}

type StaticSitesDataSourceModel struct {
	TeamID types.Int64       `tfsdk:"team_id"`
	Sites  []StaticSiteModel `tfsdk:"sites"`
}

type StaticSiteModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Slug      types.String `tfsdk:"slug"`
	URL       types.String `tfsdk:"url"`
	Plan      types.String `tfsdk:"plan"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewStaticSitesDataSource() datasource.DataSource {
	return &StaticSitesDataSource{}
}

func (d *StaticSitesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_static_sites"
}

func (d *StaticSitesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all static sites for a given team.",
		Attributes: map[string]schema.Attribute{
			"team_id": schema.Int64Attribute{
				Description: "ID of the team to list static sites for.",
				Required:    true,
			},
			"sites": schema.ListNestedAttribute{
				Description: "List of static sites.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":         schema.StringAttribute{Computed: true},
						"name":       schema.StringAttribute{Computed: true},
						"slug":       schema.StringAttribute{Computed: true},
						"url":        schema.StringAttribute{Computed: true},
						"plan":       schema.StringAttribute{Computed: true},
						"status":     schema.StringAttribute{Computed: true},
						"created_at": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *StaticSitesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StaticSitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StaticSitesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sites, err := d.client.ListStaticSites(ctx, int(data.TeamID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Failed to list static sites", err.Error())
		return
	}

	data.Sites = make([]StaticSiteModel, len(sites))
	for i, s := range sites {
		data.Sites[i] = StaticSiteModel{
			ID:        types.StringValue(s.ID),
			Name:      types.StringValue(s.Name),
			Slug:      types.StringValue(s.Slug),
			URL:       types.StringValue(s.URL),
			Plan:      types.StringValue(s.Plan),
			Status:    types.StringValue(s.Status),
			CreatedAt: types.StringValue(s.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
