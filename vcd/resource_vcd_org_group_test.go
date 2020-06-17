// +build user functional ALL

package vcd

import (
	"bytes"
	"fmt"
	"net"
	"regexp"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// TestAccVcdOrgGroup tests INTEGRATED (LDAP) group management in Terraform.
// In step 0 it spawns its own testing LDAP container with Terraform config held in `ldapSetup` var.
// In step 1 PreConfig it uses SDK to configure LDAP settings in vCD and tests group management
// LDAP configuration will be removed after test run
// More about LDAP testing container - https://github.com/rroemhild/docker-test-openldap
//
// Note. External network must be properly configured (including DNS records) and must be able to
// access internet so that it can download docker image. Also the environment where test is run must
// be able to access external network IP so that monitoring is possible.
func TestAccVcdOrgGroup(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip("TestAccVcdOrgGroup requires system admin privileges")
		return
	}
	// LDAP is being configured using go-vcloud-director - binary test cannot be run
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	ldapConfigParams := struct {
		ExternalNetwork string
		GuestImage      string
		CatalogName     string
	}{
		ExternalNetwork: testConfig.Networking.ExternalNetwork,
		GuestImage:      testConfig.VCD.Catalog.CatalogItem,
		CatalogName:     testConfig.VCD.Catalog.Name,
	}
	// getLdapSetupTemplate does not use regular templateFill because this part is used for
	// automated LDAP configuration setup
	ldapSetupConfig, err := getLdapSetupTemplate(ldapSetup, ldapConfigParams)
	if err != nil {
		t.Errorf("failed processing LDAP setup template: %s", err)
	}
	debugPrintf("#[DEBUG] CONFIGURATION for step 0 (LDAP server configuration): %s", ldapSetupConfig)

	role1 := govcd.OrgUserRoleOrganizationAdministrator
	role2 := govcd.OrgUserRoleVappAuthor

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"ProviderType": "INTEGRATED",
		"RoleName":     role1,
		"Tags":         "user",
		"FuncName":     t.Name() + "-Step1",
		"Description":  "Description1",
	}

	groupConfigText := templateFill(testAccOrgGroup, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", groupConfigText)

	params["FuncName"] = t.Name() + "-Step2"
	params["RoleName"] = role2
	params["Description"] = "Description2"
	groupConfigText2 := templateFill(testAccOrgGroup, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", groupConfigText2)

	nic0Ip := testCachedFieldValue{}

	// Instantiate LDAP configuration functions.
	// Variables passed:
	// t *testing.T - inserted because TestCase `PreConfig` function does not allow parameters but
	// we want to fail the test if something breaks in between
	// nic0Ip - using a `testCachedFieldValue` the IP will be captured after step 0 and will be
	// available in Step 1 to configure LDAP server
	ldapConfig := ldapConfigurator{
		t:      t,
		nic0Ip: &nic0Ip,
	}

	// make a function with no arguments to suit signature of TestStep.PreConfig
	configureLdapFunc := func() {
		ldapConfig.configureOrgLdap()
	}

	// Remove LDAP settings at the end of test
	defer func() {
		fmt.Printf("# Removing LDAP settings for Org '%s'\n", ldapConfig.org.AdminOrg.Name)
		err := ldapConfig.org.LdapDisable()
		if err != nil {
			ldapConfig.t.Errorf("error removing LDAP settings for Org '%s': %s", ldapConfig.org.AdminOrg.Name, err)
		}
	}()

	// groupIdRegex is reused a few times in tests to match IDs
	groupIdRegex := regexp.MustCompile(`^urn:vcloud:group:`)

	resource.Test(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdGroupDestroy("admin_staff"),
			testAccCheckVcdGroupDestroy("ship_crew"),
		),
		// CheckDestroy: testAccCheckVcdGroupDestroy(params["GroupName"].(string)),
		Steps: []resource.TestStep{
			// Step 0 - sets up direct network, vApp and VM with LDAP server and captures NIC 0 IP
			// so that before step 1 LDAP can be configured (using TestStep.PreConfig)
			resource.TestStep{
				PreConfig: func() { fmt.Println("# Setting up LDAP server") },
				Config:    ldapSetupConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_direct.net", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp.ldap-server", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_org_network.direct-net", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm.ldap-container", "id"),
					nic0Ip.cacheTestResourceFieldValue("vcd_vapp_vm.ldap-container", "network.0.ip"),
				),
			},
			resource.TestStep{
				// Step 1 - the configureLdapFunc has all required variables stored to perform LDAP
				// configuration on Org defined in testing config after testing LDAP server was
				// configured in Step0. `Config` is the same as for step 0
				PreConfig: configureLdapFunc,
				// ldapSetupConfig is still used in Config so that Terraform does not destroy LDAP
				// server built in Step 0
				Config: ldapSetupConfig + groupConfigText,
				Check: resource.ComposeAggregateTestCheckFunc(
					// sleepTester(),
					resource.TestMatchResourceAttr("vcd_org_group.group1", "id", groupIdRegex),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "name", "ship_crew"),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "role", role1),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "description", "Description1"),
					resource.TestMatchResourceAttr("vcd_org_group.group2", "id", groupIdRegex),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "name", "admin_staff"),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "role", role1),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "description", "Description1"),
				),
			},
			resource.TestStep{
				// ldapSetupConfig is still used in Config so that Terraform does not destroy LDAP
				// server built in Step 0
				Config: ldapSetupConfig + groupConfigText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// sleepTester(),
					resource.TestMatchResourceAttr("vcd_org_group.group1", "id", groupIdRegex),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "name", "ship_crew"),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "role", role2),
					resource.TestCheckResourceAttr("vcd_org_group.group1", "description", "Description2"),
					resource.TestMatchResourceAttr("vcd_org_group.group2", "id", groupIdRegex),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "name", "admin_staff"),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "role", role2),
					resource.TestCheckResourceAttr("vcd_org_group.group2", "description", "Description2"),
				),
			},

			resource.TestStep{
				ResourceName:      "vcd_org_group.group-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, "ship_crew"),
			},
		},
	})
}

