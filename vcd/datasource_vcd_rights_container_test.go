//go:build role || ALL || functional
// +build role ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type containerInfo struct {
	name      string
	rights    int
	tenants   int
	published bool
}

func getRightsContainerInfo() (map[string]containerInfo, error) {

	var containers = make(map[string]containerInfo)

	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting client configuration: %s", err)
	}
	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %s", err)
	}
	org, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		return nil, fmt.Errorf("org not found : %s", err)
	}

	roles, err := org.GetAllRoles(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving roles: %s", err)
	}
	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles found in Org %s: %s", testConfig.VCD.Org, err)
	}
	rights, err := roles[0].GetRights(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving role %s rights: %s", roles[0].Role.Name, err)
	}
	containers["vcd_role"] = containerInfo{
		name:   roles[0].Role.Name,
		rights: len(rights),
	}

	globalRoles, err := vcdClient.Client.GetAllGlobalRoles(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving global roles: %s", err)
	}
	if len(globalRoles) == 0 {
		return nil, fmt.Errorf("no global roles found:  %s", err)
	}
	rights, err = globalRoles[0].GetRights(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving global role %s rights: %s", globalRoles[0].GlobalRole.Name, err)
	}
	published := false
	if globalRoles[0].GlobalRole.PublishAll != nil {
		published = *globalRoles[0].GlobalRole.PublishAll
	}
	tenants, err := globalRoles[0].GetTenants(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving global role %s tenants: %s", globalRoles[0].GlobalRole.Name, err)
	}
	containers["vcd_global_role"] = containerInfo{
		name:      globalRoles[0].GlobalRole.Name,
		rights:    len(rights),
		tenants:   len(tenants),
		published: published,
	}

	rightBundles, err := vcdClient.Client.GetAllRightsBundles(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving rights bundles : %s", err)
	}
	if len(rightBundles) == 0 {
		return nil, fmt.Errorf("no rights bundles found:  %s", err)
	}
	rights, err = rightBundles[0].GetRights(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving rights bundle %s rights: %s", rightBundles[0].RightsBundle.Name, err)
	}
	tenants, err = rightBundles[0].GetTenants(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving rights bundle %s tenants: %s", globalRoles[0].GlobalRole.Name, err)
	}
	published = false
	if rightBundles[0].RightsBundle.PublishAll != nil {
		published = *rightBundles[0].RightsBundle.PublishAll
	}

	containers["vcd_rights_bundle"] = containerInfo{
		name:      rightBundles[0].RightsBundle.Name,
		rights:    len(rights),
		tenants:   len(tenants),
		published: published,
	}

	simpleRights, err := vcdClient.Client.GetAllRights(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving rights: %s", err)
	}

	for _, right := range simpleRights {
		if len(right.ImpliedRights) > 2 {
			containers["vcd_right"] = containerInfo{
				name:   right.Name,
				rights: len(right.ImpliedRights),
			}
			break
		}
	}

	return containers, nil
}

func TestAccVcdRightsContainers(t *testing.T) {
	preTestChecks(t)

	if !usingSysAdmin() {
		t.Skip("TestAccVcdRightsContainers requires system admin privileges")
		return
	}

	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	containers, err := getRightsContainerInfo()
	if err != nil {
		t.Logf("error retrieving containers: %s", err)
		t.FailNow()
	}
	if len(containers) != 4 {
		t.Logf("retrieved containers don't have all the required elements")
		t.FailNow()
	}

	if vcdTestVerbose {
		for k, v := range containers {
			fmt.Printf("%-20s %-35s %3d (%d)\n", k, v.name, v.rights, v.tenants)
		}
	}
	var (
		roleDataSource            = containers["vcd_role"].name
		rightDataSource           = containers["vcd_right"].name
		rightsBundleDataSource    = containers["vcd_rights_bundle"].name
		roleDef                   = "role-ds"
		globalRoleDef             = "global-role-ds"
		rightDef                  = "right-ds"
		rightsBundleDef           = "rb-ds"
		datasourceRightDef        = "data.vcd_right." + rightDef
		datasourceRoleDef         = "data.vcd_role." + roleDef
		datasourceGlobalRoleDef   = "data.vcd_global_role." + globalRoleDef
		datasourceRightsBundleDef = "data.vcd_rights_bundle." + rightsBundleDef
	)

	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"RightDef":         rightDef,
		"RoleDef":          roleDef,
		"GlobalRoleDef":    globalRoleDef,
		"RightsBundleDef":  rightsBundleDef,
		"RoleName":         roleDataSource,
		"RightName":        rightDataSource,
		"GlobalRoleName":   roleDataSource,
		"RightsBundleName": rightsBundleDataSource,
		"FuncName":         t.Name(),
		"Tags":             "role",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccRightsContainerDS, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(datasourceRoleDef),
					resource.TestCheckResourceAttr(datasourceRoleDef, "name", roleDataSource),
					resource.TestCheckResourceAttr(datasourceRoleDef, "rights.#", fmt.Sprintf("%d", containers["vcd_role"].rights)),

					testAccCheckGlobalRoleExists(datasourceGlobalRoleDef),
					resource.TestCheckResourceAttr(datasourceGlobalRoleDef, "name", roleDataSource),
					resource.TestCheckResourceAttr(datasourceGlobalRoleDef, "rights.#", fmt.Sprintf("%d", containers["vcd_global_role"].rights)),
					resource.TestCheckResourceAttr(datasourceGlobalRoleDef, "publish_to_all_tenants", fmt.Sprintf("%v", containers["vcd_global_role"].published)),
					resource.TestCheckResourceAttr(datasourceGlobalRoleDef, "tenants.#", fmt.Sprintf("%d", containers["vcd_global_role"].tenants)),

					testAccCheckRightsBundleExists(datasourceRightsBundleDef),
					resource.TestCheckResourceAttr(datasourceRightsBundleDef, "name", rightsBundleDataSource),
					resource.TestCheckResourceAttr(datasourceRightsBundleDef, "rights.#", fmt.Sprintf("%d", containers["vcd_rights_bundle"].rights)),
					resource.TestCheckResourceAttr(datasourceRightsBundleDef, "publish_to_all_tenants", fmt.Sprintf("%v", containers["vcd_rights_bundle"].published)),
					resource.TestCheckResourceAttr(datasourceRightsBundleDef, "tenants.#", fmt.Sprintf("%d", containers["vcd_rights_bundle"].tenants)),

					testAccCheckRightExists(datasourceRightDef),
					resource.TestCheckResourceAttr(datasourceRightDef, "name", rightDataSource),
					resource.TestCheckResourceAttr(datasourceRightDef, "implied_rights.#", fmt.Sprintf("%d", containers["vcd_right"].rights)),
				),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckRightExists(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no role ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, err := conn.Client.GetRightById(rs.Primary.ID)
		if err != nil {
			return err
		}
		return err
	}
}

const testAccRightsContainerDS = `
data "vcd_right" "{{.RightDef}}" {
	name = "{{.RightName}}"
}

data "vcd_role" "{{.RoleDef}}" {
  org  = "{{.Org}}"
  name = "{{.RoleName}}"
}

data "vcd_global_role" "{{.GlobalRoleDef}}" {
  name = "{{.GlobalRoleName}}"
}

data "vcd_rights_bundle" "{{.RightsBundleDef}}" {
  name = "{{.RightsBundleName}}"
}
`
