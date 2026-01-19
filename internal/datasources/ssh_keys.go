package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SshKeysDataSource{}
var _ datasource.DataSourceWithConfigure = &SshKeysDataSource{}

type SshKeysDataSource struct {
	client *client.Client
}

type SshKeysDataSourceModel struct {
	Keys []SshKeyModel `tfsdk:"keys"`
}

type SshKeyModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Fingerprint types.String `tfsdk:"fingerprint"`
	PublicKey   types.String `tfsdk:"public_key"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func NewSshKeysDataSource() datasource.DataSource {
	return &SshKeysDataSource{}
}

func (d *SshKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keys"
}

func (d *SshKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all SSH keys available for VPS authentication.",
		Attributes: map[string]schema.Attribute{
			"keys": schema.ListNestedAttribute{
				Description: "List of SSH keys.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier for the SSH key.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the SSH key.",
							Computed:    true,
						},
						"fingerprint": schema.StringAttribute{
							Description: "SHA256 fingerprint of the SSH key.",
							Computed:    true,
						},
						"public_key": schema.StringAttribute{
							Description: "The SSH public key.",
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

func (d *SshKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SshKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var data SshKeysDataSourceModel

	keys, err := d.client.ListSshKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list SSH keys", err.Error())
		return
	}

	data.Keys = make([]SshKeyModel, len(keys))
	for i, key := range keys {
		data.Keys[i] = SshKeyModel{
			ID:          types.StringValue(fmt.Sprintf("%d", key.ID)),
			Name:        types.StringValue(key.Name),
			Fingerprint: types.StringValue(key.Fingerprint),
			PublicKey:   types.StringValue(key.PublicKey),
			CreatedAt:   types.StringValue(key.CreatedAt),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
