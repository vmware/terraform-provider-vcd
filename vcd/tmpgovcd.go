package vcd

import (
	"fmt"
	"net/http"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// TO BE PORTED TO go-vcloud-director as vcdClient.GetOrgList()

func GetOrgList(vcdCli *govcd.VCDClient) (*types.OrgList, error) {
	orgListHREF := vcdCli.Client.VCDHREF
	orgListHREF.Path += "/org"

	orgList := new(types.OrgList)

	_, err := vcdCli.Client.ExecuteRequest(orgListHREF.String(), http.MethodGet,
		"", "error getting list of organizations: %s", nil, orgList)
	if err != nil {
		return nil, err
	}
	return orgList, nil
}

// GetEdgeGatewayRecordsType retrieves a list of edge gateways from VDC
func GetEdgeGatewayRecordsType(client *govcd.Client, vdc *govcd.Vdc, refresh bool) (*types.QueryResultEdgeGatewayRecordsType, error) {

	if refresh {
		err := vdc.Refresh()
		if err != nil {
			return nil, fmt.Errorf("error refreshing vdc: %s", err)
		}
	}
	for _, av := range vdc.Vdc.Link {
		if av.Rel == "edgeGateways" && av.Type == "application/vnd.vmware.vcloud.query.records+xml" {

			edgeGatewayRecordsType := new(types.QueryResultEdgeGatewayRecordsType)

			_, err := client.ExecuteRequest(av.HREF, http.MethodGet,
				"", "error querying edge gateways: %s", nil, edgeGatewayRecordsType)
			if err != nil {
				return nil, err
			}
			return edgeGatewayRecordsType, nil
		}
	}
	return nil, fmt.Errorf("no edge gateway query link found in VDC %s", vdc.Vdc.Name)
}
