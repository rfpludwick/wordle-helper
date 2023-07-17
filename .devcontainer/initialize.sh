#!/usr/bin/env bash

set -e

# shellcheck disable=SC2154
git config --global --add safe.directory "${containerWorkspaceFolder}"

# Local initializer?
if [ -x .devcontainer/initialize-local.sh ]; then
	.devcontainer/initialize-local.sh
fi
