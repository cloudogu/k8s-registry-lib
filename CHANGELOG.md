# k8s-registry-lib Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
