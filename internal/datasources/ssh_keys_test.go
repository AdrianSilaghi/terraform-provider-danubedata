package datasources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSshKeysDataSource_basic(t *testing.T) {
	name := acctest.RandomName("tf-key")
	pubKey := acctest.RandomSSHPublicKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshKeysDataSourceConfig_basic(name, pubKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.danubedata_ssh_keys.test", "keys.#"),
				),
			},
		},
	})
}

func TestAccSshKeysDataSource_withMultipleKeys(t *testing.T) {
	name1 := acctest.RandomName("tf-key")
	name2 := acctest.RandomName("tf-key")
	pubKey1 := acctest.RandomSSHPublicKey()
	pubKey2 := acctest.RandomSSHPublicKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSshKeysDataSourceConfig_multiple(name1, pubKey1, name2, pubKey2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.danubedata_ssh_keys.test", "keys.#"),
				),
			},
		},
	})
}

func testAccSshKeysDataSourceConfig_basic(name, pubKey string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_ssh_key" "test" {
  name       = %q
  public_key = %q
}

data "danubedata_ssh_keys" "test" {
  depends_on = [danubedata_ssh_key.test]
}
`, name, pubKey),
	)
}

func testAccSshKeysDataSourceConfig_multiple(name1, pubKey1, name2, pubKey2 string) string {
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

data "danubedata_ssh_keys" "test" {
  depends_on = [
    danubedata_ssh_key.test1,
    danubedata_ssh_key.test2,
  ]
}
`, name1, pubKey1, name2, pubKey2),
	)
}
