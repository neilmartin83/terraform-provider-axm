# terraform-provider-axm

Terraform Provider for the Apple Business/School Manager API

[Read the documentation to get started with using it in Terraform!](https://registry.terraform.io/providers/neilmartin83/axm/latest/docs)

This is in very early development - there be dragons! Make sure you use a test/disposable API Key, as there may be rate limiting/auth issues.

As well as the Terraform provider data sources/resources, a comprehensive set of client functions are also implemented. These may be of interest to anyone who wishes to interract with this API using Go:

* [client](internal/provider/client) contains a complete client implementation to interract with all the [endpoints documented here](https://developer.apple.com/documentation/apple-school-and-business-manager-api). OAuth functions are in a separate [client_oauth.go](internal/provider/client/client_oauth.go) file due to their complexity.

Check out the [example Go scripts](examples/client) showing how the client functions can be used in other projects. The intention here is to break these out into a separate SDK eventually.
