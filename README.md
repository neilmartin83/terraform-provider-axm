# terraform-provider-axm

Terraform Provider for the Apple School and Business Manager API

If you're new to Terraform, [check out this excellent tutorial series](https://macadminmusings.com/blog/2025/10/14/terraform-101-introduction/) by Scott Blake.

[Read the documentation to get stuck in!](https://registry.terraform.io/providers/neilmartin83/axm/latest/docs)

## API Rate Limit Information

The Apple School and Business Manager API is rate limited. These are currently undocumented by Apple but through trial and error appear to be:

* For endpoints using GET: 20 requests per minute
* For endpoints using POST: 10 requests per hour

When triggered, the API returns a 429 (too many requests) response with a `Retry-After` header containing a value in seconds. The provider behaves as follows:

* If `Retry-After` is less than or equal to `60`, wait for the number of seconds specified and retry the request
* If `Retry-After` is greater than `60`, return an error and abort the current operation.

This avoids excessive execution time, being mindful of CI/CD pipelines that are priced based on runtime.

Please be mindful of these limits as running the provider too frequently may trigger rate limiting, especially with device assignment operations. Operations may take longer than expected or fail.

A general recommendation is to allow 1 hour between provider runs in a production environment.

## Go Client

As well as the Terraform provider data sources/resources, a comprehensive set of client functions are also implemented. These may be of interest to anyone who wishes to interact with this API using Go:

* [client](internal/client) contains a complete client implementation to interract with all the [endpoints documented here](https://developer.apple.com/documentation/apple-school-and-business-manager-api). OAuth functions are in a separate [client_oauth.go](internal/client/client_oauth.go) file due to their complexity.

Check out the [example Go scripts](examples/client) showing how the client functions can be used in other projects. The intention here is to break these out into a separate SDK eventually.
