package test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCloudWatchMetricsDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_metrics" "metrics" {
						namespace = "AWS/EC2"
						metric_name = "CPUUtilization"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_metrics.metrics", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "metric_name", "CPUUtilization"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricsDataSource_WithNamespace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_metrics" "metrics" {
						namespace = "AWS/EC2"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_metrics.metrics", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "namespace", "AWS/EC2"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricsDataSource_WithMetricName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_metrics" "metrics" {
						metric_name = "CPUUtilization"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_metrics.metrics", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "metric_name", "CPUUtilization"),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricsDataSource_WithDimensions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}

					data "test_cloudwatch_metrics" "metrics" {
						namespace = "AWS/EC2"
						
						dimension {
							name = "InstanceType"
							value = "t4g.small"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_metrics.metrics", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "dimension.0.name", "InstanceType"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "dimension.0.value", "t4g.small"),
          checkMinMetricsCount(10),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricsDataSource_WithMultipleDimensions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}

					data "test_cloudwatch_metrics" "metrics" {
						namespace = "AWS/EC2"
						
						dimension {
              name = "ImageId"
              value = "ami-03820227fb3e4ffad"
            }
            dimension {
              name = "fstype"
              value = "vfat"
            }
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_metrics.metrics", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "namespace", "AWS/EC2"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "dimension.0.name", "ImageId"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "dimension.0.value", "ami-03820227fb3e4ffad"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "dimension.1.name", "fstype"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_metrics.metrics", "dimension.1.value", "vfat"),
          checkMinMetricsCount(1),
				),
			},
		},
	})
}

func TestAccCloudWatchMetricsDataSource_NoFilters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}

					data "test_cloudwatch_metrics" "metrics" {
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_metrics.metrics", "id"),
          checkMinMetricsCount(250),
				),
			},
		},
	})
}

// checkMinMetricsCount crée une fonction de vérification pour un nombre minimum de métriques
func checkMinMetricsCount(minCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources["data.test_cloudwatch_metrics.metrics"]
		if !ok {
			return fmt.Errorf("data source not found")
		}

		metricsCount := ds.Primary.Attributes["metrics.#"]
		if metricsCount == "" {
			return fmt.Errorf("metrics count not found")
		}

		count, err := strconv.Atoi(metricsCount)
		if err != nil {
			return fmt.Errorf("failed to parse metrics count: %s", err)
		}

		if count < minCount {
			return fmt.Errorf("expected at least %d metrics, got %d", minCount, count)
		}

		return nil
	}
}
