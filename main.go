package main

import (
	"context"
	"flag"
	"log"

	"github.com/cirro-bio/terraform-provider-cirro/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Run "go generate ./..." to regenerate provider docs.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name cirro

var version string = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/cirro-bio/cirro",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
