# k8s-registry-lib Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Fixed
- [#22] map every error coming from the repo to a domain error

## [v0.4.1] - 2024-09-25
### Fixed
- [#16] map k8s-errors in version-registry & descriptor-repo

## [v0.4.0] - 2024-09-19
### Changed
- [#14] Relicense to AGPL-3.0-only

## [v0.3.1] - 2024-09-06
### Changed
- [#11] Use retry watcher because the regular kubernetes watches will interrupt in half an hour.
  - See https://blogs.gnome.org/dcbw/2020/08/05/kubernetes-watches-will-ghost-you-without-warning/

## [v0.3.0] - 2024-08-21
### Changed
- [#5] Refactor and split the dogu registry in a dogu version registry and a dogu spec repository.

## [v0.2.2] - 2024-08-02
### Added
- [#8] Exposed function to create config entries from map

## [v0.2.1] - 2024-08-01
### Fixed
- [#6] DoguRegistry method IsEnabled returns an error when the configmap for the spec is not available - now it returns false instead

## [v0.2.0] - 2024-07-12
### Added
- [#3] Add config repositories for global, dogu and sensitive config
- [#3] Add option to watch for config-changes

### Changed
- Refactor local dogu registry 
- Remove fallback option to etcd

## [v0.1.0] - 2024-05-08
### Added
- [#1] Local dogu registry

## [v0.0.1] - 2024-05-08
### Added
- Initialized project
