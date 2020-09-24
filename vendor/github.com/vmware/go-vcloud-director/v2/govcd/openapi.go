package govcd

/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/peterhellberg/link"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// This file contains generalised low level methods to interact with VCD OpenAPI REST endpoints as documented in
// https://{VCD_HOST}/docs. In addition to this there are OpenAPI browser endpoints for tenant and provider
// respectively https://{VCD_HOST}/api-explorer/tenant/tenant-name and https://{VCD_HOST}/api-explorer/provider .
// OpenAPI has functions supporting below REST methods:
// GET /items (gets a slice of types like `[]types.OpenAPIEdgeGateway` or even `[]json.RawMessage` to process JSON as text.
// POST /items - creates an item
// PUT /items/URN - updates an item with specified URN
// GET /items/URN - retrieves an item with specified URN
// DELETE /items/URN - deletes an item with specified URN
//
// GET endpoints support FIQL for filtering in field `filter`. (FIQL IETF doc - https://tools.ietf.org/html/draft-nottingham-atompub-fiql-00)
// Not all API fields are supported for FIQL filtering and sometimes they return odd errors when filtering is
// unsupported. No exact documentation exists so far.
//
// OpenAPI versioning.
// OpenAPI was introduced in VCD 9.5 (with API version 31.0). Endpoints are being added with each VCD iteration.
// Internally hosted documentation (https://HOSTNAME/docs/) can be used to check which endpoints where introduced in
// which VCD API version.
// Additionally each OpenAPI endpoint has a semantic version in its path (e.g.
// https://HOSTNAME/cloudapi/1.0.0/auditTrail). This versioned endpoint should ensure compatibility as VCD evolves.

// OpenApiIsSupported allows to check whether VCD supports OpenAPI. Each OpenAPI endpoint however is introduced with
// different VCD API versions so this is just a general check if OpenAPI is supported at all. Particular endpoint
// introduction version can be checked in self hosted docs (https://HOSTNAME/docs/)
func (client *Client) OpenApiIsSupported() bool {
	// OpenAPI was introduced in VCD 9.5+ (API version 31.0+)
	return client.APIVCDMaxVersionIs(">= 31")
}

// OpenApiBuildEndpoint helps to construct OpenAPI endpoint by using already configured VCD HREF while requiring only
// the last bit for endpoint. This is a variadic function and multiple pieces can be supplied for convenience. Leading
// '/' is added automatically.
// Sample URL construct: https://HOST/cloudapi/endpoint
func (client *Client) OpenApiBuildEndpoint(endpoint ...string) (*url.URL, error) {
	endpointString := client.VCDHREF.Scheme + "://" + client.VCDHREF.Host + "/cloudapi/" + strings.Join(endpoint, "")
	urlRef, err := url.ParseRequestURI(endpointString)
	if err != nil {
		return nil, fmt.Errorf("error formatting OpenAPI endpoint: %s", err)
	}
	return urlRef, nil
}

// OpenApiGetAllItems retrieves and accumulates all pages then parsing them to a single 'outType' object. It works by at
// first crawling pages and accumulating all responses into []json.RawMessage (as strings). Because there is no
// intermediate unmarshalling to exact `outType` for every page it unmarshals into response struct in one go. 'outType'
// must be a slice of object (e.g. []*types.OpenAPIEdgeGateway) because this response contains slice of structs.
func (client *Client) OpenApiGetAllItems(apiVersion string, urlRef *url.URL, queryParams url.Values, outType interface{}) error {
	util.Logger.Printf("[TRACE] Getting all items from endpoint %s for parsing into %s type\n",
		urlRef.String(), reflect.TypeOf(outType))

	if !client.OpenApiIsSupported() {
		return fmt.Errorf("OpenAPI is not supported on this VCD version")
	}

	// Perform API call to initial endpoint. The function call recursively follows pages using Link headers "nextPage"
	// until it crawls all results
	responses, err := client.openApiGetAllPages(apiVersion, urlRef, queryParams, outType, nil)
	if err != nil {
		return fmt.Errorf("error getting all pages for endpoint %s: %s", urlRef.String(), err)
	}

	// Create a slice of raw JSON messages in text so that they can be unmarshalled to specified `outType` after multiple
	// calls are executed
	var rawJsonBodies []string
	for _, singleObject := range responses {
		rawJsonBodies = append(rawJsonBodies, string(singleObject))
	}

	// rawJsonBodies contains a slice of all response objects and they must be formatted as a JSON slice (wrapped
	// into `[]`, separated with semicolons) so that unmarshalling to specified `outType` works in one go
	allResponses := `[` + strings.Join(rawJsonBodies, ",") + `]`

	// Unmarshal all accumulated responses into `outType`
	if err = json.Unmarshal([]byte(allResponses), &outType); err != nil {
		return fmt.Errorf("error decoding values into type: %s", err)
	}

	return nil
}

