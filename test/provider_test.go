package test

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jd-ucpa/terraform-provider-test/internal"
)

// testAccProtoV5ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV5ProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
	"test": providerserver.NewProtocol5WithError(internal.Provider()),
}

func testAccPreCheck(t *testing.T) {
	// Vérifier les variables d'environnement nécessaires pour les tests
	// Par exemple, vérifier que AWS_PROFILE est défini
}

// TestAccProvider_InvalidRegion vérifie que le provider rejette correctement une région AWS invalide.
// Ce test configure le provider avec la région "eu-waste-1" (qui n'existe pas) et vérifie
// que le provider génère l'erreur attendue "invalid AWS Region: eu-waste-1" lors de la validation.
// Cela confirme que la validation des régions fonctionne comme le provider AWS officiel.
func TestAccProvider_InvalidRegion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigInvalidRegion(),
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
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigInvalidAssumeRole(),
				ExpectError: regexp.MustCompile(`"assume_role.0.role_arn" \(arn:aws:iam::1234s50599128:role/assumable\) is an invalid ARN: invalid account ID value`),
			},
		},
	})
}



func testAccProviderConfigInvalidRegion() string {
	return `
		provider "test" {
			region = "eu-waste-1"
		}

		# Utiliser une ressource pour déclencher la validation du provider
		resource "test_ssm_send_command" "test" {
			instance_ids = ["` + os.Getenv("INSTANCE_ID") + `"]
			document_name = "AWS-RunShellScript"
			parameters = {
				commands = "echo 'test'"
			}
		}
	`
}

func testAccProviderConfigInvalidAssumeRole() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "arn:aws:iam::1234s50599128:role/assumable"
			}
		}

		# Utiliser une ressource pour déclencher la validation du provider
		resource "test_ssm_send_command" "test" {
			instance_ids = ["` + os.Getenv("INSTANCE_ID") + `"]
			document_name = "AWS-RunShellScript"
			parameters = {
				commands = "echo 'test'"
			}
		}
	`
}


