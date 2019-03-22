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
			var ibody interface{}
			ibody = body
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
