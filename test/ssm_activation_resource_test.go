package test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSSMActivationResource_ValidationError teste la gestion d'erreur avec un rôle IAM qui ne satisfait pas les contraintes SSM.
// Ce test configure le provider avec assume_role, tente de créer une ressource SSM Activation avec
// un rôle IAM qui ne satisfait pas le pattern regex de l'API SSM, et vérifie que l'erreur est correctement gérée.
func TestAccSSMActivationResource_ValidationError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "invalid role name with spaces"
						secret_name = "` + getVar("SECRET_NAME") + `"
						description = "Test SSM activation with validation error"
					}
				`,
				ExpectError: regexp.MustCompile(`ValidationException`),
			},
		},
	})
}

// TestAccSSMActivationResource_Basic teste la création d'une activation SSM basique avec un rôle IAM simple.
// Ce test configure le provider avec assume_role, crée une ressource SSM Activation avec
// un rôle IAM simple (sans tirets), puis vérifie que l'activation est créée avec succès.
func TestAccSSMActivationResource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "` + getVar("SECRET_NAME") + `"
						description = "Test SSM activation from Terraform provider"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_code"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "iam_role", getVar("ACTIVATION_ROLE_NAME")),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "description", "Test SSM activation from Terraform provider"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "registration_limit", "1"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "expired", "false"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "managed", "false"),
				),
			},
		},
	})
}

// TestAccSSMActivationResource_WithExpirationDate teste la création d'une activation SSM avec une date d'expiration personnalisée.
// Ce test configure le provider avec assume_role, crée une ressource SSM Activation avec
// un bloc expiration_date spécifiant 10 jours, 8 heures et 6 minutes, puis vérifie que
// l'activation est créée avec succès avec la bonne configuration d'expiration.
func TestAccSSMActivationResource_WithExpirationDate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "` + getVar("SECRET_NAME") + `"
						description = "Test SSM activation with expiration date"
						
						expiration_date {
							days = 10
							hours = 8
							minutes = 6
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_code"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "iam_role", getVar("ACTIVATION_ROLE_NAME")),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "description", "Test SSM activation with expiration date"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "registration_limit", "1"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "expired", "false"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "managed", "false"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "expiration_date.days", "10"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "expiration_date.hours", "8"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "expiration_date.minutes", "6"),
				),
			},
		},
	})
}

// TestAccSSMActivationResource_WithSecretManaged teste la création d'une activation SSM avec un secret géré.
// Ce test configure le provider avec assume_role, crée une ressource SSM Activation avec
// un secret_name et managed=true, puis vérifie que l'activation est créée avec succès
// et que le secret est créé et géré par la ressource.
func TestAccSSMActivationResource_WithSecretManaged(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						description = "Test SSM activation with managed secret"
						secret_name = "test/ssm/activation/managed"
						managed = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_code"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "iam_role", getVar("ACTIVATION_ROLE_NAME")),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "description", "Test SSM activation with managed secret"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "secret_name", "test/ssm/activation/managed"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "managed", "true"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "secret_arn"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "secret_version"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "expired", "false"),
				),
			},
		},
	})
}

// TestAccSSMActivationResource_WithSecretNotManaged teste la création d'une activation SSM avec un secret non géré.
// Ce test configure le provider avec assume_role, crée une ressource SSM Activation avec
// un secret_name et managed=false, puis vérifie que l'activation est créée avec succès
// et que le secret est mis à jour mais pas géré par la ressource.
func TestAccSSMActivationResource_WithSecretNotManaged(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "` + getVar("SECRET_NAME") + `"
						description = "Test SSM activation with unmanaged secret"
						managed = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_code"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "iam_role", getVar("ACTIVATION_ROLE_NAME")),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "description", "Test SSM activation with unmanaged secret"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "secret_name", getVar("SECRET_NAME")),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "managed", "false"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "secret_arn"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "secret_version"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "expired", "false"),
				),
			},
		},
	})
}

// TestAccSSMActivationResource_WithTags teste la création d'une activation SSM avec des tags.
// Ce test configure le provider avec assume_role, crée une ressource SSM Activation avec
// des tags personnalisés, puis vérifie que l'activation est créée avec succès avec les bons tags.
func TestAccSSMActivationResource_WithTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "` + getVar("SECRET_NAME") + `"
						description = "Test SSM activation with tags"
						
						tags = {
							Environment = "test"
							Project     = "terraform-provider-test"
							Owner       = "test-team"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_code"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "iam_role", getVar("ACTIVATION_ROLE_NAME")),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "description", "Test SSM activation with tags"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "registration_limit", "1"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "tags.Environment", "test"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "tags.Project", "terraform-provider-test"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "tags.Owner", "test-team"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "expired", "false"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "managed", "false"),
				),
			},
		},
	})
}

// TestAccSSMActivationResource_WithRegistrationLimit teste la création d'une activation SSM avec une limite d'enregistrement.
// Ce test configure le provider avec assume_role, crée une ressource SSM Activation avec
// un registration_limit spécifique, puis vérifie que l'activation est créée avec succès
// avec la bonne limite d'enregistrement.
func TestAccSSMActivationResource_WithRegistrationLimit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "` + getVar("SECRET_NAME") + `"
						description = "Test SSM activation with registration limit"
						registration_limit = 5
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_id"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "activation_code"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "iam_role", getVar("ACTIVATION_ROLE_NAME")),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "description", "Test SSM activation with registration limit"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "registration_limit", "5"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "expired", "false"),
				),
			},
		},
	})
}

// TestAccSSMActivationResource_UpdateDescription teste la mise à jour de la description d'une activation SSM.
// Ce test crée d'abord une activation SSM avec une description, puis la met à jour avec une nouvelle description.
// Il vérifie que la ressource est recréée correctement avec la nouvelle description.
func TestAccSSMActivationResource_UpdateDescription(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "` + getVar("SECRET_NAME") + `"
						description = "Original description"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "description", "Original description"),
				),
			},
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "` + getVar("SECRET_NAME") + `"
						description = "Updated description"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "description", "Updated description"),
				),
			},
		},
	})
}

// TestAccSSMActivationResource_UpdateSecretName teste la mise à jour du nom du secret d'une activation SSM.
// Ce test crée d'abord une activation SSM avec un secret, puis met à jour le nom du secret.
// Il vérifie que la ressource est mise à jour correctement avec le nouveau nom de secret.
func TestAccSSMActivationResource_UpdateSecretName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "test/ssm/activation/update1"
						managed = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "secret_name", "test/ssm/activation/update1"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "managed", "true"),
				),
			},
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "test/ssm/activation/update2"
						managed = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "secret_name", "test/ssm/activation/update2"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "managed", "true"),
				),
			},
		},
	})
}

// TestAccSSMActivationResource_InvalidExpirationDate teste la validation des dates d'expiration invalides.
// Ce test tente de créer une activation SSM avec une date d'expiration qui dépasse 30 jours,
// et vérifie que l'erreur de validation appropriée est retournée.
func TestAccSSMActivationResource_InvalidExpirationDate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "` + getVar("SECRET_NAME") + `"
						
						expiration_date {
							days = 31
						}
					}
				`,
				ExpectError: regexp.MustCompile(`Total duration cannot exceed 30 days`),
			},
		},
	})
}

