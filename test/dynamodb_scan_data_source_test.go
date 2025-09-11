package test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccDynamoDBScanDataSource tests the DynamoDB scan data source
func TestAccDynamoDBScanDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_dynamodb_scan" "test" {
						table_name = "allowed_products"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "id", "dynamodb_scan_allowed_products"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "table_name", "allowed_products"),
					checkMinItemsCount(10),
					resource.TestCheckResourceAttrSet("data.test_dynamodb_scan.test", "items_count"),
					resource.TestCheckResourceAttrSet("data.test_dynamodb_scan.test", "scanned_count"),
				),
			},
		},
	})
}

// TestAccDynamoDBScanDataSource_WithIndex tests the DynamoDB scan data source with index
func TestAccDynamoDBScanDataSource_WithIndex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_dynamodb_scan" "test" {
						table_name = "allowed_products"
						projection_expression = "product_uuid, customer_uuid"
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "id", "dynamodb_scan_allowed_products"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "table_name", "allowed_products"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "projection_expression", "product_uuid, customer_uuid"),
					checkMinItemsCount(1),
				),
			},
		},
	})
}

// TestAccDynamoDBScanDataSource_WithAlias tests the DynamoDB scan data source with alias
func TestAccDynamoDBScanDataSource_WithAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_dynamodb_scan" "test" {
						table_name = "allowed_products"
						projection_expression = "cu, pu"
						expression_attribute_names = {
							cu = "customer_uuid"
							pu = "product_uuid"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "id", "dynamodb_scan_allowed_products"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "table_name", "allowed_products"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "projection_expression", "cu, pu"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "expression_attribute_names.cu", "customer_uuid"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "expression_attribute_names.pu", "product_uuid"),
					checkMinItemsCount(1),
				),
			},
		},
	})
}

// TestAccDynamoDBScanDataSource_WithPrefixedAlias tests the DynamoDB scan data source with prefixed alias
func TestAccDynamoDBScanDataSource_WithPrefixedAlias(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_dynamodb_scan" "test" {
						table_name = "allowed_products"
						projection_expression = "#cu, pu"
						expression_attribute_names = {
							"#cu" = "customer_uuid"
							pu = "product_uuid"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "id", "dynamodb_scan_allowed_products"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "table_name", "allowed_products"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "projection_expression", "#cu, pu"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "expression_attribute_names.#cu", "customer_uuid"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "expression_attribute_names.pu", "product_uuid"),
					checkMinItemsCount(1),
				),
			},
		},
	})
}

// TestAccDynamoDBScanDataSource_WithFilter tests the DynamoDB scan data source with filter
func TestAccDynamoDBScanDataSource_WithFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_dynamodb_scan" "test" {
						table_name = "allowed_products"
						filter_expression = "st = accepted"
						expression_attribute_names = {
							st = "status"
						}
						expression_attribute_values = {
							accepted = "ACCEPTED"
						}
					}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "id", "dynamodb_scan_allowed_products"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "table_name", "allowed_products"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "filter_expression", "st = accepted"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "expression_attribute_names.st", "status"),
					resource.TestCheckResourceAttr("data.test_dynamodb_scan.test", "expression_attribute_values.accepted", "ACCEPTED"),
					checkMinItemsCount(1),
					resource.TestCheckResourceAttrSet("data.test_dynamodb_scan.test", "items_count"),
				),
			},
		},
	})
}

// checkMinItemsCount crée une fonction de vérification pour un nombre minimum d'éléments
func checkMinItemsCount(minCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources["data.test_dynamodb_scan.test"]
		if !ok {
			return fmt.Errorf("data source not found")
		}
		
		itemsCountStr := ds.Primary.Attributes["items.#"]
		itemsCount, err := strconv.Atoi(itemsCountStr)
		if err != nil {
			return fmt.Errorf("impossible de convertir le nombre d'éléments: %v", err)
		}
		
		if itemsCount < minCount {
			return fmt.Errorf("nombre d'éléments insuffisant: %d (minimum attendu: %d)", itemsCount, minCount)
		}
		
		return nil
	}
}
