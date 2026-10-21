package datasources

import (
	"context"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func configuredDataSources() []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVpsImagesDataSource, NewVpssDataSource, NewSshKeysDataSource,
		NewCachesDataSource, NewDatabasesDataSource, NewFirewallsDataSource,
		NewServerlessContainersDataSource, NewParameterGroupsDataSource,
		NewStorageBucketsDataSource, NewStorageAccessKeysDataSource,
		NewVpsSnapshotsDataSource, NewCacheSnapshotsDataSource,
		NewDatabaseSnapshotsDataSource, NewStaticSitesDataSource,
	}
}

func TestDataSourceConfigureRejectsUnexpectedProviderData(t *testing.T) {
	ctx := context.Background()
	for _, factory := range configuredDataSources() {
		d := factory()
		var metadata datasource.MetadataResponse
		d.Metadata(ctx, datasource.MetadataRequest{}, &metadata)
		t.Run(metadata.TypeName, func(t *testing.T) {
			for _, value := range []interface{}{nil, "invalid client", client.New(client.Config{})} {
				var resp datasource.ConfigureResponse
				d.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{ProviderData: value}, &resp)
				if resp.Diagnostics.HasError() != (value == "invalid client") {
					t.Fatalf("Configure(%T): %v", value, resp.Diagnostics)
				}
			}
		})
	}
}

func TestDataSourcesRequireConfiguredClient(t *testing.T) {
	ctx := context.Background()
	for _, factory := range configuredDataSources() {
		d := factory()
		var metadata datasource.MetadataResponse
		d.Metadata(ctx, datasource.MetadataRequest{}, &metadata)
		t.Run(metadata.TypeName, func(t *testing.T) {
			defer func() {
				if value := recover(); value != nil {
					t.Errorf("unconfigured Read panicked: %v", value)
				}
			}()
			var resp datasource.ReadResponse
			d.Read(ctx, datasource.ReadRequest{}, &resp)
			if !resp.Diagnostics.HasError() {
				t.Fatal("unconfigured Read succeeded")
			}
		})
	}
}
