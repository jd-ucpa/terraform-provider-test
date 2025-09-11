package test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSSMActivationsDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					data "test_ssm_activations" "test" {
						expired = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_ssm_activations.test", "id"),
				),
			},
		},
	})
}

func TestAccSSMActivationsDataSource_WithIamRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					data "test_ssm_activations" "test" {
						expired = true
						filter {
							name   = "iam-role"
							values = ["` + getVar("ACTIVATION_ROLE_NAME") + `"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_ssm_activations.test", "id"),
				),
			},
		},
	})
}

func TestAccSSMActivationsDataSource_WithFilters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					data "test_ssm_activations" "test" {
						expired = true
						filter {
							name   = "iam-role"
							values = ["` + getVar("ACTIVATION_ROLE_NAME") + `"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_ssm_activations.test", "id"),
				),
			},
		},
	})
}

func TestAccSSMActivationsDataSource_Validation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					data "test_ssm_activations" "test" {
						filter {
							name   = "InvalidFilterName"
							values = ["value1", "value2"]
						}
					}
				`,
				ExpectError: regexp.MustCompile(`Filter name 'InvalidFilterName' is invalid`),
			},
		},
	})
}

func TestAccSSMActivationsDataSource_FilterValuesValidation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					data "test_ssm_activations" "test" {
						filter {
							name   = "iam-role"
							values = []
						}
					}
				`,
				ExpectError: regexp.MustCompile(`Filter requires at least one value`),
			},
		},
	})
}

func TestAccSSMActivationsDataSource_WithMultipleFilters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					data "test_ssm_activations" "test" {
						expired = false
						filter {
							name   = "iam-role"
							values = ["` + getVar("ACTIVATION_ROLE_NAME") + `"]
						}
						filter {
							name   = "default-instance-name"
							values = ["ssm-managed"]
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_ssm_activations.test", "id"),
				),
			},
		},
	})
}
