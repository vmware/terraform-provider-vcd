#!/bin/bash
REPO=$1
PROVIDER_PATH=$2
PKG_NAME=$3
operation=$4

if [ -z "$operation" ]
then
    echo "This script is intended to be called from GNUmakefile"
    echo "(such as 'make website' or 'make website-test')"
    echo ""
    echo "# Syntax: web-site.sh website-repository provider-path package-name operation"
    echo "# for example:"
    echo "# web-site.sh github.com/hashicorp/terraform-website $(pwd) vcd website-provider"
    echo "# web-site.sh github.com/hashicorp/terraform-website $(pwd) vcd website-provider-test"
    exit 1
fi

if [ -z "${GOPATH}" ]
then
    export GOPATH=${HOME}/go
fi

website_repo=${GOPATH}/src/${REPO}

if [ ! -d ${website_repo} ]
then
    echo "Attempting to download ${website_repo}"
    git clone https://${REPO} ${GOPATH}/src/${REPO}
fi

if [ ! -d ${website_repo} ]
then
    echo "${website_repo} not found"
    exit 1
fi

make -C ${website_repo} ${operation} PROVIDER_PATH=${PROVIDER_PATH} PROVIDER_NAME=${PKG_NAME}
