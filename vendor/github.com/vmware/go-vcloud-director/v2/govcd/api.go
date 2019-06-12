/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

// Package govcd provides a simple binding for vCloud Director REST APIs.
package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Client provides a client to vCloud Director, values can be populated automatically using the Authenticate method.
type Client struct {
	APIVersion    string      // The API version required
	VCDToken      string      // Access Token (authorization header)
	VCDAuthHeader string      // Authorization header
	VCDHREF       url.URL     // VCD API ENDPOINT
	Http          http.Client // HttpClient is the client to use. Default will be used if not provided.
	IsSysAdmin    bool        // flag if client is connected as system administrator

	// MaxRetryTimeout specifies a time limit (in seconds) for retrying requests made by the SDK
	// where vCloud director may take time to respond and retry mechanism is needed.
	// This must be >0 to avoid instant timeout errors.
	MaxRetryTimeout int
}

// Function allow to pass complex values params which shouldn't be encoded like for queries. e.g. /query?filter=(name=foo)
func (cli *Client) NewRequestWitNotEncodedParams(params map[string]string, notEncodedParams map[string]string, method string, reqUrl url.URL, body io.Reader) *http.Request {
	reqValues := url.Values{}

	// Build up our request parameters
	for key, value := range params {
		reqValues.Add(key, value)
	}

	// Add the params to our URL
	reqUrl.RawQuery = reqValues.Encode()

	for key, value := range notEncodedParams {
		if key != "" && value != "" {
			reqUrl.RawQuery += "&" + key + "=" + value
		}
	}

	// Build the request, no point in checking for errors here as we're just
	// passing a string version of an url.URL struct and http.NewRequest returns
	// error only if can't process an url.ParseRequestURI().
	req, _ := http.NewRequest(method, reqUrl.String(), body)

	if cli.VCDAuthHeader != "" && cli.VCDToken != "" {
		// Add the authorization header
		req.Header.Add(cli.VCDAuthHeader, cli.VCDToken)
		// Add the Accept header for VCD
		req.Header.Add("Accept", "application/*+xml;version="+cli.APIVersion)
	}

	// Avoids passing data if the logging of requests is disabled
	if util.LogHttpRequest {
		// Makes a safe copy of the request body, and passes it
		// to the processing function.
		payload := ""
		if req.ContentLength > 0 {
			// We try to convert body to a *bytes.Buffer
			var ibody interface{} = body
			bbody, ok := ibody.(*bytes.Buffer)
			// If the inner object is a bytes.Buffer, we get a safe copy of the data.
			// If it is really just an io.Reader, we don't, as the copy would empty the reader
			if ok {
				payload = bbody.String()
			} else {
				// With this content, we'll know that the payload is not really empty, but
				// it was unavailable due to the body type.
				payload = fmt.Sprintf("<Not retrieved from type %s>", reflect.TypeOf(body))
			}
		}
		util.ProcessRequestOutput(util.FuncNameCallStack(), method, reqUrl.String(), payload, req)
	}
	return req

}

// NewRequest creates a new HTTP request and applies necessary auth headers if
// set.
func (cli *Client) NewRequest(params map[string]string, method string, reqUrl url.URL, body io.Reader) *http.Request {
	return cli.NewRequestWitNotEncodedParams(params, nil, method, reqUrl, body)
}

// ParseErr takes an error XML resp and returns a single string for use in error messages.
func ParseErr(resp *http.Response) error {

	errBody := new(types.Error)

	// if there was an error decoding the body, just return that
	if err := decodeBody(resp, errBody); err != nil {
		util.Logger.Printf("[ParseErr]: unhandled response <--\n%+v\n-->\n", resp)
		return fmt.Errorf("[ParseErr]: error parsing error body for non-200 request: %s (%+v)", err, resp)
	}

	return fmt.Errorf("API Error: %d: %s", errBody.MajorErrorCode, errBody.Message)
}

// decodeBody is used to XML decode a response body
func decodeBody(resp *http.Response, out interface{}) error {

	body, err := ioutil.ReadAll(resp.Body)

	util.ProcessResponseOutput(util.FuncNameCallStack(), resp, fmt.Sprintf("%s", body))
	if err != nil {
		return err
	}

	// Unmarshal the XML.
	if err = xml.Unmarshal(body, &out); err != nil {
		return err
	}

	return nil
}

