#!/usr/bin/env bash

# This script is used at the start of a new release cycle, to
# initialize the CHANGELOG
# Run at the top of the repository, as '.changes/init.sh'

version=$1
if [ -z "${version}" ]
then
    if [ ! -f VERSION ]
    then
        echo "File ./VERSION not found"
        echo "run this command from the repository root directory as .changes/init.sh"
        echo ""
        echo "or supply a version on the command line: ./init.sh v4.0.0"
        exit 1
    fi

    version_from_file=$(cat VERSION | tr -d '\n' | tr -d '\r' )

    if [ -z "$version_from_file" ]
    then
        echo "No version found in ./VERSION file"
        exit 1
    fi
    version=$version_from_file
fi

starts_with_v=$(echo ${version} | grep '^v')
if [ -n "$starts_with_v" ]
then
    version=$(echo ${version} | tr -d 'v')
fi

echo "Copy the following lines at the top of CHANGELOG.md"
echo ""
echo ""
echo "## ${version} (Unreleased)"
echo ""
echo "Changes in progress for v${version} are available at [.changes/v${version}](https://github.com/vmware/terraform-provider-vcd/tree/main/.changes/v${version}) until the release."
echo ""

