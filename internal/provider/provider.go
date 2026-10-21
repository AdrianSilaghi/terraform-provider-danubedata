package provider

import (
	"context"
	"os"
	"strconv"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/datasources"
	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &DanubeDataProvider{}

type DanubeDataProvider struct {
	version string
}

type DanubeDataProviderModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	APIToken types.String `tfsdk:"api_token"`
	TeamID   types.Int64  `tfsdk:"team_id"`
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
			"team_id": schema.Int64Attribute{
				Description: "Project ID sent as X-Team-Id on every request. Can also be set via DANUBEDATA_TEAM_ID. If omitted, the API uses the token's project or the account's current project.",
				Optional:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
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
	for name, unknown := range map[string]bool{
		"base_url":  config.BaseURL.IsUnknown(),
		"api_token": config.APIToken.IsUnknown(),
		"team_id":   config.TeamID.IsUnknown(),
	} {
		if unknown {
			resp.Diagnostics.AddAttributeError(path.Root(name), "Unknown Provider Configuration", "The "+name+" value must be known before configuring the provider.")
		}
	}
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

	teamID := config.TeamID.ValueInt64()
	if config.TeamID.IsNull() {
		if value := os.Getenv("DANUBEDATA_TEAM_ID"); value != "" {
			var err error
			teamID, err = strconv.ParseInt(value, 10, 64)
			if err != nil || teamID <= 0 {
				resp.Diagnostics.AddAttributeError(path.Root("team_id"), "Invalid Project ID", "DANUBEDATA_TEAM_ID must be a positive integer project ID.")
				return
			}
		}
	} else if teamID <= 0 {
		resp.Diagnostics.AddAttributeError(path.Root("team_id"), "Invalid Project ID", "team_id must be a positive integer project ID.")
		return
	}

	// Create client
	c := client.New(client.Config{
		BaseURL:   baseURL,
		APIToken:  apiToken,
		UserAgent: "terraform-provider-danubedata/" + p.version,
		TeamID:    teamID,
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
		resources.NewDatabaseReplicaResource,
		resources.NewParameterGroupResource,

		// Storage
		resources.NewStorageBucketResource,
		resources.NewStorageAccessKeyResource,

		// Security
		resources.NewSshKeyResource,
		resources.NewFirewallResource,

		// Snapshots
		resources.NewVpsSnapshotResource,
		resources.NewCacheSnapshotResource,
		resources.NewDatabaseSnapshotResource,

		// Static sites
		resources.NewStaticSiteResource,
		resources.NewStaticSiteDomainResource,
	}
}

func (p *DanubeDataProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Listing data sources
		datasources.NewVpsImagesDataSource,
		datasources.NewCacheProvidersDataSource,
		datasources.NewDatabaseProvidersDataSource,
		datasources.NewSshKeysDataSource,
		datasources.NewParameterGroupsDataSource,

		// Resource listing data sources
		datasources.NewVpssDataSource,
		datasources.NewDatabasesDataSource,
		datasources.NewCachesDataSource,
		datasources.NewFirewallsDataSource,
		datasources.NewServerlessContainersDataSource,
		datasources.NewStorageBucketsDataSource,
		datasources.NewStorageAccessKeysDataSource,
		datasources.NewVpsSnapshotsDataSource,
		datasources.NewCacheSnapshotsDataSource,
		datasources.NewDatabaseSnapshotsDataSource,
		datasources.NewStaticSitesDataSource,
	}
}
