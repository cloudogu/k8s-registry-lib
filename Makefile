ARTIFACT_ID=k8s-apply-lib
VERSION=0.1.0
GOTAG?=1.22
MAKEFILES_VERSION=9.0.4
.DEFAULT_GOAL:=default
LINT_VERSION=v1.57.2

include build/make/variables.mk
include build/make/self-update.mk
include build/make/dependencies-gomod.mk
include build/make/build.mk
include build/make/test-common.mk
include build/make/test-unit.mk
include build/make/static-analysis.mk
include build/make/clean.mk
include build/make/release.mk
include build/make/mocks.mk

.PHONY: default
default: unit-test