package resources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccParameterGroupResource_basic(t *testing.T) {
	name := acctest.RandomName("tf-pg")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_parameter_group.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_parameter_group.test", "type", "cache"),
					resource.TestCheckResourceAttr("danubedata_parameter_group.test", "provider_type", "redis"),
					resource.TestCheckResourceAttr("danubedata_parameter_group.test", "parameters.maxmemory-policy", "allkeys-lru"),
					resource.TestCheckResourceAttr("danubedata_parameter_group.test", "is_system", "false"),
					resource.TestCheckResourceAttrSet("danubedata_parameter_group.test", "id"),
				),
			},
			{
				ResourceName:      "danubedata_parameter_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccParameterGroupResource_update(t *testing.T) {
	name := acctest.RandomName("tf-pg")
	renamed := name + "-v2"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccParameterGroupResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_parameter_group.test", "name", name),
				),
			},
			{
				Config: testAccParameterGroupResourceConfig_updated(renamed),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_parameter_group.test", "name", renamed),
					resource.TestCheckResourceAttr("danubedata_parameter_group.test", "parameters.maxmemory-policy", "volatile-lru"),
					resource.TestCheckResourceAttr("danubedata_parameter_group.test", "parameters.timeout", "300"),
				),
			},
		},
	})
}

func testAccParameterGroupResourceConfig_basic(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_parameter_group" "test" {
  name          = %q
  type          = "cache"
  provider_type = "redis"
  parameters = {
    "maxmemory-policy" = "allkeys-lru"
  }
}
`, name),
	)
}

func testAccParameterGroupResourceConfig_updated(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_parameter_group" "test" {
  name          = %q
  type          = "cache"
  provider_type = "redis"
  description   = "Updated description"
  parameters = {
    "maxmemory-policy" = "volatile-lru"
    "timeout"          = "300"
  }
}
`, name),
	)
}