// TestAccSSMActivationResource_NegativeExpirationDate teste la validation des valeurs négatives dans la date d'expiration.
// Ce test tente de créer une activation SSM avec des valeurs négatives dans expiration_date,
// et vérifie que l'erreur de validation appropriée est retournée.
func TestAccSSMActivationResource_NegativeExpirationDate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "` + getVar("SECRET_NAME") + `"
						
						expiration_date {
							days = -1
						}
					}
				`,
				ExpectError: regexp.MustCompile(`Days must be a positive number`),
			},
		},
	})
}

// TestAccSSMActivationResource_UpdateManaged teste le changement de managed de true à false.
// Ce test crée d'abord une activation SSM avec managed=true, puis met à jour pour managed=false
// avec un secret existant pour vérifier que la recréation fonctionne correctement.
func TestAccSSMActivationResource_UpdateManaged(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "test/ssm/activation/managed-update"
						managed = true
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "secret_name", "test/ssm/activation/managed-update"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "managed", "true"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "secret_arn"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "secret_version"),
				),
			},
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					resource "test_ssm_activation" "test" {
						iam_role = "` + getVar("ACTIVATION_ROLE_NAME") + `"
						secret_name = "` + getVar("SECRET_NAME") + `"
						managed = false
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "id"),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "secret_name", getVar("SECRET_NAME")),
					resource.TestCheckResourceAttr("test_ssm_activation.test", "managed", "false"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "secret_arn"),
					resource.TestCheckResourceAttrSet("test_ssm_activation.test", "secret_version"),
				),
			},
		},
	})
}
