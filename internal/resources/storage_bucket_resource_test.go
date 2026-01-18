package resources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStorageBucketResource_basic(t *testing.T) {
	name := acctest.RandomName("tf-bucket")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccStorageBucketResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "region", "fsn1"),
					resource.TestCheckResourceAttrSet("danubedata_storage_bucket.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_storage_bucket.test", "status"),
					resource.TestCheckResourceAttrSet("danubedata_storage_bucket.test", "endpoint_url"),
				),
			},
			// Import
			{
				ResourceName:      "danubedata_storage_bucket.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStorageBucketResource_withVersioning(t *testing.T) {
	name := acctest.RandomName("tf-bucket")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketResourceConfig_withVersioning(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "versioning_enabled", "true"),
				),
			},
		},
	})
}

func TestAccStorageBucketResource_publicAccess(t *testing.T) {
	name := acctest.RandomName("tf-bucket")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketResourceConfig_publicAccess(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "public_access", "true"),
				),
			},
		},
	})
}

func TestAccStorageBucketResource_update(t *testing.T) {
	name := acctest.RandomName("tf-bucket")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketResourceConfig_withVersioning(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "versioning_enabled", "false"),
				),
			},
			{
				Config: testAccStorageBucketResourceConfig_withVersioning(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_storage_bucket.test", "versioning_enabled", "true"),
				),
			},
		},
	})
}

func testAccStorageBucketResourceConfig_basic(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_storage_bucket" "test" {
  name   = %q
  region = "fsn1"

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`, name),
	)
}

func testAccStorageBucketResourceConfig_withVersioning(name string, versioning bool) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_storage_bucket" "test" {
  name               = %q
  region             = "fsn1"
  versioning_enabled = %t

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`, name, versioning),
	)
}

func testAccStorageBucketResourceConfig_publicAccess(name string, publicAccess bool) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_storage_bucket" "test" {
  name          = %q
  region        = "fsn1"
  public_access = %t

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`, name, publicAccess),
	)
}
