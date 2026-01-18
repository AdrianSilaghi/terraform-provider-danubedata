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
	_ resource.Resource                = &DatabaseResource{}
	_ resource.ResourceWithConfigure   = &DatabaseResource{}
	_ resource.ResourceWithImportState = &DatabaseResource{}
)

type DatabaseResource struct {
	client *client.Client
}

type DatabaseResourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	Status             types.String   `tfsdk:"status"`
	DatabaseProviderID types.Int64    `tfsdk:"database_provider_id"`
	ProviderType       types.String   `tfsdk:"provider_type"`
	DatabaseName       types.String   `tfsdk:"database_name"`
	ResourceProfile    types.String   `tfsdk:"resource_profile"`
	StorageSizeGB      types.Int64    `tfsdk:"storage_size_gb"`
	MemorySizeMB       types.Int64    `tfsdk:"memory_size_mb"`
	CPUCores           types.Int64    `tfsdk:"cpu_cores"`
	Version            types.String   `tfsdk:"version"`
	Datacenter         types.String   `tfsdk:"datacenter"`
	Endpoint           types.String   `tfsdk:"endpoint"`
	Port               types.Int64    `tfsdk:"port"`
	Username           types.String   `tfsdk:"username"`
	MonthlyCostCents   types.Int64    `tfsdk:"monthly_cost_cents"`
	MonthlyCost        types.Float64  `tfsdk:"monthly_cost"`
	DeployedAt         types.String   `tfsdk:"deployed_at"`
	CreatedAt          types.String   `tfsdk:"created_at"`
	UpdatedAt          types.String   `tfsdk:"updated_at"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
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
			"database_provider_id": schema.Int64Attribute{
				Description: "ID of the database provider (1=MySQL, 2=PostgreSQL, 3=MariaDB).",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(1, 3),
				},
			},
			"provider_type": schema.StringAttribute{
				Description: "Type of database provider (mysql, postgresql, mariadb). Computed from database_provider_id.",
				Computed:    true,
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
				Description: "Storage size in GB (10-500).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(10, 500),
				},
			},
			"memory_size_mb": schema.Int64Attribute{
				Description: "Memory size in MB (1024-32768).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(2048),
				Validators: []validator.Int64{
					int64validator.Between(1024, 32768),
				},
			},
			"cpu_cores": schema.Int64Attribute{
				Description: "Number of CPU cores (1-16).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(2),
				Validators: []validator.Int64{
					int64validator.Between(1, 16),
				},
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
		Name:               data.Name.ValueString(),
		DatabaseProviderID: int(data.DatabaseProviderID.ValueInt64()),
		StorageSizeGB:      int(data.StorageSizeGB.ValueInt64()),
		MemorySizeMB:       int(data.MemorySizeMB.ValueInt64()),
		CPUCores:           int(data.CPUCores.ValueInt64()),
		HetznerDatacenter:  data.Datacenter.ValueString(),
		ResourceProfile:    data.ResourceProfile.ValueString(),
	}

	if !data.DatabaseName.IsNull() && !data.DatabaseName.IsUnknown() {
		createReq.DatabaseName = data.DatabaseName.ValueString()
	}

	if !data.Version.IsNull() && !data.Version.IsUnknown() {
		createReq.Version = data.Version.ValueString()
	}

	tflog.Debug(ctx, "Creating database instance", map[string]interface{}{
		"name": data.Name.ValueString(),
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

	if !data.StorageSizeGB.Equal(state.StorageSizeGB) {
		storageSize := int(data.StorageSizeGB.ValueInt64())
		updateReq.StorageSizeGB = &storageSize
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
		tflog.Debug(ctx, "Updating database instance", map[string]interface{}{
			"id": data.ID.ValueString(),
		})

		database, err := r.client.UpdateDatabase(ctx, data.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update database instance", err.Error())
			return
		}

		// Wait for database to be running again after update
		err = r.client.WaitForDatabaseStatus(ctx, database.ID, "running", updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Database instance failed to reach running state after update",
				fmt.Sprintf("Database %s did not reach running state: %s", database.ID, err),
			)
			return
		}

		// Refresh state
		database, err = r.client.GetDatabase(ctx, database.ID)
		if err != nil {
			resp.Diagnostics.AddError("Failed to read database instance after update", err.Error())
			return
		}

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

	// If running, stop it first
	if database.Status == "running" {
		tflog.Info(ctx, "Database is running, stopping before deletion", map[string]interface{}{
			"id": databaseID,
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
	data.DatabaseProviderID = types.Int64Value(int64(database.ProviderID))
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

	// Map provider type from the loaded provider info
	if database.Provider != nil {
		data.ProviderType = types.StringValue(database.Provider.Type)
	} else {
		// Fallback based on provider ID
		switch database.ProviderID {
		case 1:
			data.ProviderType = types.StringValue("mysql")
		case 2:
			data.ProviderType = types.StringValue("postgresql")
		case 3:
			data.ProviderType = types.StringValue("mariadb")
		default:
			data.ProviderType = types.StringNull()
		}
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
}
