package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FirewallsDataSource{}
var _ datasource.DataSourceWithConfigure = &FirewallsDataSource{}

type FirewallsDataSource struct {
	client *client.Client
}

type FirewallsDataSourceModel struct {
	Firewalls []FirewallModel `tfsdk:"firewalls"`
}

type FirewallModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Status        types.String `tfsdk:"status"`
	DefaultAction types.String `tfsdk:"default_action"`
	IsDefault     types.Bool   `tfsdk:"is_default"`
	RulesCount    types.Int64  `tfsdk:"rules_count"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func NewFirewallsDataSource() datasource.DataSource {
	return &FirewallsDataSource{}
}

func (d *FirewallsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewalls"
}

func (d *FirewallsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all firewalls in your account.",
		Attributes: map[string]schema.Attribute{
			"firewalls": schema.ListNestedAttribute{
				Description: "List of firewalls.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier for the firewall.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the firewall.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the firewall.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the firewall.",
							Computed:    true,
						},
						"default_action": schema.StringAttribute{
							Description: "Default action for unmatched traffic (allow or deny).",
							Computed:    true,
						},
						"is_default": schema.BoolAttribute{
							Description: "Whether this is the default firewall.",
							Computed:    true,
						},
						"rules_count": schema.Int64Attribute{
							Description: "Number of rules in the firewall.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the firewall was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *FirewallsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FirewallsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var data FirewallsDataSourceModel

	firewalls, err := d.client.ListFirewalls(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list firewalls", err.Error())
		return
	}

	data.Firewalls = make([]FirewallModel, len(firewalls))
	for i, fw := range firewalls {
		data.Firewalls[i] = FirewallModel{
			ID:            types.StringValue(fw.ID),
			Name:          types.StringValue(fw.Name),
			Description:   types.StringValue(fw.Description),
			Status:        types.StringValue(fw.Status),
			DefaultAction: types.StringValue(fw.DefaultAction),
			IsDefault:     types.BoolValue(fw.IsDefault),
			RulesCount:    types.Int64Value(int64(len(fw.Rules))),
			CreatedAt:     types.StringValue(fw.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
