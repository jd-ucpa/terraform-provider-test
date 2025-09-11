package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccSSMSendFilesResource_Basic teste l'envoi de fichiers SSM basique en utilisant instance_ids.
// Ce test configure le provider avec assume_role, crée une ressource SSM Send Files avec une liste
// d'instance_ids spécifique, puis vérifie que les fichiers sont créés avec succès. Le test valide
// que les attributs de base sont correctement définis : id, command_id, platform, et status.
func TestAccSSMSendFilesResource_Basic(t *testing.T) {
	
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
            profile = "` + getVar("AWS_PROFILE") + `"
						assume_role {
							role_arn = "` + getVar("ROLE_ARN") + `"
						}
					}
					
					resource "test_ssm_send_files" "test" {
						platform = "linux"
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						working_directory = "/tmp"
						
						script_before_files = "pwd && ls -lA"
						script_after_files = "ls -lA && cat file1.txt && cat file2.txt"
						
						file {
							name = "file1.txt"
							content = "Hello from Terraform SSM Send Files!"
							permissions = "644"
							owner = "root"
							group = "root"
						}
					
						file {
							name = "file2.txt"
							content = "Hello from provider-test!"
							permissions = "755"
							owner = "ec2-user"
							group = "ec2-user"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "platform", "linux"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "working_directory", "/tmp"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "status", "Success"),
				),
			},
		},
	})
}

// TestAccSSMSendFilesResource_Targets teste l'envoi de fichiers SSM en utilisant le bloc targets
// au lieu d'instance_ids. Ce test configure le provider avec assume_role, crée une ressource SSM
// Send Files avec un bloc targets spécifiant InstanceIds, puis vérifie que les fichiers sont créés
// avec succès. Ce test valide que la fonctionnalité de ciblage par targets fonctionne correctement.
func TestAccSSMSendFilesResource_Targets(t *testing.T) {
	
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
            profile = "` + getVar("AWS_PROFILE") + `"
						assume_role {
							role_arn = "` + getVar("ROLE_ARN") + `"
						}
					}
					
					resource "test_ssm_send_files" "test" {
						platform = "linux"
						working_directory = "/tmp"
						
						script_before_files = ""
						script_after_files = ""
						
						targets {
							key    = "InstanceIds"
							values = ["` + getVar("INSTANCE_ID") + `"]
						}
						
						file {
							name = "test_file.txt"
							content = "Hello from Terraform SSM Send Files with targets!"
							permissions = "644"
							owner = "root"
							group = "root"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "platform", "linux"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "working_directory", "/tmp"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "status", "Success"),
				),
			},
		},
	})
}

// TestAccSSMSendFilesResource_WithScripts teste l'envoi de fichiers SSM avec des scripts avant et après.
// Ce test vérifie que les scripts_before_files et script_after_files sont exécutés correctement
// et que les fichiers sont créés avec les bonnes permissions et propriétaires.
func TestAccSSMSendFilesResource_WithScripts(t *testing.T) {
	
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
            profile = "` + getVar("AWS_PROFILE") + `"
						assume_role {
							role_arn = "` + getVar("ROLE_ARN") + `"
						}
					}
					
					resource "test_ssm_send_files" "test" {
						platform = "linux"
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						working_directory = "/tmp"
						
						script_before_files = "pwd && ls -lA"
						script_after_files = "ls -lA && cat file1.txt && cat file2.txt"
						
						file {
							name = "file1.txt"
							content = "Hello from Terraform!"
							permissions = "644"
							owner = "root"
							group = "root"
						}
						
						file {
							name = "file2.txt"
							content = "Hello from provider-test!"
							permissions = "644"
							owner = "root"
							group = "root"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "platform", "linux"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "working_directory", "/tmp"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "script_before_files", "pwd && ls -lA"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "script_after_files", "ls -lA && cat file1.txt && cat file2.txt"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "status", "Success"),
				),
			},
		},
	})
}

