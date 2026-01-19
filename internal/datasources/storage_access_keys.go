package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &StorageAccessKeysDataSource{}
var _ datasource.DataSourceWithConfigure = &StorageAccessKeysDataSource{}

type StorageAccessKeysDataSource struct {
	client *client.Client
}

type StorageAccessKeysDataSourceModel struct {
	Keys []StorageAccessKeyModel `tfsdk:"keys"`
}

type StorageAccessKeyModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	AccessKeyID    types.String `tfsdk:"access_key_id"`
	Status         types.String `tfsdk:"status"`
	AccessType     types.String `tfsdk:"access_type"`
	IsPrefixScoped types.Bool   `tfsdk:"is_prefix_scoped"`
	ExpiresAt      types.String `tfsdk:"expires_at"`
	LastUsedAt     types.String `tfsdk:"last_used_at"`
	IsExpired      types.Bool   `tfsdk:"is_expired"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func NewStorageAccessKeysDataSource() datasource.DataSource {
	return &StorageAccessKeysDataSource{}
}

func (d *StorageAccessKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_access_keys"
}

func (d *StorageAccessKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all S3 storage access keys in your account.",
		Attributes: map[string]schema.Attribute{
			"keys": schema.ListNestedAttribute{
				Description: "List of storage access keys.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier for the access key.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the access key.",
							Computed:    true,
						},
						"access_key_id": schema.StringAttribute{
							Description: "The S3 access key ID for authentication.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the key.",
							Computed:    true,
						},
						"access_type": schema.StringAttribute{
							Description: "Access type (full or restricted).",
							Computed:    true,
						},
						"is_prefix_scoped": schema.BoolAttribute{
							Description: "Whether the key is scoped to specific bucket prefixes.",
							Computed:    true,
						},
						"expires_at": schema.StringAttribute{
							Description: "Expiration timestamp (if set).",
							Computed:    true,
						},
						"last_used_at": schema.StringAttribute{
							Description: "Timestamp when the key was last used.",
							Computed:    true,
						},
						"is_expired": schema.BoolAttribute{
							Description: "Whether the key has expired.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the key was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *StorageAccessKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StorageAccessKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var data StorageAccessKeysDataSourceModel

	keys, err := d.client.ListStorageAccessKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list storage access keys", err.Error())
		return
	}

	data.Keys = make([]StorageAccessKeyModel, len(keys))
	for i, k := range keys {
		data.Keys[i] = StorageAccessKeyModel{
			ID:             types.StringValue(k.ID),
			Name:           types.StringValue(k.Name),
			AccessKeyID:    types.StringValue(k.AccessKeyID),
			Status:         types.StringValue(k.Status),
			AccessType:     types.StringValue(k.AccessType),
			IsPrefixScoped: types.BoolValue(k.IsPrefixScoped),
			IsExpired:      types.BoolValue(k.IsExpired),
			CreatedAt:      types.StringValue(k.CreatedAt),
		}
		if k.ExpiresAt != nil {
			data.Keys[i].ExpiresAt = types.StringValue(*k.ExpiresAt)
		} else {
			data.Keys[i].ExpiresAt = types.StringNull()
		}
		if k.LastUsedAt != nil {
			data.Keys[i].LastUsedAt = types.StringValue(*k.LastUsedAt)
		} else {
			data.Keys[i].LastUsedAt = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
