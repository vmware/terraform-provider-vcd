package govcloudair

import (
	"fmt"

	types "github.com/ukcloud/govcloudair/types/v56"
)

func GetVersionHeader(version types.ApiVersionType) (key, value string) {
	return "Accept", fmt.Sprintf("application/*+xml;version=%s", version)
}
