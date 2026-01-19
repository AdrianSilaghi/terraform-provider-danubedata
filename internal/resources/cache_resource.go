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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &CacheResource{}
	_ resource.ResourceWithConfigure   = &CacheResource{}
	_ resource.ResourceWithImportState = &CacheResource{}
)

type CacheResource struct {
	client *client.Client
}

type CacheResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	Status           types.String   `tfsdk:"status"`
	CacheProvider    types.String   `tfsdk:"cache_provider"`
	ResourceProfile  types.String   `tfsdk:"resource_profile"`
	MemorySizeMB     types.Int64    `tfsdk:"memory_size_mb"`
	CPUCores         types.Int64    `tfsdk:"cpu_cores"`
	Version          types.String   `tfsdk:"version"`
	Datacenter       types.String   `tfsdk:"datacenter"`
	ParameterGroupID types.String   `tfsdk:"parameter_group_id"`
	Endpoint         types.String   `tfsdk:"endpoint"`
	Port             types.Int64    `tfsdk:"port"`
	MonthlyCostCents types.Int64    `tfsdk:"monthly_cost_cents"`
	MonthlyCost      types.Float64  `tfsdk:"monthly_cost"`
	DeployedAt       types.String   `tfsdk:"deployed_at"`
	CreatedAt        types.String   `tfsdk:"created_at"`
	UpdatedAt        types.String   `tfsdk:"updated_at"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

func NewCacheResource() resource.Resource {
	return &CacheResource{}
}

func (r *CacheResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cache"
}

func (r *CacheResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData cache instance (Redis, Valkey, or Dragonfly).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the cache instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the cache instance.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`),
						"must be alphanumeric with hyphens, starting and ending with alphanumeric",
					),
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"status": schema.StringAttribute{
				Description: "Current status of the cache instance (pending, provisioning, running, stopped, error).",
				Computed:    true,
			},
			"cache_provider": schema.StringAttribute{
				Description: "Cache provider type (redis, valkey, dragonfly).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("redis", "valkey", "dragonfly"),
				},
			},
			"resource_profile": schema.StringAttribute{
				Description: "Resource profile for the cache (micro, small, medium, large).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("micro", "small", "medium", "large"),
				},
			},
			"memory_size_mb": schema.Int64Attribute{
				Description: "Memory size in MB. If not specified, determined by resource_profile.",
				Optional:    true,
				Computed:    true,
			},
			"cpu_cores": schema.Int64Attribute{
				Description: "Number of CPU cores. If not specified, determined by resource_profile.",
				Optional:    true,
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Version of the cache software.",
				Optional:    true,
				Computed:    true,
			},
			"datacenter": schema.StringAttribute{
				Description: "Datacenter location (ash, fsn1, nbg1, hel1).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("fsn1", "nbg1", "hel1", "ash"),
				},
			},
			"parameter_group_id": schema.StringAttribute{
				Description: "ID of the parameter group to use for custom configuration.",
				Optional:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "Connection endpoint for the cache instance.",
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port number for the cache instance.",
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
				Description: "Timestamp when the cache instance was deployed.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the cache instance was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the cache instance was last updated.",
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

func (r *CacheResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CacheResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CacheResourceModel
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
	createReq := client.CreateCacheRequest{
		Name:              data.Name.ValueString(),
		Provider:          data.CacheProvider.ValueString(),
		MemorySizeMB:      int(data.MemorySizeMB.ValueInt64()),
		CPUCores:          int(data.CPUCores.ValueInt64()),
		Datacenter:        data.Datacenter.ValueString(),
		ResourceProfile:   data.ResourceProfile.ValueString(),
	}

	if !data.Version.IsNull() && !data.Version.IsUnknown() {
		createReq.Version = data.Version.ValueString()
	}

	if !data.ParameterGroupID.IsNull() && !data.ParameterGroupID.IsUnknown() {
		paramGroupID := data.ParameterGroupID.ValueString()
		createReq.ParameterGroupID = &paramGroupID
	}

	tflog.Debug(ctx, "Creating cache instance", map[string]interface{}{
		"name":           data.Name.ValueString(),
		"cache_provider": data.CacheProvider.ValueString(),
	})

	// Create cache instance
	cache, err := r.client.CreateCache(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create cache instance", err.Error())
		return
	}

	data.ID = types.StringValue(cache.ID)

	tflog.Info(ctx, "Cache instance created, waiting for running state", map[string]interface{}{
		"id":   cache.ID,
		"name": cache.Name,
	})

	// Wait for cache to be running
	err = r.client.WaitForCacheStatus(ctx, cache.ID, "running", createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cache instance failed to reach running state",
			fmt.Sprintf("Cache %s did not reach running state: %s", cache.ID, err),
		)
		return
	}

	// Refresh state after cache is running
	cache, err = r.client.GetCache(ctx, cache.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read cache instance after creation", err.Error())
		return
	}

	r.mapCacheToState(cache, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CacheResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CacheResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cache, err := r.client.GetCache(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read cache instance", err.Error())
		return
	}

	r.mapCacheToState(cache, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CacheResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CacheResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CacheResourceModel
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
	updateReq := client.UpdateCacheRequest{}
	hasChanges := false

	if !data.Name.Equal(state.Name) {
		updateReq.Name = data.Name.ValueString()
		hasChanges = true
	}

	if !data.MemorySizeMB.Equal(state.MemorySizeMB) {
		memSize := int(data.MemorySizeMB.ValueInt64())
		updateReq.MemorySizeMB = &memSize
		hasChanges = true
	}

	if !data.CPUCores.Equal(state.CPUCores) {
		cpuCores := int(data.CPUCores.ValueInt64())
		updateReq.CPUCores = &cpuCores
		hasChanges = true
	}

	if !data.ResourceProfile.Equal(state.ResourceProfile) {
		updateReq.ResourceProfile = data.ResourceProfile.ValueString()
		hasChanges = true
	}

	if !data.ParameterGroupID.Equal(state.ParameterGroupID) {
		if !data.ParameterGroupID.IsNull() {
			paramGroupID := data.ParameterGroupID.ValueString()
			updateReq.ParameterGroupID = &paramGroupID
		}
		hasChanges = true
	}

	if hasChanges {
		tflog.Debug(ctx, "Updating cache instance", map[string]interface{}{
			"id": data.ID.ValueString(),
		})

		cache, err := r.client.UpdateCache(ctx, data.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update cache instance", err.Error())
			return
		}

		// Note: We don't wait for running status here because the backend
		// processes updates asynchronously via a job that sets status to "pending".
		// The status will return to "running" after ArgoCD deploys the changes.
		// Waiting here would cause a timeout since the API returns before the job runs.

		r.mapCacheToState(cache, &data)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CacheResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CacheResourceModel
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

	cacheID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting cache instance", map[string]interface{}{
		"id": cacheID,
	})

	// Check current status - cache must be stopped before deletion
	cache, err := r.client.GetCache(ctx, cacheID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to get cache instance status", err.Error())
		return
	}

	status := strings.ToLower(cache.Status)

	// If cache is in a transitional state (pending, provisioning, restoring),
	// wait for it to reach a stable state before attempting to stop/delete
	if status == "pending" || status == "provisioning" || status == "restoring" {
		tflog.Info(ctx, "Waiting for cache to reach stable state before deletion", map[string]interface{}{
			"id":     cacheID,
			"status": cache.Status,
		})

		err = r.client.WaitForCacheStatus(ctx, cacheID, "running", deleteTimeout)
		if err != nil {
			// If we timeout waiting for running, check if it went to error state
			cache, getErr := r.client.GetCache(ctx, cacheID)
			if getErr != nil {
				if client.IsNotFound(getErr) {
					return
				}
				resp.Diagnostics.AddError("Failed to get cache status", getErr.Error())
				return
			}
			status = strings.ToLower(cache.Status)
			if status != "error" && status != "stopped" {
				resp.Diagnostics.AddError(
					"Cache instance failed to reach stable state",
					fmt.Sprintf("Cache %s is in state %s, cannot delete", cacheID, cache.Status),
				)
				return
			}
		} else {
			status = "running"
		}
	}

	// Stop the cache if it's not already stopped
	if status != "stopped" && status != "deleted" && status != "error" {
		tflog.Info(ctx, "Stopping cache before deletion", map[string]interface{}{
			"id":     cacheID,
			"status": status,
		})

		err = r.client.StopCache(ctx, cacheID)
		if err != nil {
			resp.Diagnostics.AddError("Failed to stop cache instance before deletion", err.Error())
			return
		}

		// Wait for cache to be stopped
		err = r.client.WaitForCacheStatus(ctx, cacheID, "stopped", deleteTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Cache instance failed to stop",
				fmt.Sprintf("Cache %s did not stop within the timeout: %s", cacheID, err),
			)
			return
		}
	}

	// Now delete the cache
	err = r.client.DeleteCache(ctx, cacheID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete cache instance", err.Error())
		return
	}

	// Wait for cache to be deleted
	err = r.client.WaitForCacheDeletion(ctx, cacheID, deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Cache instance failed to be deleted",
			fmt.Sprintf("Cache %s was not deleted within the timeout: %s", cacheID, err),
		)
		return
	}
}

