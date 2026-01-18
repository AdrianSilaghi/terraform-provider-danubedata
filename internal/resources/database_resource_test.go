package resources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatabaseResource_basic(t *testing.T) {
	name := acctest.RandomName("tf-db")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccDatabaseResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_database.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_database.test", "datacenter", "fsn1"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "status"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "endpoint"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "port"),
				),
			},
			// Import
			{
				ResourceName:      "danubedata_database.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDatabaseResource_mysql(t *testing.T) {
	name := acctest.RandomName("tf-mysql")
	dbName := "testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResourceConfig_mysql(name, dbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_database.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_database.test", "provider_id", "1"), // MySQL
					resource.TestCheckResourceAttr("danubedata_database.test", "database_name", dbName),
					resource.TestCheckResourceAttr("danubedata_database.test", "version", "8.0"),
				),
			},
		},
	})
}

func TestAccDatabaseResource_postgresql(t *testing.T) {
	name := acctest.RandomName("tf-pg")
	dbName := "testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResourceConfig_postgresql(name, dbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_database.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_database.test", "provider_id", "2"), // PostgreSQL
					resource.TestCheckResourceAttr("danubedata_database.test", "version", "16"),
				),
			},
		},
	})
}

func TestAccDatabaseResource_mariadb(t *testing.T) {
	name := acctest.RandomName("tf-maria")
	dbName := "testdb"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResourceConfig_mariadb(name, dbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_database.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_database.test", "provider_id", "3"), // MariaDB
				),
			},
		},
	})
}

func TestAccDatabaseResource_update(t *testing.T) {
	name := acctest.RandomName("tf-db")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseResourceConfig_withResources(name, 20, 2048, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_database.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_database.test", "storage_size_gb", "20"),
					resource.TestCheckResourceAttr("danubedata_database.test", "memory_size_mb", "2048"),
					resource.TestCheckResourceAttr("danubedata_database.test", "cpu_cores", "2"),
				),
			},
			{
				Config: testAccDatabaseResourceConfig_withResources(name, 50, 4096, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_database.test", "storage_size_gb", "50"),
					resource.TestCheckResourceAttr("danubedata_database.test", "memory_size_mb", "4096"),
					resource.TestCheckResourceAttr("danubedata_database.test", "cpu_cores", "4"),
				),
			},
		},
	})
}

func testAccDatabaseResourceConfig_basic(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_database" "test" {
  name           = %q
  provider_id    = 1  # MySQL
  storage_size_gb = 20
  memory_size_mb = 2048
  cpu_cores      = 2
  datacenter     = "fsn1"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
`, name),
	)
}

func testAccDatabaseResourceConfig_mysql(name, dbName string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_database" "test" {
  name            = %q
  database_name   = %q
  provider_id     = 1  # MySQL
  version         = "8.0"
  storage_size_gb = 20
  memory_size_mb  = 2048
  cpu_cores       = 2
  datacenter      = "fsn1"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
`, name, dbName),
	)
}

func testAccDatabaseResourceConfig_postgresql(name, dbName string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_database" "test" {
  name            = %q
  database_name   = %q
  provider_id     = 2  # PostgreSQL
  version         = "16"
  storage_size_gb = 20
  memory_size_mb  = 2048
  cpu_cores       = 2
  datacenter      = "fsn1"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
`, name, dbName),
	)
}

func testAccDatabaseResourceConfig_mariadb(name, dbName string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_database" "test" {
  name            = %q
  database_name   = %q
  provider_id     = 3  # MariaDB
  storage_size_gb = 20
  memory_size_mb  = 2048
  cpu_cores       = 2
  datacenter      = "fsn1"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
`, name, dbName),
	)
}

func testAccDatabaseResourceConfig_withResources(name string, storageGB, memoryMB, cpuCores int) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_database" "test" {
  name            = %q
  provider_id     = 1  # MySQL
  storage_size_gb = %d
  memory_size_mb  = %d
  cpu_cores       = %d
  datacenter      = "fsn1"

  timeouts {
    create = "15m"
    update = "15m"
    delete = "10m"
  }
}
`, name, storageGB, memoryMB, cpuCores),
	)
}
