/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

/*
This file implements SAML authentication flow using Microsoft Active Directory Federation Services
(ADFS). It adds support to authenticate to Cloud Director using SAML authentication (by applying
WithSamlAdfs() configuration option to NewVCDClient function). The identity provider (IdP) must be
Active Directory Federation Services (ADFS) and "/adfs/services/trust/13/usernamemixed" endpoint
must be enabled to make it work. Furthermore username must be supplied in ADFS friendly format -
test@contoso.com' or 'contoso.com\test'.

It works by finding ADFS login endpoint for vCD by querying vCD SAML redirect endpoint
for specific Org and then submits authentication request to "/adfs/services/trust/13/usernamemixed"
endpoint of ADFS server. Using ADFS response it constructs a SIGN token which vCD accepts for the
"/api/sessions". After first initial "login" it grabs the regular X-Vcloud-Authorization token and
uses it for further requests.
More information in vCD documentation:
https://code.vmware.com/docs/10000/vcloud-api-programming-guide-for-service-providers/GUID-335CFC35-7AD8-40E5-91BE-53971937A2BB.html

There is a working code example in /samples/saml_auth_adfs directory how to setup client using SAML
auth.
*/

// authorizeSamlAdfs is the main entry point for SAML authentication on ADFS endpoint
// "/adfs/services/trust/13/usernamemixed"
// Input parameters:
// user - username for authentication to ADFS server (e.g. 'test@contoso.com' or
// 'contoso.com\test')
// pass - password for authentication to ADFS server
// org  - Org to authenticate to
// override_rpt_id - override relaying party trust ID. If it is empty - vCD Entity ID will be used
// as relaying party trust ID
//
// The general concept is to get a SIGN token from ADFS IdP (Identity Provider) and exchange it with
// regular vCD token for further operations. It is documented in
// https://code.vmware.com/docs/10000/vcloud-api-programming-guide-for-service-providers/GUID-335CFC35-7AD8-40E5-91BE-53971937A2BB.html
// This is achieved with the following steps:
// 1 - Lookup vCD Entity ID to use for ADFS authentication or use custom value if overrideRptId
// field is provided
// 2 - Find ADFS server name by querying vCD SAML URL which responds with HTTP redirect (302)
// 3 - Authenticate to ADFS server using vCD SAML Entity ID or custom value if overrideRptId is
// specified Relying Party Trust Identifier
// 4 - Process received ciphers from ADFS server (gzip and base64 encode) so that data can be used
// as SIGN token in vCD
// 5 - Authenticate to vCD using SIGN token in order to receive back regular
// X-Vcloud-Authorization token
// 6 - Set the received X-Vcloud-Authorization for further usage
func (vcdCli *VCDClient) authorizeSamlAdfs(user, pass, org, overrideRptId string) error {
	// Step 1 - find SAML entity ID configured in vCD metadata URL unless overrideRptId is provided
	// Example URL: url.Scheme + "://" + url.Host + "/cloud/org/" + org + "/saml/metadata/alias/vcd"
	samlEntityId := overrideRptId
	var err error
	if overrideRptId == "" {
		samlEntityId, err = getSamlEntityId(vcdCli, org)
		if err != nil {
			return fmt.Errorf("SAML - error getting vCD SAML Entity ID: %s", err)
		}
	}

	// Step 2 - find ADFS server used for SAML by calling vCD SAML endpoint and hoping for a
	// redirect to ADFS server. Example URL:
	// url.Scheme + "://" + url.Host + "/login/my-org/saml/login/alias/vcd?service=tenant:" + org
	adfsAuthEndPoint, err := getSamlAdfsServer(vcdCli, org)
	if err != nil {
		return fmt.Errorf("SAML - error getting IdP (ADFS): %s", err)
	}

	// Step 3 - authenticate to ADFS to receive SIGN token which can be used for vCD authentication
	signToken, err := getSamlAuthToken(vcdCli, user, pass, samlEntityId, adfsAuthEndPoint, org)
	if err != nil {
		return fmt.Errorf("SAML - could not get auth token from IdP (ADFS). Did you specify "+
			"username in ADFS format ('user@contoso.com' or 'contoso.com\\user')? : %s", err)
	}

	// Step 4 - gzip and base64 encode SIGN token so that vCD can understand it
	base64GzippedSignToken, err := gzipAndBase64Encode(signToken)
	if err != nil {
		return fmt.Errorf("SAML - error encoding SIGN token: %s", err)
	}
	util.Logger.Printf("[DEBUG] SAML got SIGN token from IdP '%s' for entity with ID '%s'",
		adfsAuthEndPoint, samlEntityId)

	// Step 5 - authenticate to vCD with SIGN token and receive vCD regular token in exchange
	accessToken, err := authorizeSignToken(vcdCli, base64GzippedSignToken, org)
	if err != nil {
		return fmt.Errorf("SAML - error submitting SIGN token to vCD: %s", err)
	}

	// Step 6 - set regular vCD auth token X-Vcloud-Authorization
	err = vcdCli.SetToken(org, AuthorizationHeader, accessToken)
	if err != nil {
		return fmt.Errorf("error during token-based authentication: %s", err)
	}

	return nil
}

