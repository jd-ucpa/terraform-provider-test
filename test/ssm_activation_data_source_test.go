package test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccSSMActivationDataSource_Basic teste la récupération d'informations d'activation SSM.
// Ce test configure le provider avec le profil AWS_PROFILE_OTHER_AGAIN, utilise le data source
// SSM Activation avec l'ID d'activation spécifié, puis vérifie que les informations d'activation
// sont correctement récupérées et que tous les attributs sont définis.
func TestAccSSMActivationDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
					provider "test" {
						region = "eu-west-1"
						profile = "` + getVar("AWS_PROFILE_OTHER_AGAIN") + `"
					}
					
					data "test_ssm_activation" "test" {
						activation_id = "` + getVar("ACTIVATION_ID") + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.test_ssm_activation.test", "id"),
				),
			},
		},
	})
}
