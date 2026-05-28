# Changelog

All notable changes to this project will be documented in this file.

## [0.19.46] - 2026-05-28

## [0.19.45] - 2026-05-28

## [0.19.44] - 2026-05-27

### Changed
- refactor(generator): bump react+vue pins to vform.Form adoption

## [0.19.43] - 2026-05-27

### Changed
- refactor(generator): bump react+vue pins for view.Redirect shorthand

## [0.19.42] - 2026-05-27

### Fixed
- bump vue template pin to v0.0.8

## [0.19.41] - 2026-05-27

### Fixed
- correct react+vue template pins to migration tags

## [0.19.40] - 2026-05-27

### Fixed
- bump template pins to framework-middleware cleanup tags

## [0.19.39] - 2026-05-27

### Fixed
- bump template pins to drop gonertia + populate meta tag

## [0.19.38] - 2026-05-27

### Fixed
- bump supported template tags to latest

## [0.19.37] - 2026-05-27

## [0.19.36] - 2026-05-27

### Fixed
- prefix on APP_KEY in scaffolded .env

## [0.19.35] - 2026-05-27

## [0.19.34] - 2026-05-27

## [0.19.33] - 2026-05-27

## [0.19.32] - 2026-05-26

## [0.19.31] - 2026-05-26

## [0.19.30] - 2026-05-26

## [0.19.29] - 2026-05-26

## [0.19.28] - 2026-05-26

## [0.19.27] - 2026-05-18

## [0.19.26] - 2026-05-16

## [0.19.25] - 2026-05-16

## [0.19.24] - 2026-05-11

## [0.19.23] - 2026-05-11

## [0.19.22] - 2026-05-10

## [0.19.21] - 2026-05-10

## [0.19.20] - 2026-05-08

## [0.19.19] - 2026-05-08

## [0.19.18] - 2026-05-08

### Fixed
- 

## [0.19.17] - 2026-05-08

### Changed
- chore: bump pinned templates to v0.41.0-compatible tags

## [0.19.16] - 2026-05-08

### Changed
- refactor: pin templates by tag, drop framework version coupling

## [0.19.15] - 2026-05-08

## [0.19.14] - 2026-05-07

## [0.19.13] - 2026-05-07

## [0.19.12] - 2026-05-07

## [0.19.11] - 2026-05-06

## [0.19.10] - 2026-05-06

## [0.19.9] - 2026-05-06

## [0.19.8] - 2026-05-06

## [0.19.7] - 2026-05-05

## [0.19.6] - 2026-05-05

## [0.19.5] - 2026-05-04

## [0.19.4] - 2026-05-04

## [0.19.3] - 2026-05-03

## [0.19.2] - 2026-04-30

## [0.19.1] - 2026-04-30

### Fixed
- match goreleaser archives, verify checksum, respect brew

## [0.19.0] - 2026-04-30

### Added
- expose mysql in interactive prompt and flag help

## [0.18.3] - 2026-04-29

## [0.18.2] - 2026-04-29

## [0.18.1] - 2026-04-29

### Fixed
- align config stub and wizard hint with APP_PORT env

## [0.18.0] - 2026-04-29

### Added
- --stack flag for picking React or Vue template

### Fixed
- resolve framework version from tags, not GitHub Releases
- use cli.Tip for the bun install hint
- bump velocity-cli to v0.14.0

### Changed
- chore: rename velocity-template -> velocity-template-react
- chore(deps): bump velocity-cli to v0.13.1

## [0.17.8] - 2026-04-28

## [0.17.7] - 2026-04-28

## [0.17.6] - 2026-04-26

## [0.17.5] - 2026-04-22

### Fixed
- use cli.Tip for the bun install hint

## [0.17.4] - 2026-04-22

### Fixed
- bump velocity-cli to v0.14.0

## [0.17.3] - 2026-04-22

### Fixed
- silence inertia vite SSR warmup when --ssr is off

## [0.17.2] - 2026-04-22

