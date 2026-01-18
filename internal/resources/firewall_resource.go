package resources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &FirewallResource{}
	_ resource.ResourceWithConfigure   = &FirewallResource{}
	_ resource.ResourceWithImportState = &FirewallResource{}
)

type FirewallResource struct {
	client *client.Client
}

type FirewallResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Status        types.String `tfsdk:"status"`
	IsDefault     types.Bool   `tfsdk:"is_default"`
	DefaultAction types.String `tfsdk:"default_action"`
	Rules         types.List   `tfsdk:"rules"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

type FirewallRuleModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Action         types.String `tfsdk:"action"`
	Direction      types.String `tfsdk:"direction"`
	Protocol       types.String `tfsdk:"protocol"`
	PortRangeStart types.Int64  `tfsdk:"port_range_start"`
	PortRangeEnd   types.Int64  `tfsdk:"port_range_end"`
	SourceIPs      types.List   `tfsdk:"source_ips"`
	Priority       types.Int64  `tfsdk:"priority"`
}

func NewFirewallResource() resource.Resource {
	return &FirewallResource{}
}

func (r *FirewallResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

func (r *FirewallResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData firewall for network security.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the firewall.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the firewall.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the firewall.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the firewall (draft, active, deploying).",
				Computed:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this is the default firewall for the team.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"default_action": schema.StringAttribute{
				Description: "Default action for traffic not matching any rule: 'drop' or 'accept'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("drop"),
				Validators: []validator.String{
					stringvalidator.OneOf("drop", "accept"),
				},
			},
			"rules": schema.ListNestedAttribute{
				Description: "List of firewall rules.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Rule ID (computed by API).",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name/description of the rule.",
							Optional:    true,
						},
						"action": schema.StringAttribute{
							Description: "Action to take: 'accept' or 'drop'.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("accept", "drop"),
							},
						},
						"direction": schema.StringAttribute{
							Description: "Direction: 'inbound' or 'outbound'.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("inbound", "outbound"),
							},
						},
						"protocol": schema.StringAttribute{
							Description: "Protocol: 'tcp', 'udp', 'icmp', or 'any'.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("tcp", "udp", "icmp", "any"),
							},
						},
						"port_range_start": schema.Int64Attribute{
							Description: "Start of port range (1-65535).",
							Optional:    true,
						},
						"port_range_end": schema.Int64Attribute{
							Description: "End of port range (1-65535).",
							Optional:    true,
						},
						"source_ips": schema.ListAttribute{
							Description: "List of source IP addresses or CIDR blocks.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"priority": schema.Int64Attribute{
							Description: "Rule priority (lower numbers = higher priority).",
							Optional:    true,
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the firewall was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the firewall was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *FirewallResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *FirewallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FirewallResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating firewall", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	createReq := client.CreateFirewallRequest{
		Name:          data.Name.ValueString(),
		Description:   data.Description.ValueString(),
		IsDefault:     data.IsDefault.ValueBool(),
		DefaultAction: data.DefaultAction.ValueString(),
	}

	// Convert rules
	if !data.Rules.IsNull() && !data.Rules.IsUnknown() {
		var rules []FirewallRuleModel
		resp.Diagnostics.Append(data.Rules.ElementsAs(ctx, &rules, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		createReq.Rules = make([]client.CreateFirewallRuleRequest, len(rules))
		for i, rule := range rules {
			ruleReq := client.CreateFirewallRuleRequest{
				Name:      rule.Name.ValueString(),
				Action:    rule.Action.ValueString(),
				Direction: rule.Direction.ValueString(),
				Protocol:  rule.Protocol.ValueString(),
				Priority:  int(rule.Priority.ValueInt64()),
			}

			if !rule.PortRangeStart.IsNull() {
				port := int(rule.PortRangeStart.ValueInt64())
				ruleReq.PortRangeStart = &port
			}
			if !rule.PortRangeEnd.IsNull() {
				port := int(rule.PortRangeEnd.ValueInt64())
				ruleReq.PortRangeEnd = &port
			}
			if !rule.SourceIPs.IsNull() {
				var sourceIPs []string
				resp.Diagnostics.Append(rule.SourceIPs.ElementsAs(ctx, &sourceIPs, false)...)
				ruleReq.SourceIPs = sourceIPs
			}

			createReq.Rules[i] = ruleReq
		}
	}

	firewall, err := r.client.CreateFirewall(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create firewall", err.Error())
		return
	}

	r.mapFirewallToState(ctx, firewall, &data, &resp.Diagnostics)

	tflog.Info(ctx, "Firewall created", map[string]interface{}{
		"id":   firewall.ID,
		"name": firewall.Name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FirewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	firewall, err := r.client.GetFirewall(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read firewall", err.Error())
		return
	}

	r.mapFirewallToState(ctx, firewall, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FirewallResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating firewall", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	updateReq := client.UpdateFirewallRequest{
		Name:          data.Name.ValueString(),
		Description:   data.Description.ValueString(),
		DefaultAction: data.DefaultAction.ValueString(),
	}

	if !data.IsDefault.IsNull() {
		isDefault := data.IsDefault.ValueBool()
		updateReq.IsDefault = &isDefault
	}

	firewall, err := r.client.UpdateFirewall(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update firewall", err.Error())
		return
	}

	r.mapFirewallToState(ctx, firewall, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FirewallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FirewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting firewall", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	err := r.client.DeleteFirewall(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete firewall", err.Error())
		return
	}
}

func (r *FirewallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *FirewallResource) mapFirewallToState(ctx context.Context, firewall *client.Firewall, data *FirewallResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(firewall.ID)
	data.Name = types.StringValue(firewall.Name)
	data.Description = types.StringValue(firewall.Description)
	data.Status = types.StringValue(firewall.Status)
	data.IsDefault = types.BoolValue(firewall.IsDefault)
	data.DefaultAction = types.StringValue(firewall.DefaultAction)
	data.CreatedAt = types.StringValue(firewall.CreatedAt)
	data.UpdatedAt = types.StringValue(firewall.UpdatedAt)

	// Map rules
	if len(firewall.Rules) > 0 {
		ruleObjects := make([]attr.Value, len(firewall.Rules))
		for i, rule := range firewall.Rules {
			sourceIPsValues := make([]attr.Value, len(rule.SourceIPs))
			for j, ip := range rule.SourceIPs {
				sourceIPsValues[j] = types.StringValue(ip)
			}
			sourceIPsList, _ := types.ListValue(types.StringType, sourceIPsValues)

			var portStart, portEnd types.Int64
			if rule.PortRangeStart != nil {
				portStart = types.Int64Value(int64(*rule.PortRangeStart))
			} else {
				portStart = types.Int64Null()
			}
			if rule.PortRangeEnd != nil {
				portEnd = types.Int64Value(int64(*rule.PortRangeEnd))
			} else {
				portEnd = types.Int64Null()
			}

			ruleObj, _ := types.ObjectValue(
				map[string]attr.Type{
					"id":               types.StringType,
					"name":             types.StringType,
					"action":           types.StringType,
					"direction":        types.StringType,
					"protocol":         types.StringType,
					"port_range_start": types.Int64Type,
					"port_range_end":   types.Int64Type,
					"source_ips":       types.ListType{ElemType: types.StringType},
					"priority":         types.Int64Type,
				},
				map[string]attr.Value{
					"id":               types.StringValue(rule.ID),
					"name":             types.StringValue(rule.Name),
					"action":           types.StringValue(rule.Action),
					"direction":        types.StringValue(rule.Direction),
					"protocol":         types.StringValue(rule.Protocol),
					"port_range_start": portStart,
					"port_range_end":   portEnd,
					"source_ips":       sourceIPsList,
					"priority":         types.Int64Value(int64(rule.Priority)),
				},
			)
			ruleObjects[i] = ruleObj
		}

		rulesList, diagsRules := types.ListValue(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"id":               types.StringType,
					"name":             types.StringType,
					"action":           types.StringType,
					"direction":        types.StringType,
					"protocol":         types.StringType,
					"port_range_start": types.Int64Type,
					"port_range_end":   types.Int64Type,
					"source_ips":       types.ListType{ElemType: types.StringType},
					"priority":         types.Int64Type,
				},
			},
			ruleObjects,
		)
		diags.Append(diagsRules...)
		data.Rules = rulesList
	} else {
		data.Rules = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":               types.StringType,
				"name":             types.StringType,
				"action":           types.StringType,
				"direction":        types.StringType,
				"protocol":         types.StringType,
				"port_range_start": types.Int64Type,
				"port_range_end":   types.Int64Type,
				"source_ips":       types.ListType{ElemType: types.StringType},
				"priority":         types.Int64Type,
			},
		})
	}
}
