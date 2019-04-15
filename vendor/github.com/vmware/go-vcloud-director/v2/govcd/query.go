/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type Results struct {
	Results *types.QueryResultRecordsType
	client  *Client
}

func NewResults(cli *Client) *Results {
	return &Results{
		Results: new(types.QueryResultRecordsType),
		client:  cli,
	}
}

func (vcdCli *VCDClient) Query(params map[string]string) (Results, error) {

	req := vcdCli.Client.NewRequest(params, "GET", vcdCli.QueryHREF, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vcdCli.Client.APIVersion)

	return getResult(&vcdCli.Client, req)
}

func (vcdCli *VCDClient) QueryWithNotEncodedParams(params map[string]string, notEncodedParams map[string]string) (Results, error) {
	req := vcdCli.Client.NewRequestWitNotEncodedParams(params, notEncodedParams, "GET", vcdCli.QueryHREF, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vcdCli.Client.APIVersion)

	return getResult(&vcdCli.Client, req)
}

func (vdc *Vdc) Query(params map[string]string) (Results, error) {
	queryUrl := vdc.client.VCDHREF
	queryUrl.Path += "/query"
	req := vdc.client.NewRequest(params, "GET", queryUrl, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vdc.client.APIVersion)

	return getResult(vdc.client, req)
}

func (vdc *Vdc) QueryWithNotEncodedParams(params map[string]string, notEncodedParams map[string]string) (Results, error) {
	queryUrl := vdc.client.VCDHREF
	queryUrl.Path += "/query"
	req := vdc.client.NewRequestWitNotEncodedParams(params, notEncodedParams, "GET", queryUrl, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version="+vdc.client.APIVersion)

	return getResult(vdc.client, req)
}

func getResult(client *Client, request *http.Request) (Results, error) {
	resp, err := checkResp(client.Http.Do(request))
	if err != nil {
		return Results{}, fmt.Errorf("error retrieving query: %s", err)
	}

	results := NewResults(client)

	if err = decodeBody(resp, results.Results); err != nil {
		return Results{}, fmt.Errorf("error decoding query results: %s", err)
	}

	return *results, nil
}
