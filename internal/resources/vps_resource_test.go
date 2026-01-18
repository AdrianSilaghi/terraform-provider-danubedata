package resources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVpsResource_basic(t *testing.T) {
	name := acctest.RandomName("tf-vps")
	sshKeyName := acctest.RandomName("tf-key")
	sshPubKey := acctest.RandomSSHPublicKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccVpsResourceConfig_basic(sshKeyName, sshPubKey, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_vps.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_vps.test", "image", "ubuntu-22.04"),
					resource.TestCheckResourceAttr("danubedata_vps.test", "datacenter", "fsn1"),
					resource.TestCheckResourceAttr("danubedata_vps.test", "auth_method", "ssh_key"),
					resource.TestCheckResourceAttrSet("danubedata_vps.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_vps.test", "status"),
					resource.TestCheckResourceAttrSet("danubedata_vps.test", "cpu_cores"),
					resource.TestCheckResourceAttrSet("danubedata_vps.test", "memory_size_gb"),
					resource.TestCheckResourceAttrSet("danubedata_vps.test", "storage_size_gb"),
				),
			},
			// Import
			{
				ResourceName:      "danubedata_vps.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auth_method",
					"password",
				},
			},
		},
	})
}

func TestAccVpsResource_withResourceProfile(t *testing.T) {
	name := acctest.RandomName("tf-vps")
	sshKeyName := acctest.RandomName("tf-key")
	sshPubKey := acctest.RandomSSHPublicKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpsResourceConfig_withProfile(sshKeyName, sshPubKey, name, "vps-small"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_vps.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_vps.test", "resource_profile", "vps-small"),
				),
			},
		},
	})
}

func TestAccVpsResource_update(t *testing.T) {
	name := acctest.RandomName("tf-vps")
	sshKeyName := acctest.RandomName("tf-key")
	sshPubKey := acctest.RandomSSHPublicKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with initial profile
			{
				Config: testAccVpsResourceConfig_withProfile(sshKeyName, sshPubKey, name, "vps-small"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_vps.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_vps.test", "resource_profile", "vps-small"),
				),
			},
			// Update to larger profile
			{
				Config: testAccVpsResourceConfig_withProfile(sshKeyName, sshPubKey, name, "vps-medium"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_vps.test", "resource_profile", "vps-medium"),
				),
			},
		},
	})
}

func TestAccVpsResource_withCustomResources(t *testing.T) {
	name := acctest.RandomName("tf-vps")
	sshKeyName := acctest.RandomName("tf-key")
	sshPubKey := acctest.RandomSSHPublicKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpsResourceConfig_withCustomResources(sshKeyName, sshPubKey, name, 2, 4, 50),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_vps.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_vps.test", "cpu_cores", "2"),
					resource.TestCheckResourceAttr("danubedata_vps.test", "memory_size_gb", "4"),
					resource.TestCheckResourceAttr("danubedata_vps.test", "storage_size_gb", "50"),
				),
			},
		},
	})
}

func TestAccVpsResource_withPassword(t *testing.T) {
	name := acctest.RandomName("tf-vps")
	password := acctest.RandomPassword()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpsResourceConfig_withPassword(name, password),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_vps.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_vps.test", "auth_method", "password"),
				),
			},
		},
	})
}

func testAccVpsResourceConfig_basic(sshKeyName, sshPubKey, name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_ssh_key" "test" {
  name       = %q
  public_key = %q
}

resource "danubedata_vps" "test" {
  name        = %q
  image       = "ubuntu-22.04"
  datacenter  = "fsn1"
  auth_method = "ssh_key"
  ssh_key_id  = danubedata_ssh_key.test.id

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, sshKeyName, sshPubKey, name),
	)
}

func testAccVpsResourceConfig_withProfile(sshKeyName, sshPubKey, name, profile string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_ssh_key" "test" {
  name       = %q
  public_key = %q
}

resource "danubedata_vps" "test" {
  name             = %q
  image            = "ubuntu-22.04"
  datacenter       = "fsn1"
  resource_profile = %q
  auth_method      = "ssh_key"
  ssh_key_id       = danubedata_ssh_key.test.id

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, sshKeyName, sshPubKey, name, profile),
	)
}

func testAccVpsResourceConfig_withCustomResources(sshKeyName, sshPubKey, name string, cpuCores, memoryGB, storageGB int) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_ssh_key" "test" {
  name       = %q
  public_key = %q
}

resource "danubedata_vps" "test" {
  name           = %q
  image          = "ubuntu-22.04"
  datacenter     = "fsn1"
  auth_method    = "ssh_key"
  ssh_key_id     = danubedata_ssh_key.test.id
  cpu_cores      = %d
  memory_size_gb = %d
  storage_size_gb = %d

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, sshKeyName, sshPubKey, name, cpuCores, memoryGB, storageGB),
	)
}

func testAccVpsResourceConfig_withPassword(name, password string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_vps" "test" {
  name        = %q
  image       = "ubuntu-22.04"
  datacenter  = "fsn1"
  auth_method = "password"
  password    = %q

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name, password),
	)
}
