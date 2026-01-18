package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DatabaseProvidersDataSource{}

type DatabaseProvidersDataSource struct{}

type DatabaseProvidersDataSourceModel struct {
	Providers []DatabaseProviderModel `tfsdk:"providers"`
}

type DatabaseProviderModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Version     types.String `tfsdk:"version"`
	DefaultPort types.Int64  `tfsdk:"default_port"`
}

func NewDatabaseProvidersDataSource() datasource.DataSource {
	return &DatabaseProvidersDataSource{}
}

func (d *DatabaseProvidersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_providers"
}

func (d *DatabaseProvidersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists available database providers (MySQL, PostgreSQL, MariaDB).",
		Attributes: map[string]schema.Attribute{
			"providers": schema.ListNestedAttribute{
				Description: "List of available database providers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: "Provider ID to use when creating database instances.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Provider name (MySQL, PostgreSQL, MariaDB).",
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

func (d *DatabaseProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Return static provider information
	// These match the seeded providers in the DanubeData database
	data := DatabaseProvidersDataSourceModel{
		Providers: []DatabaseProviderModel{
			{
				ID:          types.Int64Value(1),
				Name:        types.StringValue("MySQL"),
				Type:        types.StringValue("mysql"),
				Description: types.StringValue("World's most popular open source database with proven reliability and performance."),
				Version:     types.StringValue("8.0"),
				DefaultPort: types.Int64Value(3306),
			},
			{
				ID:          types.Int64Value(2),
				Name:        types.StringValue("PostgreSQL"),
				Type:        types.StringValue("postgresql"),
				Description: types.StringValue("Advanced open source relational database with powerful features and reliability."),
				Version:     types.StringValue("16"),
				DefaultPort: types.Int64Value(5432),
			},
			{
				ID:          types.Int64Value(3),
				Name:        types.StringValue("MariaDB"),
				Type:        types.StringValue("mariadb"),
				Description: types.StringValue("MySQL-compatible database with enhanced features, performance and modern architecture."),
				Version:     types.StringValue("11.4"),
				DefaultPort: types.Int64Value(3306),
			},
		},
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
