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

func TestUnitServerlessZipUploadRejectedBeforeAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("ZIP plan reached API: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"zip_upload unsupported"}`))
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "test-token"
}
resource "danubedata_serverless" "zip" {
  name = "archive"
  slug = "archive"
  deployment_type = "zip_upload"
}`, server.URL),
			ExpectError: regexp.MustCompile("zip_upload"),
		}},
	})
}

func TestUnitServerlessResource_Lifecycle(t *testing.T) {
	name := "initial"
	gitAuthType := "none"
	created := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /serverless":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["slug"] != "unit-app" {
				t.Errorf("slug = %v", body["slug"])
			}
			if _, ok := body["git_credentials"]; ok {
				t.Error("public Git create sent credentials")
			}
			if body["git_auth_type"] != "none" {
				t.Errorf("public Git create auth = %v, want none", body["git_auth_type"])
			}
			if auth, ok := body["git_auth_type"].(string); ok {
				gitAuthType = auth
			}
			created = true
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"container": map[string]interface{}{"id": "unit-1", "spec_generation": 1}})
		case "PUT /serverless/unit-1":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if body["name"] != "renamed" {
				t.Errorf("update name = %v", body["name"])
			}
			name = "renamed"
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"container": map[string]interface{}{"id": "unit-1", "spec_generation": 2}})
		case "GET /serverless/unit-1":
			if !created {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			generation := 1
			if name == "renamed" {
				generation = 2
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"container": map[string]interface{}{
				"id": "unit-1", "name": name, "slug": "unit-app", "status": "running", "resource_profile": "small",
				"deployment_type": "git_repository", "image_tag": "latest", "repository_url": "https://github.com/example/app", "repository_branch": "main", "source_type": "dockerfile", "git_auth_type": gitAuthType,
				"port": 8080, "min_scale": 0, "max_scale": 10, "spec_generation": generation, "observed_generation": generation,
				"status_details":        map[string]interface{}{"summary": "ready", "operation": map[string]bool{"terminal": true}},
				"environment_variables": map[string]string{}, "created_at": "2026-01-01", "updated_at": "2026-01-01",
			}, "url": "https://unit.example", "monthly_cost": "0.00"})
		case "DELETE /serverless/unit-1":
			created = false
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := func(name string) string {
		return fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "test-token"
}
resource "danubedata_serverless" "test" {
  name = %q
  slug = "unit-app"
  deployment_type = "git_repository"
  repository_url = "https://github.com/example/app"
  source_type = "dockerfile"
}`, server.URL, name)
	}
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("initial"), Check: resource.TestCheckResourceAttr("danubedata_serverless.test", "slug", "unit-app")},
			{Config: config("renamed"), Check: resource.TestCheckResourceAttr("danubedata_serverless.test", "name", "renamed")},
			{ResourceName: "danubedata_serverless.test", ImportState: true, ImportStateVerify: true},
		},
	})
}

func TestAccServerlessResource_basic(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccServerlessResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "deployment_type", "docker_image"),
					resource.TestCheckResourceAttrSet("danubedata_serverless.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_serverless.test", "status"),
					resource.TestCheckResourceAttrSet("danubedata_serverless.test", "url"),
				),
			},
			// Import
			{
				ResourceName:      "danubedata_serverless.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServerlessResource_docker(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessResourceConfig_docker(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "deployment_type", "docker_image"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "image", "nginx"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "image_tag", "latest"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "port", "80"),
				),
			},
		},
	})
}

func TestAccServerlessResource_withEnvVars(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessResourceConfig_withEnvVars(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "environment_variables.APP_ENV", "production"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "environment_variables.LOG_LEVEL", "info"),
				),
			},
		},
	})
}

func TestAccServerlessResource_withScaling(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessResourceConfig_withScaling(name, 1, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "min_scale", "1"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "max_scale", "5"),
				),
			},
		},
	})
}

func TestAccServerlessResource_update(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessResourceConfig_withScaling(name, 0, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "min_scale", "0"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "max_scale", "5"),
				),
			},
			{
				Config: testAccServerlessResourceConfig_withScaling(name, 1, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "min_scale", "1"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "max_scale", "10"),
				),
			},
		},
	})
}

func TestAccServerlessResource_git(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessResourceConfig_git(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "deployment_type", "git_repository"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "repository_url", "https://github.com/danubedata/example-app"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "repository_branch", "main"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "source_type", "dockerfile"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "git_auth_type", "none"),
				),
			},
		},
	})
}

func testAccServerlessResourceConfig_basic(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_serverless" "test" {
  name            = %q
  slug            = %q
  deployment_type = "docker_image"
  image           = "nginx"
  port            = 80

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name, name),
	)
}

func testAccServerlessResourceConfig_docker(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_serverless" "test" {
  name             = %q
  slug             = %q
  resource_profile = "small"
  deployment_type  = "docker_image"
  image            = "nginx"
  image_tag        = "latest"
  port             = 80
  min_scale        = 0
  max_scale        = 10

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name, name),
	)
}

func testAccServerlessResourceConfig_withEnvVars(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_serverless" "test" {
  name            = %q
  slug            = %q
  deployment_type = "docker_image"
  image           = "nginx"
  port            = 80

  environment_variables = {
    APP_ENV   = "production"
    LOG_LEVEL = "info"
  }

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name, name),
	)
}

func testAccServerlessResourceConfig_withScaling(name string, minScale, maxScale int) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_serverless" "test" {
  name            = %q
  slug            = %q
  deployment_type = "docker_image"
  image           = "nginx"
  port            = 80
  min_scale       = %d
  max_scale       = %d

  timeouts {
    create = "10m"
    update = "10m"
    delete = "10m"
  }
}
`, name, name, minScale, maxScale),
	)
}

func testAccServerlessResourceConfig_git(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_serverless" "test" {
  name              = %q
  slug              = %q
  deployment_type   = "git_repository"
  repository_url    = "https://github.com/danubedata/example-app"
  repository_branch = "main"
  source_type       = "dockerfile"
  git_auth_type     = "none"
  port              = 8080

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name, name),
	)
}
