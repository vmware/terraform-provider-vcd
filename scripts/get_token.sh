#!/bin/bash
# This script will connect to the vCD using username and password,
# and show the headers that contain a bearer or authorization token.
#
user=$1
password=$2
org=$3
IP=$4

if [ -z "$IP" ]
then
    echo "Syntax $0 user password organization hostname_or_IP_address"
    exit 1
fi

auth=$(echo -n "$user@$org:$password" |base64)

curl -I -k --header "Accept: application/*;version=32.0" \
    --header "Authorization: Basic $auth" \
    --request POST https://$IP/api/sessions

# If successful, the output of this command will include lines like the following
# X-VCLOUD-AUTHORIZATION: 08a321735de84f1d9ec80c3b3e18fa8b
# X-VMWARE-VCLOUD-ACCESS-TOKEN: eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhZG1pbmlzdHJhdG9yI[562 more characters]
#
# The string after `X-VCLOUD-AUTHORIZATION:` is the old (deprecated) token.
# The 612-character string after `X-VMWARE-VCLOUD-ACCESS-TOKEN` is the bearer token

# For VCD version 10.0+, you can use one of the following commands instead.

# PROVIDER:
# curl -I -k --header "Accept: application/*;version=33.0" \
#    --header "Authorization: Basic $auth" \
#    --request POST https://$IP/cloudapi/1.0.0/sessions/provider

# TENANT
# curl -I -k --header "Accept: application/*;version=33.0" \
#    --header "Authorization: Basic $auth" \
#    --request POST https://$IP/cloudapi/1.0.0/sessions
#
# The cloudapi requests will only return the bearer token

