package resources_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Read uses imported IDs: these are independently captured API response shapes,
// not response bodies manufactured from the resource's own state mapping.
func TestPublishedResourceReadback(t *testing.T) {
	var published struct {
		Responses map[string]json.RawMessage `json:"responses"`
	}
	raw, err := os.ReadFile("../client/testdata/api-responses.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &published); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		fixture, route, envelope string
		newResource              func() resource.Resource
		initial                  map[string]any
		want                     map[string]any
	}{
		{"VpsInstanceResource", "/vps/sample", "instance", resources.NewVpsResource, nil, map[string]any{"cpu_allocation_type": "sample", "cpu_platform": "sample", "network_tier": "standard", "primary_ipv4_floating_ip_id": "sample", "primary_ipv6_floating_ip_id": "sample"}},
		{"CacheInstanceResource", "/cache/sample", "instance", resources.NewCacheResource, nil, map[string]any{"network_tier": "standard", "memory_size_mb": int64(1)}},
		{"DatabaseInstanceResource", "/database/sample", "instance", resources.NewDatabaseResource, nil, map[string]any{"network_tier": "standard", "storage_autoscaling_enabled": true, "storage_autoscale_threshold_percent": int64(1), "storage_autoscale_increment_gib": int64(1)}},
		{"FirewallResource", "/firewalls/sample", "firewall", resources.NewFirewallResource, nil, map[string]any{"description": "sample", "default_action": "sample"}},
		{"ParameterGroupResource", "/parameter-groups/1", "parameter_group", resources.NewParameterGroupResource, nil, map[string]any{"provider_type": "sample", "is_default": true, "parameters.max_connections": "100"}},
		{"ServerlessContainerResource", "/serverless/sample", "container", resources.NewServerlessResource, nil, map[string]any{"slug": "sample", "source_type": "sample", "repository_branch": "sample", "environment_variables.APP_ENV": "production"}},
		{"SshKeyResource", "/ssh-keys/1", "key", resources.NewSshKeyResource, nil, map[string]any{"is_default": true, "fingerprint": "sample"}},
		{"StaticSiteResource", "/static-sites/sample", "data", resources.NewStaticSiteResource, nil, map[string]any{"slug": "sample", "plan": "sample"}},
		{"StaticSiteDomainResource", "/static-sites/sample/domains", "data", resources.NewStaticSiteDomainResource, map[string]any{"static_site_id": "sample", "domain": "sample"}, map[string]any{"domain_id": "sample", "is_primary": true, "dns_instructions": "sample"}},
		{"StorageAccessKeyResource", "/storage/access-keys/sample", "access_key", resources.NewStorageAccessKeyResource, nil, map[string]any{"scope": "buckets", "bucket_permissions.sample": "sample", "access_key_id": "sample"}},
		{"StorageBucketResource", "/storage/buckets/sample", "bucket", resources.NewStorageBucketResource, nil, map[string]any{"object_lock_enabled": true, "object_lock_default_retention_mode": "sample", "object_lock_default_retention_days": int64(1), "lifecycle_rules": `[{"ID":"expire-temp","Status":"Enabled","Filter":{"Prefix":"tmp/"},"Expiration":{"Days":30}}]`, "tags.environment": "test"}},
		{"VpsSnapshotResource", "/snapshots/vps", "data", resources.NewVpsSnapshotResource, nil, map[string]any{"size_gb": 1.25, "size_bytes": int64(1), "size_basis": "sample"}},
		{"CacheSnapshotResource", "/snapshots/cache", "data", resources.NewCacheSnapshotResource, nil, map[string]any{"size_mb": 1.25, "size_bytes": int64(1), "size_basis": "sample"}},
		{"DatabaseSnapshotResource", "/snapshots/database", "data", resources.NewDatabaseSnapshotResource, nil, map[string]any{"size_gb": 1.25, "size_bytes": int64(1), "size_basis": "sample"}},
	}
	for _, tc := range tests {
		t.Run(tc.fixture, func(t *testing.T) {
			body, ok := published.Responses[tc.fixture]
			if !ok {
				t.Fatalf("published fixture %s missing", tc.fixture)
			}
			var fixture struct {
				ID any `json:"id"`
			}
			if err := json.Unmarshal(body, &fixture); err != nil {
				t.Fatal(err)
			}
			id := "sample"
			if _, numeric := fixture.ID.(float64); numeric {
				id = "1"
			}
			if tc.fixture == "StaticSiteDomainResource" {
				id = "sample:sample"
			}
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				w.Header().Set("Content-Type", "application/json")
				if r.Method == http.MethodGet && r.Header.Get("Authorization") == "Bearer offline-token" &&
					(r.URL.Path == "/vps/sample/password" || r.URL.Path == "/cache/sample/connection-info" || r.URL.Path == "/database/sample/credentials") {
					_, _ = w.Write([]byte(`{"password":"offline-only","username":"fixture","connection_info":"local"}`))
					return
				}
				if r.Method != http.MethodGet || r.URL.Path != tc.route || r.Header.Get("Authorization") != "Bearer offline-token" {
					t.Errorf("unexpected request: %s %s, authorization %q", r.Method, r.URL.String(), r.Header.Get("Authorization"))
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if strings.HasPrefix(tc.route, "/snapshots/") && r.URL.Query().Get("page") != "1" {
					t.Errorf("snapshot list omitted page=1: %s", r.URL.String())
				}
				if strings.HasSuffix(tc.route, "/domains") || strings.HasPrefix(tc.route, "/snapshots/") {
					_, _ = w.Write([]byte(`{"data":[` + string(body) + `],"pagination":{"current_page":1,"last_page":1}}`))
					return
				}
				_, _ = w.Write([]byte(`{"` + tc.envelope + `":` + string(body) + `}`))
			}))
			defer server.Close()

			ctx := context.Background()
			target := tc.newResource()
			var schema resource.SchemaResponse
			target.Schema(ctx, resource.SchemaRequest{}, &schema)
			if schema.Diagnostics.HasError() {
				t.Fatal(schema.Diagnostics)
			}
			var configured resource.ConfigureResponse
			target.(resource.ResourceWithConfigure).Configure(ctx, resource.ConfigureRequest{ProviderData: client.New(client.Config{BaseURL: server.URL, APIToken: "offline-token"})}, &configured)
			if configured.Diagnostics.HasError() {
				t.Fatal(configured.Diagnostics)
			}
			initial := map[string]any{"id": id}
			for key, value := range tc.initial {
				initial[key] = value
			}
			state := newAPIFailureState(t, ctx, schema.Schema, initial)
			resp := &resource.ReadResponse{State: state}
			target.Read(ctx, resource.ReadRequest{State: tfsdk.State{Schema: schema.Schema, Raw: state.Raw}}, resp)
			if requests == 0 || resp.Diagnostics.HasError() {
				t.Fatalf("read reached API %d times; diagnostics: %v", requests, resp.Diagnostics)
			}
			for field, want := range tc.want {
				readPublishedField(t, ctx, resp.State, field, want)
			}
			if tc.fixture == "StorageBucketResource" {
				var cors types.String
				if diags := resp.State.GetAttribute(ctx, path.Root("cors_configuration"), &cors); diags.HasError() {
					t.Fatal(diags)
				}
				var config struct {
					CORSRules []struct {
						AllowedOrigins []string `json:"AllowedOrigins"`
					} `json:"CORSRules"`
				}
				if err := json.Unmarshal([]byte(cors.ValueString()), &config); err != nil {
					t.Fatalf("invalid CORS state %q: %v", cors.ValueString(), err)
				}
				if len(config.CORSRules) != 1 || !reflect.DeepEqual(config.CORSRules[0].AllowedOrigins, []string{"https://example.com"}) {
					t.Fatalf("CORS allowed origins = %#v, want https://example.com", config.CORSRules)
				}
			}
		})
	}
}

func readPublishedField(t *testing.T, ctx context.Context, state tfsdk.State, field string, want any) {
	t.Helper()
	parts := strings.Split(field, ".")
	attribute := path.Root(parts[0])
	if len(parts) == 2 {
		attribute = attribute.AtMapKey(parts[1])
	}
	var got any
	switch want.(type) {
	case string:
		var value types.String
		if diags := state.GetAttribute(ctx, attribute, &value); diags.HasError() {
			t.Fatal(diags)
		}
		got = value.ValueString()
	case bool:
		var value types.Bool
		if diags := state.GetAttribute(ctx, attribute, &value); diags.HasError() {
			t.Fatal(diags)
		}
		got = value.ValueBool()
	case int64:
		var value types.Int64
		if diags := state.GetAttribute(ctx, attribute, &value); diags.HasError() {
			t.Fatal(diags)
		}
		got = value.ValueInt64()
	case float64:
		var value types.Float64
		if diags := state.GetAttribute(ctx, attribute, &value); diags.HasError() {
			t.Fatal(diags)
		}
		got = value.ValueFloat64()
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s = %#v, want %#v", field, got, want)
	}
}
