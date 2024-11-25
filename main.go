package main

import (
	"context"
	"flag"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/inwx/terraform-provider-inwx/internal/provider"
	"log"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name inwx

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary.
	version string = "dev"

	// goreleaser can pass other information to the main package, such as the specific commit
	// https://goreleaser.com/cookbooks/using-main.version/
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	serveOpts := providerserver.ServeOpts{
		Address: "registry.terraform.io/inwx/inwx",
		Debug:   debug,
	}
	err := providerserver.Serve(context.Background(), provider.New(version), serveOpts)

	if err != nil {
		log.Fatal(err.Error())
	}

}