### Fixed
- generate AUTH_JWT_SECRET to match framework env name

## [0.17.1] - 2026-04-22

### Changed
- perf(generator): download template as tarball and build vel in parallel

## [0.17.0] - 2026-04-22

### Added
- suggest installing bun when only npm is available

## [0.16.1] - 2026-04-22

### Fixed
- show raw tool output when no package pattern matches

## [0.16.0] - 2026-04-22

### Added
- show each dep under its group as it installs

## [0.15.0] - 2026-04-22

### Added
- stream dep install output in real time

## [0.14.32] - 2026-04-22

### Fixed
- bump velocity-cli to v0.13.0 for Ctrl+C cancel

## [0.14.31] - 2026-04-22

### Fixed
- detect bun/npm availability before install

## [0.14.30] - 2026-04-21

### Fixed
- publish as Homebrew cask instead of formula

### Changed
- refactor(generator): remove dead setupTemplatesAndHotReload

## [0.14.29] - 2026-04-21

## [0.14.28] - 2026-04-21

## [0.14.27] - 2026-04-19

## [0.14.26] - 2026-04-19

### Changed
- style(cli): switch primary to terminal default; reserve green for success

## [0.14.25] - 2026-04-19

## [0.14.24] - 2026-04-19

### Changed
- style(cli): switch theme to Velocity brand greens

## [0.14.23] - 2026-04-19

### Changed
- chore(version): bump minimum Go requirement to 1.26

## [0.14.22] - 2026-04-19

### Fixed
- uncomment DB_SSL_MODE=disable for postgres scaffolds

## [0.14.21] - 2026-04-19

## [0.14.20] - 2026-04-19

## [0.14.19] - 2026-04-19

## [0.14.18] - 2026-04-19

## [0.14.17] - 2026-04-19

## [0.14.16] - 2026-04-19

## [0.14.15] - 2026-04-18

## [0.14.14] - 2026-04-18

## [0.14.13] - 2026-04-18

## [0.14.12] - 2026-04-18

## [0.14.11] - 2026-04-18

## [0.14.10] - 2026-04-17

## [0.14.9] - 2026-04-15

## [0.14.8] - 2026-04-14

## [0.14.7] - 2026-04-14

## [0.14.6] - 2026-04-14

## [0.14.5] - 2026-04-14

## [0.14.4] - 2026-04-14

## [0.14.3] - 2026-04-14

## [0.14.2] - 2026-04-13

## [0.14.1] - 2026-04-13

## [0.14.0] - 2026-04-13

### Added
- hand serve startup to the user

## [0.13.1] - 2026-04-12

## [0.13.0] - 2026-04-12

### Added
- generate APP_KEY, QUEUE_SIGNING_KEY, and JWT_SECRET

## [0.12.4] - 2026-04-12

## [0.12.3] - 2026-04-12

### Changed
- ci: allow manual dispatch of the release workflow

## [0.12.2] - 2026-04-12

### Fixed
- generate

## [0.12.1] - 2026-04-12

### Changed
- chore: rename 'Official CLI' tagline to 'Official Installer'

## [0.12.0] - 2026-04-12

### Added
- thin-wrapper over ./vel + UX polish

## [0.11.1] - 2026-04-12

### Changed
- chore: bump go to 1.26.2 (stdlib crypto/x509 CVE)

## [0.11.0] - 2026-04-12

### Added
- preflight database + drop air dependency

## [0.10.2] - 2026-04-12

### Changed
- refactor(ci): single ci job for push + PR

## [0.10.1] - 2026-04-12

### Fixed
- portable sed so tests pass on GNU + BSD
- consume published velocity-cli and embed theme file outside gitignore glob

## [0.10.0] - 2026-04-12

### Added
- prompt for database and SSR in `velocity new`

## [0.9.1] - 2026-04-12

### Changed
- refactor: migrate UI to velocity-cli SDK

## [0.9.0] - 2026-04-12

### Added
- --ssr flag for velocity new