func (r *CacheResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *CacheResource) mapCacheToState(cache *client.CacheInstance, data *CacheResourceModel) {
	data.ID = types.StringValue(cache.ID)
	data.Name = types.StringValue(cache.Name)
	data.Status = types.StringValue(cache.Status)
	data.CacheProvider = types.StringValue(strings.ToLower(cache.Provider.Name))
	data.ResourceProfile = types.StringValue(cache.ResourceProfile)
	data.MemorySizeMB = types.Int64Value(int64(cache.MemorySizeMB))
	data.CPUCores = types.Int64Value(int64(cache.CPUCores))
	data.MonthlyCostCents = types.Int64Value(int64(cache.MonthlyCostCents))
	data.MonthlyCost = types.Float64Value(cache.MonthlyCostDollars)
	data.CreatedAt = types.StringValue(cache.CreatedAt)
	data.UpdatedAt = types.StringValue(cache.UpdatedAt)

	// Map datacenter - preserve from state if not returned by API
	if cache.Datacenter != "" {
		data.Datacenter = types.StringValue(cache.Datacenter)
	}

	if cache.Version != "" {
		data.Version = types.StringValue(cache.Version)
	} else {
		data.Version = types.StringNull()
	}

	if cache.Endpoint != nil && *cache.Endpoint != "" {
		data.Endpoint = types.StringValue(*cache.Endpoint)
	} else {
		data.Endpoint = types.StringNull()
	}

	if cache.Port != nil && *cache.Port != 0 {
		data.Port = types.Int64Value(int64(*cache.Port))
	} else {
		data.Port = types.Int64Null()
	}

	if cache.DeployedAt != nil {
		data.DeployedAt = types.StringValue(*cache.DeployedAt)
	} else {
		data.DeployedAt = types.StringNull()
	}

	if cache.ParameterGroupID != nil && *cache.ParameterGroupID != "" {
		data.ParameterGroupID = types.StringValue(*cache.ParameterGroupID)
	} else {
		data.ParameterGroupID = types.StringNull()
	}
}
