package resources_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitDatabaseResource_RejectsNamesOutsideAPIContract(t *testing.T) {
	for _, name := range []string{"1database", "d" + strings.Repeat("a", 50)} {
		t.Run(name[:1], func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()
			config := fmt.Sprintf(`provider "danubedata" {
  base_url = %q
  api_token = "offline-token"
}
resource "danubedata_database" "test" {
  name = %q
  engine = "mysql"
  resource_profile = "small"
  datacenter = "fsn1"
}`, server.URL, name)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
				Steps:                    []resource.TestStep{{Config: config, ExpectError: regexp.MustCompile(`(?i)name.*(lowercase|50|length)`)}},
			})
			if requests != 0 {
				t.Fatalf("invalid name sent %d API requests", requests)
			}
		})
	}
}
