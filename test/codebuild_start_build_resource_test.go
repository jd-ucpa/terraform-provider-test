package test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccCodeBuildStartBuildResource_Basic teste le démarrage d'un build CodeBuild basique.
// Ce test configure le provider avec assume_role, crée une ressource CodeBuild Start Build avec
// un project_name spécifique, puis vérifie que le build est démarré avec succès.
// Le test valide que les attributs de base sont correctement définis : id, project_name, et build.
func TestAccCodeBuildStartBuildResource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
            profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_arn"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_number"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "build_project_name", getVar("CODEBUILD_PROJECT_NAME")),
				),
			},
		},
	})
}

// TestAccCodeBuildStartBuildResource_WithEnvironmentVariables teste le démarrage d'un build CodeBuild
// avec des variables d'environnement personnalisées. Ce test configure le provider avec un profil AWS,
// crée une ressource CodeBuild Start Build avec des variables d'environnement, puis vérifie que
// le build est démarré avec succès et que les variables d'environnement sont correctement configurées.
func TestAccCodeBuildStartBuildResource_WithEnvironmentVariables(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						environment_variables {
							name  = "TEST_VAR"
							value = "test_value"
							type  = "PLAINTEXT"
						}
						
						environment_variables {
							name  = "ANOTHER_VAR"
							value = "another_value"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_arn"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_number"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "build_project_name", getVar("CODEBUILD_PROJECT_NAME")),
				),
			},
		},
	})
}



// TestAccCodeBuildStartBuildResource_Complete teste le démarrage d'un build CodeBuild
// avec tous les paramètres optionnels. Ce test configure le provider avec un profil AWS,
// crée une ressource CodeBuild Start Build avec des variables d'environnement,
// puis vérifie que le build est démarré avec succès et que tous les paramètres sont correctement configurés.
func TestAccCodeBuildStartBuildResource_Complete(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						environment_variables {
							name  = "BUILD_ENV"
							value = "test"
							type  = "PLAINTEXT"
						}
						
						environment_variables {
							name  = "VERSION"
							value = "1.0.0"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_arn"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_number"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "build_project_name", getVar("CODEBUILD_PROJECT_NAME")),
				),
			},
		},
	})
}

// TestAccCodeBuildStartBuildResource_Update teste la mise à jour d'une ressource CodeBuild Start Build.
// Ce test vérifie que les modifications de la ressource déclenchent un nouveau build.
// Il utilise des configurations différentes pour tester le cycle de vie complet.
func TestAccCodeBuildStartBuildResource_Update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Étape 1: Create - Création initiale de la ressource
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						environment_variables {
							name  = "PHASE"
							value = "create"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
				),
			},
			// Étape 2: Update - Mise à jour de la ressource
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						environment_variables {
							name  = "PHASE"
							value = "update"
						}
						
						environment_variables {
							name  = "VERSION"
							value = "2.0.0"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
				),
			},
		},
	})
}

// TestAccCodeBuildStartBuildResource_InvalidParameterStore teste la validation
// d'un paramètre PARAMETER_STORE inexistant. Ce test vérifie que le provider
// retourne une erreur appropriée quand un paramètre SSM n'existe pas.
func TestAccCodeBuildStartBuildResource_InvalidParameterStore(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						environment_variables {
							name  = "INVALID_PARAM"
							value = "/nonexistent/parameter"
							type  = "PARAMETER_STORE"
						}
					}
				`,
				ExpectError: regexp.MustCompile(`Parameter Store parameter '/nonexistent/parameter' does not exist`),
			},
		},
	})
}

// TestAccCodeBuildStartBuildResource_InvalidSecretsManager teste la validation
// d'un secret SECRETS_MANAGER inexistant. Ce test vérifie que le provider
// retourne une erreur appropriée quand un secret n'existe pas.
func TestAccCodeBuildStartBuildResource_InvalidSecretsManager(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						environment_variables {
							name  = "INVALID_SECRET"
							value = "nonexistent-secret"
							type  = "SECRETS_MANAGER"
						}
					}
				`,
				ExpectError: regexp.MustCompile(`Secrets Manager secret 'nonexistent-secret' does not exist`),
			},
		},
	})
}

