package vcd

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/hashicorp/terraform/helper/schema"
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

// noopValueWarningValidator is a no-op validator which only emits warning string when fieldValue
// is set to the specified one
func noopValueWarningValidator(fieldValue interface{}, warningText string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		if fieldValue == i {
			warnings = append(warnings, fmt.Sprintf("%s\n\n", warningText))
		}

		return
	}
}

// anyValueWarningValidator is a validator which only emits always warning string
func anyValueWarningValidator(fieldValue interface{}, warningText string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		warnings = append(warnings, fmt.Sprintf("%s\n\n", warningText))
		return
	}
}

func checkEmptyOrSingleIP() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if net.ParseIP(v) == nil && v != "" {
			es = append(es, fmt.Errorf(
				"expected %s to be empty or contain a valid IP, got: %s", k, v))
		}
		return
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
