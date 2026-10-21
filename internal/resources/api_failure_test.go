package resources_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/client"
	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/resources"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceSchema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type apiFailureCase struct {
	name        string
	newResource func() resource.Resource
	id          string
	extra       map[string]any
	listRead    bool
	create      map[string]any
	updateField string
	updateValue any
}

func TestResourceCreateAndUpdateRefusals(t *testing.T) {
	for _, tc := range apiFailureCases {
		for _, status := range []int{http.StatusForbidden, http.StatusUnprocessableEntity} {
			t.Run(tc.name+"/create/"+strconv.Itoa(status), func(t *testing.T) {
				var requests atomic.Int64
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					requests.Add(1)
					if r.Method != http.MethodPost {
						t.Errorf("create used %s, want POST", r.Method)
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_ = json.NewEncoder(w).Encode(map[string]any{"message": "request refused", "errors": map[string][]string{"name": {"request refused"}}})
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
				target.(resource.ResourceWithConfigure).Configure(ctx, resource.ConfigureRequest{ProviderData: client.New(client.Config{BaseURL: server.URL, APIToken: "test-token"})}, &configured)
				if configured.Diagnostics.HasError() {
					t.Fatal(configured.Diagnostics)
				}

				fixture := newAPIFailureState(t, ctx, schema.Schema, tc.create)
				resp := &resource.CreateResponse{State: fixture}
				target.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: schema.Schema, Raw: fixture.Raw}}, resp)
				if requests.Load() == 0 {
					t.Fatalf("create never reached API; diagnostics: %v", resp.Diagnostics)
				}
				if !resp.Diagnostics.HasError() || !strings.Contains(fmt.Sprint(resp.Diagnostics), strconv.Itoa(status)) {
					t.Fatalf("create %d must report API refusal: %v", status, resp.Diagnostics)
				}
				var id types.String
				if diags := resp.State.GetAttribute(ctx, path.Root("id"), &id); diags.HasError() || !id.IsNull() {
					t.Fatalf("refused create acquired an ID: %q, %v", id.ValueString(), diags)
				}
			})
		}

		if tc.updateField == "" {
			continue
		}
		t.Run(tc.name+"/update/403", func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.Method != http.MethodPut {
					t.Errorf("update used %s, want PUT", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "request refused"})
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
			target.(resource.ResourceWithConfigure).Configure(ctx, resource.ConfigureRequest{ProviderData: client.New(client.Config{BaseURL: server.URL, APIToken: "test-token"})}, &configured)
			if configured.Diagnostics.HasError() {
				t.Fatal(configured.Diagnostics)
			}

			priorFields := make(map[string]any, len(tc.create)+len(tc.extra)+1)
			for name, value := range tc.create {
				priorFields[name] = value
			}
			for name, value := range tc.extra {
				priorFields[name] = value
			}
			priorFields["id"] = tc.id
			prior := newAPIFailureState(t, ctx, schema.Schema, priorFields)
			plan := tfsdk.Plan{Schema: schema.Schema, Raw: prior.Raw}
			if diags := plan.SetAttribute(ctx, path.Root(tc.updateField), tc.updateValue); diags.HasError() {
				t.Fatal(diags)
			}
			resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schema.Schema, Raw: plan.Raw}}
			target.Update(ctx, resource.UpdateRequest{State: prior, Plan: plan, Config: tfsdk.Config{Schema: schema.Schema, Raw: plan.Raw}}, resp)
			if requests.Load() == 0 {
				t.Fatalf("update never reached API; diagnostics: %v", resp.Diagnostics)
			}
			if !resp.Diagnostics.HasError() || !strings.Contains(fmt.Sprint(resp.Diagnostics), "403") {
				t.Fatalf("update must report API refusal: %v", resp.Diagnostics)
			}
			var id types.String
			if diags := resp.State.GetAttribute(ctx, path.Root("id"), &id); diags.HasError() || id.ValueString() != tc.id {
				t.Fatalf("refused update lost tracked ID: %q, %v", id.ValueString(), diags)
			}
		})
	}
}

func newAPIFailureState(t *testing.T, ctx context.Context, schema resourceSchema.Schema, fields map[string]any) tfsdk.State {
	t.Helper()
	objectType := schema.Type().TerraformType(ctx).(tftypes.Object)
	attributes := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
	for name, attributeType := range objectType.AttributeTypes {
		attributes[name] = tftypes.NewValue(attributeType, nil)
	}
	state := tfsdk.State{Schema: schema, Raw: tftypes.NewValue(objectType, attributes)}
	for name, value := range fields {
		if diags := state.SetAttribute(ctx, path.Root(name), value); diags.HasError() {
			t.Fatalf("set %s: %v", name, diags)
		}
	}
	return state
}

