# Changelog

## [1.5.3](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.5.2...v1.5.3) (2026-02-21)


### Bug Fixes

* standardize quotes in .goreleaser.yml configuration ([147abb6](https://github.com/neilmartin83/terraform-provider-axm/commit/147abb64ae73dba4ab6901d0680e8ee2bbc1b837))

## [1.5.2](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.5.1...v1.5.2) (2026-02-21)


### Bug Fixes

* standardize quotes in release workflow and update goreleaser action version ([87cad00](https://github.com/neilmartin83/terraform-provider-axm/commit/87cad0000ec69748809b6d911f0cdbe75972d260))

## [1.5.1](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.5.0...v1.5.1) (2026-02-21)


### Bug Fixes

* change fatal errors to skips for missing environment variables in acceptance tests ([4d76514](https://github.com/neilmartin83/terraform-provider-axm/commit/4d765149254170e9e7065cb87ae91d991705f8a1))
* correct typographical errors in documentation for organization device data sources ([e5188b4](https://github.com/neilmartin83/terraform-provider-axm/commit/e5188b4b458c82e5b1da85c7a7e80781d03104f3))
* improve string formatting in downloadAndParseActivityLog function ([5bc74c3](https://github.com/neilmartin83/terraform-provider-axm/commit/5bc74c39fea065408c59780fe1d169b8a201f535))
* rename build job to Build Tests and add name to generate job ([a78115b](https://github.com/neilmartin83/terraform-provider-axm/commit/a78115bddf42d21d7363b5cc3c7eb71763976156))

## [1.4.3](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.4.2...v1.4.3) (2025-12-30)


### Features

* add device management service list resource and computed name/type fields to resource schema ([#43](https://github.com/neilmartin83/terraform-provider-axm/issues/43)) ([6e11730](https://github.com/neilmartin83/terraform-provider-axm/commit/6e1173051c484f85936eef361b2e8cc786d895bf))


### Miscellaneous Chores

* release 1.4.3 ([a850eb5](https://github.com/neilmartin83/terraform-provider-axm/commit/a850eb51af158dde427b17473fb0890cb3ecb4ee))

## [1.4.2](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.4.1...v1.4.2) (2025-12-23)


### Features

* enhance error handling and retry logic for API requests ([a2903dc](https://github.com/neilmartin83/terraform-provider-axm/commit/a2903dc4379f26144bb7a36fbd9f36debaf733d3))


### Miscellaneous Chores

* release 1.4.2 ([5e21d36](https://github.com/neilmartin83/terraform-provider-axm/commit/5e21d36e3b0b331b8c688a6059278eee6305805d))

## [1.4.1](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.4.0...v1.4.1) (2025-12-20)


### Features

* add user-configurable timeouts nested attribute to all resources and data sources ([#40](https://github.com/neilmartin83/terraform-provider-axm/issues/40)) ([a3ec6a1](https://github.com/neilmartin83/terraform-provider-axm/commit/a3ec6a18ac879e9929cd989c930bd353cb5f2706))


### Miscellaneous Chores

* release 1.4.1 ([aab5f4f](https://github.com/neilmartin83/terraform-provider-axm/commit/aab5f4fe79e3d54837cc858aa9ad673d32c2e89e))

## [1.4.0](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.3.1...v1.4.0) (2025-12-18)


### Features

* add EthernetMacAddress field to DeviceAttribute struct ([66a49a7](https://github.com/neilmartin83/terraform-provider-axm/commit/66a49a7718b3d43d94ea059ae6daeb8679e667e1))
* enhance organization device data source with Ethernet MAC address and improved descriptions ([703f5f6](https://github.com/neilmartin83/terraform-provider-axm/commit/703f5f64742afc4064f7e9311806029bb5dc57f9))

## [1.2.11](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.2.10...v1.2.11) (2025-11-03)


### Bug Fixes

* update Schema method signature to include context and request parameters ([#28](https://github.com/neilmartin83/terraform-provider-axm/issues/28)) ([ec75b50](https://github.com/neilmartin83/terraform-provider-axm/commit/ec75b50e2298613af90a0df90344ee7181da3d32))

## [1.2.10](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.2.9...v1.2.10) (2025-11-01)


### Features

* enhance logging by including response headers in LogResponse method ([520181a](https://github.com/neilmartin83/terraform-provider-axm/commit/520181acb27b76429854f1a984886bd3a542390a))


### Miscellaneous Chores

* release 1.2.10 ([bf57779](https://github.com/neilmartin83/terraform-provider-axm/commit/bf577795389bc77e7139e424539865fcbceb5fad))

## [1.2.1](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.2.0...v1.2.1) (2025-08-04)


### Bug Fixes

* various client fixes ([#11](https://github.com/neilmartin83/terraform-provider-axm/issues/11)) ([ccb9434](https://github.com/neilmartin83/terraform-provider-axm/commit/ccb9434d46b4341af8c9ec56324f49c824089e4a))

## [1.2.0](https://github.com/neilmartin83/terraform-provider-axm/compare/v1.1.2...v1.2.0) (2025-07-25)


### Features

* add environment variable support for provider configuration ([cb8844c](https://github.com/neilmartin83/terraform-provider-axm/commit/cb8844c210ea2a48df407e5be37df97b3c3c6922))
* add environment variable support for provider configuration ([0719938](https://github.com/neilmartin83/terraform-provider-axm/commit/07199382a98c0181a5a08f42ee6797b8925cfe06))
* add GolangCI linting workflow for pull requests ([87d3214](https://github.com/neilmartin83/terraform-provider-axm/commit/87d321484fa0412d52739b5e52f9ad6dcbd1028b))
* add GolangCI linting workflow for pull requests ([8a762ed](https://github.com/neilmartin83/terraform-provider-axm/commit/8a762edef5842ddd4243e0dba516e915e6c76b0e))
* add release-please GitHub Actions workflow for automated releases ([10cebb8](https://github.com/neilmartin83/terraform-provider-axm/commit/10cebb81d76856fe8c8daa4539ee2a049deb3154))
* add release-please GitHub Actions workflow for automated releases ([7e151e7](https://github.com/neilmartin83/terraform-provider-axm/commit/7e151e7d4f20fa2e7ce1fb33155bd1e3aa3aa50c))
