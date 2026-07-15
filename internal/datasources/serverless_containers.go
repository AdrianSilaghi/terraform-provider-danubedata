package datasources

import (
	"context"
	"fmt"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ServerlessContainersDataSource{}
var _ datasource.DataSourceWithConfigure = &ServerlessContainersDataSource{}

type ServerlessContainersDataSource struct {
	client *client.Client
}

type ServerlessContainersDataSourceModel struct {
	Containers []ServerlessContainerModel `tfsdk:"containers"`
}

type ServerlessContainerModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Status           types.String `tfsdk:"status"`
	DeploymentType   types.String `tfsdk:"deployment_type"`
	Image            types.String `tfsdk:"image"`
	RepositoryURL    types.String `tfsdk:"repository_url"`
	RepositoryBranch types.String `tfsdk:"repository_branch"`
	URL              types.String `tfsdk:"url"`
	Port             types.Int64  `tfsdk:"port"`
	MinScale         types.Int64  `tfsdk:"min_scale"`
	MaxScale         types.Int64  `tfsdk:"max_scale"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

func NewServerlessContainersDataSource() datasource.DataSource {
	return &ServerlessContainersDataSource{}
}

func (d *ServerlessContainersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_serverless_containers"
}

func (d *ServerlessContainersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all serverless containers in your account.",
		Attributes: map[string]schema.Attribute{
			"containers": schema.ListNestedAttribute{
				Description: "List of serverless containers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Unique identifier for the serverless container.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the serverless container.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "Current status of the container.",
							Computed:    true,
						},
						"deployment_type": schema.StringAttribute{
							Description: "Deployment type (docker_image, git_repository, or zip_upload).",
							Computed:    true,
						},
						"image": schema.StringAttribute{
							Description: "Container image reference (for docker_image deployments).",
							Computed:    true,
						},
						"repository_url": schema.StringAttribute{
							Description: "Git repository URL (for git_repository deployments).",
							Computed:    true,
						},
						"repository_branch": schema.StringAttribute{
							Description: "Git branch (for git_repository deployments).",
							Computed:    true,
						},
						"url": schema.StringAttribute{
							Description: "Public HTTPS URL for the container.",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "Container port.",
							Computed:    true,
						},
						"min_scale": schema.Int64Attribute{
							Description: "Minimum number of instances.",
							Computed:    true,
						},
						"max_scale": schema.Int64Attribute{
							Description: "Maximum number of instances.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "Timestamp when the container was created.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ServerlessContainersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerlessContainersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Expected configured client. Please report this issue to the provider developers.",
		)
		return
	}

	var data ServerlessContainersDataSourceModel

	containers, err := d.client.ListServerless(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list serverless containers", err.Error())
		return
	}

	data.Containers = make([]ServerlessContainerModel, len(containers))
	for i, c := range containers {
		data.Containers[i] = ServerlessContainerModel{
			ID:               types.StringValue(c.ID),
			Name:             types.StringValue(c.Name),
			Status:           types.StringValue(c.Status),
			DeploymentType:   types.StringValue(c.DeploymentType),
			RepositoryBranch: types.StringValue(c.RepositoryBranch),
			URL:              types.StringValue(c.URL),
			Port:             types.Int64Value(int64(c.Port)),
			MinScale:         types.Int64Value(int64(c.MinScale)),
			MaxScale:         types.Int64Value(int64(c.MaxScale)),
			CreatedAt:        types.StringValue(c.CreatedAt),
		}

		if c.Image != nil {
			data.Containers[i].Image = types.StringValue(*c.Image)
		} else {
			data.Containers[i].Image = types.StringNull()
		}

		if c.RepositoryURL != nil {
			data.Containers[i].RepositoryURL = types.StringValue(*c.RepositoryURL)
		} else {
			data.Containers[i].RepositoryURL = types.StringNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
