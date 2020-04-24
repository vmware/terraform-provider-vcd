#!/bin/bash
# This script will connect to the vCD using username and password,
# and show the header that contains an authorization token.
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

curl -I -k --header "Accept: application/*;version=31.0" \
    --header "Authorization: Basic $auth" \
    --request POST https://$IP/api/sessions

# If successful, the output of this command will include a line like the following
# x-vcloud-authorization: 08a321735de84f1d9ec80c3b3e18fa8b
#
# The string after `x-vcloud-authorization:` is the token.
