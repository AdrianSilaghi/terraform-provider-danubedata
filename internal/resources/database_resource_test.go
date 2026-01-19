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
					resource.TestCheckResourceAttr("danubedata_database.test", "resource_profile", "small"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "status"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "endpoint"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "port"),
				),
			},
			// Import
			{
				ResourceName:            "danubedata_database.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"updated_at"},
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
					resource.TestCheckResourceAttr("danubedata_database.test", "engine", "mysql"),
					resource.TestCheckResourceAttr("danubedata_database.test", "database_name", dbName),
					resource.TestCheckResourceAttr("danubedata_database.test", "version", "8.0"),
					resource.TestCheckResourceAttr("danubedata_database.test", "resource_profile", "small"),
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
					resource.TestCheckResourceAttr("danubedata_database.test", "engine", "postgresql"),
					resource.TestCheckResourceAttr("danubedata_database.test", "version", "16"),
					resource.TestCheckResourceAttr("danubedata_database.test", "resource_profile", "small"),
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
					resource.TestCheckResourceAttr("danubedata_database.test", "engine", "mariadb"),
					resource.TestCheckResourceAttr("danubedata_database.test", "resource_profile", "small"),
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
				Config: testAccDatabaseResourceConfig_withProfile(name, "small"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_database.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_database.test", "resource_profile", "small"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "storage_size_gb"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "memory_size_mb"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "cpu_cores"),
				),
			},
			{
				Config: testAccDatabaseResourceConfig_withProfile(name, "medium"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_database.test", "resource_profile", "medium"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "storage_size_gb"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "memory_size_mb"),
					resource.TestCheckResourceAttrSet("danubedata_database.test", "cpu_cores"),
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
  name             = %q
  engine           = "mysql"
  resource_profile = "small"
  datacenter       = "fsn1"

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
  name             = %q
  database_name    = %q
  engine           = "mysql"
  version          = "8.0"
  resource_profile = "small"
  datacenter       = "fsn1"

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
  name             = %q
  database_name    = %q
  engine           = "postgresql"
  version          = "16"
  resource_profile = "small"
  datacenter       = "fsn1"

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
  name             = %q
  database_name    = %q
  engine           = "mariadb"
  resource_profile = "small"
  datacenter       = "fsn1"

  timeouts {
    create = "15m"
    delete = "10m"
  }
}
`, name, dbName),
	)
}

func testAccDatabaseResourceConfig_withProfile(name string, resourceProfile string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_database" "test" {
  name             = %q
  engine           = "mysql"
  resource_profile = %q
  datacenter       = "fsn1"

  timeouts {
    create = "15m"
    update = "15m"
    delete = "10m"
  }
}
`, name, resourceProfile),
	)
}