// getSamlAdfsServer finds out Active Directory Federation Service (ADFS) server to use
// for SAML authentication
// It works by temporarily patching existing http.Client behavior to avoid automatically
// following HTTP redirects and searches for Location header after the request to vCD SAML redirect
// address. The URL to search redirect location is:
// url.Scheme + "://" + url.Host + "/login/my-org/saml/login/alias/vcd?service=tenant:" + org
//
// Concurrency note. This function temporarily patches `vcdCli.Client.Http` therefore http.Client
// would not follow redirects during this time. It is however safe as vCDClient is not expected to
// use `http.Client` in any other place before authentication occurs.
func getSamlAdfsServer(vcdCli *VCDClient, org string) (string, error) {
	url := vcdCli.Client.VCDHREF

	// Backup existing http.Client redirect behavior so that it does not follow HTTP redirects
	// automatically and restore it right after this function by using defer. A new http.Client
	// could be spawned here, but the existing one is re-used on purpose to inherit all other
	// settings used for client (timeouts, etc).
	backupRedirectChecker := vcdCli.Client.Http.CheckRedirect

	defer func() {
		vcdCli.Client.Http.CheckRedirect = backupRedirectChecker
	}()

	// Patch http client to avoid following redirects
	vcdCli.Client.Http.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// Construct SAML login URL which should return a redirect to ADFS server
	loginURLString := url.Scheme + "://" + url.Host + "/login/" + org + "/saml/login/alias/vcd"
	loginURL, err := url.Parse(loginURLString)
	if err != nil {
		return "", fmt.Errorf("unable to parse login URL '%s': %s", loginURLString, err)
	}
	util.Logger.Printf("[DEBUG] SAML looking up IdP (ADFS) host redirect in: %s", loginURL.String())

	// Make a request to URL adding unencoded query parameters in the format:
	// "?service=tenant:my-org"
	req := vcdCli.Client.NewRequestWitNotEncodedParams(
		nil, map[string]string{"service": "tenant:" + org}, http.MethodGet, *loginURL, nil)
	httpResponse, err := checkResp(vcdCli.Client.Http.Do(req))
	if err != nil {
		return "", fmt.Errorf("SAML - ADFS server query failed: %s", err)
	}

	err = decodeBody(httpResponse, nil)
	if err != nil {
		return "", fmt.Errorf("SAML - error decoding body: %s", err)
	}

	// httpResponse.Location() returns an error if no 'Location' header is present
	adfsEndpoint, err := httpResponse.Location()
	if err != nil {
		return "", fmt.Errorf("SAML GET request for '%s' did not return HTTP redirect. "+
			"Is SAML configured? Got error: %s", loginURL, err)
	}

	authEndPoint := adfsEndpoint.Scheme + "://" + adfsEndpoint.Host + "/adfs/services/trust/13/usernamemixed"
	util.Logger.Printf("[DEBUG] SAML got IdP login endpoint: %s", authEndPoint)

	return authEndPoint, nil
}

