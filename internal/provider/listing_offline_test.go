package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOfflineAllDataSources(t *testing.T) {
	fixture, err := os.ReadFile("../client/testdata/api-responses.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract struct {
		Responses map[string]json.RawMessage `json:"responses"`
	}
	if err := json.Unmarshal(fixture, &contract); err != nil {
		t.Fatal(err)
	}
	listings := []struct {
		kind, path, schema, attribute string
	}{
		{"vpss", "/vps", "VpsInstanceResource", "instances"},
		{"caches", "/cache", "CacheInstanceResource", "instances"},
		{"databases", "/database", "DatabaseInstanceResource", "instances"},
		{"firewalls", "/firewalls", "FirewallResource", "firewalls"},
		{"ssh_keys", "/ssh-keys", "SshKeyResource", "keys"},
		{"parameter_groups", "/parameter-groups", "ParameterGroupResource", "groups"},
		{"serverless_containers", "/serverless", "ServerlessContainerResource", "containers"},
		{"static_sites", "/teams/17/static-sites", "StaticSiteResource", "sites"},
		{"storage_buckets", "/storage/buckets", "StorageBucketResource", "buckets"},
		{"storage_access_keys", "/storage/access-keys", "StorageAccessKeyResource", "keys"},
		{"vps_snapshots", "/snapshots/vps", "VpsSnapshotResource", "snapshots"},
		{"cache_snapshots", "/snapshots/cache", "CacheSnapshotResource", "snapshots"},
		{"database_snapshots", "/snapshots/database", "DatabaseSnapshotResource", "snapshots"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.Header.Get("X-Team-Id") != "17" {
			t.Errorf("unexpected request: %s %s, project %q", r.Method, r.URL, r.Header.Get("X-Team-Id"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/vps/images" {
			fmt.Fprint(w, `{"images":[{"id":"ubuntu-24.04","image":"registry.example/ubuntu:24.04","label":"Ubuntu 24.04","description":"Ubuntu","distro":"ubuntu","version":"24.04","family":"linux","default_user":"ubuntu"}]}`)
			return
		}
		for _, listing := range listings {
			if listing.path == r.URL.Path {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"data":       []json.RawMessage{contract.Responses[listing.schema]},
					"pagination": map[string]int{"current_page": 1, "last_page": 1, "per_page": 1, "total": 1},
				})
				return
			}
		}
		t.Errorf("unexpected request: %s", r.URL)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	config := fmt.Sprintf(`
provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
  team_id = 17
}
`, server.URL)
	var checks []resource.TestCheckFunc
	for _, listing := range listings {
		body := ""
		if listing.kind == "static_sites" {
			body = "team_id = 17"
		}
		config += fmt.Sprintf("data \"danubedata_%s\" \"test\" { %s }\n", listing.kind, body)
		checks = append(checks, resource.TestCheckResourceAttr("data.danubedata_"+listing.kind+".test", listing.attribute+".#", "1"))
	}
	config += `
data "danubedata_vps_images" "test" {}
data "danubedata_cache_providers" "test" {}
data "danubedata_database_providers" "test" {}
`
	checks = append(checks,
		resource.TestCheckResourceAttr("data.danubedata_vps_images.test", "images.0.id", "ubuntu-24.04"),
		resource.TestCheckResourceAttr("data.danubedata_cache_providers.test", "providers.#", "3"),
		resource.TestCheckResourceAttr("data.danubedata_database_providers.test", "providers.#", "3"),
	)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"danubedata": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{{Config: config, Check: resource.ComposeAggregateTestCheckFunc(checks...)}},
	})
}
