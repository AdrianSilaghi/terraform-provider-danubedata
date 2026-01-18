package resources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFirewallResource_basic(t *testing.T) {
	name := acctest.RandomName("tf-fw")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccFirewallResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_firewall.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_firewall.test", "default_action", "deny"),
					resource.TestCheckResourceAttrSet("danubedata_firewall.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_firewall.test", "status"),
				),
			},
			// Import
			{
				ResourceName:      "danubedata_firewall.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccFirewallResource_withRules(t *testing.T) {
	name := acctest.RandomName("tf-fw")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallResourceConfig_withRules(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_firewall.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_firewall.test", "rules.#", "2"),
				),
			},
		},
	})
}

func TestAccFirewallResource_update(t *testing.T) {
	name := acctest.RandomName("tf-fw")
	updatedName := acctest.RandomName("tf-fw-updated")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_firewall.test", "name", name),
				),
			},
			{
				Config: testAccFirewallResourceConfig_basic(updatedName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_firewall.test", "name", updatedName),
				),
			},
		},
	})
}

func TestAccFirewallResource_withDescription(t *testing.T) {
	name := acctest.RandomName("tf-fw")
	description := "Test firewall for Terraform acceptance testing"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallResourceConfig_withDescription(name, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_firewall.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_firewall.test", "description", description),
				),
			},
		},
	})
}

func testAccFirewallResourceConfig_basic(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_firewall" "test" {
  name           = %q
  default_action = "deny"
}
`, name),
	)
}

func testAccFirewallResourceConfig_withRules(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_firewall" "test" {
  name           = %q
  default_action = "deny"

  rules {
    name             = "Allow SSH"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 22
    port_range_end   = 22
    source_ips       = ["0.0.0.0/0"]
    priority         = 100
  }

  rules {
    name             = "Allow HTTP"
    action           = "allow"
    direction        = "inbound"
    protocol         = "tcp"
    port_range_start = 80
    port_range_end   = 80
    source_ips       = ["0.0.0.0/0"]
    priority         = 200
  }
}
`, name),
	)
}

func testAccFirewallResourceConfig_withDescription(name, description string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_firewall" "test" {
  name           = %q
  description    = %q
  default_action = "deny"
}
`, name, description),
	)
}
