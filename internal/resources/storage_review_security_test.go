package resources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestStorageReviewRejectsInvalidKeyScopesBeforeCreate(t *testing.T) {
	for _, tc := range []struct {
		name, settings, diagnostic string
	}{
		{"buckets without permissions", `scope = "buckets"`, "requires 1 to 50 bucket permissions"},
		{"team with permissions", `scope = "team"
bucket_permissions = { "bucket-1" = "read" }`, "only valid with scope buckets"},
		{"invalid permission", `scope = "buckets"
bucket_permissions = { "bucket-1" = "write" }`, "Each bucket ID must be nonempty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			posts := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					posts++
				}
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()
			config := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}
resource "danubedata_storage_access_key" "test" {
  name = "invalid-scope"
  %s
}`, server.URL, tc.settings)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: storageReviewFactory(),
				Steps:                    []resource.TestStep{{Config: config, ExpectError: regexp.MustCompile(tc.diagnostic)}},
			})
			if posts != 0 {
				t.Errorf("invalid scope caused %d key creates", posts)
			}
		})
	}
}

func TestStorageReviewKMSKeyChangeRoundTrip(t *testing.T) {
	created := false
	encryptionType := "sse-s3"
	kmsUpdates, sseUpdates := 0, 0
	bucket := func() map[string]any {
		return map[string]any{
			"id": "bucket-kms", "name": "kms-review", "status": "active", "region": "fsn1",
			"endpoint_url": "https://s3.test", "minio_bucket_name": "dd-17-kms-review",
			"encryption_enabled": true, "encryption_type": encryptionType,
			"tags": []any{}, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/storage/buckets":
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodGet && r.URL.Path == "/storage/buckets/bucket-kms":
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodPut && r.URL.Path == "/storage/buckets/bucket-kms":
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			switch string(body["encryption_type"]) {
			case `"sse-kms"`:
				if string(body["encryption_kms_key_id"]) != `"test-kms-key"` {
					t.Errorf("KMS request: %v", body)
				}
				kmsUpdates++
				encryptionType = "sse-kms"
			case `"sse-s3"`:
				if _, ok := body["encryption_kms_key_id"]; ok {
					t.Errorf("SSE-S3 request retained KMS key: %v", body)
				}
				sseUpdates++
				encryptionType = "sse-s3"
			default:
				t.Errorf("unexpected encryption update: %v", body)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/buckets/bucket-kms":
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
  name = "kms-review"
  region = "fsn1"
`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: storageReviewFactory(),
		Steps: []resource.TestStep{
			{Config: base + "}"},
			{Config: base + "encryption_type = \"sse-kms\"\nencryption_kms_key_id = \"test-kms-key\"\n}", Check: resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "encryption_type", "sse-kms")},
			{Config: base + "encryption_type = \"sse-s3\"\n}", Check: resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "encryption_type", "sse-s3")},
		},
	})
	if kmsUpdates != 1 || sseUpdates != 1 {
		t.Errorf("KMS updates=%d SSE-S3 updates=%d, want 1 each", kmsUpdates, sseUpdates)
	}
}
