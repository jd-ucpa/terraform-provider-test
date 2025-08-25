package test

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSSMSendCommandResource_Basic teste l'envoi d'une commande SSM basique en utilisant instance_ids.
// Ce test configure le provider avec assume_role, crée une ressource SSM Send Command avec une liste
// d'instance_ids spécifique, puis vérifie que la commande est exécutée avec succès. Le test valide
// que les attributs de base sont correctement définis : id, command_id, document_name, et comment.
func TestAccSSMSendCommandResource_Basic(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "ROLE_ARN", "INSTANCE_ID")
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSMSendCommandResourceConfig_Basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "document_name", "AWS-RunShellScript"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "comment", "Test SSM command from Terraform provider test"),
				),
			},
		},
	})
}

// TestAccSSMSendCommandResource_Targets teste l'envoi d'une commande SSM en utilisant le bloc targets
// au lieu d'instance_ids. Ce test configure le provider avec assume_role, crée une ressource SSM
// Send Command avec un bloc targets spécifiant InstanceIds, puis vérifie que la commande est exécutée
// avec succès. Ce test valide que la fonctionnalité de ciblage par targets fonctionne correctement.
func TestAccSSMSendCommandResource_Targets(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "ROLE_ARN", "INSTANCE_ID")
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSMSendCommandResourceConfig_Targets(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "document_name", "AWS-RunShellScript"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "comment", "Test SSM command with targets block"),
				),
			},
		},
	})
}

// TestAccSSMSendCommandResource_TargetsByTag teste l'envoi d'une commande SSM en utilisant le ciblage par tag EC2.
// Ce test configure le provider avec assume_role, crée une ressource SSM Send Command avec un bloc targets
// spécifiant "tag:Name" et la valeur du tag depuis EC2_TAG_NAME, puis vérifie que la commande est exécutée
// avec succès. Ce test valide que le ciblage par tags EC2 fonctionne correctement.
func TestAccSSMSendCommandResource_TargetsByTag(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "ROLE_ARN", "EC2_TAG_NAME")
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSMSendCommandResourceConfig_TargetsByTag(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "document_name", "AWS-RunShellScript"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "comment", "Test SSM command with EC2 tag targeting"),
				),
			},
		},
	})
}

// TestAccSSMSendCommandResource_StatusSuccess teste que le statut de la commande SSM reflète correctement
// le succès de l'exécution. Ce test exécute une commande valide "pwd" et vérifie que le statut
// retourné est "Success". Cela confirme que le provider gère correctement les commandes qui s'exécutent avec succès.
func TestAccSSMSendCommandResource_StatusSuccess(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "ROLE_ARN", "INSTANCE_ID")
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSMSendCommandResourceConfig_StatusSuccess(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "document_name", "AWS-RunShellScript"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "status", "Success"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "comment", "Test SSM command status success"),
				),
			},
		},
	})
}

// TestAccSSMSendCommandResource_StatusFailed teste que le statut de la commande SSM reflète correctement
// l'échec de l'exécution. Ce test exécute une commande invalide "pwdpwdpwd" et vérifie que le statut
// retourné est "Failed". Cela confirme que le provider gère correctement les commandes qui échouent.
func TestAccSSMSendCommandResource_StatusFailed(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "ROLE_ARN", "INSTANCE_ID")
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSMSendCommandResourceConfig_StatusFailed(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "document_name", "AWS-RunShellScript"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "status", "Failed"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test", "comment", "Test SSM command status failed"),
				),
			},
		},
	})
}

