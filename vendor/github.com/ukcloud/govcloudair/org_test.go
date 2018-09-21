package govcloudair

import (
	//"fmt"

	"testing"
)

func TestRetrieveVDC(t *testing.T) {
	g, err := GetConfigStruct()
	vcdClient, err := GetTestVCDFromYaml(g)
	if err != nil {
		t.Errorf("Error retrieving vcd client: %v", err)
	}
	err = vcdClient.Authenticate(g.User, g.Password, "System")
	if err != nil {
		t.Errorf("Could not authenticate with user %s password %s url %s: %v", g.User, g.Password, vcdClient.sessionHREF.Path, err)
		t.Errorf("orgname : %s, vcdname : %s", g.Orgname, g.Vdcname)
	}

	org, err := GetOrgFromName(vcdClient, g.Orgname)
	if err != nil {
		t.Errorf("Could not find Org: %v", err)
	}
	vdc, err := org.GetVDCFromName(g.Vdcname)

	if err != nil {
		t.Errorf("Could not find vdc: %v", err)
	}

	if vdc.Vdc.Name != g.Vdcname {
		t.Errorf("Got Wrong VDC: %v", err)
	}
}
