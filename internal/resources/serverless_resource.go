package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ServerlessResource{}
	_ resource.ResourceWithConfigure   = &ServerlessResource{}
	_ resource.ResourceWithImportState = &ServerlessResource{}
)

type ServerlessResource struct {
	client *client.Client
}

type ServerlessResourceModel struct {
	ID                    types.String   `tfsdk:"id"`
	Name                  types.String   `tfsdk:"name"`
	Status                types.String   `tfsdk:"status"`
	ResourceProfile       types.String   `tfsdk:"resource_profile"`
	DeploymentType        types.String   `tfsdk:"deployment_type"`
	ImageURL              types.String   `tfsdk:"image_url"`
	GitRepository         types.String   `tfsdk:"git_repository"`
	GitBranch             types.String   `tfsdk:"git_branch"`
	Port                  types.Int64    `tfsdk:"port"`
	MinInstances          types.Int64    `tfsdk:"min_instances"`
	MaxInstances          types.Int64    `tfsdk:"max_instances"`
	EnvironmentVariables  types.Map      `tfsdk:"environment_variables"`
	URL                   types.String   `tfsdk:"url"`
	CurrentMonthCostCents types.Int64    `tfsdk:"current_month_cost_cents"`
	CreatedAt             types.String   `tfsdk:"created_at"`
	UpdatedAt             types.String   `tfsdk:"updated_at"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
}

func NewServerlessResource() resource.Resource {
	return &ServerlessResource{}
}

func (r *ServerlessResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serverless"
}

func (r *ServerlessResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData serverless container with scale-to-zero capability.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the serverless container.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the serverless container.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Current status of the serverless container.",
				Computed:    true,
			},
			"resource_profile": schema.StringAttribute{
				Description: "Resource profile for the container (small, medium, large).",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("small"),
			},
			"deployment_type": schema.StringAttribute{
				Description: "Deployment type: 'image' for Docker image, 'git' for Git repository.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("image", "git"),
				},
			},
			"image_url": schema.StringAttribute{
				Description: "Docker image URL (required if deployment_type is 'image').",
				Optional:    true,
			},
			"git_repository": schema.StringAttribute{
				Description: "Git repository URL (required if deployment_type is 'git').",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"git_branch": schema.StringAttribute{
				Description: "Git branch to deploy.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("main"),
			},
			"port": schema.Int64Attribute{
				Description: "Port the container listens on.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(8080),
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"min_instances": schema.Int64Attribute{
				Description: "Minimum number of instances (0 for scale-to-zero).",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Validators: []validator.Int64{
					int64validator.Between(0, 10),
				},
			},
			"max_instances": schema.Int64Attribute{
				Description: "Maximum number of instances.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(5),
				Validators: []validator.Int64{
					int64validator.Between(1, 100),
				},
			},
			"environment_variables": schema.MapAttribute{
				Description: "Environment variables for the container.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"url": schema.StringAttribute{
				Description: "Public URL of the deployed service.",
				Computed:    true,
			},
			"current_month_cost_cents": schema.Int64Attribute{
				Description: "Current month's cost in cents.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the container was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the container was last updated.",
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

func (r *ServerlessResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServerlessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServerlessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := data.Timeouts.Create(ctx, 15*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	tflog.Debug(ctx, "Creating serverless container", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	createReq := client.CreateServerlessRequest{
		Name:            data.Name.ValueString(),
		ResourceProfile: data.ResourceProfile.ValueString(),
		DeploymentType:  data.DeploymentType.ValueString(),
		Port:            int(data.Port.ValueInt64()),
		MinInstances:    int(data.MinInstances.ValueInt64()),
		MaxInstances:    int(data.MaxInstances.ValueInt64()),
	}

	if !data.ImageURL.IsNull() && !data.ImageURL.IsUnknown() {
		createReq.ImageURL = data.ImageURL.ValueString()
	}

	if !data.GitRepository.IsNull() && !data.GitRepository.IsUnknown() {
		createReq.GitRepository = data.GitRepository.ValueString()
	}

	if !data.GitBranch.IsNull() && !data.GitBranch.IsUnknown() {
		createReq.GitBranch = data.GitBranch.ValueString()
	}

	if !data.EnvironmentVariables.IsNull() && !data.EnvironmentVariables.IsUnknown() {
		envVars := make(map[string]string)
		resp.Diagnostics.Append(data.EnvironmentVariables.ElementsAs(ctx, &envVars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.EnvironmentVariables = envVars
	}

	container, err := r.client.CreateServerless(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create serverless container", err.Error())
		return
	}

	data.ID = types.StringValue(container.ID)

	tflog.Info(ctx, "Serverless container created, waiting for ready state", map[string]interface{}{
		"id":   container.ID,
		"name": container.Name,
	})

	// Wait for container to be ready
	err = r.client.WaitForServerlessStatus(ctx, container.ID, "running", createTimeout)
	if err != nil {
		// Don't fail if just waiting times out - the container might still be deploying
		tflog.Warn(ctx, "Serverless container did not reach running state within timeout", map[string]interface{}{
			"id":    container.ID,
			"error": err.Error(),
		})
	}

	// Refresh state
	container, err = r.client.GetServerless(ctx, container.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read serverless container after creation", err.Error())
		return
	}

	r.mapContainerToState(ctx, container, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerlessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServerlessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	container, err := r.client.GetServerless(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read serverless container", err.Error())
		return
	}

	r.mapContainerToState(ctx, container, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerlessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServerlessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ServerlessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := data.Timeouts.Update(ctx, 15*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	tflog.Debug(ctx, "Updating serverless container", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	updateReq := client.UpdateServerlessRequest{}
	hasChanges := false

	if !data.ResourceProfile.Equal(state.ResourceProfile) {
		updateReq.ResourceProfile = data.ResourceProfile.ValueString()
		hasChanges = true
	}

	if !data.ImageURL.Equal(state.ImageURL) && !data.ImageURL.IsNull() {
		updateReq.ImageURL = data.ImageURL.ValueString()
		hasChanges = true
	}

	if !data.GitBranch.Equal(state.GitBranch) {
		updateReq.GitBranch = data.GitBranch.ValueString()
		hasChanges = true
	}

	if !data.Port.Equal(state.Port) {
		updateReq.Port = int(data.Port.ValueInt64())
		hasChanges = true
	}

	if !data.MinInstances.Equal(state.MinInstances) {
		minInst := int(data.MinInstances.ValueInt64())
		updateReq.MinInstances = &minInst
		hasChanges = true
	}

	if !data.MaxInstances.Equal(state.MaxInstances) {
		maxInst := int(data.MaxInstances.ValueInt64())
		updateReq.MaxInstances = &maxInst
		hasChanges = true
	}

	if !data.EnvironmentVariables.Equal(state.EnvironmentVariables) {
		envVars := make(map[string]string)
		resp.Diagnostics.Append(data.EnvironmentVariables.ElementsAs(ctx, &envVars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.EnvironmentVariables = envVars
		hasChanges = true
	}

	if hasChanges {
		_, err := r.client.UpdateServerless(ctx, data.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update serverless container", err.Error())
			return
		}

		// Wait for update to complete
		err = r.client.WaitForServerlessStatus(ctx, data.ID.ValueString(), "running", updateTimeout)
		if err != nil {
			tflog.Warn(ctx, "Serverless container did not reach running state after update", map[string]interface{}{
				"id":    data.ID.ValueString(),
				"error": err.Error(),
			})
		}
	}

	// Refresh state
	container, err := r.client.GetServerless(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read serverless container after update", err.Error())
		return
	}

	r.mapContainerToState(ctx, container, &data, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServerlessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServerlessResourceModel
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

	tflog.Debug(ctx, "Deleting serverless container", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	err := r.client.DeleteServerless(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete serverless container", err.Error())
		return
	}

	// Wait for deletion
	err = r.client.WaitForServerlessDeletion(ctx, data.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed waiting for serverless container deletion", err.Error())
		return
	}
}

func (r *ServerlessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ServerlessResource) mapContainerToState(ctx context.Context, container *client.ServerlessContainer, data *ServerlessResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(container.ID)
	data.Name = types.StringValue(container.Name)
	data.Status = types.StringValue(container.Status)
	data.ResourceProfile = types.StringValue(container.ResourceProfile)
	data.DeploymentType = types.StringValue(container.DeploymentType)
	data.Port = types.Int64Value(int64(container.Port))
	data.MinInstances = types.Int64Value(int64(container.MinInstances))
	data.MaxInstances = types.Int64Value(int64(container.MaxInstances))
	data.URL = types.StringValue(container.URL)
	data.CurrentMonthCostCents = types.Int64Value(int64(container.CurrentMonthCostCents))
	data.CreatedAt = types.StringValue(container.CreatedAt)
	data.UpdatedAt = types.StringValue(container.UpdatedAt)

	if container.ImageURL != "" {
		data.ImageURL = types.StringValue(container.ImageURL)
	} else {
		data.ImageURL = types.StringNull()
	}

	if container.GitRepository != "" {
		data.GitRepository = types.StringValue(container.GitRepository)
	} else {
		data.GitRepository = types.StringNull()
	}

	if container.GitBranch != "" {
		data.GitBranch = types.StringValue(container.GitBranch)
	} else {
		data.GitBranch = types.StringNull()
	}

	if len(container.EnvironmentVariables) > 0 {
		envVars, envDiags := types.MapValueFrom(ctx, types.StringType, container.EnvironmentVariables)
		diags.Append(envDiags...)
		data.EnvironmentVariables = envVars
	} else {
		data.EnvironmentVariables = types.MapNull(types.StringType)
	}
}
