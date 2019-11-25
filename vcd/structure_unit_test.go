// +build unit ALL

package vcd

import (
	"os"
	"strings"
	"testing"
)

func TestIsValidIp(t *testing.T) {
	type IpData struct {
		ip       string
		expected bool
	}
	var ipInfo = []IpData{
		{"255.255.255.255", true},
		{"255.255.255.0", true},
		{"0.0.0.0", true},
		{"1.1.1.1", true},
		{"1.2.3.4", true},
		{"10.10.10.1", true},
		{"010.010.010.001", true},
		{"018.010.010.008", true},
		{"192.168.10.100", true},
		{"192.168.0.0", true},
		{"127.0.0.1", true},
		{"127.00.00.01", true},
		{"127.000.000.001", true},
		{"0127.0000.0000.0001", true},

		{"", false},
		{"1", false},
		{"1.", false},
		{"1.1", false},
		{"1.1.", false},
		{"1.1.1", false},
		{"1.2.3.4.5", false},
		{"1.2.3.4.5.6", false},
		{"192.168.0.1a", false},
		{"192.468.00.100", false},
		{"192.168.400.100", false},
		{"000.168.400.100", false},
		{"1.1.1.1111", false},
		{"256.0.0.0", false},
		{"0.256.0.0", false},
		{"0.0.256.0", false},
		{"0.0.0.256", false},
	}

	for _, info := range ipInfo {
		result := isValidIp(info.ip)
		if result == info.expected {
			if os.Getenv(testVerbose) != "" {
				expectedText := "valid"
				if !info.expected {
					expectedText = "invalid"
				}
				t.Logf("ok - IP '%s' is %s", info.ip, expectedText)
			}
		} else {
			t.Fail()
			t.Logf("not ok - IP '%s' - expected %v but got %v", info.ip, info.expected, result)
		}
	}
}

func TestValidateIps(t *testing.T) {
	type ipData struct {
		gateway      string
		netmask      string
		dns1         string
		dns2         string
		expected     bool
		errorPortion string
	}
	var ipInfo = []ipData{
		{"1.1.1.1", "255.255.255.0", "4.4.4.4", "8.8.8.8", true, ""},
		{"", "255.255.255.0", "4.4.4.4", "8.8.8.8", false, "gateway"},
		{"400.1.1.1", "255.255.255.0", "4.4.4.4", "8.8.8.8", false, "gateway"},
		{"1.1.1.1", "255.255.255.256", "4.4.4.4", "8.8.8.8", false, "netmask"},
		{"1.1.1.1", "255.255.255.255", "4.4.4.400", "8.8.8.8", false, "dns1"},
		{"1.1.1.1", "255.255.255.255", "4.4.4.4", "8.8.8.800", false, "dns2"},
	}
	verboseTest := os.Getenv(testVerbose) != ""
	for _, info := range ipInfo {
		err := validateIps(info.gateway, info.netmask, info.dns1, info.dns2)
		valid := err == nil
		validText := "valid"
		if !valid {
			validText = "invalid"
		}
		errorText := ""
		if err != nil {
			errorText = err.Error()
			if verboseTest {
				t.Logf("# error message: %s", errorText)
			}
		}
		if valid == info.expected {
			if verboseTest {
				t.Logf("ok - %s - %s - %s - %s is %s", info.gateway, info.netmask, info.dns1, info.dns2, validText)
			}
			if errorText != "" {
				if !strings.Contains(errorText, info.errorPortion) {
					if verboseTest {
						t.Logf("error message does not contain expected string '%s' for invalid IP group", info.errorPortion)
					}
					t.Fail()
				}
			}
		} else {
			t.Logf("not ok - %s - %s - %s - %s is %v but expected %v",
				info.gateway, info.netmask, info.dns1, info.dns2, valid, info.expected)
			t.Fail()
		}
	}

}
