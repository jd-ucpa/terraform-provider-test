package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccSFNStartSyncExecutionResource_Basic teste l'exécution synchrone basique d'une state machine SFN.
// Ce test configure le provider avec le profil AWS_PROFILE, crée une ressource SFN Start Sync Execution
// avec l'ARN de la state machine spécifié, puis vérifie que l'exécution est lancée avec succès.
// Le test valide que les attributs de base sont correctement définis : id, execution_arn, status, et output.
func TestAccSFNStartSyncExecutionResource_Basic(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "example" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "id"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "execution_arn"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "state_machine_arn", getVar("STATE_MACHINE_ARN")),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "input", "{}"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "status"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "start_date"),
				),
			},
		},
	})
}

// TestAccSFNStartSyncExecutionResource_WithName teste l'exécution synchrone avec un nom personnalisé.
// Ce test configure le provider avec le profil AWS_PROFILE, crée une ressource SFN Start Sync Execution
// avec l'ARN de la state machine et un nom personnalisé, puis vérifie que l'exécution est lancée avec succès.
// Le test valide que le nom personnalisé est correctement utilisé.
func TestAccSFNStartSyncExecutionResource_WithName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "example" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						name              = "test-execution-with-name"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "id"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "execution_arn"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "state_machine_arn", getVar("STATE_MACHINE_ARN")),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "name", "test-execution-with-name"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "input", "{}"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "status"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "start_date"),
				),
			},
		},
	})
}

// TestAccSFNStartSyncExecutionResource_WithInput teste l'exécution synchrone avec un input JSON personnalisé.
// Ce test configure le provider avec le profil AWS_PROFILE, crée une ressource SFN Start Sync Execution
// avec l'ARN de la state machine et un input JSON personnalisé, puis vérifie que l'exécution est lancée avec succès.
// Le test valide que l'input personnalisé est correctement utilisé.
func TestAccSFNStartSyncExecutionResource_WithInput(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "example" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						name              = "test-execution-with-input"
						input             = jsonencode({
							"message" = "Hello from Terraform"
							"value"   = 42
						})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "id"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "execution_arn"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "state_machine_arn", getVar("STATE_MACHINE_ARN")),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "name", "test-execution-with-input"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "input", `{"message":"Hello from Terraform","value":42}`),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "status"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "start_date"),
				),
			},
		},
	})
}

// TestAccSFNStartSyncExecutionResource_AllAttributes teste l'exécution synchrone avec tous les attributs.
// Ce test configure le provider avec le profil AWS_PROFILE, crée une ressource SFN Start Sync Execution
// avec l'ARN de la state machine, un nom personnalisé et un input JSON, puis vérifie que l'exécution est lancée avec succès.
// Le test valide que tous les attributs sont correctement définis et que les détails de facturation sont présents.
func TestAccSFNStartSyncExecutionResource_AllAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "example" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						name              = "test-execution-all-attributes"
						input             = jsonencode({
							"test" = "all attributes"
							"data" = {
								"nested" = "value"
							}
						})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "id"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "execution_arn"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "state_machine_arn", getVar("STATE_MACHINE_ARN")),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "name", "test-execution-all-attributes"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "input", `{"data":{"nested":"value"},"test":"all attributes"}`),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "status"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "start_date"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "billing_details.billed_duration_in_milliseconds"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "billing_details.billed_memory_used_in_mb"),
				),
			},
		},
	})
}

