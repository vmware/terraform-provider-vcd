package govcloudair

import (
	"bytes"
	"encoding/xml"
	"fmt"
	types "github.com/ukcloud/govcloudair/types/v56"
	"net/url"
	"strconv"
	"strings"
)

//Creates an Organization based on settings, network, and org name
func CreateOrg(c *VCDClient, name string, fullName string, isEnabled bool, settings map[string]string) (Task, error) {

	orgSettings := getOrgSettings(settings)

	vcomp := &types.AdminOrg{
		Xmlns:       "http://www.vmware.com/vcloud/v1.5",
		Name:        name,
		IsEnabled:   isEnabled,
		FullName:    fullName,
		OrgSettings: orgSettings,
	}

	output, _ := xml.MarshalIndent(vcomp, "  ", "    ")

	u := c.Client.HREF
	u.Path += "/admin/orgs"

	b := bytes.NewBufferString(xml.Header + string(output))

	req := c.Client.NewRequest(map[string]string{}, "POST", u, b)

	req.Header.Add("Content-Type", "application/vnd.vmware.admin.organization+xml")

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return Task{}, fmt.Errorf("error instantiating a new Org: %s", err)
	}

	task := NewTask(&c.Client)

	if err = decodeBody(resp, task.Task); err != nil {
		return Task{}, fmt.Errorf("error decoding task response: %s", err)
	}

	return *task, nil

}

//Gets Org HREF as a string from the organization name
func getOrgHREF(c *VCDClient, orgname string) (string, error) {
	s := c.Client.HREF
	s.Path += "/org"
	req := c.Client.NewRequest(map[string]string{}, "GET", s, nil)

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return "", fmt.Errorf("error retreiving vdc: %s", err)
	}

	orgList := new(types.OrgList)
	if err = decodeBody(resp, orgList); err != nil {
		return "", fmt.Errorf("error decoding vdc response: %s", err)
	}

	for _, a := range orgList.Org {
		if a.Name == orgname {
			return a.HREF, nil
		}
	}

	return "", fmt.Errorf("Couldn't find org with name: %s", orgname)

}

//If user specifies valid organization name, then this returns a organization object
func GetOrgFromName(c *VCDClient, orgname string) (Org, error) {

	o := NewOrg(&c.Client)

	HREF, err := getOrgHREF(c, orgname)
	if err != nil {
		return Org{}, fmt.Errorf("Cannot find OrgHREF: %s", err)
	}

	u, err := url.ParseRequestURI(HREF)
	if err != nil {
		return Org{}, fmt.Errorf("Error parsing org href: %v", err)
	}

	req := c.Client.NewRequest(map[string]string{}, "GET", *u, nil)

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return Org{}, fmt.Errorf("error retreiving org: %s", err)
	}

	if err = decodeBody(resp, o.Org); err != nil {
		return Org{}, fmt.Errorf("error decoding org response: %s", err)
	}

	return *o, nil
}

//If user specifies admin Org object, then this returns an analogous organization object
func GetOrgFromAdminOrg(c *VCDClient, adminOrg AdminOrg) (Org, error) {
	o := NewOrg(&c.Client)

	s := c.Client.HREF
	s.Path += "/org/" + adminOrg.AdminOrg.ID[15:]

	req := c.Client.NewRequest(map[string]string{}, "GET", s, nil)

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return Org{}, fmt.Errorf("error fetching Org : %s", err)
	}

	if err = decodeBody(resp, o.Org); err != nil {
		return Org{}, fmt.Errorf("error decoding org response: %s", err)
	}

	return *o, nil
}

//If user specifies valid organization name, then this returns an admin organization object
func GetAdminOrgFromName(c *VCDClient, orgname string) (AdminOrg, error) {
	o := NewAdminOrg(&c.Client)

	HREF, err := getOrgHREF(c, orgname)
	if err != nil {
		return AdminOrg{}, fmt.Errorf("Cannot find OrgHREF: %s", err)
	}

	u := c.Client.HREF
	u.Path += "/admin/org/" + strings.Split(HREF, "/org/")[1]

	req := c.Client.NewRequest(map[string]string{}, "GET", u, nil)

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return AdminOrg{}, fmt.Errorf("error retreiving org: %s", err)
	}

	if err = decodeBody(resp, o.AdminOrg); err != nil {
		return AdminOrg{}, fmt.Errorf("error decoding org response: %s", err)
	}

	return *o, nil
}

