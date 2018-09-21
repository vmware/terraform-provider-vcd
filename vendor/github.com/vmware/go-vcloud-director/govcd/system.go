package govcd

import (
	"bytes"
	"encoding/xml"
	"fmt"
	types "github.com/vmware/go-vcloud-director/types/v56"
	"net/url"
	"strings"
)

// Creates an Organization based on settings, network, and org name.
// The Organization created will have these settings specified in the
// settings parameter. The settings variable is defined in types.go.
// Method will fail unless user has an admin token.
func CreateOrg(vcdClient *VCDClient, name string, fullName string, isEnabled bool, settings *types.OrgSettings) (Task, error) {
	vcomp := &types.AdminOrg{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        name,
		IsEnabled:   isEnabled,
		FullName:    fullName,
		OrgSettings: settings,
	}
	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")
	xmlData := bytes.NewBufferString(xml.Header + string(output))
	// Make Request
	orgCreateHREF := vcdClient.Client.VCDHREF
	orgCreateHREF.Path += "/admin/orgs"
	req := vcdClient.Client.NewRequest(map[string]string{}, "POST", orgCreateHREF, xmlData)
	req.Header.Add("Content-Type", "application/vnd.vmware.admin.organization+xml")
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error instantiating a new Org: %s", err)
	}

	task := NewTask(&vcdClient.Client)
	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}
	return *task, nil
}

// If user specifies a valid organization name, then this returns a
// organization object. If no valid org is found, it returns an empty
// org and no error. Otherwise it returns an error and an empty
// Org object
func GetOrgByName(vcdClient *VCDClient, orgname string) (Org, error) {
	orgUrl, err := getOrgHREF(vcdClient, orgname)
	if err != nil {
		return Org{}, nil
	}
	orgHREF, err := url.ParseRequestURI(orgUrl)
	if err != nil {
		return Org{}, fmt.Errorf("Error parsing org href: %v", err)
	}
	req := vcdClient.Client.NewRequest(map[string]string{}, "GET", *orgHREF, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return Org{}, fmt.Errorf("error retreiving org: %s", err)
	}

	org := NewOrg(&vcdClient.Client)
	if err = decodeBody(resp, org.Org); err != nil {
		return Org{}, fmt.Errorf("error decoding org response: %s", err)
	}
	return *org, nil
}

// If user specifies valid organization name,
// then this returns an admin organization object.
// If no valid org is found, it returns an empty
// org and no error. Otherwise returns an empty AdminOrg
// and an error.
func GetAdminOrgByName(vcdClient *VCDClient, orgname string) (AdminOrg, error) {
	orgUrl, err := getOrgHREF(vcdClient, orgname)
	if err != nil {
		return AdminOrg{}, nil
	}
	orgHREF := vcdClient.Client.VCDHREF
	orgHREF.Path += "/admin/org/" + strings.Split(orgUrl, "/org/")[1]
	req := vcdClient.Client.NewRequest(map[string]string{}, "GET", orgHREF, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return AdminOrg{}, fmt.Errorf("error retreiving org: %s", err)
	}
	org := NewAdminOrg(&vcdClient.Client)
	if err = decodeBody(resp, org.AdminOrg); err != nil {
		return AdminOrg{}, fmt.Errorf("error decoding org response: %s", err)
	}
	return *org, nil
}

// Returns the HREF of the org with the name orgname
func getOrgHREF(vcdClient *VCDClient, orgname string) (string, error) {
	orgListHREF := vcdClient.Client.VCDHREF
	orgListHREF.Path += "/org"
	req := vcdClient.Client.NewRequest(map[string]string{}, "GET", orgListHREF, nil)
	resp, err := checkResp(vcdClient.Client.Http.Do(req))
	if err != nil {
		return "", fmt.Errorf("error retreiving org list: %s", err)
	}
	orgList := new(types.OrgList)
	if err = decodeBody(resp, orgList); err != nil {
		return "", fmt.Errorf("error decoding response: %s", err)
	}
	// Look for orgname within OrgList
	for _, a := range orgList.Org {
		if a.Name == orgname {
			return a.HREF, nil
		}
	}
	return "", fmt.Errorf("Couldn't find org with name: %s", orgname)
}
