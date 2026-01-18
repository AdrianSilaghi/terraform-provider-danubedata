package resources

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	CacheProviderID  types.Int64    `tfsdk:"cache_provider_id"`
	ProviderType     types.String   `tfsdk:"provider_type"`
	ResourceProfile  types.String   `tfsdk:"resource_profile"`
	MemorySizeMB     types.Int64    `tfsdk:"memory_size_mb"`
	CPUCores         types.Int64    `tfsdk:"cpu_cores"`
	Version          types.String   `tfsdk:"version"`
	Datacenter       types.String   `tfsdk:"datacenter"`
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
			"cache_provider_id": schema.Int64Attribute{
				Description: "ID of the cache provider (1=Redis, 2=Valkey, 3=Dragonfly).",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 3),
				},
			},
			"provider_type": schema.StringAttribute{
				Description: "Type of cache provider (redis, valkey, dragonfly). Computed from cache_provider_id.",
				Computed:    true,
			},
			"resource_profile": schema.StringAttribute{
				Description: "Resource profile for the cache (micro, small, medium, large).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("micro", "small", "medium", "large"),
				},
			},
			"memory_size_mb": schema.Int64Attribute{
				Description: "Memory size in MB (128-32768).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(512),
				Validators: []validator.Int64{
					int64validator.Between(128, 32768),
				},
			},
			"cpu_cores": schema.Int64Attribute{
				Description: "Number of CPU cores (1-16).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.Between(1, 16),
				},
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
		CacheProviderID:   int(data.CacheProviderID.ValueInt64()),
		MemorySizeMB:      int(data.MemorySizeMB.ValueInt64()),
		CPUCores:          int(data.CPUCores.ValueInt64()),
		HetznerDatacenter: data.Datacenter.ValueString(),
		ResourceProfile:   data.ResourceProfile.ValueString(),
	}

	if !data.Version.IsNull() && !data.Version.IsUnknown() {
		createReq.Version = data.Version.ValueString()
	}

	tflog.Debug(ctx, "Creating cache instance", map[string]interface{}{
		"name": data.Name.ValueString(),
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

	if hasChanges {
		tflog.Debug(ctx, "Updating cache instance", map[string]interface{}{
			"id": data.ID.ValueString(),
		})

		cache, err := r.client.UpdateCache(ctx, data.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update cache instance", err.Error())
			return
		}

		// Wait for cache to be running again after update
		err = r.client.WaitForCacheStatus(ctx, cache.ID, "running", updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Cache instance failed to reach running state after update",
				fmt.Sprintf("Cache %s did not reach running state: %s", cache.ID, err),
			)
			return
		}

		// Refresh state
		cache, err = r.client.GetCache(ctx, cache.ID)
		if err != nil {
			resp.Diagnostics.AddError("Failed to read cache instance after update", err.Error())
			return
		}

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

	// Check current status - cache must be stopped before deletion (similar to VPS)
	cache, err := r.client.GetCache(ctx, cacheID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to get cache instance status", err.Error())
		return
	}

	// If running, stop it first
	if cache.Status == "running" {
		tflog.Info(ctx, "Cache is running, stopping before deletion", map[string]interface{}{
			"id": cacheID,
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
	data.CacheProviderID = types.Int64Value(int64(cache.ProviderID))
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

	// Map provider type from the loaded provider info
	if cache.Provider != nil {
		data.ProviderType = types.StringValue(cache.Provider.Type)
	} else {
		// Fallback based on provider ID
		switch cache.ProviderID {
		case 1:
			data.ProviderType = types.StringValue("redis")
		case 2:
			data.ProviderType = types.StringValue("valkey")
		case 3:
			data.ProviderType = types.StringValue("dragonfly")
		default:
			data.ProviderType = types.StringNull()
		}
	}

	if cache.Endpoint != nil {
		data.Endpoint = types.StringValue(*cache.Endpoint)
	} else {
		data.Endpoint = types.StringNull()
	}

	if cache.Port != nil {
		data.Port = types.Int64Value(int64(*cache.Port))
	} else {
		data.Port = types.Int64Null()
	}

	if cache.DeployedAt != nil {
		data.DeployedAt = types.StringValue(*cache.DeployedAt)
	} else {
		data.DeployedAt = types.StringNull()
	}
}
