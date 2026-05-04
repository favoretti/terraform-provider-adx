package adx

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type ADXClusterRequestClassificationPolicyTestResource struct{}

func TestAccADXClusterRequestClassificationPolicy_basic(t *testing.T) {
	r := ADXClusterRequestClassificationPolicyTestResource{}
	databaseName := testAccDatabaseName()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: r.checkDestroy(databaseName),
		Steps: []resource.TestStep{
			{
				Config: r.basic(databaseName),
				Check: resource.ComposeTestCheckFunc(
					r.checkExists("adx_cluster_request_classification_policy.test"),
					resource.TestCheckResourceAttr("adx_cluster_request_classification_policy.test", "database_name", databaseName),
					resource.TestCheckResourceAttr("adx_cluster_request_classification_policy.test", "is_enabled", "true"),
				),
			},
			{
				Config: r.update(databaseName),
				Check: resource.ComposeTestCheckFunc(
					r.checkExists("adx_cluster_request_classification_policy.test"),
					resource.TestCheckResourceAttr("adx_cluster_request_classification_policy.test", "database_name", databaseName),
					resource.TestCheckResourceAttr("adx_cluster_request_classification_policy.test", "is_enabled", "true"),
				),
			},
			{
				ResourceName:      "adx_cluster_request_classification_policy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func (this ADXClusterRequestClassificationPolicyTestResource) basic(databaseName string) string {
	return fmt.Sprintf(`
	resource "adx_cluster_request_classification_policy" "test" {
		database_name = "%s"
		is_enabled    = true

		classification_function = <<-EOT
			iff(request_properties.current_application == "Kusto.Explorer" and request_properties.request_type == "Query",
				"Ad-hoc queries",
				"default")
		EOT
	}
	`, databaseName)
}

func (this ADXClusterRequestClassificationPolicyTestResource) update(databaseName string) string {
	return fmt.Sprintf(`
	resource "adx_cluster_request_classification_policy" "test" {
		database_name = "%s"
		is_enabled    = true

		classification_function = <<-EOT
			case(
				request_properties.current_application == "Kusto.Explorer" and request_properties.request_type == "Query", "Ad-hoc queries",
				request_properties.current_application == "KustoQueryRunner", "Scheduled",
				"default")
		EOT
	}
	`, databaseName)
}

func (this ADXClusterRequestClassificationPolicyTestResource) checkExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("ID is not set for %s", resourceName)
		}

		id, err := parseADXClusterRequestClassificationPolicyID(rs.Primary.ID)
		if err != nil {
			return err
		}

		clusterConfig := GetTestClusterConfig()
		resultSet, err := queryADXMgmtAndParse[ADXRequestClassificationPolicy](
			context.Background(), testAccProvider.Meta(), clusterConfig, id.DatabaseName,
			".show cluster policy request_classification",
		)
		if err != nil {
			return fmt.Errorf("error reading cluster request classification policy: %+v", err)
		}
		if len(resultSet) == 0 || resultSet[0].Policy == "" || resultSet[0].Policy == "null" {
			return fmt.Errorf("cluster request classification policy does not exist")
		}

		return nil
	}
}

func (this ADXClusterRequestClassificationPolicyTestResource) checkDestroy(databaseName string) func(*terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "adx_cluster_request_classification_policy" {
				continue
			}

			clusterConfig := GetTestClusterConfig()
			resultSet, err := queryADXMgmtAndParse[ADXRequestClassificationPolicy](
				context.Background(), testAccProvider.Meta(), clusterConfig, databaseName,
				".show cluster policy request_classification",
			)
			if err != nil {
				if strings.Contains(err.Error(), "BadRequest_EntityNotFound") {
					continue
				}
				return fmt.Errorf("error checking cluster request classification policy destroyed: %+v", err)
			}
			if len(resultSet) > 0 && resultSet[0].Policy != "" && resultSet[0].Policy != "null" {
				return fmt.Errorf("cluster request classification policy still exists")
			}
		}
		return nil
	}
}
