package resources_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestUnitStaticSiteAndDomain_Lifecycle(t *testing.T) {
	siteExists, domainExists := false, false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "POST /static-sites":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if _, ok := body["plan"]; ok {
				t.Error("deprecated plan sent to API")
			}
			siteExists = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"data":{"id":"site-1","name":"docs","slug":"docs","url":"https://docs.pages.example","plan":"free","status":"pending","created_at":"2026-01-01","updated_at":"2026-01-01"}}`))
		case "GET /static-sites/site-1":
			if !siteExists {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(`{"data":{"id":"site-1","name":"docs","slug":"docs","url":"https://docs.pages.example","plan":"free","status":"pending","created_at":"2026-01-01","updated_at":"2026-01-01"}}`))
		case "DELETE /static-sites/site-1":
			siteExists = false
			w.WriteHeader(http.StatusNoContent)
		case "POST /static-sites/site-1/domains":
			domainExists = true
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"data":{"id":"domain-1","domain":"www.example.com","verification_status":"pending","tls_status":"pending","deployment_status":"pending","is_primary":false,"dns_instructions":"Add a TXT record","created_at":"2026-01-01"}}`))
		case "GET /static-sites/site-1/domains":
			if domainExists {
				_, _ = w.Write([]byte(`{"data":[{"id":"domain-1","domain":"www.example.com","verification_status":"pending","tls_status":"pending","deployment_status":"pending","is_primary":false,"dns_instructions":"Add a TXT record","created_at":"2026-01-01"}]}`))
			} else {
				_, _ = w.Write([]byte(`{"data":[]}`))
			}
		case "DELETE /static-sites/site-1/domains/domain-1":
			domainExists = false
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "test-token"
}
resource "danubedata_static_site" "test" {
  name = "docs"
}
resource "danubedata_static_site_domain" "test" {
  static_site_id = danubedata_static_site.test.id
  domain = "www.example.com"
}`, server.URL)
	legacyPlanConfig := strings.Replace(config, "name = \"docs\"", "name = \"docs\"\n  plan = \"pro\"", 1)
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config, Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("danubedata_static_site.test", "name", "docs"),
				resource.TestCheckResourceAttr("danubedata_static_site_domain.test", "dns_instructions", "Add a TXT record"),
			)},
			{Config: legacyPlanConfig, ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
				plancheck.ExpectResourceAction("danubedata_static_site.test", plancheck.ResourceActionUpdate),
			}}, Check: resource.TestCheckResourceAttr("danubedata_static_site.test", "plan", "pro")},
			{ResourceName: "danubedata_static_site_domain.test", ImportState: true, ImportStateId: "site-1:www.example.com", ImportStateVerify: true},
		},
	})
}
