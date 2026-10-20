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
	_ resource.Resource                = &CacheSnapshotResource{}
	_ resource.ResourceWithConfigure   = &CacheSnapshotResource{}
	_ resource.ResourceWithImportState = &CacheSnapshotResource{}
)

type CacheSnapshotResource struct {
	client *client.Client
}

type CacheSnapshotResourceModel struct {
	ID              types.String   `tfsdk:"id"`
	Name            types.String   `tfsdk:"name"`
	Description     types.String   `tfsdk:"description"`
	CacheInstanceID types.String   `tfsdk:"cache_instance_id"`
	Status          types.String   `tfsdk:"status"`
	SizeMB          types.Float64  `tfsdk:"size_mb"`
	SizeBytes       types.Int64    `tfsdk:"size_bytes"`
	SizeBasis       types.String   `tfsdk:"size_basis"`
	CreatedAt       types.String   `tfsdk:"created_at"`
	UpdatedAt       types.String   `tfsdk:"updated_at"`
	Timeouts        timeouts.Value `tfsdk:"timeouts"`
}

func NewCacheSnapshotResource() resource.Resource {
	return &CacheSnapshotResource{}
}

func (r *CacheSnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cache_snapshot"
}

func (r *CacheSnapshotResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData cache snapshot for backup and recovery.",
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
			"cache_instance_id": schema.StringAttribute{
				Description: "ID of the cache instance to snapshot.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Status of the snapshot (creating, ready, failed).",
				Computed:    true,
			},
			"size_mb": schema.Float64Attribute{
				Description: "Size of the snapshot in MB.",
				Computed:    true,
			},
			"size_bytes": schema.Int64Attribute{Computed: true, Description: "Exact snapshot size in bytes, when known."},
			"size_basis": schema.StringAttribute{Computed: true, Description: "How size was measured: stored, uploaded, or captured."},
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

func (r *CacheSnapshotResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CacheSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CacheSnapshotResourceModel
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

	tflog.Debug(ctx, "Creating cache snapshot", map[string]interface{}{
		"name":              data.Name.ValueString(),
		"cache_instance_id": data.CacheInstanceID.ValueString(),
	})

	snapshot, err := r.client.CreateCacheSnapshot(ctx, client.CreateCacheSnapshotRequest{
		CacheInstanceID: data.CacheInstanceID.ValueString(),
		Name:            data.Name.ValueString(),
		Description:     data.Description.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cache snapshot", err.Error())
		return
	}

	data.ID = types.StringValue(strconv.FormatInt(snapshot.ID, 10))
	r.mapSnapshotToState(snapshot, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.WaitForCacheSnapshotStatus(ctx, snapshot.ID, "ready", createTimeout); err != nil {
		resp.Diagnostics.AddError(
			"Cache snapshot failed to complete",
			fmt.Sprintf("Snapshot %d did not complete: %s", snapshot.ID, err),
		)
		return
	}

	snapshot, err = r.client.GetCacheSnapshot(ctx, snapshot.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read cache snapshot after creation", err.Error())
		return
	}

	r.mapSnapshotToState(snapshot, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CacheSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CacheSnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(data.ID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid cache snapshot ID", err.Error())
		return
	}

	snapshot, err := r.client.GetCacheSnapshot(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read cache snapshot", err.Error())
		return
	}

	r.mapSnapshotToState(snapshot, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CacheSnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Preserve computed snapshot values while accepting local timeout changes.
	var data CacheSnapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("timeouts"), &data.Timeouts)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CacheSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CacheSnapshotResourceModel
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
		resp.Diagnostics.AddError("Invalid cache snapshot ID", err.Error())
		return
	}

	err = r.client.DeleteCacheSnapshot(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete cache snapshot", err.Error())
		return
	}
}

func (r *CacheSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *CacheSnapshotResource) mapSnapshotToState(snapshot *client.CacheSnapshot, data *CacheSnapshotResourceModel) {
	data.ID = types.StringValue(strconv.FormatInt(snapshot.ID, 10))
	data.Name = types.StringValue(snapshot.Name)
	data.Description = types.StringValue(snapshot.Description)
	data.CacheInstanceID = types.StringValue(snapshot.CacheInstanceID)
	data.Status = types.StringValue(snapshot.Status)
	if snapshot.SizeKnown {
		data.SizeMB = types.Float64Value(snapshot.SizeMB)
	} else {
		data.SizeMB = types.Float64Null()
	}
	if snapshot.SizeBytes != nil {
		data.SizeBytes = types.Int64Value(*snapshot.SizeBytes)
	} else {
		data.SizeBytes = types.Int64Null()
	}
	if snapshot.SizeBasis != nil {
		data.SizeBasis = types.StringValue(*snapshot.SizeBasis)
	} else {
		data.SizeBasis = types.StringNull()
	}
	data.CreatedAt = types.StringValue(snapshot.CreatedAt)
	data.UpdatedAt = types.StringValue(snapshot.UpdatedAt)
}