//If user specifies a valid organization, then this returns an analogous admin organization object
func GetAdminOrgFromOrg(c *VCDClient, org Org) (AdminOrg, error) {
	o := NewAdminOrg(&c.Client)

	s := c.Client.HREF
	s.Path += "/admin/org/" + org.Org.ID[15:]

	req := c.Client.NewRequest(map[string]string{}, "GET", s, nil)

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return AdminOrg{}, fmt.Errorf("error fetching Org : %s", err)
	}

	if err = decodeBody(resp, o.AdminOrg); err != nil {
		return AdminOrg{}, fmt.Errorf("error decoding org response: %s", err)
	}

	return *o, nil
}

//Fetches an org using the Org ID, which is the UUID in the Org HREF. Returns it in admin org form
func GetAdminOrgById(c *VCDClient, orgId string) (AdminOrg, error) {
	u := c.Client.HREF
	u.Path += "/admin/org/" + orgId

	req := c.Client.NewRequest(map[string]string{}, "GET", u, nil)

	req.Header.Add("Content-Type", "application/vnd.vmware.vcloud.org+xml")

	resp, err := checkResp(c.Client.Http.Do(req))
	if err != nil {
		return AdminOrg{}, fmt.Errorf("error getting Org %s: %s", orgId, err)
	}

	org := NewAdminOrg(&c.Client)

	if err = decodeBody(resp, org.AdminOrg); err != nil {
		return AdminOrg{}, fmt.Errorf("error decoding org response: %s", err)
	}

	return *org, nil
}

func getOrgSettings(settings map[string]string) *types.OrgSettings {
	var orgSettings *types.OrgSettings

	generalSettings := new(types.OrgGeneralSettings)

	if val, ok := settings["CanPublishCatalogs"]; ok {
		canPublishCatalogs, _ := strconv.ParseBool(val)
		generalSettings.CanPublishCatalogs = canPublishCatalogs
	}
	if val, ok := settings["DeployedVMQuota"]; ok {
		DeployedVMQuota, _ := strconv.Atoi(val)
		generalSettings.DeployedVMQuota = DeployedVMQuota
	}
	if val, ok := settings["StoredVMQuota"]; ok {
		StoredVMQuota, _ := strconv.Atoi(val)
		generalSettings.StoredVMQuota = StoredVMQuota
	}
	if val, ok := settings["UseServerBootSequence"]; ok {
		UseServerBootSequence, _ := strconv.ParseBool(val)
		generalSettings.UseServerBootSequence = UseServerBootSequence
	}
	if val, ok := settings["DelayAfterPowerOnSeconds"]; ok {
		DelayAfterPowerOnSeconds, _ := strconv.Atoi(val)
		generalSettings.DelayAfterPowerOnSeconds = DelayAfterPowerOnSeconds
	}

	//vappLeaseSettings := &types.VAppLeaseSettings{}

	// if val, ok := settings["DeleteOnStorageLeaseExpiration"]; ok {
	// 	DeleteOnStorageLeaseExpiration, _ := strconv.ParseBool(val)
	// 	vappLeaseSettings.DeleteOnStorageLeaseExpiration = DeleteOnStorageLeaseExpiration
	// }
	// if val, ok := settings["DeploymentLeaseSeconds"]; ok {
	// 	DeploymentLeaseSeconds, _ := strconv.Atoi(val)
	// 	vappLeaseSettings.DeploymentLeaseSeconds = DeploymentLeaseSeconds
	// }
	// if val, ok := settings["StorageLeaseSeconds"]; ok {
	// 	StorageLeaseSeconds, _ := strconv.Atoi(val)
	// 	vappLeaseSettings.StorageLeaseSeconds = StorageLeaseSeconds
	// }

	vappTemplateLeaseSettings := new(types.VAppTemplateLeaseSettings)
	if val, ok := settings["DeleteOnStorageLeaseExpiration"]; ok {
		DeleteOnStorageLeaseExpiration, _ := strconv.ParseBool(val)
		vappTemplateLeaseSettings.DeleteOnStorageLeaseExpiration = DeleteOnStorageLeaseExpiration
	}
	if val, ok := settings["StorageLeaseSeconds"]; ok {
		StorageLeaseSeconds, _ := strconv.Atoi(val)
		vappTemplateLeaseSettings.StorageLeaseSeconds = StorageLeaseSeconds
	}

	orgSettings = &types.OrgSettings{
		General:      generalSettings,
		VAppTemplate: vappTemplateLeaseSettings,
		//VappLease: 	  vappLeaseSettings,
	}

	return orgSettings
}
