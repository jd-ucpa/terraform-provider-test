package test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCloudWatchAlarmsDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_alarms" "alarms" {
						alarm_type = "MetricAlarm"
					}

					output "alarms" {
						value = data.test_cloudwatch_alarms.alarms
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_alarms.alarms", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_alarms.alarms", "alarm_type", "MetricAlarm"),
					// Vérifier que nous avons au moins 2 alarmes
					checkMinAlarmsCount(2),
				),
			},
		},
	})
}

func TestAccCloudWatchAlarmsDataSource_WithPrefix(t *testing.T) {
	prefix := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_alarms" "alarms" {
						alarm_type = "MetricAlarm"
						alarm_name_prefix = "` + prefix + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_alarms.alarms", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_alarms.alarms", "alarm_name_prefix", prefix),
				),
			},
		},
	})
}

func TestAccCloudWatchAlarmsDataSource_WithStateValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_alarms" "alarms" {
						alarm_type = "MetricAlarm"
						state_value = "OK"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_alarms.alarms", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_alarms.alarms", "state_value", "OK"),
				),
			},
		},
	})
}

func TestAccCloudWatchAlarmsDataSource_CompositeAlarm(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_alarms" "alarms" {
						alarm_type = "CompositeAlarm"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_alarms.alarms", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_alarms.alarms", "alarm_type", "CompositeAlarm"),
				),
			},
		},
	})
}

func TestAccCloudWatchAlarmsDataSource_IgnoreAutoscalingAlarms(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_alarms" "alarms" {
						alarm_type = "MetricAlarm"
						ignore_autoscaling_alarms = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_alarms.alarms", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_alarms.alarms", "ignore_autoscaling_alarms", "true"),
					// Vérifier que les alarmes TargetTracking- sont filtrées
					func(s *terraform.State) error {
						ds, ok := s.RootModule().Resources["data.test_cloudwatch_alarms.alarms"]
						if !ok {
							return fmt.Errorf("data source not found")
						}
						
						// Vérifier qu'aucune alarme ne commence par "TargetTracking-"
						for i := 0; i < len(ds.Primary.Attributes); i++ {
							key := fmt.Sprintf("alarms.%d.alarm_name", i)
							if alarmName, exists := ds.Primary.Attributes[key]; exists {
								if len(alarmName) >= 16 && alarmName[:16] == "TargetTracking-" {
									return fmt.Errorf("found TargetTracking alarm that should have been filtered: %s", alarmName)
								}
							}
						}
						
						return nil
					},
				),
			},
		},
	})
}

func TestAccCloudWatchAlarmsDataSource_IncludeAutoscalingAlarms(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_alarms" "alarms" {
						alarm_type = "MetricAlarm"
						ignore_autoscaling_alarms = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_cloudwatch_alarms.alarms", "id"),
					resource.TestCheckResourceAttr("data.test_cloudwatch_alarms.alarms", "ignore_autoscaling_alarms", "false"),
					// Vérification silencieuse des alarmes TargetTracking
					func(s *terraform.State) error {
						ds, ok := s.RootModule().Resources["data.test_cloudwatch_alarms.alarms"]
						if !ok {
							return fmt.Errorf("data source not found")
						}
						
						// Vérification basique que le data source fonctionne
						if ds.Primary.ID == "" {
							return fmt.Errorf("data source ID is empty")
						}
						
						return nil
					},
				),
			},
		},
	})
}

func TestAccCloudWatchAlarmsDataSource_InvalidAlarmType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_alarms" "alarms" {
						alarm_type = "InvalidAlarmType"
					}
				`,
				ExpectError: regexp.MustCompile(`alarm_type must be either 'MetricAlarm' or 'CompositeAlarm'`),
			},
		},
	})
}

func TestAccCloudWatchAlarmsDataSource_InvalidStateValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}

					data "test_cloudwatch_alarms" "alarms" {
						state_value = "INVALID_STATE"
					}
				`,
				ExpectError: regexp.MustCompile(`state_value must be either 'OK', 'ALARM', or 'INSUFFICIENT_DATA'`),
			},
		},
	})
}

// checkMinAlarmsCount crée une fonction de vérification pour un nombre minimum d'alarmes
func checkMinAlarmsCount(minCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ds, ok := s.RootModule().Resources["data.test_cloudwatch_alarms.alarms"]
		if !ok {
			return fmt.Errorf("data source not found")
		}
		
		alarmsCountStr := ds.Primary.Attributes["alarms.#"]
		alarmsCount, err := strconv.Atoi(alarmsCountStr)
		if err != nil {
			return fmt.Errorf("impossible de convertir le nombre d'alarmes: %v", err)
		}
		
		if alarmsCount < minCount {
			return fmt.Errorf("nombre d'alarmes insuffisant: %d (minimum attendu: %d)", alarmsCount, minCount)
		}
		
		return nil
	}
}
