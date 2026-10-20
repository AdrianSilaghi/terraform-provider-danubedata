package resources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitSshKeyResourceLifecycle(t *testing.T) {
	exists := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		key := map[string]any{"id": 17, "name": "unit-key", "public_key": "ssh-ed25519 AAAA unit", "fingerprint": "SHA256:unit", "is_default": true, "created_at": "2026-01-01", "updated_at": "2026-01-01"}
		switch r.Method + " " + r.URL.Path {
		case "POST /ssh-keys":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["is_default"] != true {
				t.Errorf("default key flag not sent: %#v", body)
			}
			exists = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"key": key})
		case "GET /ssh-keys/17":
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"key": key})
		case "DELETE /ssh-keys/17":
			exists = false
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := fmt.Sprintf(`provider "danubedata" {
 base_url = %q
 api_token = "test-token"
}
resource "danubedata_ssh_key" "test" {
 name = "unit-key"
 public_key = "ssh-ed25519 AAAA unit"
 is_default = true
}`, server.URL)
	resource.Test(t, resource.TestCase{IsUnitTest: true, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config, Check: resource.TestCheckResourceAttr("danubedata_ssh_key.test", "is_default", "true")},
		{ResourceName: "danubedata_ssh_key.test", ImportState: true, ImportStateVerify: true},
	}})
}