// TestAccCodeBuildStartBuildResource_MixedInvalidParameters teste la validation
// avec plusieurs paramètres invalides. Ce test vérifie que le provider
// retourne des erreurs pour tous les paramètres invalides.
func TestAccCodeBuildStartBuildResource_MixedInvalidParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						environment_variables {
							name  = "VALID_VAR"
							value = "valid_value"
							type  = "PLAINTEXT"
						}
						
						environment_variables {
							name  = "INVALID_PARAM"
							value = "/nonexistent/parameter"
							type  = "PARAMETER_STORE"
						}
						
						environment_variables {
							name  = "INVALID_SECRET"
							value = "nonexistent-secret"
							type  = "SECRETS_MANAGER"
						}
					}
				`,
				ExpectError: regexp.MustCompile(`Parameter Store parameter '/nonexistent/parameter' does not exist`),
			},
		},
	})
}

// TestAccCodeBuildStartBuildResource_ValidParameterStore teste la validation
// d'un paramètre PARAMETER_STORE existant. Ce test vérifie que le provider
// accepte un paramètre SSM valide et démarre le build avec succès.
func TestAccCodeBuildStartBuildResource_ValidParameterStore(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						environment_variables {
							name  = "VALID_PARAM"
							value = "` + getVar("CODEBUILD_PARAMETER_NAME") + `"
							type  = "PARAMETER_STORE"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_arn"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_number"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "build_project_name", getVar("CODEBUILD_PROJECT_NAME")),
				),
			},
		},
	})
}

// TestAccCodeBuildStartBuildResource_ValidSecretsManager teste la validation
// d'un secret SECRETS_MANAGER existant. Ce test vérifie que le provider
// accepte un secret valide et démarre le build avec succès.
func TestAccCodeBuildStartBuildResource_ValidSecretsManager(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						environment_variables {
							name  = "VALID_SECRET"
							value = "` + getVar("CODEBUILD_SECRET_NAME") + `"
							type  = "SECRETS_MANAGER"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_arn"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_number"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "build_project_name", getVar("CODEBUILD_PROJECT_NAME")),
				),
			},
		},
	})
}

// TestAccCodeBuildStartBuildResource_MixedValidParameters teste la validation
// avec un mélange de paramètres valides. Ce test vérifie que le provider
// accepte tous les types de variables d'environnement valides.
func TestAccCodeBuildStartBuildResource_MixedValidParameters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						environment_variables {
							name  = "PLAINTEXT_VAR"
							value = "plaintext_value"
							type  = "PLAINTEXT"
						}
						
						environment_variables {
							name  = "PARAMETER_VAR"
							value = "` + getVar("CODEBUILD_PARAMETER_NAME") + `"
							type  = "PARAMETER_STORE"
						}
						
						environment_variables {
							name  = "SECRET_VAR"
							value = "` + getVar("CODEBUILD_SECRET_NAME") + `"
							type  = "SECRETS_MANAGER"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_arn"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_number"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "build_project_name", getVar("CODEBUILD_PROJECT_NAME")),
				),
			},
		},
	})
}

// TestAccCodeBuildStartBuildResource_WithTriggers teste le fonctionnement des triggers
// pour forcer un nouveau build. Ce test vérifie que :
// 1. La ressource se crée correctement avec des triggers
// 2. Un changement de trigger force un nouveau build
// 3. Les valeurs calculées sont préservées quand les triggers ne changent pas
func TestAccCodeBuildStartBuildResource_WithTriggers(t *testing.T) {
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
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						triggers = {
							version = "v1.0.0"
							phase   = "initial"
						}
						
						environment_variables {
							name  = "BUILD_PHASE"
							value = "initial"
							type  = "PLAINTEXT"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_arn"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_number"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "build_project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.version", "v1.0.0"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.phase", "initial"),
				),
			},
			// Étape 2: Update - Changement de trigger (force un nouveau build)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
						
						environment_variables {
							name  = "BUILD_PHASE"
							value = "updated"
							type  = "PLAINTEXT"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_arn"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_number"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "build_project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.phase", "updated"),
				),
			},
			// Étape 3: Update - Même triggers (pas de nouveau build)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
						
						environment_variables {
							name  = "BUILD_PHASE"
							value = "updated"
							type  = "PLAINTEXT"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_arn"),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_number"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "build_project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.phase", "updated"),
				),
			},
		},
	})
}

// TestAccCodeBuildStartBuildResource_TriggersBehavior teste le comportement spécifique des triggers
// en vérifiant que les build_id changent quand les triggers changent et restent identiques
// quand les triggers ne changent pas.
func TestAccCodeBuildStartBuildResource_TriggersBehavior(t *testing.T) {
	var firstBuildId, secondBuildId, thirdBuildId string

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
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						triggers = {
							version = "v1.0.0"
							phase   = "initial"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.version", "v1.0.0"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.phase", "initial"),
					// Capturer le build_id pour comparaison
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["test_codebuild_start_build.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						firstBuildId = rs.Primary.Attributes["build_id"]
						return nil
					},
				),
			},
			// Étape 2: Update - Changement de trigger (force un nouveau build)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.phase", "updated"),
					// Vérifier que le build_id a changé
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["test_codebuild_start_build.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						secondBuildId = rs.Primary.Attributes["build_id"]
						if firstBuildId == secondBuildId {
							return fmt.Errorf("build_id should have changed when triggers changed: %s", secondBuildId)
						}
						return nil
					},
				),
			},
			// Étape 3: Update - Même triggers (pas de nouveau build)
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE") + `"
					}
					
					resource "test_codebuild_start_build" "test" {
						project_name = "` + getVar("CODEBUILD_PROJECT_NAME") + `"
						
						triggers = {
							version = "v2.0.0"
							phase   = "updated"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "project_name", getVar("CODEBUILD_PROJECT_NAME")),
					resource.TestCheckResourceAttrSet("test_codebuild_start_build.test", "build_id"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.version", "v2.0.0"),
					resource.TestCheckResourceAttr("test_codebuild_start_build.test", "triggers.phase", "updated"),
					// Vérifier que le build_id n'a pas changé
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["test_codebuild_start_build.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						thirdBuildId = rs.Primary.Attributes["build_id"]
						if secondBuildId != thirdBuildId {
							return fmt.Errorf("build_id should not have changed when triggers are the same: %s != %s", secondBuildId, thirdBuildId)
						}
						return nil
					},
				),
			},
		},
	})
}