// OpenApiGetItem is a low level OpenAPI client function to perform GET request for any item.
// The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}')
// It responds with HTTP 403: Forbidden - If the user is not authorized or the entity does not exist. When HTTP 403 is
// returned this function returns "ErrorEntityNotFound: API_ERROR" so that one can use ContainsNotFound(err) to
// differentiate when an objects was not found from any other error.
func (client *Client) OpenApiGetItem(apiVersion string, urlRef *url.URL, params url.Values, outType interface{}) error {
	util.Logger.Printf("[TRACE] Getting item from endpoint %s with expected response of type %s",
		urlRef.String(), reflect.TypeOf(outType))

	if !client.OpenApiIsSupported() {
		return fmt.Errorf("OpenAPI is not supported on this VCD version")
	}

	req := client.newOpenApiRequest(apiVersion, params, http.MethodGet, urlRef, nil)
	resp, err := client.Http.Do(req)
	if err != nil {
		return fmt.Errorf("error performing GET request to %s: %s", urlRef.String(), err)
	}

	// Bypassing the regular path using function checkRespWithErrType and returning parsed error directly
	// HTTP 403: Forbidden - is returned if the user is not authorized or the entity does not exist.
	if resp.StatusCode == http.StatusForbidden {
		err := ParseErr(types.BodyTypeJSON, resp, &types.OpenApiError{})
		closeErr := resp.Body.Close()
		return fmt.Errorf("%s: %s [body close error: %s]", ErrorEntityNotFound, err, closeErr)
	}

	// resp is ignored below because it is the same as above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &types.OpenApiError{})

	// Any other error occurred
	if err != nil {
		return fmt.Errorf("error in HTTP GET request: %s", err)
	}

	if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
		return fmt.Errorf("error decoding JSON response after GET: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	return nil
}

// OpenApiPostItemSync is a low level OpenAPI client function to perform POST request for items that support synchronous
// requests. The urlRef must point to POST endpoint (e.g. '/1.0.0/edgeGateways') that supports synchronous requests. It
// will return an error when endpoint does not support synchronous requests (HTTP response status code is not 201).
// Response will be unmarshalled into outType.
//
// Note. Even though it may return error if the item does not support synchronous request - the object may still be
// created. OpenApiPostItem would handle both cases and always return created item.
func (client *Client) OpenApiPostItemSync(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}) error {
	util.Logger.Printf("[TRACE] Posting %s item to endpoint %s with expected response of type %s",
		reflect.TypeOf(payload), urlRef.String(), reflect.TypeOf(outType))

	if !client.OpenApiIsSupported() {
		return fmt.Errorf("OpenAPI is not supported on this VCD version")
	}

	resp, err := client.openApiPerformPostPut(http.MethodPost, apiVersion, urlRef, params, payload)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		util.Logger.Printf("[TRACE] Synchronous task expected (HTTP status code 201). Got %d", resp.StatusCode)

	}

	if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
		return fmt.Errorf("error decoding JSON response after POST: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	return nil
}

// OpenApiPostItemAsync is a low level OpenAPI client function to perform POST request for items that support
// asynchronous requests. The urlRef must point to POST endpoint (e.g. '/1.0.0/edgeGateways') that supports asynchronous
// requests. It will return an error if item does not support asynchronous request (does not respond with HTTP 202).
//
// Note. Even though it may return error if the item does not support asynchronous request - the object may still be
// created. OpenApiPostItem would handle both cases and always return created item.
func (client *Client) OpenApiPostItemAsync(apiVersion string, urlRef *url.URL, params url.Values, payload interface{}) (Task, error) {
	util.Logger.Printf("[TRACE] Posting async %s item to endpoint %s with expected task response",
		reflect.TypeOf(payload), urlRef.String())

	if !client.OpenApiIsSupported() {
		return Task{}, fmt.Errorf("OpenAPI is not supported on this VCD version")
	}

	resp, err := client.openApiPerformPostPut(http.MethodPost, apiVersion, urlRef, params, payload)
	if err != nil {
		return Task{}, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return Task{}, fmt.Errorf("POST request expected async task (HTTP response 202), got %d", resp.StatusCode)
	}

	err = resp.Body.Close()
	if err != nil {
		return Task{}, fmt.Errorf("error closing response body: %s", err)
	}

	// Asynchronous case returns "Location" header pointing to XML task
	taskUrl := resp.Header.Get("Location")
	if taskUrl == "" {
		return Task{}, fmt.Errorf("unexpected empty task HREF")
	}
	task := NewTask(client)
	task.Task.HREF = taskUrl

	return *task, nil
}

