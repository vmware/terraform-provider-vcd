#!/bin/bash
# This script will connect to VCD using username and password,
# and go through the steps to create/authorize/activate a service account.
# Provider is the only one that can create and authorize service accounts, 
# so the credentials must be for the provider.

user=$1
password=$2
org=$3
IP=$4
client_name=$5
role=$6
software_id=$7
software_version=$8
client_uri=$9

if [ -z "$software_id" ]
then
    echo "Syntax $0 user password organization hostname_or_IP_address 
    client_name role(in URN format) software_id(in UUID) [software_version client_uri]"
    exit 1
fi

if [ "$org" = "System" ]
then
    tenant="provider"
else
    tenant="tenant/$org"
fi

# Check if jq is installed with which and exit if not
if ! which jq > /dev/null
then
    echo "jq is not installed. Please install jq and try again."
    exit 1
fi

options=""
os=$(uname -s)
is_linux=$(echo "$os" | grep -i linux)
if [ -n "$is_linux" ]
then
  options="-w 0"
fi

auth=$(echo -n "$user"@System:"$password" |base64 $options)

bearer=$(curl -I -s -k --header "Accept: application/*;version=37.0" \
    --header "Authorization: Basic $auth" \
    --request POST https://$IP/api/login | grep 'x-vmware-vcloud-access-token' \
    | awk -F":" '{print $2}' | sed 's/^ *//g')

auth_header="Authorization: Bearer $bearer"

json="{
    \"client_name\":\"$client_name\",
    \"software_id\":\"$software_id\",
    \"scope\":\"$role\",
    \"software_version\":\"$software_version\",
    \"client_uri\":\"$client_uri\"
}"

echo "Creating service account..."
client_id=$(echo $json | curl -s -k --http1.1 \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    -H "$auth_header" -X POST --data-binary @- \
    https://$IP/oauth/"$tenant"/register | jq -r '.client_id')



if [ "$client_id" = "null" ]
then
    echo "Error: Could not create service account. Check if you have supplied the correct arguments."
    exit 1
fi

echo "Authorizing service account..."
authorization_details=$(curl -k -s --http1.1 \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -H "Accept: application/json" -H "$auth_header" \
    -d "client_id=$client_id" \
    -X POST https://$IP/oauth/"$tenant"/device_authorization \
    | jq -r '.user_code, .device_code, .verification_uri')

user_code=$(echo $authorization_details | awk '{print $1}')
device_code=$(echo $authorization_details | awk '{print $2}')
verification_uri=$(echo $authorization_details | awk '{print $3}')
# Check if any of the three values is equal to null and exit if so
if [ "$user_code" = "null" ] || [ "$device_code" = "null" ] || [ "$verification_uri" = "null" ]
then
    echo "Error: Could not authorize service account. Check if you have supplied the correct arguments."
    exit 1
fi



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
api_token=$(curl -k -s --http1.1 \
    -H "Content-Type: application/x-www-form-urlencoded" \
    -H "Accept: application/json" \
    -d "client_id=$client_id" \
    -d "device_code=$device_code" \
    -d "grant_type=urn:ietf:params:oauth:grant-type:device_code" \
    -X POST https://$IP/oauth/"$tenant"/token | jq -r '.refresh_token' )

# discard all api_token fields except refresh_token but keep it in json format
token_file="{\"refresh_token\":\"$api_token\"}"

echo $token_file > ./token.json 
echo "Service account created successfully and saved as token.json in the current directory."
echo "The client_name is: $client_name"

