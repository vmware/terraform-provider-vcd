package vcd

import (
	"flag"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// suppressWordToEmptyString is a DiffSuppressFunc which ignore the change from word to empty string "".
// This is useful when API returns some default value but it is not set (and not sent via API) in config.
func suppressWordToEmptyString(word string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		if old == word && new == "" {
			return true
		}
		return false
	}

}

// suppressNetworkUpgradedInterface is used to silence the changes in
// property "interface_type" in routed networks.
// In the old the version, the "internal" interface was implicit,
// while in the new one it is one of several.
// This function only considers the "internal" value, as the other interface types
// were not possible in the previous version
func suppressNetworkUpgradedInterface() schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		if old == "" && new == "internal" {
			return true
		}
		return false
	}
}

// TODO v3.0 remove once `ip` and `network_name` attributes are removed
func suppressIfIPIsOneOf() schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		switch {
		case new == "dhcp" && (old == "na" || net.ParseIP(old) != nil):
			return true
		case new == "allocated" && net.ParseIP(old) != nil:
			return true
		case new == "" && net.ParseIP(old) != nil:
			return true
		default:
			return false
		}
	}
}

// falseBoolSuppress suppresses change if value is set to false or is empty
func falseBoolSuppress() schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		_, isTrue := d.GetOk(k)
		return !isTrue
	}
}

// suppressNewFalse always suppresses when new value is false
func suppressFalse() schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return new == "false"
	}
}

// getTerraformStdout returns std out to write message in terraform output
func getTerraformStdout() *os.File {
	// Needed to avoid errors when uintptr(4) is used
	if v := flag.Lookup("test.v"); v == nil || v.Value.String() != "true" {
		return os.NewFile(uintptr(4), "stdout")
	} else {
		return os.Stdout
	}
}

// suppressCase is a schema.SchemaDiffSuppressFunc which ignore case changes
func suppressCase(k, old, new string, d *schema.ResourceData) bool {
	return strings.EqualFold(old, new)
}