// TestAccSSMSendCommandResource_Lifecycle teste le cycle de vie complet d'une ressource SSM Send Command.
// Ce test vérifie les trois phases principales : Create (création initiale), Update (mise à jour avec
// triggers modifiés), et Delete (suppression propre). Il utilise le mécanisme de triggers pour forcer
// la re-exécution de la commande lors de l'update et vérifie que chaque étape fonctionne correctement.
func TestAccSSMSendCommandResource_Lifecycle(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "ROLE_ARN", "INSTANCE_ID")
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Étape 1: Create - Création initiale de la ressource
			{
				Config: testAccSSMSendCommandResourceConfig_Lifecycle_Create(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_command.lifecycle_test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_command.lifecycle_test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_command.lifecycle_test", "document_name", "AWS-RunShellScript"),
					resource.TestCheckResourceAttr("test_ssm_send_command.lifecycle_test", "comment", "Lifecycle test - Create phase"),
					resource.TestCheckResourceAttr("test_ssm_send_command.lifecycle_test", "triggers.phase", "create"),
					resource.TestCheckResourceAttr("test_ssm_send_command.lifecycle_test", "status", "Success"),
				),
			},
			// Étape 2: Update - Mise à jour de la ressource (changement des triggers)
			{
				Config: testAccSSMSendCommandResourceConfig_Lifecycle_Update(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_command.lifecycle_test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_command.lifecycle_test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_command.lifecycle_test", "document_name", "AWS-RunShellScript"),
					resource.TestCheckResourceAttr("test_ssm_send_command.lifecycle_test", "comment", "Lifecycle test - Update phase"),
					resource.TestCheckResourceAttr("test_ssm_send_command.lifecycle_test", "triggers.phase", "update"),
					resource.TestCheckResourceAttr("test_ssm_send_command.lifecycle_test", "triggers.timestamp", "updated"),
					resource.TestCheckResourceAttr("test_ssm_send_command.lifecycle_test", "status", "Success"),
				),
			},
			// Étape 3: Delete - La suppression est automatiquement testée par le framework Terraform
			// Pas besoin d'étape explicite, le framework vérifie que Delete() ne génère pas d'erreur
		},
	})
}

// TestAccSSMSendCommandResource_Validation teste les validations du provider SSM Send Command.
// Ce test vérifie deux cas d'erreur : 1) Quand instance_ids ET targets sont spécifiés simultanément
// (mutual exclusion), et 2) Quand aucun des deux n'est spécifié (requis). Le test confirme que
// le provider génère les messages d'erreur appropriés pour ces cas de validation.
func TestAccSSMSendCommandResource_Validation(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "ROLE_ARN", "INSTANCE_ID")
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccSSMSendCommandResourceConfig_BothInstanceIdsAndTargets(),
				ExpectError: regexp.MustCompile(`Cannot specify both instance_ids and targets`),
			},
			{
				Config:      testAccSSMSendCommandResourceConfig_NoInstanceIdsOrTargets(),
				ExpectError: regexp.MustCompile(`Either instance_ids or targets must be specified`),
			},
		},
	})
}

// TestAccSSMSendCommandResource_DefaultProfile teste l'envoi d'une commande SSM basique en utilisant
// le profil AWS_PROFILE_OTHER (sans assume_role). Ce test configure le provider avec l'attribut profile,
// utilise le profil AWS_PROFILE_OTHER=3098, crée une ressource SSM Send Command avec l'instance
// INSTANCE_ID, puis vérifie que la commande est exécutée avec succès.
// Le test valide que les attributs de base sont correctement définis : id, command_id, document_name, et comment.
func TestAccSSMSendCommandResource_DefaultProfile(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "AWS_PROFILE_OTHER", "INSTANCE_ID")
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSSMSendCommandResourceConfig_DefaultProfile(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test_default", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_command.test_default", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test_default", "document_name", "AWS-RunShellScript"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test_default", "comment", "Test SSM command with AWS_PROFILE_OTHER profile"),
					resource.TestCheckResourceAttr("test_ssm_send_command.test_default", "status", "Success"),
				),
			},
		},
	})
}

func testAccSSMSendCommandResourceConfig_DefaultProfile() string {
	return `
		provider "test" {
			region = "eu-west-1"
			profile = "` + os.Getenv("AWS_PROFILE_OTHER") + `"
		}
		
		resource "test_ssm_send_command" "test_default" {
			document_name = "AWS-RunShellScript"
			instance_ids  = ["` + os.Getenv("INSTANCE_ID") + `"]
			
			parameters = {
				"commands" = "echo 'Test successful with AWS_PROFILE_OTHER profile' && pwd"
			}
			
			comment = "Test SSM command with AWS_PROFILE_OTHER profile"
		}
	`
}

