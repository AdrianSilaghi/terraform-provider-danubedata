package provider

import (
	"context"
	"os"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/datasources"
	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &DanubeDataProvider{}

type DanubeDataProvider struct {
	version string
}

type DanubeDataProviderModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	APIToken types.String `tfsdk:"api_token"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DanubeDataProvider{
			version: version,
		}
	}
}

func (p *DanubeDataProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "danubedata"
	resp.Version = p.version
}

func (p *DanubeDataProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for DanubeData managed infrastructure. Manage VPS instances, databases, caches, and object storage.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "Base URL for the DanubeData API. Defaults to https://danubedata.ro/api/v1. Can also be set via DANUBEDATA_BASE_URL environment variable.",
				Optional:    true,
			},
			"api_token": schema.StringAttribute{
				Description: "API token for DanubeData authentication. Can also be set via DANUBEDATA_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *DanubeDataProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DanubeDataProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Environment variable fallbacks
	baseURL := os.Getenv("DANUBEDATA_BASE_URL")
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() {
		baseURL = config.BaseURL.ValueString()
	}
	if baseURL == "" {
		baseURL = "https://danubedata.ro/api/v1"
	}

	apiToken := os.Getenv("DANUBEDATA_API_TOKEN")
	if !config.APIToken.IsNull() && !config.APIToken.IsUnknown() {
		apiToken = config.APIToken.ValueString()
	}

	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing API Token",
			"The provider cannot create the DanubeData API client as there is a missing or empty value for the API token. "+
				"Set the api_token value in the configuration or use the DANUBEDATA_API_TOKEN environment variable.",
		)
		return
	}

	// Create client
	c := client.New(client.Config{
		BaseURL:   baseURL,
		APIToken:  apiToken,
		UserAgent: "terraform-provider-danubedata/" + p.version,
	})

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *DanubeDataProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Compute
		resources.NewVpsResource,
		resources.NewServerlessResource,

		// Data Services
		resources.NewCacheResource,
		resources.NewDatabaseResource,

		// Storage
		resources.NewStorageBucketResource,
		resources.NewStorageAccessKeyResource,

		// Security
		resources.NewSshKeyResource,
		resources.NewFirewallResource,

		// Snapshots
		resources.NewVpsSnapshotResource,
	}
}

func (p *DanubeDataProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Listing data sources
		datasources.NewVpsImagesDataSource,
		datasources.NewCacheProvidersDataSource,
		datasources.NewDatabaseProvidersDataSource,
		datasources.NewSshKeysDataSource,

		// Resource listing data sources
		datasources.NewVpssDataSource,
		datasources.NewDatabasesDataSource,
		datasources.NewCachesDataSource,
		datasources.NewFirewallsDataSource,
		datasources.NewServerlessContainersDataSource,
		datasources.NewStorageBucketsDataSource,
		datasources.NewStorageAccessKeysDataSource,
		datasources.NewVpsSnapshotsDataSource,
	}
}
