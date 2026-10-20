package resources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitParameterGroupResource_FamilyChangeRecreatesGroup(t *testing.T) {
	groups := map[string]map[string]interface{}{}
	creates := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		path := strings.TrimPrefix(r.URL.Path, "/parameter-groups")
		switch {
		case r.Method == http.MethodPost && path == "":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			creates++
			id := strconv.Itoa(creates)
			body["id"] = creates
			body["is_default"] = false
			body["is_active"] = true
			body["is_system"] = false
			groups[id] = body
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"parameter_group": body})
		case r.Method == http.MethodGet:
			id := strings.TrimPrefix(path, "/")
			group, ok := groups[id]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"parameter_group": group})
		case r.Method == http.MethodDelete:
			delete(groups, strings.TrimPrefix(path, "/"))
		case r.Method == http.MethodPut:
			t.Error("family is immutable in the API; update must replace the group")
			w.WriteHeader(http.StatusBadRequest)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := func(family string) string {
		return fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}
resource "danubedata_parameter_group" "test" {
  name = "unit-group"
  type = "cache"
  provider_type = "redis"
  family = %q
  parameters = { "maxmemory-policy" = "allkeys-lru" }
}`, server.URL, family)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("redis7.x"), Check: resource.TestCheckResourceAttr("danubedata_parameter_group.test", "id", "1")},
			{Config: config("redis8.x"), Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("danubedata_parameter_group.test", "id", "2"),
				resource.TestCheckResourceAttr("danubedata_parameter_group.test", "family", "redis8.x"),
			)},
		},
	})
	if creates != 2 {
		t.Fatalf("created %d parameter groups, want replacement", creates)
	}
}