var apiFailureCases = []apiFailureCase{
	{"vps", resources.NewVpsResource, "vps-1", nil, false, map[string]any{"name": "test-vps", "resource_profile": "nano_shared", "cpu_platform": "amd", "image": "ubuntu-24.04", "datacenter": "fsn1", "auth_method": "password", "password": "StrongPassword123!"}, "resource_profile", "micro_shared"},
	{"serverless", resources.NewServerlessResource, "srv-1", nil, false, map[string]any{"name": "test-serverless", "slug": "test-serverless", "deployment_type": "docker_image", "resource_profile": "small", "image": "nginx", "image_tag": "latest", "port": int64(80), "min_scale": int64(0), "max_scale": int64(1)}, "name", "renamed-serverless"},
	{"cache", resources.NewCacheResource, "cache-1", nil, false, map[string]any{"name": "test-cache", "cache_provider": "redis", "resource_profile": "micro", "datacenter": "fsn1"}, "name", "renamed-cache"},
	{"database", resources.NewDatabaseResource, "db-1", nil, false, map[string]any{"name": "test-db", "engine": "postgresql", "resource_profile": "micro", "datacenter": "fsn1"}, "resource_profile", "small"},
	{"database_replica", resources.NewDatabaseReplicaResource, "db-1:1", map[string]any{"database_instance_id": "db-1", "replica_index": int64(1)}, true, map[string]any{"database_instance_id": "db-1"}, "", nil},
	{"parameter_group", resources.NewParameterGroupResource, "17", nil, false, map[string]any{"name": "test-group", "type": "cache", "provider_type": "redis", "parameters": map[string]string{}}, "name", "renamed-group"},
	{"storage_bucket", resources.NewStorageBucketResource, "bucket-1", nil, false, map[string]any{"name": "test-bucket", "region": "fsn1", "versioning_enabled": false, "public_access": false, "encryption_enabled": false, "encryption_type": "none", "object_lock_enabled": false}, "public_access", true},
	{"storage_access_key", resources.NewStorageAccessKeyResource, "key-1", nil, false, map[string]any{"name": "test-key", "scope": "team"}, "", nil},
	{"ssh_key", resources.NewSshKeyResource, "42", nil, false, map[string]any{"name": "test-key", "public_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl test@example.com"}, "", nil},
	{"firewall", resources.NewFirewallResource, "fw-1", nil, false, map[string]any{"name": "test-firewall"}, "name", "renamed-firewall"},
	{"vps_snapshot", resources.NewVpsSnapshotResource, "42", nil, true, map[string]any{"name": "test-snapshot", "vps_instance_id": "vps-1"}, "", nil},
	{"cache_snapshot", resources.NewCacheSnapshotResource, "42", nil, true, map[string]any{"name": "test-snapshot", "cache_instance_id": "cache-1"}, "", nil},
	{"database_snapshot", resources.NewDatabaseSnapshotResource, "42", nil, true, map[string]any{"name": "test-snapshot", "database_instance_id": "db-1"}, "", nil},
	{"static_site", resources.NewStaticSiteResource, "site-1", nil, false, map[string]any{"name": "test-site"}, "", nil},
	{"static_site_domain", resources.NewStaticSiteDomainResource, "site-1:domain-1", map[string]any{"static_site_id": "site-1", "domain_id": "domain-1", "domain": "www.example.com"}, true, map[string]any{"static_site_id": "site-1", "domain": "www.example.com"}, "", nil},
}

// These checks protect existing state when an API token loses access to a resource.
func TestResourceAPIFailures(t *testing.T) {
	for _, tc := range apiFailureCases {
		for _, operation := range []string{"read", "delete"} {
			for _, status := range []int{http.StatusForbidden, http.StatusNotFound} {
				t.Run(tc.name+"/"+operation+"/"+http.StatusText(status), func(t *testing.T) {
					var requests atomic.Int64
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						requests.Add(1)
						w.Header().Set("Content-Type", "application/json")
						if r.Method == http.MethodGet && status == http.StatusNotFound && tc.listRead {
							if tc.name == "database_replica" {
								_ = json.NewEncoder(w).Encode(map[string]any{"replicas": []any{}})
							} else {
								_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "pagination": map[string]int{"current_page": 1, "last_page": 1}})
							}
							return
						}
						w.WriteHeader(status)
						_ = json.NewEncoder(w).Encode(map[string]string{"message": http.StatusText(status)})
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
					target.(resource.ResourceWithConfigure).Configure(ctx, resource.ConfigureRequest{
						ProviderData: client.New(client.Config{BaseURL: server.URL, APIToken: "test-token"}),
					}, &configured)
					if configured.Diagnostics.HasError() {
						t.Fatal(configured.Diagnostics)
					}
					fields := make(map[string]any, len(tc.extra)+1)
					fields["id"] = tc.id
					for name, value := range tc.extra {
						fields[name] = value
					}
					state := newAPIFailureState(t, ctx, schema.Schema, fields)

					if operation == "read" {
						resp := &resource.ReadResponse{State: state}
						target.Read(ctx, resource.ReadRequest{State: state}, resp)
						if status == http.StatusForbidden {
							if !resp.Diagnostics.HasError() || resp.State.Raw.IsNull() {
								t.Fatalf("403 must diagnose and retain state: %v", resp.Diagnostics)
							}
							var id string
							if diags := resp.State.GetAttribute(ctx, path.Root("id"), &id); diags.HasError() || id != tc.id {
								t.Fatalf("403 lost tracked ID: %q, %v", id, diags)
							}
						} else if resp.Diagnostics.HasError() || !resp.State.Raw.IsNull() {
							t.Fatalf("missing resource should leave state: %v", resp.Diagnostics)
						}
					} else {
						resp := &resource.DeleteResponse{State: state}
						target.Delete(ctx, resource.DeleteRequest{State: state}, resp)
						if (status == http.StatusForbidden) != resp.Diagnostics.HasError() {
							t.Fatalf("delete status %d diagnostics: %v", status, resp.Diagnostics)
						}
					}
					if requests.Load() == 0 {
						t.Fatal("resource did not call the API")
					}
				})
			}
		}
	}
}
