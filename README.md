# terraform-provider-axm
Terraform Provider for the Apple Business/School Manager API

[Read the documentation to get started with using it in Terraform!](https://registry.terraform.io/providers/neilmartin83/axm/latest/docs)

This is in very early development - there be dragons! We are only working with read-only endpoints but ensure you use a test/disposable API Key, as there may be rate limiting/auth issues.

As well as the Terraform provider data sources/resources, a comprehensive set of client functions are also implemented. These may be of interest to anyone who wishes to interract with this API using Go:

* [client.go](internal/provider/client.go) contains a complete client implementation to interract with all the [endpoints documented here](https://developer.apple.com/documentation/apple-school-and-business-manager-api).
* [client_oauth.go](internal/provider/client_oauth.go) contains the [authentication implementation](https://developer.apple.com/documentation/apple-school-and-business-manager-api/implementing-oauth-for-the-apple-school-and-business-manager-api) to handle assertation and token lifecycle management. It was complex enough to warrant breaking out into its own thing!
