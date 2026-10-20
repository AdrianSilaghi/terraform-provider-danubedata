package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDataSourceEmptyAndNullableListings(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		factory   func() datasource.DataSource
		list      string
		item      string
		nullAttrs []string
	}{
		{NewVpsImagesDataSource, "images", `{"id":"ubuntu-24.04","version":24.04}`, []string{"family"}},
		{NewVpssDataSource, "instances", `{"id":"vps-1"}`, []string{"vnc_access_url", "public_ip", "private_ip", "ipv6_address", "cpu_platform", "primary_ipv4_floating_ip_id", "primary_ipv6_floating_ip_id"}},
		{NewCachesDataSource, "instances", `{"id":"cache-1","provider":{"name":"Redis"}}`, []string{"endpoint", "port"}},
		{NewDatabasesDataSource, "instances", `{"id":"database-1"}`, []string{"database_name", "endpoint", "port", "username"}},
		{NewFirewallsDataSource, "firewalls", `{"id":"firewall-1","rules":[]}`, nil},
		{NewSshKeysDataSource, "keys", `{"id":42}`, nil},
		{NewServerlessContainersDataSource, "containers", `{"id":"container-1"}`, []string{"image", "repository_url"}},
		{NewParameterGroupsDataSource, "groups", `{"id":42}`, []string{"family", "description", "created_at"}},
		{NewStorageBucketsDataSource, "buckets", `{"id":"bucket-1","tags":[]}`, []string{"encryption_type", "object_lock_default_retention_mode", "object_lock_default_retention_days", "size_limit_bytes", "tags", "display_name", "public_url"}},
		{NewStorageAccessKeysDataSource, "keys", `{"id":"key-1","bucket_permissions":[]}`, []string{"arn", "expires_at", "last_used_at"}},
		{NewVpsSnapshotsDataSource, "snapshots", `{"id":42}`, []string{"size_gb", "size_bytes", "size_basis", "expires_at", "description"}},
		{NewCacheSnapshotsDataSource, "snapshots", `{"id":42}`, []string{"size_mb", "size_bytes", "size_basis"}},
		{NewDatabaseSnapshotsDataSource, "snapshots", `{"id":42}`, []string{"size_gb", "size_bytes", "size_basis"}},
		{NewStaticSitesDataSource, "sites", `{"id":"site-1"}`, nil},
	} {
		d := tc.factory()
		var metadata datasource.MetadataResponse
		d.Metadata(ctx, datasource.MetadataRequest{}, &metadata)
		t.Run(metadata.TypeName, func(t *testing.T) {
			for _, populated := range []bool{false, true} {
				t.Run(fmt.Sprint(populated), func(t *testing.T) {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						items := []json.RawMessage{}
						if populated {
							items = append(items, json.RawMessage(tc.item))
						}
						if r.URL.Path == "/parameter-groups" && (r.URL.Query().Get("type") != "cache" || r.URL.Query().Get("provider_type") != "redis") {
							t.Errorf("parameter group filters not sent: %s", r.URL)
						}
						if r.URL.Path == "/vps/images" {
							_ = json.NewEncoder(w).Encode(map[string]interface{}{"images": items})
							return
						}
						_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": items, "pagination": map[string]int{"last_page": 1}})
					}))
					defer server.Close()
					var configured datasource.ConfigureResponse
					d.(datasource.DataSourceWithConfigure).Configure(ctx, datasource.ConfigureRequest{ProviderData: client.New(client.Config{BaseURL: server.URL, APIToken: "offline-token"})}, &configured)
					var schema datasource.SchemaResponse
					d.Schema(ctx, datasource.SchemaRequest{}, &schema)
					typ := schema.Schema.Type().TerraformType(ctx).(tftypes.Object)
					values := make(map[string]tftypes.Value, len(typ.AttributeTypes))
					for name, attrType := range typ.AttributeTypes {
						values[name] = tftypes.NewValue(attrType, nil)
					}
					if _, ok := values["team_id"]; ok {
						values["team_id"] = tftypes.NewValue(tftypes.Number, int64(17))
					}
					if _, ok := values["provider_type"]; ok {
						values["type"] = tftypes.NewValue(tftypes.String, "cache")
						values["provider_type"] = tftypes.NewValue(tftypes.String, "redis")
					}
					resp := datasource.ReadResponse{State: tfsdk.State{Schema: schema.Schema}}
					d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Raw: tftypes.NewValue(typ, values), Schema: schema.Schema}}, &resp)
					if resp.Diagnostics.HasError() {
						t.Fatal(resp.Diagnostics)
					}
					var list types.List
					if diags := resp.State.GetAttribute(ctx, path.Root(tc.list), &list); diags.HasError() {
						t.Fatal(diags)
					}
					want := 0
					if populated {
						want = 1
					}
					if list.IsNull() || len(list.Elements()) != want {
						t.Fatalf("listing = %v, want %d elements", list, want)
					}
					if populated {
						item := list.Elements()[0].(types.Object).Attributes()
						for _, name := range tc.nullAttrs {
							if value, ok := item[name]; !ok || !value.IsNull() {
								t.Errorf("%s = %v, want null", name, value)
							}
						}
						if tc.list == "instances" && metadata.TypeName == "_caches" && item["cache_provider"] != types.StringValue("redis") {
							t.Errorf("cache provider fallback = %v", item["cache_provider"])
						}
					}
				})
			}
		})
	}
}
