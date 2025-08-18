ARTIFACT_ID=k8s-registry-lib
VERSION=0.5.1
GOTAG?=1.24.6
MAKEFILES_VERSION=10.2.0
.DEFAULT_GOAL:=default

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