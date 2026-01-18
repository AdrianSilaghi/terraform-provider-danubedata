package resources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStorageAccessKeyResource_basic(t *testing.T) {
	name := acctest.RandomName("tf-key")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccStorageAccessKeyResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_storage_access_key.test", "name", name),
					resource.TestCheckResourceAttrSet("danubedata_storage_access_key.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_storage_access_key.test", "access_key_id"),
					resource.TestCheckResourceAttrSet("danubedata_storage_access_key.test", "secret_access_key"),
				),
			},
			// Import - note: secret_access_key is only returned on creation
			{
				ResourceName:      "danubedata_storage_access_key.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"secret_access_key",
				},
			},
		},
	})
}

func TestAccStorageAccessKeyResource_multiple(t *testing.T) {
	name1 := acctest.RandomName("tf-key")
	name2 := acctest.RandomName("tf-key")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageAccessKeyResourceConfig_multiple(name1, name2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_storage_access_key.test1", "name", name1),
					resource.TestCheckResourceAttr("danubedata_storage_access_key.test2", "name", name2),
					resource.TestCheckResourceAttrSet("danubedata_storage_access_key.test1", "access_key_id"),
					resource.TestCheckResourceAttrSet("danubedata_storage_access_key.test2", "access_key_id"),
				),
			},
		},
	})
}

func testAccStorageAccessKeyResourceConfig_basic(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_storage_access_key" "test" {
  name = %q
}
`, name),
	)
}

func testAccStorageAccessKeyResourceConfig_multiple(name1, name2 string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_storage_access_key" "test1" {
  name = %q
}

resource "danubedata_storage_access_key" "test2" {
  name = %q
}
`, name1, name2),
	)
}
