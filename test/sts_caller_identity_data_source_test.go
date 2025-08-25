package test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSTSCallerIdentityDataSource teste le data source STS Caller Identity qui récupère les informations
// du compte AWS actuellement authentifié. Ce test charge les variables d'environnement depuis test.env,
// configure le provider avec assume_role, puis vérifie que le data source retourne bien les attributs
// attendus : account_id, arn, et user_id. Le test confirme que l'authentification AWS fonctionne correctement.
func TestAccSTSCallerIdentityDataSource(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "ROLE_ARN", "ACCOUNT_ID_OTHER")
	
	expectedAccountID := os.Getenv("ACCOUNT_ID_OTHER")
	t.Logf("Expected account_id: %s", expectedAccountID)
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSTSCallerIdentityDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_caller_identity.test", "account_id"),
					resource.TestCheckResourceAttrSet("data.test_caller_identity.test", "arn"),
					resource.TestCheckResourceAttrSet("data.test_caller_identity.test", "user_id"),
					// Vérifier que l'account_id correspond à ACCOUNT_ID_OTHER (le compte du rôle assumé)
					resource.TestCheckResourceAttr("data.test_caller_identity.test", "account_id", expectedAccountID),
				),
			},
		},
	})
}



func testAccSTSCallerIdentityDataSourceConfig() string {
	return `
		provider "test" {
			region = "eu-west-1"
			assume_role {
				role_arn = "` + os.Getenv("ROLE_ARN") + `"
			}
		}
		
		data "test_caller_identity" "test" {}
		
		output "account_id" {
			value = data.test_caller_identity.test.account_id
		}
	`
}

// TestAccSTSCallerIdentityDataSourceDefaultProfile teste le data source STS Caller Identity
// en utilisant le profil AWS par défaut (sans assume_role). Ce test vérifie que l'account_id
// retourné correspond à ACCOUNT_ID_DEFAULT du profil AWS_PROFILE.
func TestAccSTSCallerIdentityDataSourceDefaultProfile(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "ACCOUNT_ID_DEFAULT")
	
	expectedAccountID := os.Getenv("ACCOUNT_ID_DEFAULT")
	t.Logf("Expected account_id (default profile): %s", expectedAccountID)
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSTSCallerIdentityDataSourceDefaultProfileConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_caller_identity.test_default", "account_id"),
					resource.TestCheckResourceAttrSet("data.test_caller_identity.test_default", "arn"),
					resource.TestCheckResourceAttrSet("data.test_caller_identity.test_default", "user_id"),
					// Vérifier que l'account_id correspond à ACCOUNT_ID_DEFAULT (le compte du profil par défaut)
					resource.TestCheckResourceAttr("data.test_caller_identity.test_default", "account_id", expectedAccountID),
				),
			},
		},
	})
}

func testAccSTSCallerIdentityDataSourceDefaultProfileConfig() string {
	return `
		provider "test" {
			region = "eu-west-1"
			# Pas d'assume_role - utilise le profil AWS par défaut
		}
		
		data "test_caller_identity" "test_default" {}
		
		output "account_id_default" {
			value = data.test_caller_identity.test_default.account_id
		}
	`
}

// TestAccSTSCallerIdentityDataSourceOtherProfile teste le data source STS Caller Identity
// en utilisant le profil AWS_PROFILE_OTHER=3098. Ce test vérifie que l'account_id
// retourné correspond à ACCOUNT_ID_OTHER du profil AWS_PROFILE_OTHER.
func TestAccSTSCallerIdentityDataSourceOtherProfile(t *testing.T) {
	// Charger les variables d'environnement et valider les variables requises
	SetupTestEnv(t, "AWS_PROFILE_OTHER", "ACCOUNT_ID_OTHER")
	
	expectedAccountID := os.Getenv("ACCOUNT_ID_OTHER")
	t.Logf("Expected account_id (other profile): %s", expectedAccountID)
	
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: testAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSTSCallerIdentityDataSourceOtherProfileConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_caller_identity.test_other", "account_id"),
					resource.TestCheckResourceAttrSet("data.test_caller_identity.test_other", "arn"),
					resource.TestCheckResourceAttrSet("data.test_caller_identity.test_other", "user_id"),
					// Vérifier que l'account_id correspond à ACCOUNT_ID_OTHER (le compte du profil 3098)
					resource.TestCheckResourceAttr("data.test_caller_identity.test_other", "account_id", expectedAccountID),
				),
			},
		},
	})
}

func testAccSTSCallerIdentityDataSourceOtherProfileConfig() string {
	return `
		provider "test" {
			region = "eu-west-1"
			profile = "` + os.Getenv("AWS_PROFILE_OTHER") + `"
		}
		
		data "test_caller_identity" "test_other" {}
		
		output "account_id_other" {
			value = data.test_caller_identity.test_other.account_id
		}
	`
}
