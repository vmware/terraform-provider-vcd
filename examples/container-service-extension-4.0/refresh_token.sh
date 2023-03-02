#!/usr/bin/env bash

# This script manages API tokens, used in CSE to manage the CSE Server.

# ----------------------------------------
# Pre-checks
# ----------------------------------------

! which jq > /dev/null && echo "[ERROR] No 'jq' command found, please install it" && exit 1
[ -z "$VCD_PASSWORD" ] && echo "[ERROR] Environment variable VCD_PASSWORD is not set" && exit 1
if [ $# -lt 5 ]; then
  echo "Usage: VCD_PASSWORD=mypass ./$(basename $0) create|destroy vcd_url username org token_name"
  echo "Example to create an API token: VCD_PASSWORD=mypass ./$(basename $0) create https://vcd.my-company.com administrator System TestToken"
  echo "Example to delete it: VCD_PASSWORD=mypass ./$(basename $0) destroy https://vcd.my-company.com administrator System TestToken"
  exit 1
fi

operation="$1"
vcd_url="$2"
user="$3"
org="$4"
token_name="$5"
version_header='Accept: application/*;version=37.0' # 37.0 corresponds to VCD 10.4.0

if [[ ! "$operation" == 'create' ]] && [[ ! "$operation" == 'destroy' ]]; then
  echo "[ERROR] Operation '$operation' not supported, it should be either 'create' or 'destroy'"
  exit 1
fi

# ----------------------------------------
# Auxiliary functions
# ----------------------------------------

# login Sets the global variable bearer_token for other operations to use
login() {
  bearer_token=$(curl -s --insecure --location -I -H "$version_header" \
       -H "Authorization: Basic $(printf '%s@%s:%s' "$user" "$org" "$VCD_PASSWORD" | base64)" \
       -X POST "$vcd_url/api/login" | tr -d '\r' | grep -i 'x-vmware-vcloud-access-token' | cut -d' ' -f2)
  if [ -z "$bearer_token" ]; then
       echo "[ERROR] Authentication error. Check the organization + user + password credentials and try again."
       exit 1
  fi
}

# create Creates an API token and saves it into a file
create() {
  register_response=$(curl -s --insecure --location -H "$version_header" -H "Authorization: Bearer $bearer_token" -H 'Content-Type: application/json' \
           --data "{\"client_name\": \"$token_name\"}" \
           -X POST "$vcd_url/oauth/provider/register")
  client_id=$(echo "$register_response" | jq -r '.client_id')
  if [[ -z "$client_id" ]] || [[ "$client_id" == 'null' ]]; then
       echo "[ERROR] Could not register the token '$token_name'. The client ID was not present in response: $register_response"
       exit 1
  fi

  token_response=$(curl -s --insecure --location -H "$version_header" -H "Authorization: Bearer $bearer_token" -H 'Content-Type: application/x-www-form-urlencoded' \
           --data-urlencode 'grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer' \
           --data-urlencode "client_id=$client_id" \
           --data-urlencode "assertion=$bearer_token" \
           -X POST "$vcd_url/oauth/provider/token")
  result_token=$(echo "$token_response" | jq -r '.refresh_token')
  rollback=''
  if [ -z "$result_token" ] || [[ "$result_token" == 'null' ]]; then
       echo "[ERROR] Could not retrieve the token '$token_name'. The token was not present in response: $token_response"
       rollback='true'
  fi

  # In this situation, the token was created but not retrieved, so we need to remove it to restore VCD to its previous state
  if [ -n "$rollback" ]; then
    echo "[INFO] Deleting token '$token_name' as it could not be retrieved after creation..."
    destroy
    exit 1
  fi
  echo "[INFO] Created token with name '$token_name'. Saving into .${user}_${token_name}"
  echo "$result_token" > ".${user}_${token_name}"
}

# destroy Destroys the given API token that should be present in VCD
destroy() {
  get_token_response=$(curl -s --insecure --location -H "$version_header" -H "Authorization: Bearer $bearer_token" -H 'Content-Type: application/json' \
       -X GET "$vcd_url/cloudapi/1.0.0/tokens?page=1&pageSize=1&filterEncoded=true&filter=(name==$token_name;(type==PROXY,type==REFRESH))&sortAsc=name")
  token_id=$(echo "$get_token_response" | jq -r '.values[0].id')
  if [[ -z "$token_id" ]] || [[ "$token_id" == 'null' ]]; then
       echo "[ERROR] Could not get the token '$token_name'. The token ID was not present in response: $get_token_response"
       exit 1
  fi

  delete_token_response=$(curl -s -I --insecure --location -H "$version_header" -H "Authorization: Bearer $bearer_token" -H 'Content-Type: application/json' \
       -X DELETE "$vcd_url/cloudapi/1.0.0/tokens/$token_id")
  delete_result=$(echo "$delete_token_response" | grep '204')
  if [ -z "$delete_result" ]; then
       echo "[ERROR] Could not delete the token '$token_name': $delete_token_response"
       exit 1
  fi
  echo "[INFO] Deleted token '$token_name' with ID '$token_id'"
  rm -f ".${user}_${token_name}"
}

# ----------------------------------------
# Main
# ----------------------------------------

login
if [ "$operation" == 'create' ]; then
  create
fi

if [ "$operation" == 'destroy' ]; then
  destroy
fi