// OpenApiPostItem is a low level OpenAPI client function to perform POST request for item supporting synchronous or
// asynchronous requests. The urlRef must point to POST endpoint (e.g. '/1.0.0/edgeGateways'). When a task is
// synchronous - it will track task until it is finished and pick reference to marshal outType.
func (client *Client) OpenApiPostItem(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}) error {
	util.Logger.Printf("[TRACE] Posting %s item to endpoint %s with expected response of type %s",
		reflect.TypeOf(payload), urlRef.String(), reflect.TypeOf(outType))

	if !client.OpenApiIsSupported() {
		return fmt.Errorf("OpenAPI is not supported on this VCD version")
	}

	resp, err := client.openApiPerformPostPut(http.MethodPost, apiVersion, urlRef, params, payload)
	if err != nil {
		return err
	}

	// Handle two cases of API behaviour - synchronous (response status code is 201) and asynchronous (response status
	// code 202)
	switch resp.StatusCode {
	// Asynchronous case - must track task and get item HREF from there
	case http.StatusAccepted:
		taskUrl := resp.Header.Get("Location")
		util.Logger.Printf("[TRACE] Asynchronous task detected, tracking task with HREF: %s", taskUrl)
		task := NewTask(client)
		task.Task.HREF = taskUrl
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error waiting completion of task (%s): %s", taskUrl, err)
		}

		// Here we have to find the resource once more to return it populated.
		// Task Owner ID is the ID of created object. ID must be used (although HREF exists in task) because HREF points to
		// old XML API and here we need to pull data from OpenAPI.

		newObjectUrl, _ := url.ParseRequestURI(urlRef.String() + task.Task.Owner.ID)
		err = client.OpenApiGetItem(apiVersion, newObjectUrl, nil, outType)
		if err != nil {
			return fmt.Errorf("error retrieving item after creation: %s", err)
		}

		// Synchronous task - new item body is returned in response of HTTP POST request
	case http.StatusCreated:
		util.Logger.Printf("[TRACE] Synchronous task detected, marshalling outType '%s'", reflect.TypeOf(outType))
		if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
			return fmt.Errorf("error decoding JSON response after POST: %s", err)
		}
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	return nil
}

// OpenApiPutItemSync is a low level OpenAPI client function to perform PUT request for items that support synchronous
// requests. The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}') and support synchronous
// requests. It will return an error when endpoint does not support synchronous requests (HTTP response status code is not 201).
// Response will be unmarshalled into outType.
//
// Note. Even though it may return error if the item does not support synchronous request - the object may still be
// updated. OpenApiPutItem would handle both cases and always return updated item.
func (client *Client) OpenApiPutItemSync(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}) error {
	util.Logger.Printf("[TRACE] Putting %s item to endpoint %s with expected response of type %s",
		reflect.TypeOf(payload), urlRef.String(), reflect.TypeOf(outType))

	if !client.OpenApiIsSupported() {
		return fmt.Errorf("OpenAPI is not supported on this VCD version")
	}

	resp, err := client.openApiPerformPostPut(http.MethodPut, apiVersion, urlRef, params, payload)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		util.Logger.Printf("[TRACE] Synchronous task expected (HTTP status code 201). Got %d", resp.StatusCode)

	}

	if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
		return fmt.Errorf("error decoding JSON response after PUT: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	return nil
}