// testAccCheckVcdGroupDestroy verifies if Org Group with given name does not exist in vCD
func testAccCheckVcdGroupDestroy(groupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return err
		}
		group, err := adminOrg.GetGroupByName(groupName, false)
		if err != govcd.ErrorEntityNotFound {
			return fmt.Errorf("group %s was not destroyed", groupName)
		}
		if group != nil {
			return fmt.Errorf("group %s was found in %s ", groupName, adminOrg.AdminOrg.Name)
		}
		return nil
	}
}

const testAccOrgGroup = `
resource "vcd_org_group" "group1" {
  provider_type = "INTEGRATED"
  name          = "ship_crew"
  role          = "{{.RoleName}}"
  description   = "{{.Description}}"
}

resource "vcd_org_group" "group2" {
  provider_type = "INTEGRATED"
  name          = "admin_staff"
  role          = "{{.RoleName}}"
  description   = "{{.Description}}"
}
`

const ldapSetup = `
resource "vcd_network_direct" "net" {
  name             = "TestAccVcdOrgGroup"
  external_network = "{{.ExternalNetwork}}"
}

resource "vcd_vapp" "ldap-server" {
  name = "ldap-server"
}

resource "vcd_vapp_org_network" "direct-net" {
  vapp_name        = vcd_vapp.ldap-server.name
  org_network_name = vcd_network_direct.net.name
}

resource "vcd_vapp_vm" "ldap-container" {
  vapp_name     = vcd_vapp.ldap-server.name
  name          = "ldap-host"
  catalog_name  = "{{.CatalogName}}"
  template_name = "{{.GuestImage}}"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  customization {
    initscript = <<-EOT
		{
			until ip a show eth0 | grep "inet "
			do
				sleep 1
			done
			systemctl enable docker
			systemctl start docker
			docker run --name ldap-server --restart=always --privileged -d -p 389:389 rroemhild/test-openldap
		} &
		EOT
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.direct-net.org_network_name
    ip_allocation_mode = "POOL"
    is_primary         = true
  }
}
`

// getLdapSetupTemplate
func getLdapSetupTemplate(templateText string, params interface{}) (string, error) {
	var ldapSetupCfg bytes.Buffer
	tmpl, err := template.New("test").Parse(templateText)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(&ldapSetupCfg, params)
	if err != nil {
		return "", err
	}

	return ldapSetupCfg.String(), nil
}

