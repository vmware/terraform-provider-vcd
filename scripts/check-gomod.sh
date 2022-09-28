#!/bin/bash

# This script is intended to check whether go.mod has a replaced go-vcloud-director module, exiting with error if so.
# If the environment variable RELEASE is set, it also checks whether go-vcloud-director is using a released version.

GO_VCLOUD_DIRECTOR_URI='github.com/vmware/go-vcloud-director/v2'

if [ ! -f go.mod ]
then
    echo "go.mod file doesn't exist"
    exit 1
fi

grepVersion=''
if [ -n "$RELEASE" ]; then
  echo "INFO: Will check that no unreleased versions of go-vcloud-director are present in go.mod"
  grepVersion=$(grep -E "$GO_VCLOUD_DIRECTOR_URI v[1-9]+\.[0-9]+\.[0-9]+\S" go.mod)
  if [ -n "$grepVersion" ]
  then
    printf "ERROR: Found a non-released version of go-vcloud-director: %s\n" "$(echo "$grepVersion" | cut -d' ' -f 2)"
  fi
fi

grepReplace=$(grep "replace $GO_VCLOUD_DIRECTOR_URI" go.mod | grep -v '^//')
if [ -n "$grepReplace" ]; then
  echo "ERROR: Found this replaced module: $grepReplace"
fi

if [ -z "$grepVersion" ] && [ -z "$grepReplace" ]; then
  echo "No errors found: SUCCESS"
  exit 0
fi

echo "Errors found: FAILURE"
exit 1