## [0.8.0] - 2026-04-09

### Added
- build vel binary after scaffolding

## [0.7.2] - 2026-04-09

### Fixed
- remove unused os/exec imports

## [0.7.1] - 2026-04-09

### Fixed
- remove vel binary build step, update tips for baked-in CLI

## [0.7.0] - 2026-04-09

### Added
- remove cmd/vel scaffolding, simplify migration runner

## [0.6.50] - 2026-04-08

### Changed
- chore: update velocity framework to v0.20.3

## [0.6.49] - 2026-02-23

### Changed
- Update docs link to velocity.velocitykode.com/docs

## [0.6.48] - 2026-02-21

### Changed
- chore: add workflow_dispatch trigger to auto-release

## [0.6.47] - 2026-02-19

### Changed
- chore: update velocity framework to v0.16.0

## [0.6.46] - 2026-02-17

### Changed
- chore: update velocity framework to v0.15.0

## [0.6.45] - 2026-02-15

### Changed
- chore: update velocity framework to v0.14.1

## [0.6.44] - 2026-02-15

### Changed
- chore: update velocity framework to v0.14.0

## [0.6.43] - 2026-02-15

### Changed
- chore: update velocity framework to v0.13.0

## [0.6.42] - 2026-02-15

### Changed
- chore: update velocity framework to v0.12.0

## [0.6.41] - 2026-02-15

### Changed
- chore: update velocity framework to v0.11.0

## [0.6.40] - 2026-02-14

### Changed
- refactor: update import paths after velocity pkg/ promotion

## [0.6.39] - 2026-02-14

### Changed
- chore: update velocity framework to v0.10.2

## [0.6.38] - 2026-02-14

### Changed
- chore: update velocity framework to v0.10.1

## [0.6.37] - 2026-02-14

### Changed
- chore: update velocity framework to v0.10.0

## [0.6.36] - 2026-02-13

### Changed
- chore: update velocity framework to v0.9.17

## [0.6.35] - 2026-02-13

### Changed
- chore: update velocity framework to v0.9.16

## [0.6.34] - 2026-02-12

### Fixed
- skip default migrations when template already provides them

## [0.6.33] - 2026-02-12

### Changed
- chore: update velocity framework to v0.9.15

## [0.6.32] - 2026-02-12

### Changed
- refactor: migrate to instance-based DI - remove package-level globals

## [0.6.31] - 2026-02-12

### Changed
- chore: update velocity framework to v0.9.13

## [0.6.30] - 2026-02-12

### Changed
- chore: update velocity framework to v0.9.12

## [0.6.29] - 2026-02-12

### Changed
- chore: update velocity framework to v0.9.11

## [0.6.28] - 2026-02-11

### Changed
- chore: update velocity framework to v0.9.10

## [0.6.27] - 2026-02-11

### Changed
- chore: update velocity framework to v0.9.9

## [0.6.26] - 2026-02-11

### Changed
- chore: update velocity framework to v0.9.8

## [0.6.25] - 2026-02-11

### Changed
- chore: update velocity framework to v0.9.7

## [0.6.24] - 2026-02-11

### Changed
- chore: update velocity framework to v0.9.6

## [0.6.23] - 2026-02-11

### Fixed
- require Go 1.25.7 to resolve crypto/tls vulnerability

## [0.6.22] - 2026-02-11

### Changed
- refactor: update stubs to use velocity.Default() DI pattern

## [0.6.21] - 2026-02-11

### Changed
- chore: update velocity framework to v0.9.5

## [0.6.20] - 2026-02-11

### Changed
- chore: update velocity framework to v0.9.4

## [0.6.19] - 2026-02-11

### Changed
- chore: update velocity framework to v0.9.3

## [0.6.18] - 2026-02-11

### Changed
- chore: update velocity framework to v0.9.2

## [0.6.17] - 2026-02-08

### Changed
- chore: update velocity framework to v0.9.1

## [0.6.16] - 2026-02-08

