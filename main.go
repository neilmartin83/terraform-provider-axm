package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	axm "github.com/neilmartin83/terraform-provider-axm/internal/provider"
)

func main() {
	err := providerserver.Serve(context.Background(), axm.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/neilmartin83/axm",
	})
	if err != nil {
		log.Fatalf("Error serving provider: %v", err)
	}
}
