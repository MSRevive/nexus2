# Changelog

## v1.3.0
### Added
* Config options to modify the driver for SQLite.
  * ``maxidleconns: 0 # Max idle connections``
  * ``maxopenconns: 0 # Max open connections``
  * ``connmaxlifetime: "" # Max lifetime of connections.``
* Rate limiting by IP instead.

### Changed
* Changed the router to chi instead of gorilla/mux
  * RiP gorilla/mux since it's no longer maintained.
* Better error handling for database opening.
* ``/character/rollback`` endpoints are now in their own routing group. This will break any automated rollback systems.

### Fixed
* Fixed SQL connection not being closed.

## v1.2.1
### Added
* Timeout for database in config.

### Fixed
* Gradefully shutdown server on signal.

## v1.2.0
### Added
* Added config options ``BackupTime`` and ``MaxBackups``
  * Customize how max character backups and how often to do character backups
* Added new endpoint ``PATCH /character/{steamid}/{slot}/rollback/latest`` which restore the character to via latest backup.

### Changed
* Refactor the code and the project structure.

## v1.1.0
### Added
* Add migration system (with -m flag) to migrate TOML/INI config file to YAML
* Add authentication to check useragent as well.
* Add database migration system, thanks to [@pbarnum](https://github.com/pbarnum)
* Add new endpoints for rollback system, thanks to [@pbarnum](https://github.com/pbarnum)
  * ``PATCH /character/{uid}/restore`` - Restore character to original version.
  * ``GET /character/{steamid}/{slot}/versions`` - Get all character versions.
  * ``PATCH /character/{steamid}/{slot}/rollback/{version}`` - Rollback character to specified version.
  
### Fixed
* Fix issue of creating character not overwriting, thanks to [@pbarnum](https://github.com/pbarnum)


## v1.0.4
### Added
* Add isAdmin field for FN admins.
* Add API endpoint to generate character file from data on via database for steamid64 and slot.
* Add API endpoint to check if FN connection is valid.
* Add new middleware method to bypass authorization so we can do manual authorization.
* Add support for YAML config file.

### Fixed
* Fix GET endpoint for steamid and slot to return single character instead of array of characters.

## v1.0.3
### Changes
* Make sure slot is always emitted, and defaults to 0 if there's no slot.

## v1.0.3-rc1
### Added
* Add debug message to see if request is received by server.

## v1.0.2
### Added
* Add (bool)isBanned field to JSON response for ``api/v1/character/{steamid64}``, ``api/v1/character/{steamid64}/{slot}``, ``api/v1/character/id/{id}``

## v1.0.1
### Added
* Add size field

### Changes
* Data is now text instead of blob type.

## v1.0.0
Initial Release

### Removed 
* Removed race field since it's no longer used.
* Removed all unnecessary fields and replaced them with a single blob field.

### Changes
* Say the actual license in startup instead of just linking to license.

## v1.0.0-rc11
### Changes
* Changed all JSON fields to be text instead of varchar(255)

## v1.0.0-rc10
### Removed
* Got rid of sheaths field.

### Changes
* Switched bags, spells, and equipped to text to support large JSON data.
* Switched to new migration engine.

## v1.0.0-rc9
### Changes
* Updated auralog and entgo dependencies to the latest version
* Switched to a faster json package (go-json) instead of using encoding/json
* Switch race to string because MSC.

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