// getSamlEntityId attempts to load vCD hosted SAML metadata from URL:
// url.Scheme + "://" + url.Host + "/cloud/org/" + org + "/saml/metadata/alias/vcd"
// Returns an error if Entity ID is empty
// Sample response body can be found in saml_auth_unit_test.go
func getSamlEntityId(vcdCli *VCDClient, org string) (string, error) {
	url := vcdCli.Client.VCDHREF
	samlMetadataUrl := url.Scheme + "://" + url.Host + "/cloud/org/" + org + "/saml/metadata/alias/vcd"

	metadata := types.VcdSamlMetadata{}
	errString := fmt.Sprintf("SAML - unable to load metadata from URL %s: %%s", samlMetadataUrl)
	_, err := vcdCli.Client.ExecuteRequest(samlMetadataUrl, http.MethodGet, "", errString, nil, &metadata)
	if err != nil {
		return "", err
	}

	samlEntityId := metadata.EntityID
	util.Logger.Printf("[DEBUG] SAML got entity ID: %s", samlEntityId)

	if samlEntityId == "" {
		return "", errors.New("SAML - got empty entity ID")
	}

	return samlEntityId, nil
}

// getSamlAuthToken generates a token request payload using function
// getSamlTokenRequestBody. This request is submitted to ADFS server endpoint
// "/adfs/services/trust/13/usernamemixed" and `RequestedSecurityTokenTxt` is expected in response
// Sample response body can be found in saml_auth_unit_test.go
func getSamlAuthToken(vcdCli *VCDClient, user, pass, samlEntityId, authEndpoint, org string) (string, error) {
	requestBody := getSamlTokenRequestBody(user, pass, samlEntityId, authEndpoint)
	samlTokenRequestBody := strings.NewReader(requestBody)
	tokenRequestResponse := types.AdfsAuthResponseEnvelope{}

	// Post to ADFS endpoint "/adfs/services/trust/13/usernamemixed"
	authEndpointUrl, err := url.Parse(authEndpoint)
	if err != nil {
		return "", fmt.Errorf("SAML - error parsing authentication endpoint %s: %s", authEndpoint, err)
	}
	req := vcdCli.Client.NewRequest(nil, http.MethodPost, *authEndpointUrl, samlTokenRequestBody)
	req.Header.Add("Content-Type", types.SoapXML)
	resp, err := vcdCli.Client.Http.Do(req)
	resp, err = checkRespWithErrType(resp, err, &types.AdfsAuthErrorEnvelope{})
	if err != nil {
		return "", fmt.Errorf("SAML - ADFS token request query failed for RPT ID ('%s'): %s",
			samlEntityId, err)
	}

	err = decodeBody(resp, &tokenRequestResponse)
	if err != nil {
		return "", fmt.Errorf("SAML - error decoding ADFS token request response: %s", err)
	}

	tokenString := tokenRequestResponse.Body.RequestSecurityTokenResponseCollection.RequestSecurityTokenResponse.RequestedSecurityTokenTxt.Text

	return tokenString, nil
}

// authorizeSignToken submits a SIGN token received from ADFS server and gets regular vCD
// "X-Vcloud-Authorization" token in exchange
// Sample response body can be found in saml_auth_unit_test.go
func authorizeSignToken(vcdCli *VCDClient, base64GzippedSignToken, org string) (string, error) {
	url, err := url.Parse(vcdCli.Client.VCDHREF.Scheme + "://" + vcdCli.Client.VCDHREF.Host + "/api/sessions")
	if err != nil {
		return "", fmt.Errorf("SAML error - could not parse URL for posting SIGN token: %s", err)
	}

	signHeader := http.Header{}
	signHeader.Add("Authorization", `SIGN token="`+base64GzippedSignToken+`",org="`+org+`"`)

	req := vcdCli.Client.newRequest(nil, nil, http.MethodPost, *url, nil, vcdCli.Client.APIVersion, signHeader)
	resp, err := checkResp(vcdCli.Client.Http.Do(req))
	if err != nil {
		return "", fmt.Errorf("SAML - error submitting SIGN token for authentication to %s: %s", req.URL.String(), err)
	}
	err = decodeBody(resp, nil)
	if err != nil {
		return "", fmt.Errorf("SAML - error decoding body SIGN token auth response: %s", err)
	}

	accessToken := resp.Header.Get("X-Vcloud-Authorization")
	util.Logger.Printf("[DEBUG] SAML - setting access token for further requests")
	return accessToken, nil
}