// isTcpPortOpen checks if remote TCP port is open or closed every 8 seconds until timeout is
// reached
func isTcpPortOpen(host, port string, timeout int) bool {
	retryTimeout := timeout
	// due to the VMs taking long time to boot it needs to be at least 5 minutes
	// may be even more in slower environments
	if timeout < 5*60 { // 5 minutes
		retryTimeout = 5 * 60 // 5 minutes
	}
	timeOutAfterInterval := time.Duration(retryTimeout) * time.Second
	timeoutAfter := time.After(timeOutAfterInterval)
	tick := time.NewTicker(time.Duration(8) * time.Second)

	for {
		select {
		case <-timeoutAfter:
			fmt.Printf(" Failed\n")
			return false
		case <-tick.C:
			timeout := time.Second * 3
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
			if err != nil {
				fmt.Printf(".")
			}
			// Connection established - the port is open
			if conn != nil {
				defer conn.Close()
				fmt.Printf(" Done\n")
				return true
			}
		}
	}
}

// ldapConfigurator is a struct holding required data to ease go-vcloud-director SDK operation
// integration with Terraform acceptance test framework. It allows to use *testing.T inside some
// Terraform accpentace functions which accept only functions with no parameters (such as
// "PreConfig)
type ldapConfigurator struct {
	t         *testing.T
	vcdClient *VCDClient
	org       *govcd.AdminOrg
	nic0Ip    *testCachedFieldValue
}

// configureOrgLdap checks that LDAP TCP port 389 is open for ldapConfigurator.nic0Ip and then
// configures vCD Org to use that LDAP server
func (l *ldapConfigurator) configureOrgLdap() {
	var err error
	// Step 0 - collect needed connections
	l.vcdClient = createTemporaryVCDConnection()
	l.org, err = l.vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		l.t.Errorf("could not get Org '%s': %s", testConfig.VCD.Org, err)
	}

	// Step 1 - ensure LDAP is already UP and serving connections on TCP 389
	fmt.Printf("# Waiting until LDAP responds on %s:389 :", l.nic0Ip.fieldValue)
	isPortOpen := isTcpPortOpen(l.nic0Ip.fieldValue, "389", testConfig.Provider.MaxRetryTimeout)
	if !isPortOpen {
		l.t.Error("error waiting for LDAP to respond")
	}

	// Step 2 - configure LDAP
	l.orgConfigureLdap(l.nic0Ip.fieldValue)
}

func (l *ldapConfigurator) orgConfigureLdap(ldapServerIp string) {
	fmt.Printf("# Configuring LDAP settings for Org '%s'", l.org.AdminOrg.Name)

	// The below settings are tailored for LDAP docker testing image
	// https://github.com/rroemhild/docker-test-openldap
	ldapSettings := &types.OrgLdapSettingsType{
		OrgLdapMode: types.LdapModeCustom,
		CustomOrgLdapSettings: &types.CustomOrgLdapSettings{
			HostName:                ldapServerIp,
			Port:                    389,
			SearchBase:              "dc=planetexpress,dc=com",
			AuthenticationMechanism: "SIMPLE",
			ConnectorType:           "OPEN_LDAP",
			Username:                "cn=admin,dc=planetexpress,dc=com",
			Password:                "GoodNewsEveryone",
			UserAttributes: &types.OrgLdapUserAttributes{
				ObjectClass:               "inetOrgPerson",
				ObjectIdentifier:          "uid",
				Username:                  "uid",
				Email:                     "mail",
				FullName:                  "cn",
				GivenName:                 "givenName",
				Surname:                   "sn",
				Telephone:                 "telephoneNumber",
				GroupMembershipIdentifier: "dn",
			},
			GroupAttributes: &types.OrgLdapGroupAttributes{
				ObjectClass:          "group",
				ObjectIdentifier:     "cn",
				GroupName:            "cn",
				Membership:           "member",
				MembershipIdentifier: "dn",
			},
		},
	}

	_, err := l.org.LdapConfigure(ldapSettings)
	if err != nil {
		fmt.Println(" Failed")
		l.t.Errorf("failed configuring LDAP for Org '%s': %s", testConfig.VCD.Org, err)
	}
	fmt.Println(" Done")
}
