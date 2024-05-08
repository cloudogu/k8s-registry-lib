#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

sonarProps='sonar-project.properties'

# this function will be sourced from release.sh and be called from release_functions.sh
update_versions_modify_files() {
  newReleaseVersion="${1}"
  sed -i "s/\(sonar.projectVersion=\).*/\1${newReleaseVersion}/" "${sonarProps}"
}

update_versions_stage_modified_files() {
  git add "${sonarProps}"
}