// OpenApiPutItemAsync is a low level OpenAPI client function to perform PUT request for items that support asynchronous
// requests. The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}') that supports asynchronous
// requests. It will return an error if item does not support asynchronous request (does not respond with HTTP 202).
//
// Note. Even though it may return error if the item does not support asynchronous request - the object may still be
// created. OpenApiPutItem would handle both cases and always return created item.
func (client *Client) OpenApiPutItemAsync(apiVersion string, urlRef *url.URL, params url.Values, payload interface{}) (Task, error) {
	util.Logger.Printf("[TRACE] Putting async %s item to endpoint %s with expected task response",
		reflect.TypeOf(payload), urlRef.String())

	if !client.OpenApiIsSupported() {
		return Task{}, fmt.Errorf("OpenAPI is not supported on this VCD version")
	}
	resp, err := client.openApiPerformPostPut(http.MethodPut, apiVersion, urlRef, params, payload)
	if err != nil {
		return Task{}, err
	}

	if resp.StatusCode != http.StatusAccepted {
		return Task{}, fmt.Errorf("PUT request expected async task (HTTP response 202), got %d", resp.StatusCode)
	}

	err = resp.Body.Close()
	if err != nil {
		return Task{}, fmt.Errorf("error closing response body: %s", err)
	}

	// Asynchronous case returns "Location" header pointing to XML task
	taskUrl := resp.Header.Get("Location")
	if taskUrl == "" {
		return Task{}, fmt.Errorf("unexpected empty task HREF")
	}
	task := NewTask(client)
	task.Task.HREF = taskUrl

	return *task, nil
}

// OpenApiPutItem is a low level OpenAPI client function to perform PUT request for any item.
// The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}')
// It handles synchronous and asynchronous tasks. When a task is synchronous - it will block until it is finished.
func (client *Client) OpenApiPutItem(apiVersion string, urlRef *url.URL, params url.Values, payload, outType interface{}) error {
	util.Logger.Printf("[TRACE] Putting %s item to endpoint %s with expected response of type %s",
		reflect.TypeOf(payload), urlRef.String(), reflect.TypeOf(outType))

	if !client.OpenApiIsSupported() {
		return fmt.Errorf("OpenAPI is not supported on this VCD version")
	}
	resp, err := client.openApiPerformPostPut(http.MethodPut, apiVersion, urlRef, params, payload)

	if err != nil {
		return err
	}

	// Handle two cases of API behaviour - synchronous (response status code is 201) and asynchronous (response status
	// code 202)
	switch resp.StatusCode {
	// Asynchronous case - must track task and get item HREF from there
	case http.StatusAccepted:
		taskUrl := resp.Header.Get("Location")
		util.Logger.Printf("[TRACE] Asynchronous task detected, tracking task with HREF: %s", taskUrl)
		task := NewTask(client)
		task.Task.HREF = taskUrl
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error waiting completion of task (%s): %s", taskUrl, err)
		}

		// Here we have to find the resource once more to return it populated. Provided params ir ignored for retrieval.
		err = client.OpenApiGetItem(apiVersion, urlRef, nil, outType)
		if err != nil {
			return fmt.Errorf("error retrieving item after updating: %s", err)
		}

		// Synchronous task - new item body is returned in response of HTTP PUT request
	case http.StatusOK:
		util.Logger.Printf("[TRACE] Synchronous task detected, marshalling outType '%s'", reflect.TypeOf(outType))
		if err = decodeBody(types.BodyTypeJSON, resp, outType); err != nil {
			return fmt.Errorf("error decoding JSON response after PUT: %s", err)
		}
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing HTTP PUT response body: %s", err)
	}

	return nil
}

