package resources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCacheResource_basic(t *testing.T) {
	name := acctest.RandomName("tf-cache")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccCacheResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_cache.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_cache.test", "datacenter", "fsn1"),
					resource.TestCheckResourceAttrSet("danubedata_cache.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_cache.test", "status"),
					resource.TestCheckResourceAttrSet("danubedata_cache.test", "endpoint"),
					resource.TestCheckResourceAttrSet("danubedata_cache.test", "port"),
				),
			},
			// Import
			{
				ResourceName:      "danubedata_cache.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCacheResource_redis(t *testing.T) {
	name := acctest.RandomName("tf-redis")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCacheResourceConfig_redis(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_cache.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_cache.test", "provider_id", "1"), // Redis
					resource.TestCheckResourceAttr("danubedata_cache.test", "memory_size_mb", "512"),
				),
			},
		},
	})
}

func TestAccCacheResource_valkey(t *testing.T) {
	name := acctest.RandomName("tf-valkey")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCacheResourceConfig_valkey(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_cache.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_cache.test", "provider_id", "2"), // Valkey
				),
			},
		},
	})
}

func TestAccCacheResource_update(t *testing.T) {
	name := acctest.RandomName("tf-cache")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCacheResourceConfig_withResources(name, 512, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_cache.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_cache.test", "memory_size_mb", "512"),
					resource.TestCheckResourceAttr("danubedata_cache.test", "cpu_cores", "1"),
				),
			},
			{
				Config: testAccCacheResourceConfig_withResources(name, 1024, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_cache.test", "memory_size_mb", "1024"),
					resource.TestCheckResourceAttr("danubedata_cache.test", "cpu_cores", "2"),
				),
			},
		},
	})
}

func testAccCacheResourceConfig_basic(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_cache" "test" {
  name           = %q
  provider_id    = 1  # Redis
  memory_size_mb = 512
  cpu_cores      = 1
  datacenter     = "fsn1"

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name),
	)
}

func testAccCacheResourceConfig_redis(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_cache" "test" {
  name           = %q
  provider_id    = 1  # Redis
  memory_size_mb = 512
  cpu_cores      = 1
  datacenter     = "fsn1"
  version        = "7.2"

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name),
	)
}

func testAccCacheResourceConfig_valkey(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_cache" "test" {
  name           = %q
  provider_id    = 2  # Valkey
  memory_size_mb = 512
  cpu_cores      = 1
  datacenter     = "fsn1"

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name),
	)
}

func testAccCacheResourceConfig_withResources(name string, memoryMB, cpuCores int) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_cache" "test" {
  name           = %q
  provider_id    = 1  # Redis
  memory_size_mb = %d
  cpu_cores      = %d
  datacenter     = "fsn1"

  timeouts {
    create = "10m"
    update = "10m"
    delete = "10m"
  }
}
`, name, memoryMB, cpuCores),
	)
}