### Changed
- chore: update velocity framework to v0.9.0

## [0.6.15] - 2026-02-05

### Changed
- chore: update velocity framework to v0.8.0

## [0.6.14] - 2026-01-29

### Changed
- chore: update velocity framework to v0.7.0

## [0.6.13] - 2026-01-29

### Changed
- chore: update velocity framework to v0.6.8

## [0.6.12] - 2026-01-29

### Changed
- chore: update velocity framework to v0.6.6

## [0.6.11] - 2026-01-29

### Changed
- chore: update velocity framework to v0.6.5

## [0.6.10] - 2026-01-29

### Changed
- chore: update velocity framework to v0.6.4

## [0.6.9] - 2026-01-29

### Changed
- chore: update velocity framework to v0.6.3

## [0.6.8] - 2026-01-29

### Changed
- chore: update velocity framework to v0.6.2

## [0.6.7] - 2026-01-29

### Changed
- chore: update velocity framework to v0.6.1

## [0.6.6] - 2026-01-29

### Changed
- chore: update velocity framework to v0.6.0

## [0.6.5] - 2026-01-29

### Changed
- chore: update velocity framework to v0.5.0

## [0.6.4] - 2026-01-27

### Changed
- chore: update velocity framework to v0.4.0

## [0.6.3] - 2026-01-26

### Changed
- chore: update velocity framework to v0.3.1

## [0.6.2] - 2026-01-26

### Changed
- chore: update velocity framework to v0.3.0

## [0.6.1] - 2026-01-26

### Changed
- test: add coverage for --api flag

## [0.6.0] - 2026-01-26

### Added
- add --api flag for API-only projects

### Changed
- style: fix gofmt formatting
- test: add comprehensive test coverage for config, commands, generator, and detector packages

## [0.5.9] - 2026-01-09

### Changed
- chore: update velocity framework to v0.2.5

## [0.5.8] - 2026-01-09

### Changed
- chore: update velocity framework to v0.2.4

## [0.5.7] - 2026-01-09

### Changed
- chore: update velocity framework to v0.2.3

## [0.5.6] - 2026-01-09

### Changed
- chore: update velocity framework to v0.2.2

## [0.5.5] - 2026-01-09

### Changed
- chore: update velocity framework to v0.2.1

## [0.5.4] - 2026-01-09

### Changed
- chore: update velocity framework to v0.2.0

## [0.5.3] - 2026-01-02

### Changed
- chore: update velocity framework to v0.1.2

## [0.5.2] - 2026-01-02

### Changed
- chore: update velocity framework to v0.1.1

## [0.5.1] - 2026-01-01

### Fixed
- use go run for module-aware migration execution

## [0.5.0] - 2026-01-01

### Added
- 

## [0.4.5] - 2026-01-01

### Changed
- chore: update velocity framework to v0.1.0

## [0.4.4] - 2026-01-01

### Fixed
- remove cmd/vel stub reference that broke build

## [0.4.3] - 2026-01-01

### Fixed
- correct stub paths to app/http/, disable init command

## [0.4.2] - 2026-01-01

### Changed
- chore: update velocity framework to v0.0.5

## [0.4.1] - 2026-01-01

### Changed
- docs: add README and LICENSE

## [0.4.0] - 2026-01-01

### Added
- add one-liner to setup vel shell function

## [0.3.0] - 2026-01-01

### Added
- add shell function tip after project creation

## [0.2.0] - 2026-01-01

### Added
- initial velocity-installer

### Fixed
- sync version with release tag

### Changed
- chore: add package doc
- chore: enable workflows
- refactor: rename .velocity to .vel
- chore: disable workflows temporarily
- chore: add .gitignore
- refactor: move main.go to root

## [0.0.1] - 2026-01-01

### Added
- Initial release of velocity-installer
- `velocity new` command to create new projects
- `velocity init` command to initialize existing projects
- `velocity config` command for global configuration
- `velocity self-update` command to update installer
- Automatic `vel` binary building after project creation