// OpenApiDeleteItem is a low level OpenAPI client function to perform DELETE request for any item.
// The urlRef must point to ID of exact item (e.g. '/1.0.0/edgeGateways/{EDGE_ID}')
// It handles synchronous and asynchronous tasks. When a task is synchronous - it will block until it is finished.
func (client *Client) OpenApiDeleteItem(apiVersion string, urlRef *url.URL, params url.Values) error {
	util.Logger.Printf("[TRACE] Deleting item at endpoint %s", urlRef.String())

	if !client.OpenApiIsSupported() {
		return fmt.Errorf("OpenAPI is not supported on this VCD version")
	}

	// Perform request
	req := client.newOpenApiRequest(apiVersion, params, http.MethodDelete, urlRef, nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		return err
	}

	// resp is ignored below because it would be the same as above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &types.OpenApiError{})
	if err != nil {
		return fmt.Errorf("error in HTTP DELETE request: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	// OpenAPI may work synchronously or asynchronously. When working asynchronously - it will return HTTP 202 and
	// `Location` header will contain reference to task so that it can be tracked. In DELETE case we do not care about any
	// ID so if DELETE operation is synchronous (returns HTTP 201) - the request has already succeeded.
	if resp.StatusCode == http.StatusAccepted {
		taskUrl := resp.Header.Get("Location")
		task := NewTask(client)
		task.Task.HREF = taskUrl
		err = task.WaitTaskCompletion()
		if err != nil {
			return fmt.Errorf("error waiting completion of task (%s): %s", taskUrl, err)
		}
	}

	return nil
}

// openApiPerformPostPut is a shared function for all public PUT and POST function parts - OpenApiPostItemSync,
// OpenApiPostItemAsync, OpenApiPostItem, OpenApiPutItemSync, OpenApiPutItemAsync, OpenApiPutItem
func (client *Client) openApiPerformPostPut(httpMethod string, apiVersion string, urlRef *url.URL, params url.Values, payload interface{}) (*http.Response, error) {
	// Marshal payload if we have one
	var body *bytes.Buffer
	if payload != nil {
		marshaledJson, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("error marshalling JSON data for %s request %s", httpMethod, err)
		}
		body = bytes.NewBuffer(marshaledJson)
	}

	req := client.newOpenApiRequest(apiVersion, params, httpMethod, urlRef, body)
	resp, err := client.Http.Do(req)
	if err != nil {
		return nil, err
	}

	// resp is ignored below because it is the same the one above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &types.OpenApiError{})
	if err != nil {
		return nil, fmt.Errorf("error in HTTP %s request: %s", httpMethod, err)
	}
	return resp, nil
}

// openApiGetAllPages is a recursive function that helps to accumulate responses from multiple pages for GET query. It
// works by at first crawling pages and accumulating all responses into []json.RawMessage (as strings). Because there is
// no intermediate unmarshalling to exact `outType` for every page it can unmarshal into direct `outType` supplied.
// outType must be a slice of object (e.g. []*types.OpenApiRole) because accumulated responses are in JSON list
func (client *Client) openApiGetAllPages(apiVersion string, urlRef *url.URL, queryParams url.Values, outType interface{}, responses []json.RawMessage) ([]json.RawMessage, error) {
	if responses == nil {
		responses = []json.RawMessage{}
	}

	// Perform request
	req := client.newOpenApiRequest(apiVersion, queryParams, http.MethodGet, urlRef, nil)

	resp, err := client.Http.Do(req)
	if err != nil {
		return nil, err
	}

	// resp is ignored below because it is the same as above
	_, err = checkRespWithErrType(types.BodyTypeJSON, resp, err, &types.OpenApiError{})
	if err != nil {
		return nil, fmt.Errorf("error in HTTP GET request: %s", err)
	}

	// Pages will unwrap pagination and keep a slice of raw json message to marshal to specific types
	pages := &types.OpenApiPages{}

	if err = decodeBody(types.BodyTypeJSON, resp, pages); err != nil {
		return nil, fmt.Errorf("error decoding JSON page response: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing response body: %s", err)
	}

	// Accumulate all responses in a single page as JSON text using json.RawMessage
	// After pages are unwrapped one can marshal response into specified type
	// singleQueryResponses := &json.RawMessage{}
	var singleQueryResponses []json.RawMessage
	if err = json.Unmarshal(pages.Values, &singleQueryResponses); err != nil {
		return nil, fmt.Errorf("error decoding values into accumulation type: %s", err)
	}
	responses = append(responses, singleQueryResponses...)

	// Check if there is still 'nextPage' linked and continue accumulating responses if so
	nextPageUrlRef, err := findRelLink("nextPage", resp.Header)
	if err != nil && !IsNotFound(err) {
		return nil, fmt.Errorf("error looking for 'nextPage' in 'Link' header: %s", err)
	}

	if nextPageUrlRef != nil {
		responses, err = client.openApiGetAllPages(apiVersion, nextPageUrlRef, url.Values{}, outType, responses)
		if err != nil {
			return nil, fmt.Errorf("got error on page %d: %s", pages.Page, err)
		}
	}

	return responses, nil
}

// newOpenApiRequest is a low level function used in upstream OpenAPI functions which handles logging and
// authentication for each API request
func (client *Client) newOpenApiRequest(apiVersion string, params url.Values, method string, reqUrl *url.URL, body io.Reader) *http.Request {

	// Add the params to our URL
	reqUrl.RawQuery += params.Encode()

	// If the body contains data - try to read all contents for logging and re-create another
	// io.Reader with all contents to use it down the line
	var readBody []byte
	if body != nil {
		readBody, _ = ioutil.ReadAll(body)
		body = bytes.NewReader(readBody)
	}

	// Build the request, no point in checking for errors here as we're just
	// passing a string version of an url.URL struct and http.NewRequest returns
	// error only if can't process an url.ParseRequestURI().
	req, _ := http.NewRequest(method, reqUrl.String(), body)

	if client.VCDAuthHeader != "" && client.VCDToken != "" {
		// Add the authorization header
		req.Header.Add(client.VCDAuthHeader, client.VCDToken)
		// Add the Accept header for VCD
		acceptMime := types.JSONMime + ";version=" + apiVersion
		req.Header.Add("Accept", acceptMime)
	}

	// Inject JSON mime type
	req.Header.Add("Content-Type", types.JSONMime)

	// Avoids passing data if the logging of requests is disabled
	if util.LogHttpRequest {
		payload := ""
		if req.ContentLength > 0 {
			payload = string(readBody)
		}
		util.ProcessRequestOutput(util.FuncNameCallStack(), method, reqUrl.String(), payload, req)
		debugShowRequest(req, payload)
	}

	return req
}

// findRelLink looks for link to "nextPage" in "Link" header. It will return when first occurrence is found.
// Sample Link header:
// Link: [<https://HOSTNAME/cloudapi/1.0.0/auditTrail?sortAsc=&pageSize=25&sortDesc=&page=7>;rel="lastPage";
// type="application/json";model="AuditTrailEvents" <https://HOSTNAME/cloudapi/1.0.0/auditTrail?sortAsc=&pageSize=25&sortDesc=&page=2>;
// rel="nextPage";type="application/json";model="AuditTrailEvents"]
// Returns *url.Url or ErrorEntityNotFound
func findRelLink(relFieldName string, header http.Header) (*url.URL, error) {
	headerLinks := link.ParseHeader(header)

	for relKeyName, linkAddress := range headerLinks {
		switch {
		// When map key has more than one name (separated by space). In such cases it can have map key as
		// "lastPage nextPage" when nextPage==lastPage or similar and one specific field needs to be matched.
		case strings.Contains(relKeyName, " "):
			relNameSlice := strings.Split(relKeyName, " ")
			for _, oneRelName := range relNameSlice {
				if oneRelName == relFieldName {
					return url.Parse(linkAddress.String())
				}
			}
		case relKeyName == relFieldName:
			return url.Parse(linkAddress.String())
		}
	}

	return nil, ErrorEntityNotFound
}

// jsonRawMessagesToStrings converts []*json.RawMessage to []string
func jsonRawMessagesToStrings(messages []json.RawMessage) []string {
	resultString := make([]string, len(messages))
	for index, message := range messages {
		resultString[index] = string(message)
	}

	return resultString
}

// copyOrNewUrlValues either creates a copy of parameters or instantiates a new url.Values if nil parameters are
// supplied. It helps to avoid mutating supplied parameter when additional values must be injected internally.
func copyOrNewUrlValues(parameters url.Values) url.Values {
	parameterCopy := make(map[string][]string)

	// if supplied parameters are nil - we just return new initialized
	if parameters == nil {
		return parameterCopy
	}

	// Copy URL values
	for key, value := range parameters {
		parameterCopy[key] = value
	}

	return parameterCopy
}

// queryParameterFilterAnd is a helper to append "AND" clause to FIQL filter by using ';' (semicolon) if any values are
// already set in 'filter' value of parameters. If none existed before then 'filter' value will be set.
//
// Note. It does a copy of supplied 'parameters' value and does not mutate supplied original parameters.
func queryParameterFilterAnd(filter string, parameters url.Values) url.Values {
	newParameters := copyOrNewUrlValues(parameters)

	existingFilter := newParameters.Get("filter")
	if existingFilter == "" {
		newParameters.Set("filter", filter)
		return newParameters
	}

	newParameters.Set("filter", existingFilter+";"+filter)
	return newParameters
}
