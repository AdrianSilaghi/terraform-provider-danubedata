package resources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
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
	_ resource.Resource                = &StorageAccessKeyResource{}
	_ resource.ResourceWithConfigure   = &StorageAccessKeyResource{}
	_ resource.ResourceWithImportState = &StorageAccessKeyResource{}
)

type StorageAccessKeyResource struct {
	client *client.Client
}

type StorageAccessKeyResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Status          types.String `tfsdk:"status"`
	ExpiresAt       types.String `tfsdk:"expires_at"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func NewStorageAccessKeyResource() resource.Resource {
	return &StorageAccessKeyResource{}
}

func (r *StorageAccessKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_access_key"
}

func (r *StorageAccessKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData S3-compatible storage access key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the access key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the access key.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 255),
				},
			},
			"access_key_id": schema.StringAttribute{
				Description: "The S3 access key ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_access_key": schema.StringAttribute{
				Description: "The S3 secret access key. Only available after creation.",
				Computed:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Description: "Current status of the access key (active, revoked).",
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "Optional expiration date for the access key (ISO 8601 format).",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the access key was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the access key was last updated.",
				Computed:    true,
			},
		},
	}
}

func (r *StorageAccessKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StorageAccessKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StorageAccessKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request
	createReq := client.CreateStorageAccessKeyRequest{
		Name: data.Name.ValueString(),
	}

	if !data.ExpiresAt.IsNull() && !data.ExpiresAt.IsUnknown() {
		expiresAt := data.ExpiresAt.ValueString()
		createReq.ExpiresAt = &expiresAt
	}

	tflog.Debug(ctx, "Creating storage access key", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	// Create access key
	createResp, err := r.client.CreateStorageAccessKey(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create storage access key", err.Error())
		return
	}

	// Map the response to state
	data.ID = types.StringValue(createResp.ID)
	data.AccessKeyID = types.StringValue(createResp.AccessKeyID)
	data.SecretAccessKey = types.StringValue(createResp.SecretAccessKey)
	data.Status = types.StringValue("active") // Newly created keys are always active

	if createResp.ExpiresAt != nil {
		data.ExpiresAt = types.StringValue(*createResp.ExpiresAt)
	} else {
		data.ExpiresAt = types.StringNull()
	}

	// Fetch the full access key to get created_at/updated_at
	accessKey, err := r.client.GetStorageAccessKey(ctx, createResp.ID)
	if err != nil {
		// Non-fatal - we have the critical data from create response
		tflog.Warn(ctx, "Could not fetch access key details after creation", map[string]interface{}{
			"id":    createResp.ID,
			"error": err.Error(),
		})
		data.CreatedAt = types.StringNull()
		data.UpdatedAt = types.StringNull()
	} else {
		data.CreatedAt = types.StringValue(accessKey.CreatedAt)
		data.UpdatedAt = types.StringValue(accessKey.UpdatedAt)
	}

	tflog.Info(ctx, "Storage access key created", map[string]interface{}{
		"id":            createResp.ID,
		"access_key_id": createResp.AccessKeyID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StorageAccessKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StorageAccessKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessKey, err := r.client.GetStorageAccessKey(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read storage access key", err.Error())
		return
	}

	// Update state from API (but preserve secret key from state since API doesn't return it)
	data.Name = types.StringValue(accessKey.Name)
	data.AccessKeyID = types.StringValue(accessKey.AccessKeyID)
	data.Status = types.StringValue(accessKey.Status)
	data.CreatedAt = types.StringValue(accessKey.CreatedAt)
	data.UpdatedAt = types.StringValue(accessKey.UpdatedAt)

	if accessKey.ExpiresAt != nil {
		data.ExpiresAt = types.StringValue(*accessKey.ExpiresAt)
	} else {
		data.ExpiresAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StorageAccessKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Access keys cannot be updated, all changes require replacement
	resp.Diagnostics.AddError(
		"Storage Access Key Update Not Supported",
		"Storage access keys cannot be updated. Changes require creating a new access key.",
	)
}

func (r *StorageAccessKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StorageAccessKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessKeyID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting storage access key", map[string]interface{}{
		"id": accessKeyID,
	})

	// Delete (revoke) the access key
	err := r.client.DeleteStorageAccessKey(ctx, accessKeyID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete storage access key", err.Error())
		return
	}
}

func (r *StorageAccessKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Note: The secret access key cannot be recovered after import
	resp.Diagnostics.AddWarning(
		"Secret Access Key Not Recoverable",
		"The secret access key cannot be recovered during import. It was only shown once when the key was created.",
	)
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
