package govcloudair

import (
	"fmt"
)

func GetVersionHeader(version string) (key, value string) {
	return "Accept", fmt.Sprintf("application/*+xml;version=%s", version)
}
