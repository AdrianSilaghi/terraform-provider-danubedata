package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &VpsSnapshotResource{}
	_ resource.ResourceWithConfigure   = &VpsSnapshotResource{}
	_ resource.ResourceWithImportState = &VpsSnapshotResource{}
)

type VpsSnapshotResource struct {
	client *client.Client
}

type VpsSnapshotResourceModel struct {
	ID            types.String   `tfsdk:"id"`
	Name          types.String   `tfsdk:"name"`
	Description   types.String   `tfsdk:"description"`
	VpsInstanceID types.String   `tfsdk:"vps_instance_id"`
	Status        types.String   `tfsdk:"status"`
	SizeGB        types.Float64  `tfsdk:"size_gb"`
	CreatedAt     types.String   `tfsdk:"created_at"`
	UpdatedAt     types.String   `tfsdk:"updated_at"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

func NewVpsSnapshotResource() resource.Resource {
	return &VpsSnapshotResource{}
}

func (r *VpsSnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_snapshot"
}

func (r *VpsSnapshotResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData VPS snapshot for backup and recovery.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the snapshot.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the snapshot.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the snapshot.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vps_instance_id": schema.StringAttribute{
				Description: "ID of the VPS instance to snapshot.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Status of the snapshot (pending, completed, failed).",
				Computed:    true,
			},
			"size_gb": schema.Float64Attribute{
				Description: "Size of the snapshot in GB.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the snapshot was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the snapshot was last updated.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *VpsSnapshotResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VpsSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VpsSnapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 30*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	tflog.Debug(ctx, "Creating VPS snapshot", map[string]interface{}{
		"name":            data.Name.ValueString(),
		"vps_instance_id": data.VpsInstanceID.ValueString(),
	})

	createReq := client.CreateVpsSnapshotRequest{
		VpsInstanceID: data.VpsInstanceID.ValueString(),
		Name:          data.Name.ValueString(),
		Description:   data.Description.ValueString(),
	}

	snapshot, err := r.client.CreateVpsSnapshot(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create VPS snapshot", err.Error())
		return
	}

	data.ID = types.StringValue(snapshot.ID)

	tflog.Info(ctx, "VPS snapshot created, waiting for completion", map[string]interface{}{
		"id":   snapshot.ID,
		"name": snapshot.Name,
	})

	// Wait for snapshot to complete
	err = r.client.WaitForVpsSnapshotStatus(ctx, snapshot.ID, "completed", createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"VPS snapshot failed to complete",
			fmt.Sprintf("Snapshot %s did not complete: %s", snapshot.ID, err),
		)
		return
	}

	// Refresh state
	snapshot, err = r.client.GetVpsSnapshot(ctx, snapshot.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read VPS snapshot after creation", err.Error())
		return
	}

	r.mapSnapshotToState(snapshot, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpsSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VpsSnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshot, err := r.client.GetVpsSnapshot(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read VPS snapshot", err.Error())
		return
	}

	r.mapSnapshotToState(snapshot, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpsSnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Snapshots cannot be updated, all changes require replacement
	var data VpsSnapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VpsSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VpsSnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := data.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	tflog.Debug(ctx, "Deleting VPS snapshot", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	err := r.client.DeleteVpsSnapshot(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete VPS snapshot", err.Error())
		return
	}
}

func (r *VpsSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *VpsSnapshotResource) mapSnapshotToState(snapshot *client.VpsSnapshot, data *VpsSnapshotResourceModel) {
	data.ID = types.StringValue(snapshot.ID)
	data.Name = types.StringValue(snapshot.Name)
	data.Description = types.StringValue(snapshot.Description)
	data.VpsInstanceID = types.StringValue(snapshot.VpsInstanceID)
	data.Status = types.StringValue(snapshot.Status)
	data.SizeGB = types.Float64Value(snapshot.SizeGB)
	data.CreatedAt = types.StringValue(snapshot.CreatedAt)
	data.UpdatedAt = types.StringValue(snapshot.UpdatedAt)
}
