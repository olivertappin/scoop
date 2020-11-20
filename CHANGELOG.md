# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.2] - 2020-11-20

### Changed
- Address lack of additional message detail. ([#1](https://github.com/olivertappin/scoop/issues/1))

## [2.0.1] - 2020-04-15
### Added
- Video tutorial to show basic usage and embedded within `README.md`.

### Changed
- Change 'Notes' anchor to 'FAQs' within `README.md`.
- Add additional sections under new FAQs anchor in `README.md`.

## [2.0.0] - 2020-01-04
### Added
- New `to-arg` and `from-arg` arguments to define individual, additional arguments for the queue decelerations.
- New `from-durable` and `to-durable` arguments to define durable queues individually for the queue decelerations.
- Queue deceleration arguments to `README.md` with documentation references.

### Changed
- Slight tidy up in `CHANGELOG.md`.
- Order of verbose output to be more logical.
- Relocated `scoop.go` script, inline with GoLang community recommendations.

### Removed
- `durable` argument to be replaced by `from-durable` and `to-durable`.

## [1.2.1] - 2020-01-04
### Changed
- Correct `x-dead-letter-exchange` and `x-dead-letter-routing-key` type casting.

## [1.2.0] - 2019-10-28
### Added
- New `arg` argument to define additional arguments to queue decelerations.
- Queue deceleration arguments to `README.md` with documentation references.
- Missing exit statuses after failed validation.
- Validation to prevent users from moving messages between the same queue.
- Install and uninstall commands to `README.md`.

### Changed
- Correct typo in `durable` argument description for `deceleration`.

## [1.1.0] - 2019-10-10
### Added
- New parameter to define durable queues.
- Missing `exchange` parameter to `README.md`.

### Changed
- Remove unnecessary whitespace from `src/scoop.go` source.
- Add full-stops to all `CHANGELOG.md` lines for consistency.

## [1.0.0] - 2019-10-03
### Added
- Basic arguments for command line scooping.
- Native workaround documented within the `README.md` as suggested by [@EagleEyeJohn](https://github.com/EagleEyeJohn).