// getSamlTokenRequestBody returns a SAML Token request body which is accepted by ADFS server
// endpoint "/adfs/services/trust/13/usernamemixed".
// The payload is not configured as a struct and unmarshalled because Go's unmarshalling changes
// structure so that ADFS does not accept the payload
func getSamlTokenRequestBody(user, password, samlEntityIdReference, adfsAuthEndpoint string) string {
	return `<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" 
	xmlns:a="http://www.w3.org/2005/08/addressing" 
	xmlns:u="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd">
	<s:Header>
		<a:Action s:mustUnderstand="1">http://docs.oasis-open.org/ws-sx/ws-trust/200512/RST/Issue</a:Action>
		<a:ReplyTo>
			<a:Address>http://www.w3.org/2005/08/addressing/anonymous</a:Address>
		</a:ReplyTo>
		<a:To s:mustUnderstand="1">` + adfsAuthEndpoint + `</a:To>
		<o:Security s:mustUnderstand="1" 
			xmlns:o="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd">
			<u:Timestamp u:Id="_0">
				<u:Created>` + time.Now().Format(time.RFC3339) + `</u:Created>
				<u:Expires>` + time.Now().Add(1*time.Minute).Format(time.RFC3339) + `</u:Expires>
			</u:Timestamp>
			<o:UsernameToken>
				<o:Username>` + user + `</o:Username>
				<o:Password o:Type="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-username-token-profile-1.0#PasswordText">` + password + `</o:Password>
			</o:UsernameToken>
		</o:Security>
	</s:Header>
	<s:Body>
		<trust:RequestSecurityToken xmlns:trust="http://docs.oasis-open.org/ws-sx/ws-trust/200512">
			<wsp:AppliesTo xmlns:wsp="http://schemas.xmlsoap.org/ws/2004/09/policy">
				<a:EndpointReference>
					<a:Address>` + samlEntityIdReference + `</a:Address>
				</a:EndpointReference>
			</wsp:AppliesTo>
			<trust:KeySize>0</trust:KeySize>
			<trust:KeyType>http://docs.oasis-open.org/ws-sx/ws-trust/200512/Bearer</trust:KeyType>
			<i:RequestDisplayToken xml:lang="en" 
				xmlns:i="http://schemas.xmlsoap.org/ws/2005/05/identity" />
			<trust:RequestType>http://docs.oasis-open.org/ws-sx/ws-trust/200512/Issue</trust:RequestType>
			<trust:TokenType>http://docs.oasis-open.org/wss/oasis-wss-saml-token-profile-1.1#SAMLV2.0</trust:TokenType>
		</trust:RequestSecurityToken>
	</s:Body>
</s:Envelope>`
}

// gzipAndBase64Encode accepts a string, gzips it and encodes in base64
func gzipAndBase64Encode(text string) (string, error) {
	var gzipBuffer bytes.Buffer
	gz := gzip.NewWriter(&gzipBuffer)
	if _, err := gz.Write([]byte(text)); err != nil {
		return "", fmt.Errorf("error writing to gzip buffer: %s", err)
	}
	if err := gz.Close(); err != nil {
		return "", fmt.Errorf("error closing gzip buffer: %s", err)
	}
	base64GzippedToken := base64.StdEncoding.EncodeToString(gzipBuffer.Bytes())

	return base64GzippedToken, nil
}