// TestAccSSMSendFilesResource_DefaultProfile teste l'envoi de fichiers SSM en utilisant
// le profil AWS_PROFILE_OTHER (sans assume_role). Ce test configure le provider avec l'attribut profile,
// utilise le profil AWS_PROFILE_OTHER=3098, crée une ressource SSM Send Files avec l'instance
// INSTANCE_ID, puis vérifie que les fichiers sont créés avec succès.
func TestAccSSMSendFilesResource_DefaultProfile(t *testing.T) {
	
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}
					
					resource "test_ssm_send_files" "test_default" {
						platform = "linux"
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						working_directory = "/tmp"
						
						script_before_files = ""
						script_after_files = "cat test_file.txt; ls -lA test_file.txt"
						
						file {
							name = "test_file.txt"
							content = "Hello from Terraform SSM Send Files with AWS_PROFILE_OTHER profile!"
							permissions = "644"
							owner = "   ec2-user   "
							group = "   root   "
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test_default", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test_default", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test_default", "platform", "linux"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test_default", "working_directory", "/tmp"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test_default", "status", "Success"),
				),
			},
		},
	})
}

func TestAccSSMSendFilesResource_Windows(t *testing.T) {

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
            profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}
					
					resource "test_ssm_send_files" "test_windows" {
						platform = "windows"
						instance_ids = ["` + getVar("INSTANCE_ID_WIN") + `"]
						working_directory = "C:/Users/Default/Documents"
						
						script_before_files = "Get-ChildItem -Force | Format-Table Mode, Length, LastWriteTime, Name"
						script_after_files = "Get-ChildItem -Force | Format-Table Mode, Length, LastWriteTime, Name; cat test_file.txt"
						
						file {
							name = "test_file.txt"
							content = "Hello from Terraform SSM Send Files on Windows!"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test_windows", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test_windows", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test_windows", "platform", "windows"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test_windows", "working_directory", "C:/Users/Default/Documents"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test_windows", "status", "Success"),
				),
			},
		},
	})
}

// TestAccSSMSendFilesResource_Lifecycle teste le cycle de vie complet d'une ressource SSM Send Files.
// Ce test vérifie les trois phases principales : Create (création initiale), Update (mise à jour avec
// triggers modifiés), et Delete (suppression propre). Il utilise le mécanisme de triggers pour forcer
// la re-exécution de la commande lors de l'update et vérifie que chaque étape fonctionne correctement.
func TestAccSSMSendFilesResource_Lifecycle(t *testing.T) {
	
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Étape 1: Create - Création initiale de la ressource
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
            profile = "` + getVar("AWS_PROFILE") + `"
						assume_role {
							role_arn = "` + getVar("ROLE_ARN") + `"
						}
					}
					
					resource "test_ssm_send_files" "lifecycle_test" {
						platform = "linux"
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						working_directory = "/tmp"
						
						script_before_files = "echo 'Lifecycle test - Create phase' && pwd && date"
						script_after_files = "echo 'Files created successfully' && ls -la test_lifecycle.txt && cat test_lifecycle.txt"
						
						file {
							name = "test_lifecycle.txt"
							content = "Hello from Terraform SSM Send Files - Create phase!"
							permissions = "644"
							owner = "ec2-user"
							group = "ec2-user"
						}
						
						triggers = {
							phase = "create"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.lifecycle_test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.lifecycle_test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_test", "platform", "linux"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_test", "working_directory", "/tmp"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_test", "triggers.phase", "create"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_test", "status", "Success"),
				),
			},
			// Étape 2: Update - Mise à jour de la ressource (changement des triggers et contenu)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
            profile = "` + getVar("AWS_PROFILE") + `"
						assume_role {
							role_arn = "` + getVar("ROLE_ARN") + `"
						}
					}
					
					resource "test_ssm_send_files" "lifecycle_test" {
						platform = "linux"
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						working_directory = "/tmp"
						
						script_before_files = "echo 'Lifecycle test - Update phase' && pwd && date && echo 'Starting file update...' && echo 'Update phase' >> test_lifecycle.txt"
						script_after_files = "echo 'Files updated successfully' && ls -la test_lifecycle_update.txt && cat test_lifecycle.txt && cat test_lifecycle_update.txt && echo 'Update phase completed'"
						
						file {
							name = "test_lifecycle_update.txt"
							content = "Hello from Terraform SSM Send Files - Update phase! File content has been modified."
							permissions = "755"
							owner = "ec2-user"
							group = "root"
						}
						
						triggers = {
							phase = "update"
							timestamp = "updated"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.lifecycle_test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.lifecycle_test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_test", "platform", "linux"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_test", "working_directory", "/tmp"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_test", "triggers.phase", "update"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_test", "triggers.timestamp", "updated"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_test", "status", "Success"),
				),
			},
			// Étape 3: Delete - La suppression est automatiquement testée par le framework Terraform
			// Pas besoin d'étape explicite, le framework vérifie que Delete() ne génère pas d'erreur
		},
	})
}

