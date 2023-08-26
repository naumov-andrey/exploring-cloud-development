#!/bin/bash
# Recursively finds all directories with a go.mod file and creates
# a GitHub Actions JSON output. This is used by the linter action.

echo "Resolving modules in $(pwd)"

PATHS=$(find . -mindepth 2 -type f -name go.mod -printf '"%h",')
echo "matrix={\"workdir\":[${PATHS%?}]}" >> $GITHUB_OUTPUT
