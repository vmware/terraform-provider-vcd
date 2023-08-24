#!/usr/bin/env bash
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
    echo -e "Syntax $0 sys_admin_user sys_admin_password organization hostname_or_IP_address 
    client_name role(in URN format) software_id(in UUID) [software_version client_uri]\n"
    echo "Example: $0 admin password my-org my-vcd.com serviceAccount1 urn:vcloud:role:vApp%20Author f0359776-67ec-4198-a70d-14ce2abba232"
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
    --request POST https://$IP/api/login | grep -i 'x-vmware-vcloud-access-token' \
    | awk -F":" '{print $2}' | sed 's/^ *//g' | tr -d '\n' | tr -d '\r' )

auth_header="Authorization: Bearer $bearer"
if [ "$org" != "System" ]
then
    orgjson=$(curl -s -k --http1.1 \
    -H "Content-Type: application/json" \
    -H "Accept: application/json;version=37.0" \
    -H "$auth_header" \
    https://$IP/cloudapi/1.0.0/orgs)

    # iterate over orgjson values with jq and get the id of the org that matches the org name
    org_id=$(echo $orgjson | jq -r '.values[] | select(.name=="'"$org"'") | .id')
fi

json="{
    \"client_name\":\"$client_name\",
    \"software_id\":\"$software_id\",
    \"scope\":\"$role\",
    \"software_version\":\"$software_version\",
    \"client_uri\":\"$client_uri\"
}"

headers=()
headers+=("-H" "Content-Type: application/json")
headers+=("-H" "Accept: application/json")
headers+=("-H" "$auth_header")
headers+=("--data-binary" "@-")
headers+=("--http1.1")

echo "Creating service account..."
client_id=$(echo $json| curl -k -s \
    "${headers[@]}" \
    https://"$IP"/oauth/"$tenant"/register | jq -r .client_id  | tr -d '\n' | tr -d '\r' )

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
    -X POST https://$IP/oauth/"$tenant"/device_authorization | jq -r '.user_code, .device_code, .verification_uri') 



user_code=$(echo $authorization_details | awk '{print $1}')
device_code=$(echo $authorization_details | awk '{print $2}')
verification_uri=$(echo $authorization_details | awk '{print $3}')
# Check if any of the three values is equal to null and exit if so
if [ "$user_code" = "null" ] || [ "$device_code" = "null" ] || [ "$verification_uri" = "null" ]
then
    echo "Error: Could not authorize service account. Check if you have supplied the correct arguments."
    exit 1
fi

# Send a grant request to cloudapi/1.0.0/deviceLookup/grant
#
activate_json="{
    \"userCode\":\"$user_code\"
}"

if [ -n "$org_id" ]
then
    org_header="X-VMWARE-VCLOUD-TENANT-CONTEXT: $org_id" 
fi

echo "Activating service account..."

echo $activate_json | curl -k -s --http1.1 \
    -H "Content-Type: application/json" \
    -H "Accept: application/json;version=37.0" \
    -H "$org_header" \
    -H "$auth_header" -X POST --data-binary @- \
    -X POST https://$IP/cloudapi/1.0.0/deviceLookup/grant

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

