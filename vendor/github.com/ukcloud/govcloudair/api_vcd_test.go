package govcloudair

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"strings"
	"testing"
	//. "gopkg.in/check.v1"
	//types "github.com/ukcloud/govcloudair/types/v56"
)

type TestConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Url      string `yaml:"url"`
	Orgname  string `yaml:"org"`
	Vdcname  string `yaml:"vdc"`
}

// tests the ability to authenticate user with given username, password, org, and vdc
func TestAuthenticate(t *testing.T) {
	g, err := GetConfigStruct()
	vcdClient, err := GetTestVCDFromYaml(g)
	if err != nil {
		t.Errorf("Error retrieving vcd client: %v ", err)
	}

	err = vcdClient.Authenticate(g.User, g.Password, "System")
	if err != nil {
		t.Errorf("Could not authenticate with user %s password %s url %s: %v", g.User, g.Password, g.Url, err)
		t.Errorf("orgname : %s, vdcname : %s", g.Orgname, g.Vdcname)
	}

}

func GetConfigStruct() (TestConfig, error) {
	g := TestConfig{}
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return TestConfig{}, fmt.Errorf("could not read config file: %v", err)
	}
	err = yaml.Unmarshal(yamlFile, &g)
	if err != nil {
		return TestConfig{}, fmt.Errorf("could not unmarshal yaml file: %v", err)
	}
	return g, nil
}

func GetTestVCDFromYaml(g TestConfig) (*VCDClient, error) {
	u, err := url.ParseRequestURI(g.Url)
	if err != nil {
		return &VCDClient{}, fmt.Errorf("could not parse Url: %s", err)
	}
	vcdClient := NewVCDClient(*u, true)
	return vcdClient, nil
}

// Creates an organization test, gets it, and then deletes it
func TestCreateOrg(t *testing.T) {
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

	task, err := CreateOrg(vcdClient, orgName, orgName, true, map[string]string{"DeployedVMQuota": "13"})

	if err != nil {
		t.Errorf("Error while creating org: %v", err)
	}

	err = task.WaitTaskCompletion()

	if err != nil {
		t.Errorf("Task could not complete %v", err)
	}

	org, err := GetAdminOrgById(vcdClient, task.Task.ID[15:])

	if err != nil {
		t.Errorf("Org was not created")
	}

	err = org.Delete(true, true)

	if err != nil {
		t.Errorf("Org could not be deleted")
	}
}

// If the user specifies a valid org name, getAdminOrgByID will return a valid AdminOrg object
func TestGetOrg(t *testing.T) {
	g, err := GetConfigStruct()
	vcdClient, err := GetTestVCDFromYaml(g)
	if err != nil {
		t.Errorf("Error retrieving vcd client: %v", err)
	}
	err = vcdClient.Authenticate(g.User, g.Password, "System")
	if err != nil {
		t.Errorf("Could not authenticate: %v", err)
	}

	returnOrg, err := GetAdminOrgFromName(vcdClient, g.Orgname)
	if err != nil {
		t.Errorf("Error getting org %v", err)
	}

	if returnOrg.AdminOrg.Name != g.Orgname {
		t.Error("org ID do not match")
	}
}

func TestRetrieveOrg(t *testing.T) {
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
		t.Errorf("Error retrieving Org: %v", err)
	}
	if org.Org.Name != g.Orgname {
		t.Errorf("Got Wrong Org: %v", err)
	}

	//tests getting of adminOrg type

	adminorg, err := GetAdminOrgFromName(vcdClient, g.Orgname)

	if err != nil {
		t.Errorf("Error converting org to adminOrg : %v", err)
	}

	if adminorg.AdminOrg.Name != g.Orgname || !strings.Contains(adminorg.AdminOrg.HREF, "/admin/") {
		t.Errorf("Did not get AdminOrg Type from Name")
	}

	//tests conversion of org to adminOrg
	adminorg, err = GetAdminOrgFromOrg(vcdClient, org)

	if err != nil {
		t.Errorf("Error converting org to adminOrg : %v", err)
	}

	if adminorg.AdminOrg.Name != g.Orgname || !strings.Contains(adminorg.AdminOrg.HREF, "/admin/") {
		t.Errorf("Did not get AdminOrg Type from Conversion")
	}

	//tests conversion of adminOrg to org

	org, err = GetOrgFromAdminOrg(vcdClient, adminorg)

	if err != nil {
		t.Errorf("Error converting org to adminOrg : %v", err)
	}

	if org.Org.Name != g.Orgname || strings.Contains(org.Org.HREF, "/admin/") {
		t.Errorf("Did not get Org Type From Conversion")
	}
}

const orgName = "test"
