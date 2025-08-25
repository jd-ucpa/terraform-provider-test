package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/jd-ucpa/terraform-provider-test/internal"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/jd-ucpa/test",
	}

	err := providerserver.Serve(context.Background(), internal.Provider, opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
