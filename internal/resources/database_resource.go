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
	_ resource.Resource                = &DatabaseResource{}
	_ resource.ResourceWithConfigure   = &DatabaseResource{}
	_ resource.ResourceWithImportState = &DatabaseResource{}
)

type DatabaseResource struct {
	client *client.Client
}

type DatabaseResourceModel struct {
	ID               types.String   `tfsdk:"id"`
	Name             types.String   `tfsdk:"name"`
	Status           types.String   `tfsdk:"status"`
	Engine           types.String   `tfsdk:"engine"`
	DatabaseName     types.String   `tfsdk:"database_name"`
	ResourceProfile  types.String   `tfsdk:"resource_profile"`
	StorageSizeGB    types.Int64    `tfsdk:"storage_size_gb"`
	MemorySizeMB     types.Int64    `tfsdk:"memory_size_mb"`
	CPUCores         types.Int64    `tfsdk:"cpu_cores"`
	Version          types.String   `tfsdk:"version"`
	Datacenter       types.String   `tfsdk:"datacenter"`
	ParameterGroupID types.String   `tfsdk:"parameter_group_id"`
	Endpoint         types.String   `tfsdk:"endpoint"`
	Port             types.Int64    `tfsdk:"port"`
	Username         types.String   `tfsdk:"username"`
	MonthlyCostCents types.Int64    `tfsdk:"monthly_cost_cents"`
	MonthlyCost      types.Float64  `tfsdk:"monthly_cost"`
	DeployedAt       types.String   `tfsdk:"deployed_at"`
	CreatedAt        types.String   `tfsdk:"created_at"`
	UpdatedAt        types.String   `tfsdk:"updated_at"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

func (r *DatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData database instance (MySQL, PostgreSQL, or MariaDB).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the database instance.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the database instance. Must be lowercase alphanumeric with hyphens (DNS compatible).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`),
						"must be lowercase alphanumeric with hyphens (DNS compatible)",
					),
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"status": schema.StringAttribute{
				Description: "Current status of the database instance (pending, provisioning, running, stopped, error).",
				Computed:    true,
			},
			"engine": schema.StringAttribute{
				Description: "Database engine (mysql, postgresql, mariadb).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("mysql", "postgresql", "mariadb"),
				},
			},
			"database_name": schema.StringAttribute{
				Description: "Name of the initial database to create. Must start with a letter and contain only letters, numbers, and underscores.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
						"must start with a letter and contain only letters, numbers, and underscores",
					),
					stringvalidator.LengthAtMost(64),
				},
			},
			"resource_profile": schema.StringAttribute{
				Description: "Resource profile for the database (small, medium, large).",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("small", "medium", "large"),
				},
			},
			"storage_size_gb": schema.Int64Attribute{
				Description: "Storage size in GB. Derived from resource_profile.",
				Computed:    true,
			},
			"memory_size_mb": schema.Int64Attribute{
				Description: "Memory size in MB. Derived from resource_profile.",
				Computed:    true,
			},
			"cpu_cores": schema.Int64Attribute{
				Description: "Number of CPU cores. Derived from resource_profile.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Version of the database software.",
				Optional:    true,
				Computed:    true,
			},
			"datacenter": schema.StringAttribute{
				Description: "Datacenter location (fsn1, nbg1, hel1).",
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
				Description: "Connection endpoint for the database instance.",
				Computed:    true,
			},
			"port": schema.Int64Attribute{
				Description: "Port number for the database instance.",
				Computed:    true,
			},
			"username": schema.StringAttribute{
				Description: "Root username for the database instance.",
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
				Description: "Timestamp when the database instance was deployed.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the database instance was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the database instance was last updated.",
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

func (r *DatabaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseResourceModel
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
	createReq := client.CreateDatabaseRequest{
		Name:            data.Name.ValueString(),
		Provider:        data.Engine.ValueString(),
		Datacenter:      data.Datacenter.ValueString(),
		ResourceProfile: data.ResourceProfile.ValueString(),
	}

	if !data.DatabaseName.IsNull() && !data.DatabaseName.IsUnknown() {
		createReq.DatabaseName = data.DatabaseName.ValueString()
	}

	if !data.Version.IsNull() && !data.Version.IsUnknown() {
		createReq.Version = data.Version.ValueString()
	}

	if !data.ParameterGroupID.IsNull() && !data.ParameterGroupID.IsUnknown() {
		paramGroupID := data.ParameterGroupID.ValueString()
		createReq.ParameterGroupID = &paramGroupID
	}

	tflog.Debug(ctx, "Creating database instance", map[string]interface{}{
		"name":   data.Name.ValueString(),
		"engine": data.Engine.ValueString(),
	})

	// Create database instance
	database, err := r.client.CreateDatabase(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create database instance", err.Error())
		return
	}

	data.ID = types.StringValue(database.ID)

	tflog.Info(ctx, "Database instance created, waiting for running state", map[string]interface{}{
		"id":   database.ID,
		"name": database.Name,
	})

	// Wait for database to be running
	err = r.client.WaitForDatabaseStatus(ctx, database.ID, "running", createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Database instance failed to reach running state",
			fmt.Sprintf("Database %s did not reach running state: %s", database.ID, err),
		)
		return
	}

	// Refresh state after database is running
	database, err = r.client.GetDatabase(ctx, database.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read database instance after creation", err.Error())
		return
	}

	r.mapDatabaseToState(database, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	database, err := r.client.GetDatabase(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read database instance", err.Error())
		return
	}

	r.mapDatabaseToState(database, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DatabaseResourceModel
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
	updateReq := client.UpdateDatabaseRequest{}
	hasChanges := false

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
		tflog.Debug(ctx, "Updating database instance", map[string]interface{}{
			"id": data.ID.ValueString(),
		})

		database, err := r.client.UpdateDatabase(ctx, data.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update database instance", err.Error())
			return
		}

		// Note: We don't wait for running status here because the backend
		// processes updates asynchronously via a job that sets status to "pending".
		// The status will return to "running" after ArgoCD deploys the changes.
		// Waiting here would cause a timeout since the API returns before the job runs.

		r.mapDatabaseToState(database, &data)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseResourceModel
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

	databaseID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting database instance", map[string]interface{}{
		"id": databaseID,
	})

	// Check current status - database must be stopped before deletion
	database, err := r.client.GetDatabase(ctx, databaseID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to get database instance status", err.Error())
		return
	}

	status := strings.ToLower(database.Status)

	// If database is in a transitional state (pending, provisioning, restoring),
	// wait for it to reach a stable state before attempting to stop/delete
	if status == "pending" || status == "provisioning" || status == "restoring" {
		tflog.Info(ctx, "Waiting for database to reach stable state before deletion", map[string]interface{}{
			"id":     databaseID,
			"status": database.Status,
		})

		err = r.client.WaitForDatabaseStatus(ctx, databaseID, "running", deleteTimeout)
		if err != nil {
			// If we timeout waiting for running, check if it went to error state
			database, getErr := r.client.GetDatabase(ctx, databaseID)
			if getErr != nil {
				if client.IsNotFound(getErr) {
					return
				}
				resp.Diagnostics.AddError("Failed to get database status", getErr.Error())
				return
			}
			status = strings.ToLower(database.Status)
			if status != "error" && status != "stopped" {
				resp.Diagnostics.AddError(
					"Database instance failed to reach stable state",
					fmt.Sprintf("Database %s is in state %s, cannot delete", databaseID, database.Status),
				)
				return
			}
		} else {
			status = "running"
		}
	}

	// Stop the database if it's not already stopped
	if status != "stopped" && status != "deleted" && status != "error" {
		tflog.Info(ctx, "Stopping database before deletion", map[string]interface{}{
			"id":     databaseID,
			"status": status,
		})

		err = r.client.StopDatabase(ctx, databaseID)
		if err != nil {
			resp.Diagnostics.AddError("Failed to stop database instance before deletion", err.Error())
			return
		}

		// Wait for database to be stopped
		err = r.client.WaitForDatabaseStatus(ctx, databaseID, "stopped", deleteTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Database instance failed to stop",
				fmt.Sprintf("Database %s did not stop within the timeout: %s", databaseID, err),
			)
			return
		}
	}

	// Now delete the database
	err = r.client.DeleteDatabase(ctx, databaseID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete database instance", err.Error())
		return
	}

	// Wait for database to be deleted
	err = r.client.WaitForDatabaseDeletion(ctx, databaseID, deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Database instance failed to be deleted",
			fmt.Sprintf("Database %s was not deleted within the timeout: %s", databaseID, err),
		)
		return
	}
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *DatabaseResource) mapDatabaseToState(database *client.DatabaseInstance, data *DatabaseResourceModel) {
	data.ID = types.StringValue(database.ID)
	data.Name = types.StringValue(database.Name)
	data.Status = types.StringValue(database.Status)
	data.Engine = types.StringValue(strings.ToLower(database.Engine.Name))
	data.ResourceProfile = types.StringValue(database.ResourceProfile)
	data.StorageSizeGB = types.Int64Value(int64(database.StorageSizeGB))
	data.MemorySizeMB = types.Int64Value(int64(database.MemorySizeMB))
	data.CPUCores = types.Int64Value(int64(database.CPUCores))
	data.MonthlyCostCents = types.Int64Value(int64(database.MonthlyCostCents))
	data.MonthlyCost = types.Float64Value(database.MonthlyCostDollars)
	data.CreatedAt = types.StringValue(database.CreatedAt)
	data.UpdatedAt = types.StringValue(database.UpdatedAt)

	// Map datacenter - preserve from state if not returned by API
	if database.Datacenter != "" {
		data.Datacenter = types.StringValue(database.Datacenter)
	}

	// Map database_name if returned by API
	if database.DatabaseName != nil {
		data.DatabaseName = types.StringValue(*database.DatabaseName)
	}

	if database.Version != "" {
		data.Version = types.StringValue(database.Version)
	} else {
		data.Version = types.StringNull()
	}

	if database.Endpoint != nil {
		data.Endpoint = types.StringValue(*database.Endpoint)
	} else {
		data.Endpoint = types.StringNull()
	}

	if database.Port != nil {
		data.Port = types.Int64Value(int64(*database.Port))
	} else {
		data.Port = types.Int64Null()
	}

	if database.Username != nil {
		data.Username = types.StringValue(*database.Username)
	} else {
		data.Username = types.StringNull()
	}

	if database.DeployedAt != nil {
		data.DeployedAt = types.StringValue(*database.DeployedAt)
	} else {
		data.DeployedAt = types.StringNull()
	}

	if database.ParameterGroupID != nil && *database.ParameterGroupID != "" {
		data.ParameterGroupID = types.StringValue(*database.ParameterGroupID)
	} else {
		data.ParameterGroupID = types.StringNull()
	}
}
