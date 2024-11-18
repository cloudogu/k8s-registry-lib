ARTIFACT_ID=k8s-registry-lib
VERSION=0.5.0
GOTAG?=1.23.1
MAKEFILES_VERSION=9.3.2
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