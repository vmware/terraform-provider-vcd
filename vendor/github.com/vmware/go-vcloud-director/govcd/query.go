/*
 * Copyright 2016 Skyscape Cloud Services.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"

	types "github.com/vmware/go-vcloud-director/types/v56"
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

func (vdcCli *VCDClient) Query(params map[string]string) (Results, error) {

	req := vdcCli.Client.NewRequest(params, "GET", vdcCli.QueryHREF, nil)
	req.Header.Add("Accept", "vnd.vmware.vcloud.org+xml;version=5.5")

	resp, err := checkResp(vdcCli.Client.Http.Do(req))
	if err != nil {
		return Results{}, fmt.Errorf("error retreiving query: %s", err)
	}

	results := NewResults(&vdcCli.Client)

	if err = decodeBody(resp, results.Results); err != nil {
		return Results{}, fmt.Errorf("error decoding query results: %s", err)
	}

	return *results, nil
}
