//go:build role || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// This file contains the improved version of TestAccVcdRightsContainers.
//
// The old version of the test created one data source for each type (right, role, global role, rights bundle)
// with the risk of taking a different element in different takes, and missing some nuances.
//
// The new version of the test creates one data source for each element of each type (except vcd_right, for which we only
// take the ones with 3 or more implied rights)

type containerInfo struct {
	name      string
	rights    int
	tenants   int
	published bool
}

// getAllRightsContainerInfo retrieves all containers of rights (roles, global roles, rights bundles), and return some
// information about each of them for further processing.
// Additionally, it returns information about 10 rights (using all of them in a test would be impractical)
func getAllRightsContainerInfo() (map[string][]containerInfo, error) {

	var containers = make(map[string][]containerInfo)

	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting client configuration: %s", err)
	}
	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken, testConfig.Provider.ApiTokenFile, testConfig.Provider.ServiceAccountTokenFile)
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
	var rolesInfo []containerInfo
	for _, role := range roles {
		rights, err := role.GetRights(nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving role %s rights: %s", role.Role.Name, err)
		}
		rolesInfo = append(rolesInfo, containerInfo{
			name:   role.Role.Name,
			rights: len(rights),
		})
	}
	containers["vcd_role"] = rolesInfo

	globalRoles, err := vcdClient.Client.GetAllGlobalRoles(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving global roles: %s", err)
	}
	if len(globalRoles) == 0 {
		return nil, fmt.Errorf("no global roles found:  %s", err)
	}
	var globalRolesInfo []containerInfo
	for _, gr := range globalRoles {
		rights, err := gr.GetRights(nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving global role %s rights: %s", gr.GlobalRole.Name, err)
		}
		published := false
		if gr.GlobalRole.PublishAll != nil {
			published = *gr.GlobalRole.PublishAll
		}
		grTenants, err := gr.GetTenants(nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving global role %s tenants: %s", gr.GlobalRole.Name, err)
		}
		numGrTenants := len(grTenants)
		globalRolesInfo = append(globalRolesInfo, containerInfo{
			name:      gr.GlobalRole.Name,
			rights:    len(rights),
			tenants:   numGrTenants,
			published: published,
		})
	}
	containers["vcd_global_role"] = globalRolesInfo

	rightsBundles, err := vcdClient.Client.GetAllRightsBundles(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving rights bundles : %s", err)
	}
	if len(rightsBundles) == 0 {
		return nil, fmt.Errorf("no rights bundles found:  %s", err)
	}
	var rightsBundleInfo []containerInfo
	for _, rightsBundle := range rightsBundles {
		rights, err := rightsBundle.GetRights(nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving rights bundle %s rights: %s", rightsBundle.RightsBundle.Name, err)
		}
		rbTenants, err := rightsBundle.GetTenants(nil)
		if err != nil {
			return nil, fmt.Errorf("error retrieving rights bundle %s tenants: %s", rightsBundle.RightsBundle.Name, err)
		}
		published := false
		if rightsBundle.RightsBundle.PublishAll != nil {
			published = *rightsBundle.RightsBundle.PublishAll
		}

		numRbTenants := len(rbTenants)
		if published {
			numRbTenants = 0
		}
		rightsBundleInfo = append(rightsBundleInfo, containerInfo{
			name:      rightsBundle.RightsBundle.Name,
			rights:    len(rights),
			tenants:   numRbTenants,
			published: published,
		})
	}
	containers["vcd_rights_bundle"] = rightsBundleInfo

	simpleRights, err := vcdClient.Client.GetAllRights(nil)
	if err != nil {
		return nil, fmt.Errorf("error retrieving rights: %s", err)
	}

	var rightsInfo []containerInfo
	for _, right := range simpleRights {
		// Collecting all the right containing at least 3 implied rights.
		// Collecting all would be impractical: there are > 400 rights, of which > 250 have implied rights
		if len(right.ImpliedRights) > 2 {
			rightsInfo = append(rightsInfo, containerInfo{
				name:   right.Name,
				rights: len(right.ImpliedRights),
			})
		}
	}
	containers["vcd_right"] = rightsInfo

	return containers, nil
}

