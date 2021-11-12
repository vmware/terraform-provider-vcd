package vcd

import (
	"hash/crc32"
)

// imported from Hashicorp (https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html)

// hashcodeString hashes a string to a unique hashcode.
//
// crc32 returns a uint32, but for our use we need
// a non-negative integer. Here we cast to an integer
// and invert it if the result is negative.
func hashcodeString(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	return 0
}
