package vcd

import (
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

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
func anyValueWarningValidator(warningText string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (warnings []string, errors []error) {
		warnings = append(warnings, fmt.Sprintf("%s\n\n", warningText))
		return
	}
}

// checkEmptyOrSingleIP validates if the field is set to empty or a valid IP address
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

// validateCase checks if a string is of caseType "upper" or "lower"
func validateCase(caseType string) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		switch caseType {
		case "upper":
			if strings.ToUpper(v) != v {
				es = append(es, fmt.Errorf(
					"expected string to be upper cased, got: %s", v))
			}
		case "lower":
			if strings.ToLower(v) != v {
				es = append(es, fmt.Errorf(
					"expected string to be lower cased, got: %s", v))
			}
		default:
			panic("unsupported validation type for validateCase() function")
		}
		return
	}
}

// validateBusType checks if bus type is in correct format for independent disk
func validateBusType(v interface{}, k string) (warnings []string, errors []error) {
	value := v.(string)

	if busTypes[strings.ToUpper(value)] == "" {
		errors = append(errors, fmt.Errorf("%q (%q) value isn't valid", k, value))
	}

	return
}

// validateBusSubType checks if bus subtype is in correct format for independent disk
func validateBusSubType(v interface{}, k string) (warnings []string, errors []error) {
	value := v.(string)

	if busSubTypes[strings.ToLower(value)] == "" {
		errors = append(errors, fmt.Errorf("%q (%q) value isn't valid", k, value))
	}

	return
}
