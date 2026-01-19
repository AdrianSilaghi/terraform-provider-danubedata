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
					resource.TestCheckResourceAttr("danubedata_cache.test", "cache_provider", "redis"),
					resource.TestCheckResourceAttrSet("danubedata_cache.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_cache.test", "status"),
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
					resource.TestCheckResourceAttr("danubedata_cache.test", "cache_provider", "redis"),
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
					resource.TestCheckResourceAttr("danubedata_cache.test", "cache_provider", "valkey"),
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
				Config: testAccCacheResourceConfig_withProfile(name, "micro"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_cache.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_cache.test", "resource_profile", "micro"),
					resource.TestCheckResourceAttrSet("danubedata_cache.test", "memory_size_mb"),
					resource.TestCheckResourceAttrSet("danubedata_cache.test", "cpu_cores"),
				),
			},
			{
				Config: testAccCacheResourceConfig_withProfile(name, "small"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_cache.test", "resource_profile", "small"),
					resource.TestCheckResourceAttrSet("danubedata_cache.test", "memory_size_mb"),
					resource.TestCheckResourceAttrSet("danubedata_cache.test", "cpu_cores"),
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
  name             = %q
  cache_provider   = "redis"
  resource_profile = "micro"
  memory_size_mb   = 256
  cpu_cores        = 1
  datacenter       = "fsn1"

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
  name             = %q
  cache_provider   = "redis"
  resource_profile = "micro"
  memory_size_mb   = 256
  cpu_cores        = 1
  datacenter       = "fsn1"
  version          = "7.2"

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
  name             = %q
  cache_provider   = "valkey"
  resource_profile = "micro"
  memory_size_mb   = 256
  cpu_cores        = 1
  datacenter       = "fsn1"

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name),
	)
}

func testAccCacheResourceConfig_withProfile(name string, resourceProfile string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_cache" "test" {
  name             = %q
  cache_provider   = "redis"
  resource_profile = %q
  datacenter       = "fsn1"

  timeouts {
    create = "10m"
    update = "10m"
    delete = "10m"
  }
}
`, name, resourceProfile),
	)
}
