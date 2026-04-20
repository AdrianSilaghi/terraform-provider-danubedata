package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ParameterGroupsDataSource{}
var _ datasource.DataSourceWithConfigure = &ParameterGroupsDataSource{}

type ParameterGroupsDataSource struct {
	client *client.Client
}

type ParameterGroupsDataSourceModel struct {
	Type         types.String         `tfsdk:"type"`
	ProviderType types.String         `tfsdk:"provider_type"`
	Groups       []ParameterGroupItem `tfsdk:"groups"`
}

type ParameterGroupItem struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	ProviderType types.String `tfsdk:"provider_type"`
	Family       types.String `tfsdk:"family"`
	Description  types.String `tfsdk:"description"`
	IsDefault    types.Bool   `tfsdk:"is_default"`
	IsActive     types.Bool   `tfsdk:"is_active"`
	IsSystem     types.Bool   `tfsdk:"is_system"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

func NewParameterGroupsDataSource() datasource.DataSource {
	return &ParameterGroupsDataSource{}
}

func (d *ParameterGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_parameter_groups"
}

func (d *ParameterGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists parameter groups available for cache, database, or queue instances (includes system groups).",
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Description: "Filter by parameter group type (cache, database, queue).",
				Optional:    true,
			},
			"provider_type": schema.StringAttribute{
				Description: "Filter by provider type (e.g., redis, mysql).",
				Optional:    true,
			},
			"groups": schema.ListNestedAttribute{
				Description: "List of parameter groups matching the filters.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":            schema.StringAttribute{Computed: true},
						"name":          schema.StringAttribute{Computed: true},
						"type":          schema.StringAttribute{Computed: true},
						"provider_type": schema.StringAttribute{Computed: true},
						"family":        schema.StringAttribute{Computed: true},
						"description":   schema.StringAttribute{Computed: true},
						"is_default":    schema.BoolAttribute{Computed: true},
						"is_active":     schema.BoolAttribute{Computed: true},
						"is_system":     schema.BoolAttribute{Computed: true},
						"created_at":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *ParameterGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ParameterGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ParameterGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := client.ListParameterGroupsOptions{}
	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		opts.Type = data.Type.ValueString()
	}
	if !data.ProviderType.IsNull() && !data.ProviderType.IsUnknown() {
		opts.ProviderType = data.ProviderType.ValueString()
	}

	groups, err := d.client.ListParameterGroups(ctx, opts)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list parameter groups", err.Error())
		return
	}

	data.Groups = make([]ParameterGroupItem, len(groups))
	for i, g := range groups {
		item := ParameterGroupItem{
			ID:           types.StringValue(fmt.Sprintf("%d", g.ID)),
			Name:         types.StringValue(g.Name),
			Type:         types.StringValue(g.Type),
			ProviderType: types.StringValue(g.ProviderType),
			IsDefault:    types.BoolValue(g.IsDefault),
			IsActive:     types.BoolValue(g.IsActive),
			IsSystem:     types.BoolValue(g.IsSystem),
		}
		if g.Family != nil {
			item.Family = types.StringValue(*g.Family)
		} else {
			item.Family = types.StringNull()
		}
		if g.Description != nil {
			item.Description = types.StringValue(*g.Description)
		} else {
			item.Description = types.StringNull()
		}
		if g.CreatedAt != nil {
			item.CreatedAt = types.StringValue(*g.CreatedAt)
		} else {
			item.CreatedAt = types.StringNull()
		}
		data.Groups[i] = item
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