// TestAccSSMSendFilesResource_LifecycleWindows teste le cycle de vie complet d'une ressource SSM Send Files sur Windows.
// Ce test vérifie les trois phases principales : Create (création initiale), Update (mise à jour avec
// triggers modifiés), et Delete (suppression propre). Il utilise le mécanisme de triggers pour forcer
// la re-exécution de la commande lors de l'update et vérifie que chaque étape fonctionne correctement.
func TestAccSSMSendFilesResource_LifecycleWindows(t *testing.T) {
	
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Étape 1: Create - Création initiale de la ressource
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}
					
					resource "test_ssm_send_files" "lifecycle_windows_test" {
						platform = "windows"
						instance_ids = ["` + getVar("INSTANCE_ID_WIN") + `"]
						working_directory = "C:/Users/Default/Documents"
						
						script_before_files = "Write-Host 'Lifecycle test - Create phase on Windows' -ForegroundColor Green; Get-Date; Get-Location"
						script_after_files = "Write-Host 'Files created successfully on Windows' -ForegroundColor Green; Get-ChildItem test_lifecycle_windows.txt | Format-List; Get-Content test_lifecycle_windows.txt"
						
						file {
							name = "test_lifecycle_windows.txt"
							content = "Create phase on Windows!"
						}
						
						triggers = {
							phase = "create"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.lifecycle_windows_test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.lifecycle_windows_test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_windows_test", "platform", "windows"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_windows_test", "working_directory", "C:/Users/Default/Documents"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_windows_test", "triggers.phase", "create"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_windows_test", "status", "Success"),
				),
			},
			// Étape 2: Update - Mise à jour de la ressource (changement des triggers et contenu)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}
					
					resource "test_ssm_send_files" "lifecycle_windows_test" {
						platform = "windows"
						instance_ids = ["` + getVar("INSTANCE_ID_WIN") + `"]
						working_directory = "C:/Users/Default/Documents"
						
						script_before_files = "Write-Host 'Lifecycle test - Update phase on Windows' -ForegroundColor Yellow; Get-Date; Write-Host 'Starting file update on Windows...' -ForegroundColor Cyan; Add-Content test_lifecycle_windows.txt 'Update phase on Windows'"
						script_after_files = "Write-Host 'Files updated successfully on Windows' -ForegroundColor Green; Get-ChildItem test_lifecycle_windows_update.txt | Format-List; Get-Content test_lifecycle_windows.txt; Get-Content test_lifecycle_windows_update.txt; Write-Host 'Update phase completed on Windows' -ForegroundColor Green"
						
						file {
							name = "test_lifecycle_windows_update.txt"
							content = "Update phase on Windows!"
						}
						
						triggers = {
							phase = "update"
							timestamp = "updated_windows"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.lifecycle_windows_test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.lifecycle_windows_test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_windows_test", "platform", "windows"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_windows_test", "working_directory", "C:/Users/Default/Documents"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_windows_test", "triggers.phase", "update"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_windows_test", "triggers.timestamp", "updated_windows"),
					resource.TestCheckResourceAttr("test_ssm_send_files.lifecycle_windows_test", "status", "Success"),
				),
			},
			// Étape 3: Delete - La suppression est automatiquement testée par le framework Terraform
			// Pas besoin d'étape explicite, le framework vérifie que Delete() ne génère pas d'erreur
		},
	})
}

