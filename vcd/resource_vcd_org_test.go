package vcd

// import (
// 	"fmt"
// 	"os"
// 	"testing"

// 	"github.com/hashicorp/terraform/helper/resource"
// 	"github.com/hashicorp/terraform/terraform"
// 	govcd "github.com/ukcloud/govcloudair"
// )

// func TestAccVcdOrgBasic(t *testing.T) {
// 	if v := os.Getenv("VCD_EXTERNAL_IP"); v == "" {
// 		t.Skip("Environment variable VCD_EXTERNAL_IP must be set to run ORG tests")
// 		return
// 	}

// 	var e govcd.Org

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckVcdDNATDestroy,
// 		Steps: []resource.TestStep{
// 			resource.TestStep{
// 				Config: fmt.Sprintf(testAccCheckVcdOrg_basic),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckVcdOrgExists("vcd_org.test1", &e),
// 					resource.TestCheckResourceAttr(
// 						"vcd_org.test1", "name", "test1"),
// 					resource.TestCheckResourceAttr(
// 						"vcd_org.test1", "full_name", "test1"),
// 					resource.TestCheckResourceAttr(
// 						"vcd_org.test1", "is_enabled", "true"),
// 				),
// 			},
// 		},
// 	})
// }

// func testAccCheckVcdOrgExists(n string, org *govcd.Org) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		rs, ok := s.RootModule().Resources[n]
// 		if !ok {
// 			return fmt.Errorf("Not found: %s", n)
// 		}

// 		if rs.Primary.ID == "" {
// 			return fmt.Errorf("No ORG ID is set")
// 		}

// 		conn := testAccProvider.Meta().(*VCDClient)

// 		_, f, err := conn.(rs.Primary.ID)

// 		if err != nil {
// 			return err
// 		}
// 		org = &f

// 		return nil
// 	}
// }

// func testAccCheckOrgDestroy(s *terraform.State) error {
// 	//conn := testAccProvider.Meta().(*VCDClient)
// 	for _, rs := range s.RootModule().Resources {
// 		if rs.Type != "vcd_org" {
// 			continue
// 		}

// 		// if rs.Name == "test" {
// 		// 	return fmt.Errorf("Org Not properly deleted")
// 		// }

// 	}

// 	return nil
// }

// const testAccCheckVcdOrg_basic = `
// resource "vcd_org" "test1"{
//   name = "test1"
//   full_name = "test1"
//   is_enabled = "true"
//   force = "true"
//   recursive = "true"
// }
// `