// checkResp wraps http.Client.Do() and verifies the request, if status code
// is 2XX it passes back the response, if it's a known invalid status code it
// parses the resultant XML error and returns a descriptive error, if the
// status code is not handled it returns a generic error with the status code.
func checkResp(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return resp, err
	}

	switch resp.StatusCode {
	// Valid request, return the response.
	case
		http.StatusOK,        // 200
		http.StatusCreated,   // 201
		http.StatusAccepted,  // 202
		http.StatusNoContent: // 204
		return resp, nil
	// Invalid request, parse the XML error returned and return it.
	case
		http.StatusBadRequest,                  // 400
		http.StatusUnauthorized,                // 401
		http.StatusForbidden,                   // 403
		http.StatusNotFound,                    // 404
		http.StatusMethodNotAllowed,            // 405
		http.StatusNotAcceptable,               // 406
		http.StatusProxyAuthRequired,           // 407
		http.StatusRequestTimeout,              // 408
		http.StatusConflict,                    // 409
		http.StatusGone,                        // 410
		http.StatusLengthRequired,              // 411
		http.StatusPreconditionFailed,          // 412
		http.StatusRequestEntityTooLarge,       // 413
		http.StatusRequestURITooLong,           // 414
		http.StatusUnsupportedMediaType,        // 415
		http.StatusLocked,                      // 423
		http.StatusFailedDependency,            // 424
		http.StatusUpgradeRequired,             // 426
		http.StatusPreconditionRequired,        // 428
		http.StatusTooManyRequests,             // 429
		http.StatusRequestHeaderFieldsTooLarge, // 431
		http.StatusUnavailableForLegalReasons,  // 451
		http.StatusInternalServerError,         // 500
		http.StatusServiceUnavailable,          // 503
		http.StatusGatewayTimeout:              // 504
		return nil, ParseErr(resp)
	// Unhandled response.
	default:
		return nil, fmt.Errorf("unhandled API response, please report this issue, status code: %s", resp.Status)
	}
}

// Helper function creates request, runs it, checks response and parses task from response.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// E.g. client.ExecuteTaskRequest(updateDiskLink.HREF, http.MethodPut, updateDiskLink.Type, "error updating disk: %s", xmlPayload)
func (client *Client) ExecuteTaskRequest(pathURL, requestType, contentType, errorMessage string, payload interface{}) (Task, error) {

	if !isMessageWithPlaceHolder(errorMessage) {
		return Task{}, fmt.Errorf("error message has to include place holder for error")
	}

	resp, err := executeRequest(pathURL, requestType, contentType, payload, client)
	if err != nil {
		return Task{}, fmt.Errorf(errorMessage, err)
	}

	task := NewTask(client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding Task response: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return Task{}, fmt.Errorf(errorMessage, err)
	}

	// The request was successful
	return *task, nil
}

// Helper function creates request, runs it, checks response and do not expect any values from it.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// E.g. client.ExecuteRequestWithoutResponse(catalogItemHREF.String(), http.MethodDelete, "", "error deleting Catalog item: %s", nil)
func (client *Client) ExecuteRequestWithoutResponse(pathURL, requestType, contentType, errorMessage string, payload interface{}) error {

	if !isMessageWithPlaceHolder(errorMessage) {
		return fmt.Errorf("error message has to include place holder for error")
	}

	resp, err := executeRequest(pathURL, requestType, contentType, payload, client)
	if err != nil {
		return fmt.Errorf(errorMessage, err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("error closing response body: %s", err)
	}

	// The request was successful
	return nil
}

// Helper function creates request, runs it, check responses and parses out interface from response.
// pathURL - request URL
// requestType - HTTP method type
// contentType - value to set for "Content-Type"
// errorMessage - error message to return when error happens
// payload - XML struct which will be marshalled and added as body/payload
// out - structure to be used for unmarshalling xml
// E.g. 	unmarshalledAdminOrg := &types.AdminOrg{}
// client.ExecuteRequest(adminOrg.AdminOrg.HREF, http.MethodGet, "", "error refreshing organization: %s", nil, unmarshalledAdminOrg)
func (client *Client) ExecuteRequest(pathURL, requestType, contentType, errorMessage string, payload, out interface{}) (*http.Response, error) {

	if !isMessageWithPlaceHolder(errorMessage) {
		return &http.Response{}, fmt.Errorf("error message has to include place holder for error")
	}

	resp, err := executeRequest(pathURL, requestType, contentType, payload, client)
	if err != nil {
		return resp, fmt.Errorf(errorMessage, err)
	}

	if err = decodeBody(resp, out); err != nil {
		return resp, fmt.Errorf("error decoding response: %s", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return resp, fmt.Errorf("error closing response body: %s", err)
	}

	// The request was successful
	return resp, nil
}

func executeRequest(pathURL, requestType, contentType string, payload interface{}, client *Client) (*http.Response, error) {
	url, _ := url.ParseRequestURI(pathURL)

	var req *http.Request
	switch requestType {
	case http.MethodPost, http.MethodPut:

		marshaledXml, err := xml.MarshalIndent(payload, "  ", "    ")
		if err != nil {
			return &http.Response{}, fmt.Errorf("error marshalling xml data %v", err)
		}
		body := bytes.NewBufferString(xml.Header + string(marshaledXml))

		req = client.NewRequest(map[string]string{}, requestType, *url, body)

	default:
		req = client.NewRequest(map[string]string{}, requestType, *url, nil)
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	return checkResp(client.Http.Do(req))
}

func isMessageWithPlaceHolder(message string) bool {
	err := fmt.Errorf(message, "test error")
	return !strings.Contains(err.Error(), "%!(EXTRA")
}

// combinedTaskErrorMessage is a general purpose function
// that returns the contents of the operation error and, if found, the error
// returned by the associated task
func combinedTaskErrorMessage(task *types.Task, err error) string {
	extendedError := err.Error()
	if task.Error != nil {
		extendedError = fmt.Sprintf("operation error: %s - task error: [%d - %s] %s",
			err, task.Error.MajorErrorCode, task.Error.MinorErrorCode, task.Error.Message)
	}
	return extendedError
}
