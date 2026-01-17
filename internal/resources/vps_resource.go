package resources

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &VpsResource{}
	_ resource.ResourceWithConfigure   = &VpsResource{}
	_ resource.ResourceWithImportState = &VpsResource{}
)

type VpsResource struct {
	client *client.Client
}

type VpsResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	Status            types.String   `tfsdk:"status"`
	ResourceProfile   types.String   `tfsdk:"resource_profile"`
	CPUAllocationType types.String   `tfsdk:"cpu_allocation_type"`
	Image             types.String   `tfsdk:"image"`
	Datacenter        types.String   `tfsdk:"datacenter"`
	NetworkStack      types.String   `tfsdk:"network_stack"`
	AuthMethod        types.String   `tfsdk:"auth_method"`
	SSHKeyID          types.String   `tfsdk:"ssh_key_id"`
	Password          types.String   `tfsdk:"password"`
	CustomCloudInit   types.String   `tfsdk:"custom_cloud_init"`
	CPUCores          types.Int64    `tfsdk:"cpu_cores"`
	MemorySizeGB      types.Int64    `tfsdk:"memory_size_gb"`
	StorageSizeGB     types.Int64    `tfsdk:"storage_size_gb"`
	PublicIP          types.String   `tfsdk:"public_ip"`
	PrivateIP         types.String   `tfsdk:"private_ip"`
	IPv6Address       types.String   `tfsdk:"ipv6_address"`
	MonthlyCostCents  types.Int64    `tfsdk:"monthly_cost_cents"`
	MonthlyCost       types.Float64  `tfsdk:"monthly_cost"`
	DeployedAt        types.String   `tfsdk:"deployed_at"`
	CreatedAt         types.String   `tfsdk:"created_at"`
	UpdatedAt         types.String   `tfsdk:"updated_at"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

func NewVpsResource() resource.Resource {
	return &VpsResource{}
}

func (r *VpsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps"
}

func (r *VpsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData VPS instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the VPS instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the VPS instance. Must be lowercase alphanumeric with hyphens.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`),
						"must be lowercase alphanumeric with hyphens, starting and ending with alphanumeric",
					),
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"status": schema.StringAttribute{
				Description: "Current status of the VPS instance (pending, provisioning, running, stopped, error).",
				Computed:    true,
			},
			"resource_profile": schema.StringAttribute{
				Description: "Resource profile for the VPS (nano_shared, micro_shared, small_shared, medium_shared, large_shared, or dedicated variants).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("nano_shared"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"nano_shared", "micro_shared", "small_shared", "medium_shared", "large_shared",
						"nano", "micro", "small", "medium", "large",
					),
				},
			},
			"cpu_allocation_type": schema.StringAttribute{
				Description: "CPU allocation type: 'shared' or 'dedicated'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("shared"),
				Validators: []validator.String{
					stringvalidator.OneOf("shared", "dedicated"),
				},
			},
			"image": schema.StringAttribute{
				Description: "Operating system image (e.g., 'ubuntu-24.04', 'debian-12').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"datacenter": schema.StringAttribute{
				Description: "Datacenter location (fsn1, nbg1, hel1, ash).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("fsn1", "nbg1", "hel1", "ash"),
				},
			},
			"network_stack": schema.StringAttribute{
				Description: "Network stack: 'ipv4_only', 'ipv6_only', or 'dual_stack'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("dual_stack"),
				Validators: []validator.String{
					stringvalidator.OneOf("ipv4_only", "ipv6_only", "dual_stack"),
				},
			},
			"auth_method": schema.StringAttribute{
				Description: "Authentication method: 'ssh_key' or 'password'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("ssh_key", "password"),
				},
			},
			"ssh_key_id": schema.StringAttribute{
				Description: "SSH key ID for authentication (required if auth_method is 'ssh_key').",
				Optional:    true,
			},
			"password": schema.StringAttribute{
				Description: "Root password (required if auth_method is 'password'). Must be at least 12 characters.",
				Optional:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(12),
				},
			},
			"custom_cloud_init": schema.StringAttribute{
				Description: "Custom cloud-init configuration script.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(10000),
				},
			},
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
			"monthly_cost_cents": schema.Int64Attribute{
				Description: "Monthly cost in cents.",
				Computed:    true,
			},
			"monthly_cost": schema.Float64Attribute{
				Description: "Monthly cost in dollars.",
				Computed:    true,
			},
			"deployed_at": schema.StringAttribute{
				Description: "Timestamp when the VPS was deployed.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the VPS was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the VPS was last updated.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *VpsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VpsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VpsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get timeout
	createTimeout, diags := data.Timeouts.Create(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Build create request
	createReq := client.CreateVpsRequest{
		Name:            data.Name.ValueString(),
		ResourceProfile: data.ResourceProfile.ValueString(),
		Image:           data.Image.ValueString(),
		Datacenter:      data.Datacenter.ValueString(),
		AuthMethod:      data.AuthMethod.ValueString(),
	}

	if !data.CPUAllocationType.IsNull() && !data.CPUAllocationType.IsUnknown() {
		createReq.CPUAllocationType = data.CPUAllocationType.ValueString()
	}

	if !data.NetworkStack.IsNull() && !data.NetworkStack.IsUnknown() {
		createReq.NetworkStack = data.NetworkStack.ValueString()
	}

	if !data.SSHKeyID.IsNull() && !data.SSHKeyID.IsUnknown() {
		sshKeyID := data.SSHKeyID.ValueString()
		createReq.SSHKeyID = &sshKeyID
	}

	if !data.Password.IsNull() && !data.Password.IsUnknown() {
		password := data.Password.ValueString()
		createReq.Password = &password
		createReq.PasswordConfirm = &password
	}

	if !data.CustomCloudInit.IsNull() && !data.CustomCloudInit.IsUnknown() {
		cloudInit := data.CustomCloudInit.ValueString()
		createReq.CustomCloudInit = &cloudInit
	}

	tflog.Debug(ctx, "Creating VPS instance", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	// Create VPS
	vps, err := r.client.CreateVps(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create VPS", err.Error())
		return
	}

	data.ID = types.StringValue(vps.ID)

	tflog.Info(ctx, "VPS instance created, waiting for running state", map[string]interface{}{
		"id":   vps.ID,
		"name": vps.Name,
	})

	// Wait for VPS to be running
	err = r.client.WaitForVpsStatus(ctx, vps.ID, "running", createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"VPS failed to reach running state",
			fmt.Sprintf("VPS %s did not reach running state: %s", vps.ID, err),
		)
		return
	}

	// Refresh state after VPS is running
	vps, err = r.client.GetVps(ctx, vps.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read VPS after creation", err.Error())
		return
	}

	r.mapVpsToState(vps, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VpsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vps, err := r.client.GetVps(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read VPS", err.Error())
		return
	}

	r.mapVpsToState(vps, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VpsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state VpsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get timeout
	updateTimeout, diags := data.Timeouts.Update(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Build update request
	updateReq := client.UpdateVpsRequest{}
	hasChanges := false

	if !data.ResourceProfile.Equal(state.ResourceProfile) {
		updateReq.ResourceProfile = data.ResourceProfile.ValueString()
		hasChanges = true
	}

	if !data.Password.IsNull() && !data.Password.IsUnknown() && !data.Password.Equal(state.Password) {
		password := data.Password.ValueString()
		updateReq.Password = &password
		updateReq.PasswordConfirm = &password
		hasChanges = true
	}

	if hasChanges {
		tflog.Debug(ctx, "Updating VPS instance", map[string]interface{}{
			"id": data.ID.ValueString(),
		})

		vps, err := r.client.UpdateVps(ctx, data.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update VPS", err.Error())
			return
		}

		r.mapVpsToState(vps, &data)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VpsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get timeout
	deleteTimeout, diags := data.Timeouts.Delete(ctx, 15*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	tflog.Debug(ctx, "Deleting VPS instance", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	err := r.client.DeleteVps(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete VPS", err.Error())
		return
	}

	// Wait for VPS to be deleted
	err = r.client.WaitForVpsDeletion(ctx, data.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"VPS failed to be deleted",
			fmt.Sprintf("VPS %s was not deleted within the timeout: %s", data.ID.ValueString(), err),
		)
		return
	}
}

func (r *VpsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// extractImageID extracts the short image ID from a full registry path
// e.g., "registry.danubedata.ro/platform/kubevirt-ubuntu:ubuntu-24.04-2025.11.03" -> "ubuntu-24.04"
func extractImageID(fullPath string) string {
	// If it doesn't contain a registry path, return as-is
	if !strings.Contains(fullPath, "/") {
		return fullPath
	}

	// Extract the tag part after the colon
	// e.g., "kubevirt-ubuntu:ubuntu-24.04-2025.11.03" -> "ubuntu-24.04-2025.11.03"
	parts := strings.Split(fullPath, ":")
	if len(parts) < 2 {
		return fullPath
	}
	tag := parts[len(parts)-1]

	// Remove the date suffix (e.g., "-2025.11.03")
	// Pattern: distro-version-YYYY.MM.DD
	datePattern := regexp.MustCompile(`-\d{4}\.\d{2}\.\d{2}$`)
	imageID := datePattern.ReplaceAllString(tag, "")

	return imageID
}

func (r *VpsResource) mapVpsToState(vps *client.VpsInstance, data *VpsResourceModel) {
	data.ID = types.StringValue(vps.ID)
	data.Name = types.StringValue(vps.Name)
	data.Status = types.StringValue(vps.Status)
	data.ResourceProfile = types.StringValue(vps.ResourceProfile)
	// Only set image if not already set (e.g., during import)
	// The API returns the full image path, but users provide short IDs like "ubuntu-24.04"
	if data.Image.IsNull() || data.Image.IsUnknown() {
		data.Image = types.StringValue(extractImageID(vps.Image))
	}
	data.Datacenter = types.StringValue(vps.Datacenter)
	data.CPUCores = types.Int64Value(int64(vps.CPUCores))
	data.MemorySizeGB = types.Int64Value(int64(vps.MemorySizeGB))
	data.StorageSizeGB = types.Int64Value(int64(vps.StorageSizeGB))
	data.MonthlyCostCents = types.Int64Value(int64(vps.MonthlyCostCents))
	data.MonthlyCost = types.Float64Value(vps.MonthlyCost)
	data.CreatedAt = types.StringValue(vps.CreatedAt)
	data.UpdatedAt = types.StringValue(vps.UpdatedAt)

	if vps.PublicIP != nil {
		data.PublicIP = types.StringValue(*vps.PublicIP)
	} else {
		data.PublicIP = types.StringNull()
	}

	if vps.PrivateIP != nil {
		data.PrivateIP = types.StringValue(*vps.PrivateIP)
	} else {
		data.PrivateIP = types.StringNull()
	}

	if vps.IPv6Address != nil {
		data.IPv6Address = types.StringValue(*vps.IPv6Address)
	} else {
		data.IPv6Address = types.StringNull()
	}

	if vps.DeployedAt != nil {
		data.DeployedAt = types.StringValue(*vps.DeployedAt)
	} else {
		data.DeployedAt = types.StringNull()
	}

	if vps.SSHKeyID != nil {
		data.SSHKeyID = types.StringValue(*vps.SSHKeyID)
	}
}
