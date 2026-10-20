package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func storageOfflineFactory() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"danubedata": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func TestOfflineStorageBucketLifecycle(t *testing.T) {
	var mu sync.Mutex
	created, versioning, deleted := false, false, false
	encryptionEnabled, encryptionType := true, "sse-s3"
	var sizeLimit any
	platformCors := map[string]any{"AllowedOrigins": []string{"https://console.danubedata.ro"}, "AllowedMethods": []string{"GET", "PUT", "POST", "DELETE", "HEAD"}, "AllowedHeaders": []string{"*"}, "ExposeHeaders": []string{"ETag", "x-amz-request-id", "x-amz-id-2"}, "MaxAgeSeconds": 86400}
	var userCors []any
	var tags any = []any{}
	bucket := func() map[string]any {
		corsRules := append([]any{platformCors}, userCors...)
		return map[string]any{
			"id": "bucket-1", "name": "offline-bucket", "display_name": nil,
			"status": "active", "region": "fsn1", "endpoint_url": "https://s3.test",
			"minio_bucket_name": "dd-17-offline-bucket", "public_access": false,
			"versioning_enabled": versioning, "encryption_enabled": encryptionEnabled,
			"encryption_type": encryptionType, "object_lock_enabled": false,
			"object_lock_default_retention_mode": nil, "object_lock_default_retention_days": nil,
			"size_bytes": 0, "size_limit_bytes": sizeLimit, "object_count": 0,
			"tags": tags, "cors_configuration": map[string]any{"CORSRules": corsRules}, "lifecycle_rules": nil,
			"monthly_cost_cents": 0, "monthly_cost_dollars": 0,
			"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/storage/buckets":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["name"] != "offline-bucket" || body["region"] != "fsn1" {
				t.Errorf("bucket create body: %v, %v", body, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodGet && r.URL.Path == "/storage/buckets/bucket-1":
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket(), "endpoint": "https://s3.test"})
		case r.Method == http.MethodPut && r.URL.Path == "/storage/buckets/bucket-1":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("bucket update body: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if value, ok := body["versioning_enabled"].(bool); ok {
				versioning = value
			}
			if value, ok := body["size_limit_bytes"]; ok {
				sizeLimit = value
			}
			if value, ok := body["encryption_enabled"].(bool); ok {
				encryptionEnabled = value
			}
			if value, ok := body["encryption_type"].(string); ok {
				encryptionType = value
			}
			if encryptionEnabled == (encryptionType == "none") {
				t.Errorf("incoherent encryption request: %v", body)
			}
			if value, ok := body["cors_configuration"].(map[string]any); ok {
				userCors, _ = value["CORSRules"].([]any)
			}
			if value, ok := body["tags"]; ok {
				tags = value
				if len(value.(map[string]any)) == 0 {
					tags = []any{}
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket()})
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/buckets/bucket-1":
			deleted, created = true, false
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "destroying"})
		default:
			t.Errorf("unexpected bucket request %s %s", r.Method, r.URL.String())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	base := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
  team_id = 17
}
resource "danubedata_storage_bucket" "test" {
name = "offline-bucket"
region = "fsn1"`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: storageOfflineFactory(),
		Steps: []resource.TestStep{
			{Config: base + "\nsize_limit_bytes = 1073741824\n}", Check: resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "id", "bucket-1")},
			{Config: base + "\nversioning_enabled = true\nsize_limit_bytes = 2147483648\ncors_configuration = jsonencode({ CORSRules = [{ AllowedOrigins = [\"https://app.test\"], AllowedMethods = [\"GET\"] }] })\ntags = { team = \"infra\" }\n}", Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "versioning_enabled", "true"),
				resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "size_limit_bytes", "2147483648"),
				resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "tags.team", "infra"),
			)},
			{Config: base + "\nversioning_enabled = true\nsize_limit_bytes = 2147483648\n}"},
			{Config: base + "\nversioning_enabled = true\nsize_limit_bytes = 2147483648\nencryption_enabled = false\nencryption_type = \"none\"\n}", Check: resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "encryption_type", "none")},
			{Config: base + "\nversioning_enabled = true\nsize_limit_bytes = 2147483648\nencryption_enabled = true\nencryption_type = \"sse-s3\"\n}", Check: resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "encryption_enabled", "true")},
			{ResourceName: "danubedata_storage_bucket.test", ImportState: true, ImportStateVerify: true},
			{RefreshState: true},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	if !deleted || !versioning || sizeLimit != float64(2147483648) || !encryptionEnabled || encryptionType != "sse-s3" {
		t.Errorf("bucket lifecycle: deleted=%v versioning=%v quota=%v encryption=%v/%s", deleted, versioning, sizeLimit, encryptionEnabled, encryptionType)
	}
}

func TestOfflineScopedStorageAccessKeyLifecycle(t *testing.T) {
	var mu sync.Mutex
	created, deleted := false, false
	key := map[string]any{
		"id": "key-1", "name": "offline-key", "access_key_id": "AK1", "status": "active",
		"scope": "buckets", "bucket_permissions": []map[string]string{{"bucket_id": "bucket-1", "level": "read"}},
		"arn": "arn:aws:iam::17:user/key-1", "expires_at": nil,
		"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/storage/access-keys":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["scope"] != "buckets" {
				t.Errorf("key create body: %v, %v", body, err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			perms, _ := body["bucket_permissions"].([]any)
			if len(perms) != 1 {
				t.Errorf("key permissions: %v", perms)
			}
			created = true
			response := map[string]any{}
			for k, v := range key {
				response[k] = v
			}
			response["secret_access_key"] = "one-time-secret"
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(response)
		case r.Method == http.MethodGet && r.URL.Path == "/storage/access-keys/key-1":
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"access_key": key})
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/access-keys/key-1":
			created, deleted = false, true
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "revoked"})
		default:
			t.Errorf("unexpected key request %s %s", r.Method, r.URL.String())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
  team_id = 17
}

resource "danubedata_storage_access_key" "test" {
  name = "offline-key"
  scope = "buckets"
  bucket_permissions = { "bucket-1" = "read" }
}`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: storageOfflineFactory(),
		Steps: []resource.TestStep{
			{Config: config, Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("danubedata_storage_access_key.test", "scope", "buckets"),
				resource.TestCheckResourceAttr("danubedata_storage_access_key.test", "secret_access_key", "one-time-secret"),
			)},
			{RefreshState: true, Check: resource.TestCheckResourceAttr("danubedata_storage_access_key.test", "secret_access_key", "one-time-secret")},
			{ResourceName: "danubedata_storage_access_key.test", ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"secret_access_key"}},
		},
	})
	mu.Lock()
	defer mu.Unlock()
	if !deleted {
		t.Error("key was not revoked during destroy")
	}
}

