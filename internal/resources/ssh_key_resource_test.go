package resources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSshKeyResource_basic(t *testing.T) {
	name := acctest.RandomName("tf-key")
	pubKey := acctest.RandomSSHPublicKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccSshKeyResourceConfig(name, pubKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_ssh_key.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_ssh_key.test", "public_key", pubKey),
					resource.TestCheckResourceAttrSet("danubedata_ssh_key.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_ssh_key.test", "fingerprint"),
					resource.TestCheckResourceAttrSet("danubedata_ssh_key.test", "created_at"),
				),
			},
			// Import
			{
				ResourceName:      "danubedata_ssh_key.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSshKeyResource_multiple(t *testing.T) {
	name1 := acctest.RandomName("tf-key")
	name2 := acctest.RandomName("tf-key")
	pubKey1 := acctest.RandomSSHPublicKey()
	pubKey2 := acctest.RandomSSHPublicKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshKeyResourceConfig_multiple(name1, pubKey1, name2, pubKey2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_ssh_key.test1", "name", name1),
					resource.TestCheckResourceAttr("danubedata_ssh_key.test2", "name", name2),
					resource.TestCheckResourceAttrSet("danubedata_ssh_key.test1", "fingerprint"),
					resource.TestCheckResourceAttrSet("danubedata_ssh_key.test2", "fingerprint"),
				),
			},
		},
	})
}

func testAccSshKeyResourceConfig(name, publicKey string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_ssh_key" "test" {
  name       = %q
  public_key = %q
}
`, name, publicKey),
	)
}

func testAccSshKeyResourceConfig_multiple(name1, pubKey1, name2, pubKey2 string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_ssh_key" "test1" {
  name       = %q
  public_key = %q
}

resource "danubedata_ssh_key" "test2" {
  name       = %q
  public_key = %q
}
`, name1, pubKey1, name2, pubKey2),
	)
}
