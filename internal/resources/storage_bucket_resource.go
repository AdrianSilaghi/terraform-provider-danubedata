package resources

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &StorageBucketResource{}
	_ resource.ResourceWithConfigure   = &StorageBucketResource{}
	_ resource.ResourceWithImportState = &StorageBucketResource{}
)

type StorageBucketResource struct {
	client *client.Client
}

type StorageBucketResourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	Name               types.String   `tfsdk:"name"`
	DisplayName        types.String   `tfsdk:"display_name"`
	Status             types.String   `tfsdk:"status"`
	Region             types.String   `tfsdk:"region"`
	EndpointURL        types.String   `tfsdk:"endpoint_url"`
	MinioBucketName    types.String   `tfsdk:"minio_bucket_name"`
	PublicAccess       types.Bool     `tfsdk:"public_access"`
	VersioningEnabled  types.Bool     `tfsdk:"versioning_enabled"`
	EncryptionEnabled  types.Bool     `tfsdk:"encryption_enabled"`
	EncryptionType     types.String   `tfsdk:"encryption_type"`
	EncryptionKMSKeyID types.String   `tfsdk:"encryption_kms_key_id"`
	ObjectLockEnabled  types.Bool     `tfsdk:"object_lock_enabled"`
	ObjectLockMode     types.String   `tfsdk:"object_lock_default_retention_mode"`
	ObjectLockDays     types.Int64    `tfsdk:"object_lock_default_retention_days"`
	SizeLimitBytes     types.Int64    `tfsdk:"size_limit_bytes"`
	CorsConfiguration  types.String   `tfsdk:"cors_configuration"`
	LifecycleRules     types.String   `tfsdk:"lifecycle_rules"`
	Tags               types.Map      `tfsdk:"tags"`
	SizeBytes          types.Int64    `tfsdk:"size_bytes"`
	ObjectCount        types.Int64    `tfsdk:"object_count"`
	MonthlyCostCents   types.Int64    `tfsdk:"monthly_cost_cents"`
	MonthlyCost        types.Float64  `tfsdk:"monthly_cost"`
	CreatedAt          types.String   `tfsdk:"created_at"`
	UpdatedAt          types.String   `tfsdk:"updated_at"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

func NewStorageBucketResource() resource.Resource {
	return &StorageBucketResource{}
}

func (r *StorageBucketResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_bucket"
}

func (r *StorageBucketResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DanubeData S3-compatible storage bucket.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for the storage bucket.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the storage bucket. Must follow S3 bucket naming rules (3-63 chars, lowercase alphanumeric with hyphens).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`),
						"must follow S3 bucket naming rules (lowercase alphanumeric with hyphens)",
					),
					stringvalidator.LengthBetween(3, 63),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable display name for the bucket.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"status": schema.StringAttribute{
				Description: "Current status of the storage bucket (pending, active, error, destroying).",
				Computed:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region for the storage bucket (fsn1).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("fsn1"),
				},
			},
			"endpoint_url": schema.StringAttribute{
				Description: "S3 endpoint URL for accessing the bucket.",
				Computed:    true,
			},
			"minio_bucket_name": schema.StringAttribute{
				Description: "Internal MinIO bucket name (includes team prefix).",
				Computed:    true,
			},
			"public_access": schema.BoolAttribute{
				Description: "Whether the bucket has public read access enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"versioning_enabled": schema.BoolAttribute{
				Description: "Whether object versioning is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"encryption_enabled": schema.BoolAttribute{
				Description: "Whether server-side encryption is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"encryption_type": schema.StringAttribute{
				Description: "Encryption type (none, sse-s3, sse-kms).",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("none", "sse-s3", "sse-kms"),
				},
			},
			"encryption_kms_key_id": schema.StringAttribute{
				Description: "KMS key ID required for sse-kms. The API does not return this value on read.",
				Optional:    true, Sensitive: true,
			},
			"object_lock_enabled": schema.BoolAttribute{
				Description: "Enable Object Lock at bucket creation. Cannot be disabled later.",
				Optional:    true, Computed: true,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown(), boolplanmodifier.RequiresReplace()},
			},
			"object_lock_default_retention_mode": schema.StringAttribute{
				Description: "Default Object Lock retention mode: GOVERNANCE or COMPLIANCE. Set together with days.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.OneOf("GOVERNANCE", "COMPLIANCE")},
			},
			"object_lock_default_retention_days": schema.Int64Attribute{
				Description: "Default Object Lock retention in days (1 to 36500). Set together with mode.",
				Optional:    true,
			},
			"size_limit_bytes": schema.Int64Attribute{
				Description: "Bucket quota in bytes, at least 1 GiB. Applied after creation; null removes it.",
				Optional:    true,
			},
			"cors_configuration": schema.StringAttribute{
				Description: "JSON S3 CORS configuration (use jsonencode with CORSRules).",
				Optional:    true,
			},
			"lifecycle_rules": schema.StringAttribute{
				Description: "JSON array of S3 lifecycle rules (use jsonencode).",
				Optional:    true,
			},
			"tags": schema.MapAttribute{
				Description: "Bucket tags as name to value map (up to 50 tags).",
				Optional:    true, ElementType: types.StringType,
			},
			"size_bytes": schema.Int64Attribute{
				Description: "Current size of the bucket in bytes.",
				Computed:    true,
			},
			"object_count": schema.Int64Attribute{
				Description: "Number of objects in the bucket.",
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
			"created_at": schema.StringAttribute{
				Description: "Timestamp when the bucket was created.",
				Computed:    true,
			},
			"updated_at": schema.StringAttribute{
				Description: "Timestamp when the bucket was last updated.",
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

func (r *StorageBucketResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *StorageBucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StorageBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get timeout
	createTimeout, diags := data.Timeouts.Create(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Build create request
	createReq := client.CreateStorageBucketRequest{
		Name:              data.Name.ValueString(),
		Region:            data.Region.ValueString(),
		VersioningEnabled: data.VersioningEnabled.ValueBool(),
		PublicAccess:      data.PublicAccess.ValueBool(),
		EncryptionEnabled: data.EncryptionEnabled.ValueBool(),
		ObjectLockEnabled: data.ObjectLockEnabled.ValueBool(),
	}
	if data.ObjectLockEnabled.ValueBool() && !data.VersioningEnabled.ValueBool() {
		resp.Diagnostics.AddError("Invalid Object Lock settings", "Object Lock requires versioning_enabled = true.")
		return
	}
	if data.ObjectLockMode.IsNull() != data.ObjectLockDays.IsNull() || (!data.ObjectLockMode.IsNull() && !data.ObjectLockEnabled.ValueBool()) {
		resp.Diagnostics.AddError("Invalid Object Lock settings", "Retention mode and days must be set together on an Object Lock bucket.")
		return
	}
	if !data.ObjectLockMode.IsNull() {
		mode, days := data.ObjectLockMode.ValueString(), data.ObjectLockDays.ValueInt64()
		if days < 1 || days > 36500 {
			resp.Diagnostics.AddError("Invalid retention days", "Retention days must be between 1 and 36500.")
			return
		}
		createReq.ObjectLockMode, createReq.ObjectLockDays = &mode, &days
	}
	if !data.SizeLimitBytes.IsNull() && data.SizeLimitBytes.ValueInt64() < 1073741824 {
		resp.Diagnostics.AddError("Invalid size limit", "Size limit must be at least 1 GiB.")
		return
	}
	if !data.EncryptionKMSKeyID.IsNull() {
		createReq.EncryptionKMSKeyID = data.EncryptionKMSKeyID.ValueString()
	}
	if !data.Tags.IsNull() {
		resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &createReq.Tags, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	if !data.CorsConfiguration.IsNull() {
		createReq.CorsConfiguration = json.RawMessage(data.CorsConfiguration.ValueString())
		if !json.Valid(createReq.CorsConfiguration) {
			resp.Diagnostics.AddError("Invalid CORS configuration", "CORS configuration must be valid JSON.")
			return
		}
	}
	if !data.LifecycleRules.IsNull() {
		createReq.LifecycleRules = json.RawMessage(data.LifecycleRules.ValueString())
		if !json.Valid(createReq.LifecycleRules) {
			resp.Diagnostics.AddError("Invalid lifecycle rules", "Lifecycle rules must be valid JSON.")
			return
		}
	}

	if !data.DisplayName.IsNull() && !data.DisplayName.IsUnknown() {
		createReq.DisplayName = data.DisplayName.ValueString()
	}

	if !data.EncryptionType.IsNull() && !data.EncryptionType.IsUnknown() {
		createReq.EncryptionType = data.EncryptionType.ValueString()
	}
	if (createReq.EncryptionType == "sse-kms") != (createReq.EncryptionKMSKeyID != "") ||
		(createReq.EncryptionEnabled && createReq.EncryptionType == "none") ||
		(!createReq.EncryptionEnabled && createReq.EncryptionType != "none") {
		resp.Diagnostics.AddError("Invalid encryption settings", "Use type none when encryption is disabled; use sse-s3 or sse-kms when enabled. sse-kms requires a KMS key ID.")
		return
	}

	tflog.Debug(ctx, "Creating storage bucket", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	// Create storage bucket
	bucket, err := r.client.CreateStorageBucket(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create storage bucket", err.Error())
		return
	}

	data.ID = types.StringValue(bucket.ID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), data.ID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Storage bucket created, waiting for active state", map[string]interface{}{
		"id":   bucket.ID,
		"name": bucket.Name,
	})

	// Wait for bucket to be active
	err = r.client.WaitForStorageBucketStatus(ctx, bucket.ID, "active", createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Storage bucket failed to reach active state",
			fmt.Sprintf("Bucket %s did not reach active state: %s", bucket.ID, err),
		)
		return
	}

	// Refresh state after bucket is active
	bucket, err = r.client.GetStorageBucket(ctx, bucket.ID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read storage bucket after creation", err.Error())
		return
	}
	if !data.SizeLimitBytes.IsNull() {
		limit := data.SizeLimitBytes.ValueInt64()
		limitPtr := &limit
		bucket, err = r.client.UpdateStorageBucket(ctx, bucket.ID, client.UpdateStorageBucketRequest{SizeLimitBytes: &limitPtr})
		if err != nil {
			resp.Diagnostics.AddError("Failed to set bucket size limit", err.Error())
			return
		}
	}

	r.mapBucketToState(bucket, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StorageBucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StorageBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := r.client.GetStorageBucket(ctx, data.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read storage bucket", err.Error())
		return
	}

	r.mapBucketToState(bucket, &data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StorageBucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StorageBucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state StorageBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var configuredType types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("encryption_type"), &configuredType)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !data.EncryptionEnabled.Equal(state.EncryptionEnabled) && configuredType.IsNull() {
		resp.Diagnostics.AddError("Encryption type required", "Set encryption_type to none when disabling encryption, or sse-s3 or sse-kms when enabling it.")
		return
	}
	if !configuredType.IsNull() && ((data.EncryptionEnabled.ValueBool() && configuredType.ValueString() == "none") || (!data.EncryptionEnabled.ValueBool() && configuredType.ValueString() != "none")) {
		resp.Diagnostics.AddError("Invalid encryption settings", "encryption_enabled and encryption_type must agree.")
		return
	}

	// Get timeout
	updateTimeout, diags := data.Timeouts.Update(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Build update request
	updateReq := client.UpdateStorageBucketRequest{}
	hasChanges := false

	if !data.DisplayName.Equal(state.DisplayName) {
		if data.DisplayName.IsNull() {
			resp.Diagnostics.AddError("Cannot clear display name", "The API does not support clearing display_name after creation. Set a new value instead.")
			return
		}
		displayName := data.DisplayName.ValueString()
		updateReq.DisplayName = &displayName
		hasChanges = true
	}

	if !data.VersioningEnabled.Equal(state.VersioningEnabled) {
		versioning := data.VersioningEnabled.ValueBool()
		updateReq.VersioningEnabled = &versioning
		hasChanges = true
	}

	if !data.PublicAccess.Equal(state.PublicAccess) {
		publicAccess := data.PublicAccess.ValueBool()
		updateReq.PublicAccess = &publicAccess
		hasChanges = true
	}

	if !data.EncryptionEnabled.Equal(state.EncryptionEnabled) {
		encryption := data.EncryptionEnabled.ValueBool()
		updateReq.EncryptionEnabled = &encryption
		hasChanges = true
	}

	if !data.EncryptionType.IsUnknown() && !data.EncryptionType.Equal(state.EncryptionType) {
		encType := data.EncryptionType.ValueString()
		updateReq.EncryptionType = &encType
		hasChanges = true
	}
	if !data.EncryptionKMSKeyID.Equal(state.EncryptionKMSKeyID) {
		if data.EncryptionKMSKeyID.IsNull() && data.EncryptionType.ValueString() == "sse-kms" {
			resp.Diagnostics.AddError("Invalid encryption settings", "A bucket using sse-kms must retain its KMS key ID or change encryption_type.")
			return
		}
		keyID := data.EncryptionKMSKeyID.ValueString()
		if !data.EncryptionKMSKeyID.IsNull() {
			updateReq.EncryptionKMSKeyID = &keyID
		}
		hasChanges = true
	}
	if !data.ObjectLockMode.Equal(state.ObjectLockMode) || !data.ObjectLockDays.Equal(state.ObjectLockDays) {
		if !data.ObjectLockEnabled.ValueBool() || data.ObjectLockMode.IsNull() != data.ObjectLockDays.IsNull() {
			resp.Diagnostics.AddError("Invalid Object Lock settings", "Retention mode and days must be set together on an Object Lock bucket.")
			return
		}
		var mode *string
		var days *int64
		if !data.ObjectLockMode.IsNull() {
			m, d := data.ObjectLockMode.ValueString(), data.ObjectLockDays.ValueInt64()
			if d < 1 || d > 36500 {
				resp.Diagnostics.AddError("Invalid retention days", "Retention days must be between 1 and 36500.")
				return
			}
			mode, days = &m, &d
		}
		updateReq.ObjectLockMode, updateReq.ObjectLockDays = &mode, &days
		hasChanges = true
	}
	if !data.SizeLimitBytes.Equal(state.SizeLimitBytes) {
		var limit *int64
		if !data.SizeLimitBytes.IsNull() {
			value := data.SizeLimitBytes.ValueInt64()
			if value < 1073741824 {
				resp.Diagnostics.AddError("Invalid size limit", "Size limit must be at least 1 GiB.")
				return
			}
			limit = &value
		}
		updateReq.SizeLimitBytes = &limit
		hasChanges = true
	}
	if !data.Tags.Equal(state.Tags) {
		tags := map[string]string{}
		if !data.Tags.IsNull() {
			resp.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tags, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		updateReq.Tags = &tags
		hasChanges = true
	}
	if !data.CorsConfiguration.Equal(state.CorsConfiguration) {
		updateReq.CorsConfiguration = json.RawMessage(data.CorsConfiguration.ValueString())
		if data.CorsConfiguration.IsNull() {
			updateReq.CorsConfiguration = json.RawMessage(`{"CORSRules":[]}`)
		}
		if !json.Valid(updateReq.CorsConfiguration) {
			resp.Diagnostics.AddError("Invalid CORS configuration", "CORS configuration must be valid JSON.")
			return
		}
		hasChanges = true
	}
	if !data.LifecycleRules.Equal(state.LifecycleRules) {
		updateReq.LifecycleRules = json.RawMessage(data.LifecycleRules.ValueString())
		if data.LifecycleRules.IsNull() {
			updateReq.LifecycleRules = json.RawMessage(`[]`)
		}
		if !json.Valid(updateReq.LifecycleRules) {
			resp.Diagnostics.AddError("Invalid lifecycle rules", "Lifecycle rules must be valid JSON.")
			return
		}
		hasChanges = true
	}
	if updateReq.EncryptionType != nil && *updateReq.EncryptionType == "sse-kms" && data.EncryptionKMSKeyID.IsNull() {
		resp.Diagnostics.AddError("Invalid encryption settings", "sse-kms requires encryption_kms_key_id.")
		return
	}

	if hasChanges {
		tflog.Debug(ctx, "Updating storage bucket", map[string]interface{}{
			"id": data.ID.ValueString(),
		})

		bucket, err := r.client.UpdateStorageBucket(ctx, data.ID.ValueString(), updateReq)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update storage bucket", err.Error())
			return
		}

		r.mapBucketToState(bucket, &data)
	}

	// Any path that did not go through UpdateStorageBucket still holds the plan's
	// unknown values for every computed attribute, because only `id` carries
	// UseStateForUnknown. Refresh from the API before writing state, or Terraform
	// rejects the apply. This covers a timeouts-only change and any other update
	// the hasChanges set does not recognise.
	if !hasChanges {
		bucket, err := r.client.GetStorageBucket(ctx, data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read storage bucket after update", err.Error())
			return
		}

		r.mapBucketToState(bucket, &data)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *StorageBucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data StorageBucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get timeout
	deleteTimeout, diags := data.Timeouts.Delete(ctx, 10*time.Minute)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	bucketID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting storage bucket", map[string]interface{}{
		"id": bucketID,
	})

	// Delete the bucket
	err := r.client.DeleteStorageBucket(ctx, bucketID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete storage bucket", err.Error())
		return
	}

	// Wait for bucket to be deleted
	err = r.client.WaitForStorageBucketDeletion(ctx, bucketID, deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			"Storage bucket failed to be deleted",
			fmt.Sprintf("Bucket %s was not deleted within the timeout: %s", bucketID, err),
		)
		return
	}
}

func (r *StorageBucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *StorageBucketResource) mapBucketToState(bucket *client.StorageBucket, data *StorageBucketResourceModel) {
	data.ID = types.StringValue(bucket.ID)
	data.Name = types.StringValue(bucket.Name)
	data.Status = types.StringValue(bucket.Status)
	data.Region = types.StringValue(bucket.Region)
	data.EndpointURL = types.StringValue(bucket.EndpointURL)
	data.MinioBucketName = types.StringValue(bucket.MinioBucketName)
	data.PublicAccess = types.BoolValue(bucket.PublicAccess)
	data.VersioningEnabled = types.BoolValue(bucket.VersioningEnabled)
	data.EncryptionEnabled = types.BoolValue(bucket.EncryptionEnabled)
	data.ObjectLockEnabled = types.BoolValue(bucket.ObjectLockEnabled)
	data.SizeLimitBytes = types.Int64Null()
	if bucket.SizeLimitBytes != nil {
		data.SizeLimitBytes = types.Int64Value(*bucket.SizeLimitBytes)
	}
	data.ObjectLockMode = types.StringNull()
	if bucket.ObjectLockMode != nil {
		data.ObjectLockMode = types.StringValue(*bucket.ObjectLockMode)
	}
	data.ObjectLockDays = types.Int64Null()
	if bucket.ObjectLockDays != nil {
		data.ObjectLockDays = types.Int64Value(*bucket.ObjectLockDays)
	}
	configuredCors := data.CorsConfiguration
	data.CorsConfiguration = bucketCorsState(bucket.CorsConfiguration)
	if data.CorsConfiguration.IsNull() && !configuredCors.IsNull() && !configuredCors.IsUnknown() &&
		json.Valid([]byte(configuredCors.ValueString())) && bucketCorsState(json.RawMessage(configuredCors.ValueString())).IsNull() {
		data.CorsConfiguration = configuredCors
	}
	configuredLifecycle := data.LifecycleRules
	data.LifecycleRules = bucketJSONState(bucket.LifecycleRules)
	if data.LifecycleRules.IsNull() && !configuredLifecycle.IsNull() && !configuredLifecycle.IsUnknown() &&
		bucketJSONState(json.RawMessage(configuredLifecycle.ValueString())).IsNull() && json.Valid([]byte(configuredLifecycle.ValueString())) {
		data.LifecycleRules = configuredLifecycle
	}
	if len(bucket.Tags) == 0 {
		if data.Tags.IsNull() || data.Tags.IsUnknown() || len(data.Tags.Elements()) != 0 {
			data.Tags = types.MapNull(types.StringType)
		}
	} else {
		value, _ := types.MapValueFrom(context.Background(), types.StringType, map[string]string(bucket.Tags))
		data.Tags = value
	}
	data.SizeBytes = types.Int64Value(bucket.SizeBytes)
	data.ObjectCount = types.Int64Value(int64(bucket.ObjectCount))
	data.MonthlyCostCents = types.Int64Value(int64(bucket.MonthlyCostCents))
	data.MonthlyCost = types.Float64Value(bucket.MonthlyCostDollars)
	data.CreatedAt = types.StringValue(bucket.CreatedAt)
	data.UpdatedAt = types.StringValue(bucket.UpdatedAt)

	if bucket.DisplayName != nil {
		data.DisplayName = types.StringValue(*bucket.DisplayName)
	} else {
		data.DisplayName = types.StringNull()
	}

	if bucket.EncryptionType != nil {
		data.EncryptionType = types.StringValue(*bucket.EncryptionType)
	} else {
		data.EncryptionType = types.StringNull()
	}
}

func bucketJSONState(raw json.RawMessage) types.String {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "[]" || string(raw) == `{"CORSRules":[]}` {
		return types.StringNull()
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		return types.StringNull()
	}
	return types.StringValue(compact.String())
}

func bucketCorsState(raw json.RawMessage) types.String {
	var config map[string]json.RawMessage
	if err := json.Unmarshal(raw, &config); err != nil || config == nil {
		return types.StringNull()
	}
	var rules []json.RawMessage
	if err := json.Unmarshal(config["CORSRules"], &rules); err != nil {
		return bucketJSONState(raw)
	}
	if len(rules) > 0 && isPlatformCorsRule(rules[0]) {
		rules = rules[1:]
	}
	if len(rules) == 0 {
		return types.StringNull()
	}
	config["CORSRules"], _ = json.Marshal(rules)
	userConfig, _ := json.Marshal(config)
	return types.StringValue(string(userConfig))
}

func isPlatformCorsRule(raw json.RawMessage) bool {
	var rule struct {
		AllowedOrigins []string
		AllowedMethods []string
		AllowedHeaders []string
		ExposeHeaders  []string
		MaxAgeSeconds  int
	}
	if err := json.Unmarshal(raw, &rule); err != nil {
		return false
	}
	return len(rule.AllowedOrigins) == 1 &&
		(strings.HasPrefix(rule.AllowedOrigins[0], "https://console.") || strings.HasPrefix(rule.AllowedOrigins[0], "http://console.")) &&
		slices.Equal(rule.AllowedMethods, []string{"GET", "PUT", "POST", "DELETE", "HEAD"}) &&
		slices.Equal(rule.AllowedHeaders, []string{"*"}) &&
		slices.Equal(rule.ExposeHeaders, []string{"ETag", "x-amz-request-id", "x-amz-id-2"}) &&
		rule.MaxAgeSeconds == 86400
}