// TestAccSSMSendFilesResource_WithTriggers teste le fonctionnement des triggers
// pour forcer un nouvel envoi de fichiers SSM. Ce test vérifie que :
// 1. La ressource se crée correctement avec des triggers
// 2. Un changement de trigger force un nouvel envoi
// 3. Les valeurs calculées sont préservées quand les triggers ne changent pas
func TestAccSSMSendFilesResource_WithTriggers(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Étape 1: Create - Création initiale avec triggers
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}
					
					resource "test_ssm_send_files" "test" {
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						platform = "linux"
						working_directory = "/tmp"
						
						triggers = {
							version = "v1.0.0"
							phase   = "initial"
						}
						
						file {
							name = "test_file_v1.txt"
							content = "Hello from Terraform SSM Send Files v1.0.0!"
							permissions = "644"
							owner = "ec2-user"
							group = "root"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "instance_ids.0", getVar("INSTANCE_ID")),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "command_id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "status"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.version", "v1.0.0"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.phase", "initial"),
				),
			},
			// Étape 2: Update - Changement de trigger (force un nouvel envoi)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}
					
					resource "test_ssm_send_files" "test" {
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						platform = "linux"
						working_directory = "/tmp"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
						
						file {
							name = "test_file_v2.txt"
							content = "Hello from Terraform SSM Send Files v2.0.0!"
							permissions = "644"
							owner = "ec2-user"
							group = "root"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "instance_ids.0", getVar("INSTANCE_ID")),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "command_id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "status"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.phase", "updated"),
				),
			},
			// Étape 3: Update - Même triggers (pas de nouvel envoi)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}
					
					resource "test_ssm_send_files" "test" {
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						platform = "linux"
						working_directory = "/tmp"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
						
						file {
							name = "test_file_v2.txt"
							content = "Hello from Terraform SSM Send Files v2.0.0!"
							permissions = "644"
							owner = "ec2-user"
							group = "root"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "instance_ids.0", getVar("INSTANCE_ID")),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "command_id"),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "status"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.phase", "updated"),
				),
			},
		},
	})
}

// TestAccSSMSendFilesResource_TriggersBehavior teste le comportement spécifique des triggers
// en vérifiant que les command_id changent quand les triggers changent et restent identiques
// quand les triggers ne changent pas.
func TestAccSSMSendFilesResource_TriggersBehavior(t *testing.T) {
	var firstCommandId, secondCommandId, thirdCommandId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Étape 1: Create - Création initiale avec triggers
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}
					
					resource "test_ssm_send_files" "test" {
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						platform = "linux"
						working_directory = "/tmp"
						
						triggers = {
							version = "v1.0.0"
							phase   = "initial"
						}
						
						file {
							name = "test_file.txt"
							content = "Hello from Terraform SSM Send Files!"
							permissions = "644"
							owner = "ec2-user"
							group = "root"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "instance_ids.0", getVar("INSTANCE_ID")),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.version", "v1.0.0"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.phase", "initial"),
					// Capturer le command_id pour comparaison
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["test_ssm_send_files.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						firstCommandId = rs.Primary.Attributes["command_id"]
						return nil
					},
				),
			},
			// Étape 2: Update - Changement de trigger (force un nouvel envoi)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}
					
					resource "test_ssm_send_files" "test" {
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						platform = "linux"
						working_directory = "/tmp"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
						
						file {
							name = "test_file.txt"
							content = "Hello from Terraform SSM Send Files!"
							permissions = "644"
							owner = "ec2-user"
							group = "root"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "instance_ids.0", getVar("INSTANCE_ID")),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.phase", "updated"),
					// Vérifier que le command_id a changé
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["test_ssm_send_files.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						secondCommandId = rs.Primary.Attributes["command_id"]
						if firstCommandId == secondCommandId {
							return fmt.Errorf("command_id should have changed when triggers changed: %s", secondCommandId)
						}
						return nil
					},
				),
			},
			// Étape 3: Update - Même triggers (pas de nouvel envoi)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER") + `"
					}
					
					resource "test_ssm_send_files" "test" {
						instance_ids = ["` + getVar("INSTANCE_ID") + `"]
						platform = "linux"
						working_directory = "/tmp"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
						
						file {
							name = "test_file.txt"
							content = "Hello from Terraform SSM Send Files!"
							permissions = "644"
							owner = "ec2-user"
							group = "root"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "instance_ids.0", getVar("INSTANCE_ID")),
					resource.TestCheckResourceAttrSet("test_ssm_send_files.test", "command_id"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_ssm_send_files.test", "triggers.phase", "updated"),
					// Vérifier que le command_id n'a pas changé
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["test_ssm_send_files.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						thirdCommandId = rs.Primary.Attributes["command_id"]
						if secondCommandId != thirdCommandId {
							return fmt.Errorf("command_id should not have changed when triggers are the same: %s != %s", secondCommandId, thirdCommandId)
						}
						return nil
					},
				),
			},
		},
	})
}