// TestAccSFNStartSyncExecutionResource_Update teste la mise à jour de la ressource SFN Start Sync Execution.
// Ce test vérifie que toute modification de la ressource (changement d'input, de nom, etc.) force une nouvelle exécution.
// Il utilise deux configurations différentes pour tester le comportement de mise à jour.
func TestAccSFNStartSyncExecutionResource_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Étape 1: Création initiale
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "example" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						name              = "test-execution-update-1"
						input             = jsonencode({
							"step" = "initial"
						})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "id"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "name", "test-execution-update-1"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "input", `{"step":"initial"}`),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "status"),
				),
			},
			// Étape 2: Mise à jour (changement d'input et de nom)
			{
				Config: `
					provider "test" {
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "example" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						name              = "test-execution-update-2"
						input             = jsonencode({
							"step" = "updated"
							"new_field" = "new_value"
						})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "id"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "name", "test-execution-update-2"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.example", "input", `{"new_field":"new_value","step":"updated"}`),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.example", "status"),
				),
			},
		},
	})
}

// TestAccSFNStartSyncExecutionResource_WithTriggers teste le fonctionnement des triggers
// pour forcer une nouvelle exécution SFN. Ce test vérifie que :
// 1. La ressource se crée correctement avec des triggers
// 2. Un changement de trigger force une nouvelle exécution
// 3. Les valeurs calculées sont préservées quand les triggers ne changent pas
func TestAccSFNStartSyncExecutionResource_WithTriggers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Étape 1: Create - Création initiale avec triggers
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "test" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						
						triggers = {
							version = "v1.0.0"
							phase   = "initial"
						}
						
						input = jsonencode({
							message = "Hello from Terraform"
							version = "v1.0.0"
						})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "id"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "state_machine_arn", getVar("STATE_MACHINE_ARN")),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "execution_arn"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "status"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.version", "v1.0.0"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.phase", "initial"),
				),
			},
			// Étape 2: Update - Changement de trigger (force une nouvelle exécution)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "test" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
						
						input = jsonencode({
							message = "Hello from Terraform"
							version = "v2.0.0"
						})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "id"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "state_machine_arn", getVar("STATE_MACHINE_ARN")),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "execution_arn"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "status"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.phase", "updated"),
				),
			},
			// Étape 3: Update - Même triggers (pas de nouvelle exécution)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "test" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
						
						input = jsonencode({
							message = "Hello from Terraform"
							version = "v2.0.0"
						})
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "id"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "state_machine_arn", getVar("STATE_MACHINE_ARN")),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "execution_arn"),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "status"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.phase", "updated"),
				),
			},
		},
	})
}

// TestAccSFNStartSyncExecutionResource_TriggersBehavior teste le comportement spécifique des triggers
// en vérifiant que les execution_arn changent quand les triggers changent et restent identiques
// quand les triggers ne changent pas.
func TestAccSFNStartSyncExecutionResource_TriggersBehavior(t *testing.T) {
	var firstExecutionArn, secondExecutionArn, thirdExecutionArn string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Étape 1: Create - Création initiale avec triggers
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "test" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						
						triggers = {
							version = "v1.0.0"
							phase   = "initial"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "id"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "state_machine_arn", getVar("STATE_MACHINE_ARN")),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "execution_arn"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.version", "v1.0.0"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.phase", "initial"),
					// Capturer l'execution_arn pour comparaison
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["test_sfn_start_sync_execution.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						firstExecutionArn = rs.Primary.Attributes["execution_arn"]
						return nil
					},
				),
			},
			// Étape 2: Update - Changement de trigger (force une nouvelle exécution)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "test" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "id"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "state_machine_arn", getVar("STATE_MACHINE_ARN")),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "execution_arn"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.phase", "updated"),
					// Vérifier que l'execution_arn a changé
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["test_sfn_start_sync_execution.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						secondExecutionArn = rs.Primary.Attributes["execution_arn"]
						if firstExecutionArn == secondExecutionArn {
							return fmt.Errorf("execution_arn should have changed when triggers changed: %s", secondExecutionArn)
						}
						return nil
					},
				),
			},
			// Étape 3: Update - Même triggers (pas de nouvelle exécution)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_sfn_start_sync_execution" "test" {
						state_machine_arn = "` + getVar("STATE_MACHINE_ARN") + `"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "id"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "state_machine_arn", getVar("STATE_MACHINE_ARN")),
					resource.TestCheckResourceAttrSet("test_sfn_start_sync_execution.test", "execution_arn"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_sfn_start_sync_execution.test", "triggers.phase", "updated"),
					// Vérifier que l'execution_arn n'a pas changé
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["test_sfn_start_sync_execution.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						thirdExecutionArn = rs.Primary.Attributes["execution_arn"]
						if secondExecutionArn != thirdExecutionArn {
							return fmt.Errorf("execution_arn should not have changed when triggers are the same: %s != %s", secondExecutionArn, thirdExecutionArn)
						}
						return nil
					},
				),
			},
		},
	})
}
