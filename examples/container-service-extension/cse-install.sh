#!/usr/bin/env bash

! which vcd > /dev/null && echo "ERROR: No 'vcd' command found, please install VCD cli" && exit 1
! which cse > /dev/null && echo "ERROR: No 'cse' command found, please install CSE cli" && exit 1
[ ! -f ./config.yaml ] && echo "ERROR: config.yaml does not exist" && exit 1

# Execute cse install command. The config file needs to have 0600 permissions.
chmod 0400 config.yaml
cse install -s -c config.yaml
