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

options=""
os=$(uname -s)
is_linux=$(echo "$os" | grep -i linux)
if [ -n "$is_linux" ]
then
  options="-w 0"
fi

auth=$(echo -n "$user@$org:$password" |base64 $options)

bearer=$(curl -I -k --header "Accept: application/*;version=37.0" \
    --header "Authorization: Basic $auth" \
    --request POST https://$IP/api/login | grep 'x-vmware-vcloud-access-token' \
    | awk -F":" '{print $2}' | sed 's/^ *//g')

auth_header="Authorization: Bearer $bearer"

# Request user input to create a service account:
# client_name
# software_id (in UUID format), 
# VCD Role in URN format (e.g. urn:vcloud:role:System%20Administrator)
# Software version (optional)
# Software URI (optional)

echo "Enter client name: "
read client_name
echo "Enter software ID (in UUID format): "
read software_id
echo "Enter VCD Role in URN format (e.g. urn:vcloud:role:System%20Administrator): "
read role
echo "Enter software version (optional): "
read software_version
echo "Enter client URI (optional): "
read client_uri

json="{
    \"client_name\":\"$client_name\",
    \"software_id\":\"$software_id\",
    \"scope\":\"$role\",
    \"software_version\":\"$software_version\",
    \"client_uri\":\"$client_uri\"
}"

echo "Creating service account..."
client_id=$(echo $json | curl -k --http1.1 \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "$auth_header" -X POST --data-binary @- \
    https://$IP/oauth/provider/register | jq -r '.client_id')

echo "Authorizing service account..."
authorization_details=$(curl -k --http1.1 \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -H "Accept: application/json" -H "$auth_header" \
    -d "client_id=$client_id" \
    -X POST https://$IP/oauth/provider/device_authorization \
    | jq -r '.user_code, .device_code, .verification_uri')

user_code=$(echo $authorization_details | awk '{print $1}')
device_code=$(echo $authorization_details | awk '{print $2}')
verification_uri=$(echo $authorization_details | awk '{print $3}')

echo "Please accept the service account access request at $verification_uri"
echo "The user code is: $user_code"

# Ask the user to write yes to continue
# If user enters anything other than yes, ask again
while [ "$answer" != "yes" ]
do
    echo "Please enter 'yes' when the service account access was granted: "
    read answer
done

# get the access token
api_token=$(curl -k --http1.1 \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -H "Accept: application/json" \
    -d "client_id=$client_id" \
    -d "device_code=$device_code" \
    -d "grant_type=urn:ietf:params:oauth:grant-type:device_code" \
    -X POST https://$IP/oauth/provider/token | jq -r '.refresh_token' )

# discard all api_token fields except refresh_token but keep it in json format
token_file="{\"refresh_token\":\"$api_token\"}"

echo $token_file > ./token.json 
echo "Service account created successfully and saved as token.json in the current directory."

