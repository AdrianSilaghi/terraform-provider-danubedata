package resources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitServerlessResourceUpdatesRuntimeFields(t *testing.T) {
	profile, image, tag, port, minScale, maxScale := "small", "nginx", "v1", 8080, 1, 3
	env := map[string]string{"OLD": "secret"}
	generation, updates, exists := 1, 0, false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /serverless":
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body["environment_variables"] == nil || body["min_scale"] != float64(1) {
				t.Errorf("create lost runtime fields: %v", body)
			}
			exists = true
			w.WriteHeader(http.StatusCreated)
			_, _ = fmt.Fprint(w, `{"container":{"id":"app-1","spec_generation":1}}`)
		case "PUT /serverless/app-1":
			updates++
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			for key, want := range map[string]any{"resource_profile": "micro", "image": "caddy", "image_tag": "v2", "port": float64(80), "min_scale": float64(0), "max_scale": float64(5)} {
				if body[key] != want {
					t.Errorf("update %s = %v, want %v", key, body[key], want)
				}
			}
			if values, ok := body["environment_variables"].(map[string]any); !ok || len(values) != 0 {
				t.Errorf("update must clear environment variables, got %v", body["environment_variables"])
			}
			if _, ok := body["git_credentials"]; ok {
				t.Error("update sent write-only credentials")
			}
			profile, image, tag, port, minScale, maxScale, env = "micro", "caddy", "v2", 80, 0, 5, map[string]string{}
			generation++
			_, _ = fmt.Fprintf(w, `{"container":{"id":"app-1","spec_generation":%d}}`, generation)
		case "GET /serverless/app-1":
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"container": map[string]any{
				"id": "app-1", "name": "app", "slug": "app", "status": "running", "resource_profile": profile,
				"deployment_type": "docker_image", "image": image, "image_tag": tag, "repository_branch": "main", "git_auth_type": "none",
				"port": port, "min_scale": minScale, "max_scale": maxScale, "environment_variables": env,
				"spec_generation": generation, "observed_generation": generation,
				"status_details": map[string]any{"summary": "ready", "operation": map[string]bool{"terminal": true}},
			}, "url": "https://app.example", "monthly_cost": "0.00"})
		case "DELETE /serverless/app-1":
			exists = false
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := func(profile, image, tag string, port, min, max int, variables string) string {
		return fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "test-token"
}

resource "danubedata_serverless" "test" {
  name = "app"
  slug = "app"
  deployment_type = "docker_image"
  resource_profile = %q
  image = %q
  image_tag = %q
  port = %d
  min_scale = %d
  max_scale = %d
  environment_variables = %s
}`, server.URL, profile, image, tag, port, min, max, variables)
	}
	resource.Test(t, resource.TestCase{IsUnitTest: true, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{
		{Config: config("small", "nginx", "v1", 8080, 1, 3, `{ OLD = "secret" }`)},
		{Config: config("micro", "caddy", "v2", 80, 0, 5, `{}`), Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("danubedata_serverless.test", "min_scale", "0"),
			resource.TestCheckResourceAttr("danubedata_serverless.test", "environment_variables.%", "0"),
		)},
	}})
	if updates != 1 {
		t.Errorf("update requests = %d, want 1", updates)
	}
}

func TestUnitServerlessResourceRejectsPrivateGitBeforeAPI(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	for _, tc := range []struct {
		name, auth, credentials string
	}{
		{name: "private auth", auth: "access_token"},
		{name: "explicit credentials", auth: "none", credentials: `git_credentials = "unit-secret"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resource.Test(t, resource.TestCase{IsUnitTest: true, ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories, Steps: []resource.TestStep{{
				Config: fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "test-token"
}
resource "danubedata_serverless" "private" {
  name = "private"
  slug = "private"
  deployment_type = "git_repository"
  repository_url = "https://example.com/org/app"
  source_type = "dockerfile"
	git_auth_type = %q
  %s
}`, server.URL, tc.auth, tc.credentials),
				ExpectError: regexp.MustCompile(`(?i)private git.*not supported|git credentials.*not supported`),
			}}})
		})
	}
	if calls != 0 {
		t.Errorf("private Git create made %d API calls before reporting unsupported credentials", calls)
	}
}
