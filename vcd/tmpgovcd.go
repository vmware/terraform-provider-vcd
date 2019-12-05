package vcd

import (
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