func TestOfflineStorageBucketWaitFailureRetainsID(t *testing.T) {
	created, deleted := 0, false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/storage/buckets":
			created++
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"bucket":{"id":"bucket-failed","name":"failed-bucket","status":"pending"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/storage/buckets/bucket-failed":
			if deleted {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(`{"bucket":{"id":"bucket-failed","name":"failed-bucket","status":"error"}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/buckets/bucket-failed":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}

resource "danubedata_storage_bucket" "test" {
  name = "failed-bucket"
  region = "fsn1"
}`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: storageOfflineFactory(),
		Steps:                    []resource.TestStep{{Config: config, ExpectError: regexp.MustCompile("failed to reach active state")}},
	})
	if created != 1 || !deleted {
		t.Errorf("bucket error cleanup: creates=%d deleted=%v", created, deleted)
	}
}

func TestOfflineDefaultTeamStorageKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/storage/access-keys":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body["scope"] != "team" || body["bucket_permissions"] != nil {
				t.Errorf("default key body: %v, %v", body, err)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"key-team","name":"team-key","access_key_id":"AK","secret_access_key":"secret","scope":"team","bucket_permissions":[]}`))
		case r.Method == http.MethodGet && r.URL.Path == "/storage/access-keys/key-team":
			_, _ = w.Write([]byte(`{"access_key":{"id":"key-team","name":"team-key","access_key_id":"AK","status":"active","scope":"team","bucket_permissions":[],"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/access-keys/key-team":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected team key request %s %s", r.Method, r.URL.String())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}

resource "danubedata_storage_access_key" "test" {
  name = "team-key"
}`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: storageOfflineFactory(),
		Steps:                    []resource.TestStep{{Config: config, Check: resource.TestCheckResourceAttr("danubedata_storage_access_key.test", "scope", "team")}},
	})
}

func TestOfflineImportedObjectLockDoesNotReplace(t *testing.T) {
	created, posts := false, 0
	bucket := map[string]any{
		"id": "locked-bucket", "name": "locked-bucket", "status": "active", "region": "fsn1",
		"endpoint_url": "https://s3.test", "minio_bucket_name": "dd-17-locked-bucket",
		"versioning_enabled": true, "object_lock_enabled": true,
		"encryption_enabled": true, "encryption_type": "sse-s3",
		"tags": []any{}, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z",
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/storage/buckets":
			posts++
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket})
		case r.Method == http.MethodGet && r.URL.Path == "/storage/buckets/locked-bucket":
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"bucket": bucket, "endpoint": "https://s3.test"})
		case r.Method == http.MethodDelete && r.URL.Path == "/storage/buckets/locked-bucket":
			created = false
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected locked bucket request %s %s", r.Method, r.URL.String())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	base := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}
resource "danubedata_storage_bucket" "test" {
  name = "locked-bucket"
  region = "fsn1"
  versioning_enabled = true
`, server.URL)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: storageOfflineFactory(),
		Steps: []resource.TestStep{
			{Config: base + "object_lock_enabled = true\n}"},
			{Config: base + "}", Check: resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "object_lock_enabled", "true")},
			{ResourceName: "danubedata_storage_bucket.test", ImportState: true, ImportStateVerify: true},
		},
	})
	if posts != 1 {
		t.Errorf("bucket was recreated when Object Lock config was omitted: POSTs=%d", posts)
	}
}
