#!/usr/bin/env bash

! which vcd > /dev/null && echo "ERROR: No 'vcd' command found, please install VCD cli" && exit 1
! which cse > /dev/null && echo "ERROR: No 'cse' command found, please install CSE cli" && exit 1
[ ! -f ./config.yaml ] && echo "ERROR: config.yaml does not exist" && exit 1

chmod 0600 config.yaml
cse install -s -c config.yaml || true # Ignore failures, for Terraform update/destroy not to fail