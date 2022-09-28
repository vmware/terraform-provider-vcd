#!/bin/bash

# This script is intended to check whether go.mod has a replaced module, exiting with error if so.
# If the environment variable RELEASE is set, it also checks whether go-vcloud-director is using a released version.

GO_VCLOUD_DIRECTOR_URI='github.com/vmware/go-vcloud-director/v2'

if [ ! -f go.mod ]
then
    echo "go.mod file doesn't exist"
    exit 1
fi

grepAlpha=''
if [ -n "$RELEASE" ]; then
  echo "INFO: Will check that no unreleased versions of go-vcloud-director are present in go.mod"
  grepAlpha=$(grep -E "$GO_VCLOUD_DIRECTOR_URI v[1-9]+\.[0-9]+\.[0-9]+\-.*" go.mod)
  if [ -n "$grepAlpha" ]
  then
    printf "ERROR: Found a non-released version of go-vcloud-director: %s\n" "$(echo "$grepAlpha" | cut -d' ' -f 2)"
  fi
fi

grepReplace=$(grep "replace $GO_VCLOUD_DIRECTOR_URI" go.mod | grep -v '//')
if [ -n "$grepReplace" ]; then
  printf "ERROR: Found these replaced modules:\n%s\n" "$(echo "$grepReplace" | cut -d' ' -f 2)"
fi

if [ -z "$grepAlpha" ] && [ -z "$grepReplace" ]; then
  echo "No errors found: SUCCESS"
  exit 0
fi

echo "Errors found: FAILURE"
exit 1
