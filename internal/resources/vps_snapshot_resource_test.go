package resources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVpsSnapshotResource_basic(t *testing.T) {
	vpsName := acctest.RandomName("tf-vps")
	snapshotName := acctest.RandomName("tf-snap")
	sshKeyName := acctest.RandomName("tf-key")
	sshPubKey := acctest.RandomSSHPublicKey()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccVpsSnapshotResourceConfig_basic(sshKeyName, sshPubKey, vpsName, snapshotName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_vps_snapshot.test", "name", snapshotName),
					resource.TestCheckResourceAttrSet("danubedata_vps_snapshot.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_vps_snapshot.test", "status"),
					resource.TestCheckResourceAttrSet("danubedata_vps_snapshot.test", "vps_instance_id"),
				),
			},
			// Import
			{
				ResourceName:      "danubedata_vps_snapshot.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVpsSnapshotResource_withDescription(t *testing.T) {
	vpsName := acctest.RandomName("tf-vps")
	snapshotName := acctest.RandomName("tf-snap")
	sshKeyName := acctest.RandomName("tf-key")
	sshPubKey := acctest.RandomSSHPublicKey()
	description := "Test snapshot for Terraform acceptance testing"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVpsSnapshotResourceConfig_withDescription(sshKeyName, sshPubKey, vpsName, snapshotName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_vps_snapshot.test", "name", snapshotName),
					resource.TestCheckResourceAttr("danubedata_vps_snapshot.test", "description", description),
				),
			},
		},
	})
}

func testAccVpsSnapshotResourceConfig_basic(sshKeyName, sshPubKey, vpsName, snapshotName string) string {
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

resource "danubedata_vps_snapshot" "test" {
  name            = %q
  vps_instance_id = danubedata_vps.test.id

  timeouts {
    create = "10m"
    delete = "5m"
  }
}
`, sshKeyName, sshPubKey, vpsName, snapshotName),
	)
}

func testAccVpsSnapshotResourceConfig_withDescription(sshKeyName, sshPubKey, vpsName, snapshotName, description string) string {
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

resource "danubedata_vps_snapshot" "test" {
  name            = %q
  description     = %q
  vps_instance_id = danubedata_vps.test.id

  timeouts {
    create = "10m"
    delete = "5m"
  }
}
`, sshKeyName, sshPubKey, vpsName, snapshotName, description),
	)
}