func testAccSSMSendCommandResourceConfig_Basic() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "` + os.Getenv("ROLE_ARN") + `"
			}
		}
		
		resource "test_ssm_send_command" "test" {
			document_name = "AWS-RunShellScript"
			instance_ids  = ["` + os.Getenv("INSTANCE_ID") + `"]
			
			parameters = {
				"commands" = "echo 'Test successful' && pwd"
			}
			
			comment = "Test SSM command from Terraform provider test"
		}
	`
}

func testAccSSMSendCommandResourceConfig_Targets() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "` + os.Getenv("ROLE_ARN") + `"
			}
		}
		
		resource "test_ssm_send_command" "test" {
			document_name = "AWS-RunShellScript"
			
			targets {
				key    = "InstanceIds"
				values = ["` + os.Getenv("INSTANCE_ID") + `"]
			}
			
			parameters = {
				"commands" = "echo 'Test with targets' && pwd"
			}
			
			comment = "Test SSM command with targets block"
		}
	`
}

func testAccSSMSendCommandResourceConfig_TargetsByTag() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "` + os.Getenv("ROLE_ARN") + `"
			}
		}
		
		resource "test_ssm_send_command" "test" {
			document_name = "AWS-RunShellScript"
			
			targets {
				key    = "tag:Name"
				values = ["` + os.Getenv("EC2_TAG_NAME") + `"]
			}
			
			parameters = {
				"commands" = "echo 'Test with EC2 tag targeting' && pwd"
			}
			
			comment = "Test SSM command with EC2 tag targeting"
		}
	`
}

func testAccSSMSendCommandResourceConfig_StatusSuccess() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "` + os.Getenv("ROLE_ARN") + `"
			}
		}
		
		resource "test_ssm_send_command" "test" {
			document_name = "AWS-RunShellScript"
			instance_ids  = ["` + os.Getenv("INSTANCE_ID") + `"]
			
			parameters = {
				"commands" = "pwd"
			}
			
			comment = "Test SSM command status success"
		}
	`
}

func testAccSSMSendCommandResourceConfig_StatusFailed() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "` + os.Getenv("ROLE_ARN") + `"
			}
		}
		
		resource "test_ssm_send_command" "test" {
			document_name = "AWS-RunShellScript"
			instance_ids  = ["` + os.Getenv("INSTANCE_ID") + `"]
			
			parameters = {
				"commands" = "pwdpwdpwd"
			}
			
			comment = "Test SSM command status failed"
		}
	`
}

func testAccSSMSendCommandResourceConfig_BothInstanceIdsAndTargets() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "` + os.Getenv("ROLE_ARN") + `"
			}
		}
		
		resource "test_ssm_send_command" "test" {
			document_name = "AWS-RunShellScript"
			instance_ids  = ["` + os.Getenv("INSTANCE_ID") + `"]
			
			targets {
				key    = "InstanceIds"
				values = ["` + os.Getenv("INSTANCE_ID") + `"]
			}
			
			parameters = {
				"commands" = "echo 'Test validation'"
			}
		}
	`
}

func testAccSSMSendCommandResourceConfig_NoInstanceIdsOrTargets() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "` + os.Getenv("ROLE_ARN") + `"
			}
		}
		
		resource "test_ssm_send_command" "test" {
			document_name = "AWS-RunShellScript"
			
			parameters = {
				"commands" = "echo 'Test validation'"
			}
		}
	`
}

func testAccSSMSendCommandResourceConfig_Lifecycle_Create() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "` + os.Getenv("ROLE_ARN") + `"
			}
		}
		
		resource "test_ssm_send_command" "lifecycle_test" {
			document_name = "AWS-RunShellScript"
			instance_ids  = ["` + os.Getenv("INSTANCE_ID") + `"]
			
			parameters = {
				"commands" = "echo 'Lifecycle test - Create phase' && pwd && date"
			}
			
			comment = "Lifecycle test - Create phase"
			
			triggers = {
				phase = "create"
			}
		}
	`
}

func testAccSSMSendCommandResourceConfig_Lifecycle_Update() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "` + os.Getenv("ROLE_ARN") + `"
			}
		}
		
		resource "test_ssm_send_command" "lifecycle_test" {
			document_name = "AWS-RunShellScript"
			instance_ids  = ["` + os.Getenv("INSTANCE_ID") + `"]
			
			parameters = {
				"commands" = "echo 'Lifecycle test - Update phase' && pwd && date && echo 'Command updated successfully'"
			}
			
			comment = "Lifecycle test - Update phase"
			
			triggers = {
				phase = "update"
				timestamp = "updated"
			}
		}
	`
}
