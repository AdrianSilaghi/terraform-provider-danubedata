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

func TestUnitVpsResourceLifecycle(t *testing.T) {
	profile, exists := "nano_shared", false
	vpsStatus := "running"
	resizeAwaitingRefresh := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /vps":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["cpu_platform"] != "amd" || body["resource_profile"] != profile {
				t.Errorf("VPS create body: %#v", body)
			}
			exists = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"instance":{"id":"vps-1"}}`))
		case "GET /vps/vps-1/status":
			if resizeAwaitingRefresh {
				t.Error("status is still stopped before the queued resize job starts; it cannot prove resize completion")
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": vpsStatus})
		case "GET /vps/vps-1":
			resizeAwaitingRefresh = false
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"instance": map[string]any{
				"id": "vps-1", "name": "unit-vps", "status": "running", "resource_profile": profile,
				"cpu_allocation_type": "shared", "cpu_platform": "amd", "network_tier": "standard",
				"image": "registry.example/kubevirt-ubuntu:ubuntu-24.04-2026.10.01", "datacenter": "fsn1",
				"ssh_key_id": 1, "cpu_cores": 2, "memory_size_gb": 2, "storage_size_gb": 40,
				"created_at": "2026-01-01", "updated_at": "2026-01-01",
			}})
		case "POST /vps/vps-1/stop":
			vpsStatus = "stopped"
			_, _ = w.Write([]byte(`{"status":"stopping"}`))
		case "GET /vps/vps-1/password":
			w.WriteHeader(http.StatusNotFound)
		case "PUT /vps/vps-1":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["resource_profile"] != "micro_shared" {
				t.Errorf("VPS resize: %#v", body)
			}
			profile = "micro_shared"
			resizeAwaitingRefresh = true
			_, _ = w.Write([]byte(`{"instance":{"id":"vps-1"}}`))
		case "DELETE /vps/vps-1":
			exists = false
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := func(p string) string {
		return fmt.Sprintf(`provider "danubedata" {
 base_url = %q
 api_token = "test-token"
}
resource "danubedata_vps" "test" {
 name = "unit-vps"
 image = "ubuntu-24.04"
 datacenter = "fsn1"
 cpu_platform = "amd"
 resource_profile = %q
 auth_method = "ssh_key"
 ssh_key_id = "1"
}`, server.URL, p)
	}
	resource.Test(t, resource.TestCase{IsUnitTest: true, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config("nano_shared"), Check: resource.TestCheckResourceAttr("danubedata_vps.test", "cpu_platform", "amd")},
		{Config: config("micro_shared"), Check: resource.TestCheckResourceAttr("danubedata_vps.test", "resource_profile", "micro_shared")},
		{ResourceName: "danubedata_vps.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"auth_method", "password", "network_stack"}},
	}})
}

func TestUnitFirewallResourceLifecycle(t *testing.T) {
	name, exists, rules := "unit-fw", false, []map[string]any{}
	deploys := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		respond := func() {
			allRules := append(append([]map[string]any{}, rules...), map[string]any{"id": "locked", "name": "SMTP policy", "action": "deny", "direction": "outbound", "protocol": "tcp", "port": 25, "order": 100, "is_locked": true})
			_ = json.NewEncoder(w).Encode(map[string]any{"firewall": map[string]any{
				"id": "fw-1", "name": name, "description": "", "status": "active", "rules": allRules,
				"created_at": "2026-01-01", "updated_at": "2026-01-01",
			}})
		}
		switch r.Method + " " + r.URL.Path {
		case "POST /firewalls":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["name"] != name {
				t.Errorf("firewall create: %#v", body)
			}
			exists = true
			rules = []map[string]any{{"id": "r-1", "name": "SSH", "action": "allow", "direction": "inbound", "protocol": "tcp", "port": 22, "source_ips": []string{}, "source_instance_ids": []map[string]string{{"id": "source-1", "type": "vps"}}, "order": 0, "is_locked": false}}
			w.WriteHeader(http.StatusCreated)
			respond()
		case "GET /firewalls/fw-1":
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			respond()
		case "PUT /firewalls/fw-1":
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if len(body["rules"]) != 0 {
				var next []map[string]any
				if err := json.Unmarshal(body["rules"], &next); err != nil {
					t.Error(err)
				}
				if len(next) == 0 {
					rules = []map[string]any{}
				} else if len(next) == 1 && next[0]["port"] == float64(23) {
					sources, ok := next[0]["source_instance_ids"].([]any)
					if !ok || len(sources) != 1 {
						t.Errorf("typed source lost on port change: %s", body["rules"])
					}
					rules[0]["port"] = 23
				} else {
					t.Errorf("unexpected rules update: %s", body["rules"])
				}
			} else if len(rules) == 0 {
				t.Error("expected rules=[] to clear")
			}
			if len(body["name"]) != 0 {
				name = "renamed-fw"
			}
			respond()
		case "DELETE /firewalls/fw-1":
			exists = false
			w.WriteHeader(http.StatusNoContent)
		case "POST /firewalls/fw-1/deploy":
			deploys++
			w.WriteHeader(http.StatusAccepted)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := func(name string, withRule bool, port ...int) string {
		rule := "rules = []"
		if withRule {
			rule = `rules = [{ name = "SSH", action = "allow", direction = "inbound", protocol = "tcp", port = 22, source_instance_ids = [{ id = "source-1", type = "vps" }] }]`
			if len(port) > 0 {
				rule = `rules = [{ name = "SSH", action = "allow", direction = "inbound", protocol = "tcp", port = 23 }]`
			}
		}
		return fmt.Sprintf(`provider "danubedata" {
 base_url = %q
 api_token = "test-token"
}
resource "danubedata_firewall" "test" {
 name = %q
 %s
}`, server.URL, name, rule)
	}
	resource.Test(t, resource.TestCase{IsUnitTest: true, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config("unit-fw", true), Check: resource.TestCheckResourceAttr("danubedata_firewall.test", "rules.0.port", "22")},
		{ResourceName: "danubedata_firewall.test", ImportState: true, ImportStateVerify: true},
		{Config: config("renamed-fw", true), Check: resource.TestCheckResourceAttr("danubedata_firewall.test", "name", "renamed-fw")},
		{Config: config("renamed-fw", true, 23), Check: resource.TestCheckResourceAttr("danubedata_firewall.test", "rules.0.port", "23")},
		{Config: config("renamed-fw", false), Check: resource.TestCheckResourceAttr("danubedata_firewall.test", "rules.#", "0")},
		{ResourceName: "danubedata_firewall.test", ImportState: true, ImportStateVerify: true},
	}})
	if deploys != 2 {
		t.Errorf("rule changes queued %d deployments, want 2", deploys)
	}
}
