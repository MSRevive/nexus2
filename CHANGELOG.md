# Changelog

## v1.0.0
### Changes
* Updated auralog and entgo dependencies to the latest version
* Switched to a faster json package (go-json) instead of using encoding/json

## v1.0.0-rc8
### Added
* Adds ExpireTime to config so you can modify when the log rotates.
* Adds Version to program start so you know what version you're running.

### Fixed
* Fixes log rotate not working.
* Fixes log not appending when resuming log file.
* Concurrency support for lists.
* Fixes api/v1 endpoints not returning the result.

## v1.0.0-rc7
### Fixed
* Fixes TLS error when enabled.

## v1.0.0-rc6
* Reverted config for HTTP and HTTPS ports, cause it is not needed.
* Auto cert http server will now error out if it fails.
* Enabled tls-alpn ACME challenges for TLS.
* Switched from TOML to INI config format because it's more standard.
* Set control headers to further secure the server.

## v1.0.0-rc5
* Hopefully fixed autocert not working on non default ports.

## v1.0.0-rc4
* Added config for HTTP and HTTPS ports

## v1.0.0-rc3
* Added auto certificate if enabled from Let's Encrypt.
* Added domain config option for the auto cert system.

## v1.0.0-rc2
* Renamed config IP to address.
* Renamed cdir flag to cfile.

## v1.0.0-rc1
Initial release.