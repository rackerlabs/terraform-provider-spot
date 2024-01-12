package main

import (
	"context"
	"log"

	"terraform-provider-rackspacespot/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "ngpc.rxt.io/rackspace/rackspacespot",
	}

	err := providerserver.Serve(context.Background(), provider.New(), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
