package resources_test

import (
	"fmt"
	"testing"

	"github.com/AdrianSilaghi/terraform-provider-danubedata/internal/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServerlessResource_basic(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and read
			{
				Config: testAccServerlessResourceConfig_basic(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "deployment_type", "docker"),
					resource.TestCheckResourceAttrSet("danubedata_serverless.test", "id"),
					resource.TestCheckResourceAttrSet("danubedata_serverless.test", "status"),
					resource.TestCheckResourceAttrSet("danubedata_serverless.test", "url"),
				),
			},
			// Import
			{
				ResourceName:      "danubedata_serverless.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccServerlessResource_docker(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessResourceConfig_docker(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "deployment_type", "docker"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "image_url", "nginx:latest"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "port", "80"),
				),
			},
		},
	})
}

func TestAccServerlessResource_withEnvVars(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessResourceConfig_withEnvVars(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "environment_variables.APP_ENV", "production"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "environment_variables.LOG_LEVEL", "info"),
				),
			},
		},
	})
}

func TestAccServerlessResource_withScaling(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessResourceConfig_withScaling(name, 1, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "name", name),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "min_instances", "1"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "max_instances", "5"),
				),
			},
		},
	})
}

func TestAccServerlessResource_update(t *testing.T) {
	name := acctest.RandomName("tf-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessResourceConfig_withScaling(name, 0, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "min_instances", "0"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "max_instances", "5"),
				),
			},
			{
				Config: testAccServerlessResourceConfig_withScaling(name, 1, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("danubedata_serverless.test", "min_instances", "1"),
					resource.TestCheckResourceAttr("danubedata_serverless.test", "max_instances", "10"),
				),
			},
		},
	})
}

func testAccServerlessResourceConfig_basic(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_serverless" "test" {
  name            = %q
  deployment_type = "docker"
  image_url       = "nginx:latest"
  port            = 80

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name),
	)
}

func testAccServerlessResourceConfig_docker(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_serverless" "test" {
  name             = %q
  resource_profile = "small"
  deployment_type  = "docker"
  image_url        = "nginx:latest"
  port             = 80
  min_instances    = 0
  max_instances    = 10

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name),
	)
}

func testAccServerlessResourceConfig_withEnvVars(name string) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_serverless" "test" {
  name            = %q
  deployment_type = "docker"
  image_url       = "nginx:latest"
  port            = 80

  environment_variables = {
    APP_ENV   = "production"
    LOG_LEVEL = "info"
  }

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, name),
	)
}

func testAccServerlessResourceConfig_withScaling(name string, minInstances, maxInstances int) string {
	return acctest.ConfigCompose(
		acctest.ProviderConfig(),
		fmt.Sprintf(`
resource "danubedata_serverless" "test" {
  name            = %q
  deployment_type = "docker"
  image_url       = "nginx:latest"
  port            = 80
  min_instances   = %d
  max_instances   = %d

  timeouts {
    create = "10m"
    update = "10m"
    delete = "10m"
  }
}
`, name, minInstances, maxInstances),
	)
}
