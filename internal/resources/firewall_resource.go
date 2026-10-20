package resources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	IsDefault     types.String `tfsdk:"is_default"`
	IsSystem      types.Bool   `tfsdk:"is_system"`
	DefaultAction types.String `tfsdk:"default_action"`
	TeamID        types.Int64  `tfsdk:"team_id"`
	Rules         types.List   `tfsdk:"rules"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
}

type FirewallRuleModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Action            types.String `tfsdk:"action"`
	Direction         types.String `tfsdk:"direction"`
	Protocol          types.String `tfsdk:"protocol"`
	Port              types.Int64  `tfsdk:"port"`
	IsLocked          types.Bool   `tfsdk:"is_locked"`
	SourceInstanceIDs types.List   `tfsdk:"source_instance_ids"`
	PortRangeStart    types.Int64  `tfsdk:"port_range_start"`
	PortRangeEnd      types.Int64  `tfsdk:"port_range_end"`
	SourceIPs         types.List   `tfsdk:"source_ips"`
	Order             types.Int64  `tfsdk:"order"`
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
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 255)},
			},
			"description": schema.StringAttribute{
				Description: "Description of the firewall.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtMost(1000)},
			},
			"status": schema.StringAttribute{
				Description:   "Current status of the firewall (draft, active, deploying).",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"is_default":     schema.StringAttribute{Description: "Default firewall marker.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"is_system":      schema.BoolAttribute{Description: "Platform-managed firewall that cannot be detached or deleted.", Computed: true, PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}},
			"default_action": schema.StringAttribute{Description: "Default firewall action.", Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"team_id":        schema.Int64Attribute{Description: "Owning project ID.", Computed: true, PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()}},
			"rules": schema.ListNestedAttribute{
				Description:   "List of firewall rules.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description:   "Rule ID (computed by API).",
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"name": schema.StringAttribute{
							Description:   "Name/description of the rule.",
							Optional:      true,
							Computed:      true,
							PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
						},
						"action": schema.StringAttribute{
							Description: "Action to take: 'allow' or 'deny'.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("allow", "deny"),
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
							Description: "Protocol: 'tcp', 'udp', 'icmp', 'any', 'gre', or 'esp'.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("tcp", "udp", "icmp", "any", "gre", "esp"),
							},
						},
						"port": schema.Int64Attribute{
							Description: "Single destination port (1-65535).",
							Optional:    true,
							Validators:  []validator.Int64{int64validator.Between(1, 65535)},
						},
						"is_locked": schema.BoolAttribute{
							Description:   "Platform-managed rule that the API preserves on update.",
							Computed:      true,
							PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
						},
						"source_instance_ids": schema.ListNestedAttribute{
							Description:   "Source resources this rule admits traffic from.",
							Optional:      true,
							Computed:      true,
							PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
							NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
								"id":   schema.StringAttribute{Required: true},
								"type": schema.StringAttribute{Required: true},
							}},
						},
						"port_range_start": schema.Int64Attribute{
							Description: "Start of port range (1-65535).",
							Optional:    true,
							Validators:  []validator.Int64{int64validator.Between(1, 65535)},
						},
						"port_range_end": schema.Int64Attribute{
							Description: "End of port range (1-65535).",
							Optional:    true,
							Validators:  []validator.Int64{int64validator.Between(1, 65535)},
						},
						"source_ips": schema.ListAttribute{
							Description:   "List of source IP addresses or CIDR blocks.",
							Optional:      true,
							Computed:      true,
							PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
							ElementType:   types.StringType,
						},
						"order": schema.Int64Attribute{
							Description:   "Rule evaluation order (lower numbers are evaluated first).",
							Optional:      true,
							Computed:      true,
							PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
							Validators:    []validator.Int64{int64validator.AtLeast(0)},
						},
					},
				},
			},
			"created_at": schema.StringAttribute{
				Description:   "Timestamp when the firewall was created.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at": schema.StringAttribute{
				Description:   "Timestamp when the firewall was last updated.",
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
	}

	if !data.Rules.IsNull() && !data.Rules.IsUnknown() {
		rules, diags := firewallRulesToRequest(ctx, data.Rules)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Rules = rules
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
	var state FirewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateFirewallRequest{
		Name: data.Name.ValueString(),
	}
	if !data.Description.Equal(state.Description) {
		var description *string
		if !data.Description.IsNull() {
			v := data.Description.ValueString()
			description = &v
		}
		updateReq.Description = &description
	}

	rulesChanged, compareDiags := firewallRulesChanged(ctx, data.Rules, state.Rules)
	resp.Diagnostics.Append(compareDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if rulesChanged {
		rules, diags := firewallRulesToRequest(ctx, data.Rules, state.Rules)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.Rules = &rules
	}

	firewall, err := r.client.UpdateFirewall(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update firewall", err.Error())
		return
	}

	r.mapFirewallToState(ctx, firewall, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func firewallRulesChanged(ctx context.Context, planned, previous types.List) (bool, diag.Diagnostics) {
	if planned.IsNull() || planned.IsUnknown() {
		return false, nil
	}
	var planRules, stateRules []FirewallRuleModel
	diags := planned.ElementsAs(ctx, &planRules, false)
	if !previous.IsNull() && !previous.IsUnknown() {
		diags.Append(previous.ElementsAs(ctx, &stateRules, false)...)
	}
	if diags.HasError() {
		return false, diags
	}
	if len(planRules) != len(stateRules) {
		return true, diags
	}
	for i, plan := range planRules {
		state := stateRules[i]
		if (!plan.Name.IsUnknown() && !plan.Name.Equal(state.Name)) ||
			!plan.Action.Equal(state.Action) || !plan.Direction.Equal(state.Direction) || !plan.Protocol.Equal(state.Protocol) ||
			(!plan.Port.IsUnknown() && !plan.Port.Equal(state.Port)) ||
			(!plan.PortRangeStart.IsUnknown() && !plan.PortRangeStart.Equal(state.PortRangeStart)) ||
			(!plan.PortRangeEnd.IsUnknown() && !plan.PortRangeEnd.Equal(state.PortRangeEnd)) ||
			(!plan.SourceIPs.IsUnknown() && !plan.SourceIPs.Equal(state.SourceIPs)) ||
			(!plan.SourceInstanceIDs.IsUnknown() && !plan.SourceInstanceIDs.Equal(state.SourceInstanceIDs)) ||
			(!plan.Order.IsUnknown() && !plan.Order.Equal(state.Order)) {
			return true, diags
		}
	}
	return false, diags
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

func firewallRulesToRequest(ctx context.Context, value types.List, prior ...types.List) ([]client.CreateFirewallRuleRequest, diag.Diagnostics) {
	var rules []FirewallRuleModel
	diags := value.ElementsAs(ctx, &rules, false)
	if diags.HasError() {
		return nil, diags
	}
	var previous []FirewallRuleModel
	if len(prior) > 0 && !prior[0].IsNull() && !prior[0].IsUnknown() {
		diags.Append(prior[0].ElementsAs(ctx, &previous, false)...)
		if diags.HasError() {
			return nil, diags
		}
	}
	result := make([]client.CreateFirewallRuleRequest, len(rules))
	for i, rule := range rules {
		match := -1
		if !rule.Name.IsNull() && !rule.Name.IsUnknown() {
			for j, old := range previous {
				if rule.Name.Equal(old.Name) && rule.Action.Equal(old.Action) && rule.Direction.Equal(old.Direction) && rule.Protocol.Equal(old.Protocol) {
					if match != -1 {
						match = -1 // Duplicate identities cannot be matched safely.
						break
					}
					match = j
				}
			}
		}
		if match >= 0 {
			old := previous[match]
			if !rule.ID.IsUnknown() && !rule.ID.IsNull() && !rule.ID.Equal(old.ID) {
				diags.AddError("Ambiguous firewall rule identity", "The planned rule ID does not match the rule identified by its name and protocol. Set source restrictions explicitly or apply rule changes one at a time.")
				continue
			}
			if rule.ID.IsUnknown() {
				rule.ID = old.ID
			}
			if rule.Name.IsUnknown() {
				rule.Name = old.Name
			}
			if rule.Order.IsUnknown() {
				rule.Order = old.Order
			}
			if rule.SourceIPs.IsUnknown() {
				rule.SourceIPs = old.SourceIPs
			}
			if rule.SourceInstanceIDs.IsUnknown() {
				rule.SourceInstanceIDs = old.SourceInstanceIDs
			}
		} else if len(previous) > 0 && (rule.SourceIPs.IsUnknown() || rule.SourceInstanceIDs.IsUnknown()) {
			diags.AddError("Unresolved firewall rule restrictions", "A rule has unknown source restrictions and no unambiguous prior rule identity. Set source_ips and source_instance_ids explicitly before changing this rule.")
			continue
		}
		item := client.CreateFirewallRuleRequest{
			ID: rule.ID.ValueString(), Name: rule.Name.ValueString(), Action: rule.Action.ValueString(),
			Direction: rule.Direction.ValueString(), Protocol: rule.Protocol.ValueString(), Order: int(rule.Order.ValueInt64()),
		}
		if !rule.Port.IsNull() && !rule.Port.IsUnknown() {
			v := int(rule.Port.ValueInt64())
			item.Port = &v
		}
		if !rule.PortRangeStart.IsNull() && !rule.PortRangeStart.IsUnknown() {
			v := int(rule.PortRangeStart.ValueInt64())
			item.PortRangeStart = &v
		}
		if !rule.PortRangeEnd.IsNull() && !rule.PortRangeEnd.IsUnknown() {
			v := int(rule.PortRangeEnd.ValueInt64())
			item.PortRangeEnd = &v
		}
		if !rule.SourceIPs.IsNull() && !rule.SourceIPs.IsUnknown() {
			diags.Append(rule.SourceIPs.ElementsAs(ctx, &item.SourceIPs, false)...)
		}
		if !rule.SourceInstanceIDs.IsNull() && !rule.SourceInstanceIDs.IsUnknown() {
			var sources []struct {
				ID   types.String `tfsdk:"id"`
				Type types.String `tfsdk:"type"`
			}
			diags.Append(rule.SourceInstanceIDs.ElementsAs(ctx, &sources, false)...)
			for _, source := range sources {
				if source.Type.IsNull() || source.Type.IsUnknown() {
					diags.AddError("Unresolved firewall source", "A source instance has no resolvable type. Remove it explicitly before updating rules.")
					continue
				}
				sourceType := source.Type.ValueString()
				item.SourceInstanceIDs = append(item.SourceInstanceIDs, client.FirewallSourceInstance{ID: source.ID.ValueString(), Type: &sourceType})
			}
		}
		result[i] = item
	}
	return result, diags
}

var firewallSourceTypes = map[string]attr.Type{"id": types.StringType, "type": types.StringType}
var firewallRuleTypes = map[string]attr.Type{
	"id": types.StringType, "name": types.StringType, "action": types.StringType,
	"direction": types.StringType, "protocol": types.StringType, "port": types.Int64Type,
	"port_range_start": types.Int64Type, "port_range_end": types.Int64Type,
	"source_ips":          types.ListType{ElemType: types.StringType},
	"source_instance_ids": types.ListType{ElemType: types.ObjectType{AttrTypes: firewallSourceTypes}},
	"order":               types.Int64Type, "is_locked": types.BoolType,
}

func (r *FirewallResource) mapFirewallToState(ctx context.Context, firewall *client.Firewall, data *FirewallResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(firewall.ID)
	data.Name = types.StringValue(firewall.Name)
	if firewall.Description == "" && data.Description.IsNull() {
		data.Description = types.StringNull()
	} else {
		data.Description = types.StringValue(firewall.Description)
	}
	data.Status = types.StringValue(firewall.Status)
	data.IsDefault = types.StringValue(firewall.IsDefault)
	data.IsSystem = types.BoolValue(firewall.IsSystem)
	data.DefaultAction = types.StringValue(firewall.DefaultAction)
	data.TeamID = types.Int64Value(int64(firewall.TeamID))
	data.CreatedAt = types.StringValue(firewall.CreatedAt)
	data.UpdatedAt = types.StringValue(firewall.UpdatedAt)

	values := make([]attr.Value, 0, len(firewall.Rules))
	for _, rule := range firewall.Rules {
		// The API retains platform rules even when clients submit a new rule list.
		if rule.IsLocked {
			continue
		}
		ipValues := make([]attr.Value, len(rule.SourceIPs))
		for j, ip := range rule.SourceIPs {
			ipValues[j] = types.StringValue(ip)
		}
		ips, d := types.ListValue(types.StringType, ipValues)
		diags.Append(d...)
		sourceValues := make([]attr.Value, len(rule.SourceInstanceIDs))
		for j, source := range rule.SourceInstanceIDs {
			sourceType := types.StringNull()
			if source.Type != nil {
				sourceType = types.StringValue(*source.Type)
			}
			sourceValues[j], d = types.ObjectValue(firewallSourceTypes, map[string]attr.Value{
				"id": types.StringValue(source.ID), "type": sourceType,
			})
			diags.Append(d...)
		}
		sources, d := types.ListValue(types.ObjectType{AttrTypes: firewallSourceTypes}, sourceValues)
		diags.Append(d...)
		port, portStart, portEnd := types.Int64Null(), types.Int64Null(), types.Int64Null()
		if rule.Port != nil {
			port = types.Int64Value(int64(*rule.Port))
		}
		if rule.PortRangeStart != nil {
			portStart = types.Int64Value(int64(*rule.PortRangeStart))
		}
		if rule.PortRangeEnd != nil {
			portEnd = types.Int64Value(int64(*rule.PortRangeEnd))
		}
		value, d := types.ObjectValue(firewallRuleTypes, map[string]attr.Value{
			"id": types.StringValue(rule.ID), "name": types.StringValue(rule.Name),
			"action": types.StringValue(rule.Action), "direction": types.StringValue(rule.Direction),
			"protocol": types.StringValue(rule.Protocol), "port": port,
			"port_range_start": portStart, "port_range_end": portEnd,
			"source_ips": ips, "source_instance_ids": sources,
			"order": types.Int64Value(int64(rule.Order)), "is_locked": types.BoolValue(rule.IsLocked),
		})
		diags.Append(d...)
		values = append(values, value)
	}
	if len(values) == 0 && data.Rules.IsNull() {
		data.Rules = types.ListNull(types.ObjectType{AttrTypes: firewallRuleTypes})
		return
	}
	data.Rules, _ = types.ListValue(types.ObjectType{AttrTypes: firewallRuleTypes}, values)
}
