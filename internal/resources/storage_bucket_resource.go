package resources

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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
	ID                types.String   `tfsdk:"id"`
	Name              types.String   `tfsdk:"name"`
	DisplayName       types.String   `tfsdk:"display_name"`
	Status            types.String   `tfsdk:"status"`
	Region            types.String   `tfsdk:"region"`
	EndpointURL       types.String   `tfsdk:"endpoint_url"`
	MinioBucketName   types.String   `tfsdk:"minio_bucket_name"`
	PublicAccess      types.Bool     `tfsdk:"public_access"`
	VersioningEnabled types.Bool     `tfsdk:"versioning_enabled"`
	EncryptionEnabled types.Bool     `tfsdk:"encryption_enabled"`
	EncryptionType    types.String   `tfsdk:"encryption_type"`
	SizeBytes         types.Int64    `tfsdk:"size_bytes"`
	ObjectCount       types.Int64    `tfsdk:"object_count"`
	MonthlyCostCents  types.Int64    `tfsdk:"monthly_cost_cents"`
	MonthlyCost       types.Float64  `tfsdk:"monthly_cost"`
	CreatedAt         types.String   `tfsdk:"created_at"`
	UpdatedAt         types.String   `tfsdk:"updated_at"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
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
				Default:     stringdefault.StaticString("sse-s3"),
				Validators: []validator.String{
					stringvalidator.OneOf("none", "sse-s3", "sse-kms"),
				},
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
	}

	if !data.DisplayName.IsNull() && !data.DisplayName.IsUnknown() {
		createReq.DisplayName = data.DisplayName.ValueString()
	}

	if !data.EncryptionType.IsNull() && !data.EncryptionType.IsUnknown() {
		createReq.EncryptionType = data.EncryptionType.ValueString()
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

	if !data.EncryptionType.Equal(state.EncryptionType) {
		encType := data.EncryptionType.ValueString()
		updateReq.EncryptionType = &encType
		hasChanges = true
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
