package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &StorageBucketsDataSource{}
var _ datasource.DataSourceWithConfigure = &StorageBucketsDataSource{}

type StorageBucketsDataSource struct {
	client *client.Client
}

type StorageBucketsDataSourceModel struct {
	Buckets []StorageBucketModel `tfsdk:"buckets"`
}

type StorageBucketModel struct {
	ID                types.String  `tfsdk:"id"`
	Name              types.String  `tfsdk:"name"`
	DisplayName       types.String  `tfsdk:"display_name"`
	Status            types.String  `tfsdk:"status"`
	Region            types.String  `tfsdk:"region"`
	EndpointURL       types.String  `tfsdk:"endpoint_url"`
	PublicURL         types.String  `tfsdk:"public_url"`
	MinioBucketName   types.String  `tfsdk:"minio_bucket_name"`
	PublicAccess      types.Bool    `tfsdk:"public_access"`
	VersioningEnabled types.Bool    `tfsdk:"versioning_enabled"`
	EncryptionEnabled types.Bool    `tfsdk:"encryption_enabled"`
	SizeBytes         types.Int64   `tfsdk:"size_bytes"`
	ObjectCount       types.Int64   `tfsdk:"object_count"`
	MonthlyCost       types.Float64 `tfsdk:"monthly_cost"`
	CreatedAt         types.String  `tfsdk:"created_at"`
}

func NewStorageBucketsDataSource() datasource.DataSource {
	return &StorageBucketsDataSource{}
}

func (d *StorageBucketsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_buckets"
}

func (d *StorageBucketsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all S3-compatible storage buckets in your account.",
		Attributes: map[string]schema.Attribute{
			"buckets": schema.ListNestedAttribute{
				Description: "List of storage buckets.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier for the storage bucket.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the storage bucket.",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "Human-readable display name.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the bucket.",
							Computed:    true,
						},
						"region": schema.StringAttribute{
							Description: "Region where the bucket is located.",
							Computed:    true,
						},
						"endpoint_url": schema.StringAttribute{
							Description: "S3-compatible endpoint URL.",
							Computed:    true,
						},
						"public_url": schema.StringAttribute{
							Description: "Public URL (if public access enabled).",
							Computed:    true,
						},
						"minio_bucket_name": schema.StringAttribute{
							Description: "Internal bucket name.",
							Computed:    true,
						},
						"public_access": schema.BoolAttribute{
							Description: "Whether public access is enabled.",
							Computed:    true,
						},
						"versioning_enabled": schema.BoolAttribute{
							Description: "Whether versioning is enabled.",
							Computed:    true,
						},
						"encryption_enabled": schema.BoolAttribute{
							Description: "Whether encryption is enabled.",
							Computed:    true,
						},
						"size_bytes": schema.Int64Attribute{
							Description: "Current size in bytes.",
							Computed:    true,
						},
						"object_count": schema.Int64Attribute{
							Description: "Number of objects in the bucket.",
							Computed:    true,
						},
						"monthly_cost": schema.Float64Attribute{
							Description: "Estimated monthly cost.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the bucket was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *StorageBucketsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *StorageBucketsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var data StorageBucketsDataSourceModel

	buckets, err := d.client.ListStorageBuckets(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list storage buckets", err.Error())
		return
	}

	data.Buckets = make([]StorageBucketModel, len(buckets))
	for i, b := range buckets {
		data.Buckets[i] = StorageBucketModel{
			ID:                types.StringValue(b.ID),
			Name:              types.StringValue(b.Name),
			Status:            types.StringValue(b.Status),
			Region:            types.StringValue(b.Region),
			EndpointURL:       types.StringValue(b.EndpointURL),
			MinioBucketName:   types.StringValue(b.MinioBucketName),
			PublicAccess:      types.BoolValue(b.PublicAccess),
			VersioningEnabled: types.BoolValue(b.VersioningEnabled),
			EncryptionEnabled: types.BoolValue(b.EncryptionEnabled),
			SizeBytes:         types.Int64Value(b.SizeBytes),
			ObjectCount:       types.Int64Value(int64(b.ObjectCount)),
			MonthlyCost:       types.Float64Value(b.MonthlyCostDollars),
			CreatedAt:         types.StringValue(b.CreatedAt),
		}
		if b.DisplayName != nil {
			data.Buckets[i].DisplayName = types.StringValue(*b.DisplayName)
		} else {
			data.Buckets[i].DisplayName = types.StringNull()
		}
		if b.PublicURL != nil {
			data.Buckets[i].PublicURL = types.StringValue(*b.PublicURL)
		} else {
			data.Buckets[i].PublicURL = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
