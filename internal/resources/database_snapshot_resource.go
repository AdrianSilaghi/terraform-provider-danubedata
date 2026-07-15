package resources

import (
	"context"
	"fmt"
	"strconv"
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
	_ resource.Resource                = &DatabaseSnapshotResource{}
	_ resource.ResourceWithConfigure   = &DatabaseSnapshotResource{}
	_ resource.ResourceWithImportState = &DatabaseSnapshotResource{}
)

type DatabaseSnapshotResource struct {
	client *client.Client
}

type DatabaseSnapshotResourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	Description        types.String   `tfsdk:"description"`
	DatabaseInstanceID types.String   `tfsdk:"database_instance_id"`
	Status             types.String   `tfsdk:"status"`
	SizeGB             types.Float64  `tfsdk:"size_gb"`
	CreatedAt          types.String   `tfsdk:"created_at"`
	UpdatedAt          types.String   `tfsdk:"updated_at"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

func NewDatabaseSnapshotResource() resource.Resource {
	return &DatabaseSnapshotResource{}
}

func (r *DatabaseSnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_snapshot"
}

func (r *DatabaseSnapshotResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData database snapshot for backup and recovery.",
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
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"database_instance_id": schema.StringAttribute{
				Description: "ID of the database instance to snapshot.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Status of the snapshot (creating, ready, failed).",
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

func (r *DatabaseSnapshotResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatabaseSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseSnapshotResourceModel
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

	tflog.Debug(ctx, "Creating database snapshot", map[string]interface{}{
		"name":                 data.Name.ValueString(),
		"database_instance_id": data.DatabaseInstanceID.ValueString(),
	})

	snapshot, err := r.client.CreateDatabaseSnapshot(ctx, client.CreateDatabaseSnapshotRequest{
		DatabaseInstanceID: data.DatabaseInstanceID.ValueString(),
		Name:               data.Name.ValueString(),
		Description:        data.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create database snapshot", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(snapshot.ID, 10))

	if err := r.client.WaitForDatabaseSnapshotStatus(ctx, snapshot.ID, "ready", createTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Database snapshot failed to complete",
			fmt.Sprintf("Snapshot %d did not complete: %s", snapshot.ID, err),
		)
		return
	}

	snapshot, err = r.client.GetDatabaseSnapshot(ctx, snapshot.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read database snapshot after creation", err.Error())
		return
	}

	r.mapSnapshotToState(snapshot, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseSnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid database snapshot ID", err.Error())
		return
	}

	snapshot, err := r.client.GetDatabaseSnapshot(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read database snapshot", err.Error())
		return
	}

	r.mapSnapshotToState(snapshot, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseSnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Snapshots are immutable; all user-visible fields require replacement. Preserve
	// existing state rather than writing the plan (which may contain Unknown computed
	// fields when only the timeouts block changes).
	var data DatabaseSnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseSnapshotResourceModel
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

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid database snapshot ID", err.Error())
		return
	}

	err = r.client.DeleteDatabaseSnapshot(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete database snapshot", err.Error())
		return
	}
}

func (r *DatabaseSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *DatabaseSnapshotResource) mapSnapshotToState(snapshot *client.DatabaseSnapshot, data *DatabaseSnapshotResourceModel) {
	data.ID = types.StringValue(strconv.FormatInt(snapshot.ID, 10))
	data.Name = types.StringValue(snapshot.Name)
	data.Description = types.StringValue(snapshot.Description)
	data.DatabaseInstanceID = types.StringValue(snapshot.DatabaseInstanceID)
	data.Status = types.StringValue(snapshot.Status)
	data.SizeGB = types.Float64Value(snapshot.SizeGB)
	data.CreatedAt = types.StringValue(snapshot.CreatedAt)
	data.UpdatedAt = types.StringValue(snapshot.UpdatedAt)
}
