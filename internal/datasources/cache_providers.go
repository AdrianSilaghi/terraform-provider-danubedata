package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CacheProvidersDataSource{}

type CacheProvidersDataSource struct{}

type CacheProvidersDataSourceModel struct {
	Providers []CacheProviderModel `tfsdk:"providers"`
}

type CacheProviderModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Version     types.String `tfsdk:"version"`
	DefaultPort types.Int64  `tfsdk:"default_port"`
}

func NewCacheProvidersDataSource() datasource.DataSource {
	return &CacheProvidersDataSource{}
}

func (d *CacheProvidersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cache_providers"
}

func (d *CacheProvidersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists available cache providers (Redis, Valkey, Dragonfly).",
		Attributes: map[string]schema.Attribute{
			"providers": schema.ListNestedAttribute{
				Description: "List of available cache providers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "Provider ID to use when creating cache instances.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Provider name (Redis, Valkey, Dragonfly).",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "Provider type identifier.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Provider description.",
							Computed:    true,
						},
						"version": schema.StringAttribute{
							Description: "Default version.",
							Computed:    true,
						},
						"default_port": schema.Int64Attribute{
							Description: "Default port number.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *CacheProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Return static provider information
	// These match the seeded providers in the DanubeData database
	data := CacheProvidersDataSourceModel{
		Providers: []CacheProviderModel{
			{
				ID:          types.Int64Value(1),
				Name:        types.StringValue("Redis"),
				Type:        types.StringValue("redis"),
				Description: types.StringValue("High-performance in-memory data structure store, used as a database, cache, and message broker."),
				Version:     types.StringValue("7.2"),
				DefaultPort: types.Int64Value(6379),
			},
			{
				ID:          types.Int64Value(2),
				Name:        types.StringValue("Valkey"),
				Type:        types.StringValue("valkey"),
				Description: types.StringValue("Open source high-performance data store forked from Redis. Fully compatible with Redis protocol."),
				Version:     types.StringValue("8.0"),
				DefaultPort: types.Int64Value(6379),
			},
			{
				ID:          types.Int64Value(3),
				Name:        types.StringValue("Dragonfly"),
				Type:        types.StringValue("dragonfly"),
				Description: types.StringValue("A modern replacement for Redis that is fully compatible with Redis API but built for cloud workloads."),
				Version:     types.StringValue("1.15"),
				DefaultPort: types.Int64Value(6379),
			},
		},
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
