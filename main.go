package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	axm "github.com/neilmartin83/terraform-provider-axm/internal/provider"
)

func main() {
	providerserver.Serve(context.Background(), axm.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/neilmartin83/axm",
	})
}
