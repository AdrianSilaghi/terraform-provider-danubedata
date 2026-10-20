package resources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestStorageReviewClearsQuotaAndLifecycle(t *testing.T) {
	created, cleared := false, false
	quota := any(nil)
	lifecycle := any([]any{map[string]any{
		"ID": "expire", "Status": "Enabled", "Expiration": map[string]any{"Days": 30},
	}})
	bucket := func() map[string]any {
		return map[string]any{
			"id": "bucket-clear", "name": "clear-review", "status": "active", "region": "fsn1",
			"endpoint_url": "https://s3.test", "minio_bucket_name": "dd-17-clear-review",
			"encryption_enabled": true, "encryption_type": "sse-s3",
			"size_limit_bytes": quota, "lifecycle_rules": lifecycle, "tags": []any{},
			"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/storage/buckets":
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodGet && r.URL.Path == "/storage/buckets/bucket-clear":
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodPut && r.URL.Path == "/storage/buckets/bucket-clear":
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if raw, ok := body["size_limit_bytes"]; ok {
				if string(raw) == "null" {
					quota = nil
				} else {
					quota = int64(1073741824)
				}
			}
			if raw, ok := body["lifecycle_rules"]; ok && string(raw) == `[]` {
				lifecycle = []any{}
			}
			if string(body["size_limit_bytes"]) == "null" && string(body["lifecycle_rules"]) == `[]` {
				cleared = true
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/buckets/bucket-clear":
			created = false
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	base := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}
resource "danubedata_storage_bucket" "test" {
  name = "clear-review"
  region = "fsn1"
`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: storageReviewFactory(),
		Steps: []resource.TestStep{
			{Config: base + `size_limit_bytes = 1073741824
lifecycle_rules = jsonencode([{ ID = "expire", Status = "Enabled", Expiration = { Days = 30 } }])
}`},
			{Config: base + "}", Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckNoResourceAttr("danubedata_storage_bucket.test", "size_limit_bytes"),
				resource.TestCheckNoResourceAttr("danubedata_storage_bucket.test", "lifecycle_rules"),
			)},
		},
	})
	if !cleared {
		t.Error("quota and lifecycle were not cleared together")
	}
}
