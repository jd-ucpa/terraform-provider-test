package test

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/jd-ucpa/terraform-provider-test/internal"
	ini "gopkg.in/ini.v1"
)

// Variable globale pour la configuration de test
var testConfig *ini.File

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"test": providerserver.NewProtocol6WithError(internal.Provider()),
}

// getVar récupère une valeur de la configuration de test
func getVar(key string) string {
	if testConfig == nil {
		cfg, err := ini.Load("test.env")
		if err != nil {
			panic(fmt.Sprintf("Impossible de charger test.env: %v", err))
		}
		testConfig = cfg
	}

	value := testConfig.Section("").Key(key).String()
	if value == "" {
		panic(fmt.Sprintf("La variable %s est manquante dans test.env", key))
	}
	return value
}

