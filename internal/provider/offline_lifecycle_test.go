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

// This exercises Terraform's provider protocol without an account or a remote API.
func TestOfflineSSHKeyLifecycleAndPagination(t *testing.T) {
	const publicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl test@example.com"
	type sshKey struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		PublicKey   string `json:"public_key"`
		Fingerprint string `json:"fingerprint"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
		IsDefault   bool   `json:"is_default"`
		TeamID      any    `json:"team_id"`
		UserID      int    `json:"user_id"`
	}
	key := sshKey{41, "terraform-test", publicKey, "SHA256:test", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z", false, nil, 1}
	other := sshKey{42, "another-key", publicKey, "SHA256:other", "2026-01-01T00:00:00Z", "2026-01-01T00:00:00Z", false, nil, 1}

	var mu sync.Mutex
	created, reads, deletes, secondPages := false, 0, 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer offline-token" || r.Header.Get("X-Team-Id") != "17" {
			t.Errorf("unexpected API headers: authorization=%q team=%q", r.Header.Get("Authorization"), r.Header.Get("X-Team-Id"))
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/ssh-keys":
			var body struct {
				Name      string `json:"name"`
				PublicKey string `json:"public_key"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name != key.Name || body.PublicKey != key.PublicKey {
				t.Errorf("unexpected SSH key create body: %+v, %v", body, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"message": "SSH key created successfully", "key": key})
		case r.Method == http.MethodGet && r.URL.Path == "/ssh-keys/41":
			reads++
			if !created {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"message": "Not found"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"key": key})
		case r.Method == http.MethodGet && r.URL.Path == "/ssh-keys":
			page := r.URL.Query().Get("page")
			var current sshKey
			switch page {
			case "1":
				current = key
			case "2":
				current = other
				secondPages++
			default:
				t.Errorf("unexpected SSH key list page %q", page)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":       []sshKey{current},
				"pagination": map[string]int{"current_page": int(page[0] - '0'), "last_page": 2, "per_page": 1, "total": 2},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/ssh-keys/41":
			deletes++
			created = false
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected API request %s %s", r.Method, r.URL.String())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := fmt.Sprintf(`
provider "danubedata" {
  base_url  = %q
  api_token = "offline-token"
  team_id   = 17
}
resource "danubedata_ssh_key" "test" {
  name       = "terraform-test"
  public_key = %q
}
`, server.URL, publicKey)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"danubedata": providerserver.NewProtocol6WithError(provider.New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_ssh_key.test", "id", "41"),
					resource.TestCheckResourceAttr("danubedata_ssh_key.test", "fingerprint", "SHA256:test"),
				),
			},
			{
				Config: config + `data "danubedata_ssh_keys" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.danubedata_ssh_keys.all", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.danubedata_ssh_keys.all", "keys.1.id", "42"),
				),
			},
			{ResourceName: "danubedata_ssh_key.test", ImportState: true, ImportStateVerify: true},
			{RefreshState: true},
		},
	})

	mu.Lock()
	defer mu.Unlock()
	if created || reads == 0 || deletes != 1 || secondPages == 0 {
		t.Errorf("incomplete offline lifecycle: created=%v reads=%d deletes=%d page2=%d", created, reads, deletes, secondPages)
	}
}
