package resources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func storageReviewFactory() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"danubedata": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func TestStorageReviewRejectsInvalidBucketSettingsBeforeCreate(t *testing.T) {
	for _, tc := range []struct {
		name, settings, diagnostic string
	}{
		{"lock without versioning", "object_lock_enabled = true", "Object Lock requires versioning"},
		{"retention without days", "object_lock_enabled = true\nversioning_enabled = true\nobject_lock_default_retention_mode = \"GOVERNANCE\"", "Retention mode and days must be set together"},
		{"quota below minimum", "size_limit_bytes = 123", "Size limit must be at least 1 GiB"},
		{"invalid CORS JSON", "cors_configuration = \"{\"", "CORS configuration must be valid JSON"},
		{"invalid lifecycle JSON", "lifecycle_rules = \"{\"", "Lifecycle rules must be valid JSON"},
		{"missing KMS key", "encryption_type = \"sse-kms\"", "sse-kms requires a KMS key ID"},
		{"disabled with SSE", "encryption_enabled = false\nencryption_type = \"sse-s3\"", "Use type none when encryption is disabled"},
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
resource "danubedata_storage_bucket" "test" {
  name = "review-bucket"
  region = "fsn1"
  %s
}`, server.URL, tc.settings)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: storageReviewFactory(),
				Steps:                    []resource.TestStep{{Config: config, ExpectError: regexp.MustCompile(tc.diagnostic)}},
			})
			if posts != 0 {
				t.Errorf("invalid settings caused %d bucket creates", posts)
			}
		})
	}
}

func TestStorageReviewRevokesKeyOnIncompleteOrWrongScopeResponse(t *testing.T) {
	for _, tc := range []struct {
		name       string
		response   map[string]any
		diagnostic string
	}{
		{"missing secret", map[string]any{"id": "key-1", "access_key_id": "AK1", "scope": "buckets", "bucket_permissions": []map[string]string{{"bucket_id": "bucket-1", "level": "read"}}}, "Incomplete access key response"},
		{"wrong scope", map[string]any{"id": "key-1", "access_key_id": "AK1", "secret_access_key": "secret", "scope": "team"}, "scope was not applied"},
		{"missing permission", map[string]any{"id": "key-1", "access_key_id": "AK1", "secret_access_key": "secret", "scope": "buckets"}, "scope was not applied"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			posts, deletes := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodPost && r.URL.Path == "/storage/access-keys":
					posts++
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(tc.response)
				case r.Method == http.MethodDelete && r.URL.Path == "/storage/access-keys/key-1":
					deletes++
					w.WriteHeader(http.StatusNoContent)
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			config := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}

resource "danubedata_storage_access_key" "test" {
  name = "review-key"
  scope = "buckets"
  bucket_permissions = { "bucket-1" = "read" }
}`, server.URL)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: storageReviewFactory(),
				Steps:                    []resource.TestStep{{Config: config, ExpectError: regexp.MustCompile(tc.diagnostic)}},
			})
			if posts != 1 || deletes != 1 {
				t.Errorf("creates=%d revokes=%d, want 1 each", posts, deletes)
			}
		})
	}
}

func TestStorageReviewExplicitEmptyRulesRoundTrip(t *testing.T) {
	created := false
	platformRule := map[string]any{
		"AllowedOrigins": []string{"https://console.danubedata.ro"},
		"AllowedMethods": []string{"GET", "PUT", "POST", "DELETE", "HEAD"},
		"AllowedHeaders": []string{"*"},
		"ExposeHeaders":  []string{"ETag", "x-amz-request-id", "x-amz-id-2"},
		"MaxAgeSeconds":  86400,
	}
	bucket := func() map[string]any {
		return map[string]any{
			"id": "bucket-empty", "name": "empty-rules", "status": "active", "region": "fsn1",
			"endpoint_url": "https://s3.test", "minio_bucket_name": "dd-17-empty-rules",
			"encryption_enabled": true, "encryption_type": "sse-s3",
			"cors_configuration": map[string]any{"CORSRules": []any{platformRule}},
			"lifecycle_rules":    []any{}, "tags": []any{},
			"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/storage/buckets":
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if string(body["lifecycle_rules"]) != `[]` || string(body["cors_configuration"]) != `{"CORSRules":[]}` {
				t.Errorf("create rules: CORS=%s lifecycle=%s", body["cors_configuration"], body["lifecycle_rules"])
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodGet && r.URL.Path == "/storage/buckets/bucket-empty":
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/buckets/bucket-empty":
			created = false
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}

resource "danubedata_storage_bucket" "test" {
  name = "empty-rules"
  region = "fsn1"
  cors_configuration = jsonencode({ CORSRules = [] })
  lifecycle_rules = jsonencode([])
}`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: storageReviewFactory(),
		Steps: []resource.TestStep{
			{Config: config, Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "cors_configuration", `{"CORSRules":[]}`),
				resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "lifecycle_rules", `[]`),
			)},
			{RefreshState: true},
		},
	})
}

func TestStorageReviewClearsDefaultRetention(t *testing.T) {
	created, cleared := false, false
	mode := any("GOVERNANCE")
	days := any(7)
	bucket := func() map[string]any {
		return map[string]any{
			"id": "bucket-locked", "name": "locked-review", "status": "active", "region": "fsn1",
			"endpoint_url": "https://s3.test", "minio_bucket_name": "dd-17-locked-review",
			"encryption_enabled": true, "encryption_type": "sse-s3",
			"versioning_enabled": true, "object_lock_enabled": true,
			"object_lock_default_retention_mode": mode, "object_lock_default_retention_days": days,
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
		case r.Method == http.MethodGet && r.URL.Path == "/storage/buckets/bucket-locked":
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodPut && r.URL.Path == "/storage/buckets/bucket-locked":
			var body map[string]json.RawMessage
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if string(body["object_lock_default_retention_mode"]) != "null" || string(body["object_lock_default_retention_days"]) != "null" {
				t.Errorf("clear retention body: %v", body)
			}
			mode, days, cleared = nil, nil, true
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/buckets/bucket-locked":
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
  name = "locked-review"
  region = "fsn1"
  versioning_enabled = true
  object_lock_enabled = true
`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: storageReviewFactory(),
		Steps: []resource.TestStep{
			{Config: base + "object_lock_default_retention_mode = \"GOVERNANCE\"\nobject_lock_default_retention_days = 7\n}"},
			{Config: base + "}", Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckNoResourceAttr("danubedata_storage_bucket.test", "object_lock_default_retention_mode"),
				resource.TestCheckNoResourceAttr("danubedata_storage_bucket.test", "object_lock_default_retention_days"),
			)},
		},
	})
	if !cleared {
		t.Error("default retention was not cleared")
	}
}
