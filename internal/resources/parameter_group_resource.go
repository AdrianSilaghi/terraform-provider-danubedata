package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ParameterGroupResource{}
	_ resource.ResourceWithConfigure   = &ParameterGroupResource{}
	_ resource.ResourceWithImportState = &ParameterGroupResource{}
)

type ParameterGroupResource struct {
	client *client.Client
}

type ParameterGroupResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	ProviderType     types.String `tfsdk:"provider_type"`
	Family           types.String `tfsdk:"family"`
	Description      types.String `tfsdk:"description"`
	Parameters       types.Map    `tfsdk:"parameters"`
	LockedParameters types.List   `tfsdk:"locked_parameters"`
	IsDefault        types.Bool   `tfsdk:"is_default"`
	IsActive         types.Bool   `tfsdk:"is_active"`
	IsSystem         types.Bool   `tfsdk:"is_system"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
}

func NewParameterGroupResource() resource.Resource {
	return &ParameterGroupResource{}
}

func (r *ParameterGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_parameter_group"
}

func (r *ParameterGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData parameter group for cache, database, or queue instances.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the parameter group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the parameter group.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Parameter group type: cache, database, or queue.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("cache", "database", "queue"),
				},
			},
			"provider_type": schema.StringAttribute{
				Description: "Provider type (e.g., redis, valkey, dragonfly, mysql, postgresql, mariadb).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"family": schema.StringAttribute{
				Description: "Optional family label (e.g., redis7.x).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Optional description of the parameter group.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"parameters": schema.MapAttribute{
				Description: "Key/value map of parameters. Values are strings; numeric and boolean values should be expressed as strings (e.g., \"10000\", \"true\").",
				Required:    true,
				ElementType: types.StringType,
			},
			"locked_parameters": schema.ListAttribute{
				Description: "List of parameter keys that cannot be overridden by instances using this group.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this is the default parameter group for the provider type.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"is_active": schema.BoolAttribute{
				Description: "Whether this parameter group is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"is_system": schema.BoolAttribute{
				Description: "Whether this is a system-managed parameter group. System groups cannot be modified or deleted.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the parameter group was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the parameter group was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *ParameterGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ParameterGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ParameterGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	paramMap := make(map[string]string)
	resp.Diagnostics.Append(data.Parameters.ElementsAs(ctx, &paramMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// nil (not []) so JSON `omitempty` actually omits the field when the user didn't
	// set it, preventing the API from interpreting an empty array as "clear all locks".
	var locked []string
	if !data.LockedParameters.IsNull() && !data.LockedParameters.IsUnknown() {
		resp.Diagnostics.Append(data.LockedParameters.ElementsAs(ctx, &locked, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	createReq := client.CreateParameterGroupRequest{
		Name:             data.Name.ValueString(),
		Type:             data.Type.ValueString(),
		ProviderType:     data.ProviderType.ValueString(),
		Parameters:       stringMapToInterface(paramMap),
		LockedParameters: locked,
	}

	if !data.Family.IsNull() && !data.Family.IsUnknown() {
		v := data.Family.ValueString()
		createReq.Family = &v
	}
	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		v := data.Description.ValueString()
		createReq.Description = &v
	}
	if !data.IsDefault.IsNull() && !data.IsDefault.IsUnknown() {
		v := data.IsDefault.ValueBool()
		createReq.IsDefault = &v
	}

	tflog.Debug(ctx, "Creating parameter group", map[string]interface{}{
		"name":          createReq.Name,
		"type":          createReq.Type,
		"provider_type": createReq.ProviderType,
	})

	pg, err := r.client.CreateParameterGroup(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create parameter group", err.Error())
		return
	}

	resp.Diagnostics.Append(r.mapParameterGroupToState(ctx, pg, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Parameter group created", map[string]interface{}{
		"id":   pg.ID,
		"name": pg.Name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ParameterGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ParameterGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pg, err := r.client.GetParameterGroup(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read parameter group", err.Error())
		return
	}

	resp.Diagnostics.Append(r.mapParameterGroupToState(ctx, pg, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ParameterGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ParameterGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	paramMap := make(map[string]string)
	resp.Diagnostics.Append(data.Parameters.ElementsAs(ctx, &paramMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// nil (not []) so JSON `omitempty` actually omits the field when the user didn't
	// set it, preventing the API from interpreting an empty array as "clear all locks".
	var locked []string
	if !data.LockedParameters.IsNull() && !data.LockedParameters.IsUnknown() {
		resp.Diagnostics.Append(data.LockedParameters.ElementsAs(ctx, &locked, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	name := data.Name.ValueString()
	updateReq := client.UpdateParameterGroupRequest{
		Name:             &name,
		Parameters:       stringMapToInterface(paramMap),
		LockedParameters: locked,
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		v := data.Description.ValueString()
		updateReq.Description = &v
	}
	if !data.IsDefault.IsNull() && !data.IsDefault.IsUnknown() {
		v := data.IsDefault.ValueBool()
		updateReq.IsDefault = &v
	}
	if !data.IsActive.IsNull() && !data.IsActive.IsUnknown() {
		v := data.IsActive.ValueBool()
		updateReq.IsActive = &v
	}

	pg, err := r.client.UpdateParameterGroup(ctx, data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update parameter group", err.Error())
		return
	}

	resp.Diagnostics.Append(r.mapParameterGroupToState(ctx, pg, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ParameterGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ParameterGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteParameterGroup(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete parameter group", err.Error())
		return
	}
}

func (r *ParameterGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ParameterGroupResource) mapParameterGroupToState(ctx context.Context, pg *client.ParameterGroup, data *ParameterGroupResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	data.ID = types.StringValue(fmt.Sprintf("%d", pg.ID))
	data.Name = types.StringValue(pg.Name)
	data.Type = types.StringValue(pg.Type)
	data.ProviderType = types.StringValue(pg.ProviderType)
	if pg.Family != nil {
		data.Family = types.StringValue(*pg.Family)
	} else {
		data.Family = types.StringNull()
	}
	if pg.Description != nil {
		data.Description = types.StringValue(*pg.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.IsDefault = types.BoolValue(pg.IsDefault)
	data.IsActive = types.BoolValue(pg.IsActive)
	data.IsSystem = types.BoolValue(pg.IsSystem)
	if pg.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*pg.CreatedAt)
	} else {
		data.CreatedAt = types.StringNull()
	}
	if pg.UpdatedAt != nil {
		data.UpdatedAt = types.StringValue(*pg.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	paramMap := make(map[string]string, len(pg.Parameters))
	for k, v := range pg.Parameters {
		paramMap[k] = stringifyParameterValue(v)
	}
	mapValue, mapDiags := types.MapValueFrom(ctx, types.StringType, paramMap)
	diags.Append(mapDiags...)
	if mapDiags.HasError() {
		return diags
	}
	data.Parameters = mapValue

	lockedSlice := pg.LockedParameters
	if lockedSlice == nil {
		lockedSlice = []string{}
	}
	lockedValue, listDiags := types.ListValueFrom(ctx, types.StringType, lockedSlice)
	diags.Append(listDiags...)
	if listDiags.HasError() {
		return diags
	}
	data.LockedParameters = lockedValue

	return diags
}

// stringMapToInterface converts a map[string]string to map[string]interface{} for the API.
func stringMapToInterface(m map[string]string) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// stringifyParameterValue renders a parameter value (possibly number/bool/null) as a string.
func stringifyParameterValue(v interface{}) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", x)
	}
}
