package datasources

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDataSourcesDoNotHideAPIFailures(t *testing.T) {
	factories := []func() datasource.DataSource{
		NewVpsImagesDataSource, NewVpssDataSource, NewSshKeysDataSource,
		NewCachesDataSource, NewDatabasesDataSource, NewFirewallsDataSource,
		NewServerlessContainersDataSource, NewParameterGroupsDataSource,
		NewStorageBucketsDataSource, NewStorageAccessKeysDataSource,
		NewVpsSnapshotsDataSource, NewCacheSnapshotsDataSource,
		NewDatabaseSnapshotsDataSource, NewStaticSitesDataSource,
	}
	for _, status := range []int{http.StatusForbidden, http.StatusInternalServerError} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				w.WriteHeader(status)
				fmt.Fprint(w, `{"message":"Cannot read project resources"}`)
			}))
			defer server.Close()
			c := client.New(client.Config{BaseURL: server.URL, APIToken: "offline-token", TeamID: 17})
			for _, factory := range factories {
				d := factory()
				var metadata datasource.MetadataResponse
				d.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "danubedata"}, &metadata)
				t.Run(metadata.TypeName, func(t *testing.T) {
					ctx := context.Background()
					var schema datasource.SchemaResponse
					d.Schema(ctx, datasource.SchemaRequest{}, &schema)
					var configured datasource.ConfigureResponse
					d.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{ProviderData: c}, &configured)
					if configured.Diagnostics.HasError() {
						t.Fatal(configured.Diagnostics)
					}
					typ := schema.Schema.Type().TerraformType(ctx).(tftypes.Object)
					values := make(map[string]tftypes.Value, len(typ.AttributeTypes))
					for name, attrType := range typ.AttributeTypes {
						values[name] = tftypes.NewValue(attrType, nil)
					}
					if _, ok := values["team_id"]; ok {
						values["team_id"] = tftypes.NewValue(tftypes.Number, int64(17))
					}
					req := datasource.ReadRequest{Config: tfsdk.Config{Raw: tftypes.NewValue(typ, values), Schema: schema.Schema}}
					resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
					before := requests.Load()
					d.Read(ctx, req, &resp)
					if requests.Load() != before+1 {
						t.Fatal("test did not reach the API")
					}
					if !resp.Diagnostics.HasError() {
						t.Fatal("API failure reported as a successful resource listing")
					}
				})
			}
		})
	}
}
