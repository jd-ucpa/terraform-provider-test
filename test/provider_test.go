package test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccProvider_InvalidRegion vérifie que le provider rejette correctement une région AWS invalide.
// Ce test configure le provider avec la région "eu-waste-1" (qui n'existe pas) et vérifie
// que le provider génère l'erreur attendue "invalid AWS Region: eu-waste-1" lors de la validation.
// Cela confirme que la validation des régions fonctionne comme le provider AWS officiel.
func TestAccProvider_InvalidRegion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-waste-1"
					}

					resource "test_ssm_send_command" "test" {
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						document_name = "AWS-RunShellScript"
						parameters = {
							commands = "echo 'test'"
						}
					}
				`,
				ExpectError: regexp.MustCompile(`invalid AWS Region: eu-waste-1`),
			},
		},
	})
}

// TestAccProvider_InvalidAssumeRole vérifie que le provider rejette correctement un ARN de rôle invalide.
// Ce test configure le provider avec un ARN "arn:aws:iam::1234s50599128:role/assumable" qui contient
// des lettres dans l'ID de compte (1234s50599128) au lieu de chiffres uniquement. Le test vérifie
// que le provider génère l'erreur attendue avec le pattern regex correspondant au provider AWS officiel.
func TestAccProvider_InvalidAssumeRole(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						assume_role {
							role_arn = "arn:aws:iam::1234s50599128:role/assumable"
						}
					}

					resource "test_ssm_send_command" "test" {
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						document_name = "AWS-RunShellScript"
						parameters = {
							commands = "echo 'test'"
						}
					}
				`,
				ExpectError: regexp.MustCompile(`invalid account ID value`),
			},
		},
	})
}
