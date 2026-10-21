package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestOfflineParameterGroupLifecycle(t *testing.T) {
	group := map[string]any{
		"id": 73, "name": "offline-redis", "type": "cache", "provider_type": "redis",
		"family": "redis7.x", "description": "offline group", "parameters": map[string]any{"maxmemory": 1024},
		"locked_parameters": []string{"maxmemory"}, "team_id": 17, "is_default": false, "is_active": true,
		"is_system": false, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
	}
	var mu sync.Mutex
	created, reads, updates, deletes := false, 0, 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer offline-token" || r.Header.Get("X-Team-Id") != "17" {
			t.Errorf("unexpected parameter-group API headers: authorization=%q team=%q", r.Header.Get("Authorization"), r.Header.Get("X-Team-Id"))
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/parameter-groups":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["name"] != "offline-redis" || body["type"] != "cache" || body["provider_type"] != "redis" {
				t.Errorf("unexpected parameter-group create request: %+v, %v", body, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			parameters, ok := body["parameters"].(map[string]any)
			if !ok || parameters["maxmemory"] != "1024" {
				t.Errorf("expected string-valued parameters object, got %#v", body["parameters"])
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "Parameter group created", "parameter_group": group})
		case r.Method == http.MethodGet && r.URL.Path == "/parameter-groups/73":
			reads++
			if !created {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"parameter_group": group})
		case r.Method == http.MethodPut && r.URL.Path == "/parameter-groups/73":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("invalid parameter-group update request: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			parameters, parametersOK := body["parameters"].(map[string]any)
			locks, locksOK := body["locked_parameters"].([]any)
			if !parametersOK || len(parameters) != 0 || !locksOK || len(locks) != 0 {
				t.Errorf("expected explicit empty parameters object and locks array, got %#v", body)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			updates++
			group["parameters"] = []any{} // Empty parameters are serialized as [] by the API.
			group["locked_parameters"] = []string{}
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "Parameter group updated", "parameter_group": group})
		case r.Method == http.MethodDelete && r.URL.Path == "/parameter-groups/73":
			deletes++
			created = false
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "Parameter group deleted"})
		default:
			t.Errorf("unexpected parameter-group API request %s %s", r.Method, r.URL.String())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := func(parameters, locks string) string {
		return fmt.Sprintf(`
provider "danubedata" {
  base_url  = %q
  api_token = "offline-token"
  team_id   = 17
}
resource "danubedata_parameter_group" "test" {
  name              = "offline-redis"
  type              = "cache"
  provider_type     = "redis"
  family            = "redis7.x"
  description       = "offline group"
  parameters        = %s
  locked_parameters = %s
}
`, server.URL, parameters, locks)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"danubedata": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{Config: config(`{ maxmemory = "1024" }`, `["maxmemory"]`), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("danubedata_parameter_group.test", "id", "73"),
				resource.TestCheckResourceAttr("danubedata_parameter_group.test", "parameters.maxmemory", "1024"),
			)},
			{Config: config(`{}`, `[]`), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("danubedata_parameter_group.test", "parameters.%", "0"),
				resource.TestCheckResourceAttr("danubedata_parameter_group.test", "locked_parameters.#", "0"),
			)},
			{ResourceName: "danubedata_parameter_group.test", ImportState: true, ImportStateVerify: true},
			{RefreshState: true},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	if created || reads == 0 || updates != 1 || deletes != 1 {
		t.Errorf("incomplete parameter-group lifecycle: created=%v reads=%d updates=%d deletes=%d", created, reads, updates, deletes)
	}
}
