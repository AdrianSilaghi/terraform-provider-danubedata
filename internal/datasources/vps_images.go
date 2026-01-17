package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &VpsImagesDataSource{}
var _ datasource.DataSourceWithConfigure = &VpsImagesDataSource{}

type VpsImagesDataSource struct {
	client *client.Client
}

type VpsImagesDataSourceModel struct {
	Images []VpsImageModel `tfsdk:"images"`
}

type VpsImageModel struct {
	ID          types.String `tfsdk:"id"`
	Image       types.String `tfsdk:"image"`
	Label       types.String `tfsdk:"label"`
	Description types.String `tfsdk:"description"`
	Distro      types.String `tfsdk:"distro"`
	Version     types.String `tfsdk:"version"`
	Family      types.String `tfsdk:"family"`
	DefaultUser types.String `tfsdk:"default_user"`
}

func NewVpsImagesDataSource() datasource.DataSource {
	return &VpsImagesDataSource{}
}

func (d *VpsImagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_images"
}

func (d *VpsImagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves available VPS operating system images.",
		Attributes: map[string]schema.Attribute{
			"images": schema.ListNestedAttribute{
				Description: "List of available VPS images.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Image identifier (e.g., 'ubuntu-24.04').",
							Computed:    true,
						},
						"image": schema.StringAttribute{
							Description: "Full image reference.",
							Computed:    true,
						},
						"label": schema.StringAttribute{
							Description: "Human-readable label.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Image description.",
							Computed:    true,
						},
						"distro": schema.StringAttribute{
							Description: "Distribution (ubuntu, debian, alma, rocky, fedora, alpine).",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Distribution version.",
							Computed:    true,
						},
						"family": schema.StringAttribute{
							Description: "OS family (debian, redhat, fedora, alpine).",
							Computed:    true,
						},
						"default_user": schema.StringAttribute{
							Description: "Default SSH user.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *VpsImagesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VpsImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VpsImagesDataSourceModel

	images, err := d.client.ListVpsImages(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list VPS images", err.Error())
		return
	}

	data.Images = make([]VpsImageModel, len(images))
	for i, img := range images {
		data.Images[i] = VpsImageModel{
			ID:          types.StringValue(img.ID),
			Image:       types.StringValue(img.Image),
			Label:       types.StringValue(img.Label),
			Description: types.StringValue(img.Description),
			Distro:      types.StringValue(img.Distro),
			Version:     types.StringValue(img.GetVersion()),
			DefaultUser: types.StringValue(img.DefaultUser),
		}
		if img.Family != nil {
			data.Images[i].Family = types.StringValue(*img.Family)
		} else {
			data.Images[i].Family = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