func TestAccVcdRightsContainers(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)

	containers, err := getAllRightsContainerInfo()
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
			fmt.Printf("%-20s\n", k)
			for _, c := range v {
				fmt.Printf("    %-55s %3d (%d)\n", c.name, c.rights, c.tenants)
			}
		}
	}

	quotedString := func(info containerInfo) string {
		return fmt.Sprintf(`"%s"`, info.name)
	}
	roles := ObjectMap[containerInfo, string](containers["vcd_role"], quotedString)
	globalRoles := ObjectMap[containerInfo, string](containers["vcd_global_role"], quotedString)
	rights := ObjectMap[containerInfo, string](containers["vcd_right"], quotedString)
	rightsBundles := ObjectMap[containerInfo, string](containers["vcd_rights_bundle"], quotedString)

	var (
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
		"Org":             testConfig.VCD.Org,
		"RightDef":        rightDef,
		"RoleDef":         roleDef,
		"GlobalRoleDef":   globalRoleDef,
		"RightsBundleDef": rightsBundleDef,
		"Roles":           strings.Join(roles, ","),
		"Rights":          strings.Join(rights, ","),
		"GlobalRoles":     strings.Join(globalRoles, ","),
		"RightsBundles":   strings.Join(rightsBundles, ","),
		"FuncName":        t.Name(),
		"Tags":            "role",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccRightContainerDS, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.ComposeTestCheckFunc(
						testAccCheckGroupRightDataSourceInfo(
							datasourceRoleDef,
							containers["vcd_role"],
							testAccCheckRoleExists)...,
					),
					resource.ComposeTestCheckFunc(
						testAccCheckGroupRightDataSourceInfo(
							datasourceGlobalRoleDef,
							containers["vcd_global_role"],
							testAccCheckGlobalRoleExists)...,
					),
					resource.ComposeTestCheckFunc(
						testAccCheckGroupRightDataSourceInfo(
							datasourceRightDef,
							containers["vcd_right"],
							testAccCheckRightExists)...,
					),
					resource.ComposeTestCheckFunc(
						testAccCheckGroupRightDataSourceInfo(
							datasourceRightsBundleDef,
							containers["vcd_rights_bundle"],
							testAccCheckRightsBundleExists)...,
					),
				),
			},
		},
	})
	postTestChecks(t)
}

type existFunc func(identifier string) resource.TestCheckFunc

// testAccCheckGroupRightDataSourceInfo produces an array of resource.TestCheckFunc, suitable to be fed to resource.ComposeTestCheckFunc
func testAccCheckGroupRightDataSourceInfo(prefix string, data []containerInfo, checkExists existFunc) []resource.TestCheckFunc {
	var checkFuncs []resource.TestCheckFunc
	for i, info := range data {
		def := fmt.Sprintf("%s.%d", prefix, i)
		checkFuncs = append(checkFuncs, checkExists(def))
		checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr(def, "name", info.name))
		if !strings.Contains(prefix, "vcd_right.") {
			checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr(def, "rights.#", fmt.Sprintf("%d", info.rights)))
		}
		if strings.Contains(prefix, "global") || strings.Contains(prefix, "bundle") {
			checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr(def, "publish_to_all_tenants", fmt.Sprintf("%v", info.published)))
			checkFuncs = append(checkFuncs, resource.TestCheckResourceAttr(def, "tenants.#", fmt.Sprintf("%d", info.tenants)))
		}
	}
	return checkFuncs
}

const testAccRightContainerDS = `
variable "rights_bundles" {
  default = [{{.RightsBundles}}]
}

variable "global_roles" {
  default = [{{.GlobalRoles}}]
}

variable "roles" {
  default = [{{.Roles}}]
}

variable "rights" {
  default = [{{.Rights}}]
}

data "vcd_right" "{{.RightDef}}" {
  count = length(var.rights)
  name  = var.rights[count.index]
}

data "vcd_role" "{{.RoleDef}}" {
  count =  length(var.roles)
  org   = "{{.Org}}"
  name  = var.roles[count.index]
}

data "vcd_global_role" "{{.GlobalRoleDef}}" {
  count = length(var.global_roles)
  name  = var.global_roles[count.index]
}

data "vcd_rights_bundle" "{{.RightsBundleDef}}" {
  count = length(var.rights_bundles)
  name  = var.rights_bundles[count.index]
}
`

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
