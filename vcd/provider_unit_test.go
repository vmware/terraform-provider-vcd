// +build unit ALL

package vcd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"testing"
)

// Checks that the provider header in index.html.markdown
// has the version defined in the VERSION file
func TestProviderVersion(t *testing.T) {
	indexFile := path.Join(getCurrentDir(), "..", "website", "docs", "index.html.markdown")
	_, err := os.Stat(indexFile)
	if os.IsNotExist(err) {
		fmt.Printf("%s\n", indexFile)
		panic("Could not find index.html.markdown file")
	}

	indexText, err := ioutil.ReadFile(indexFile)
	if err != nil {
		panic(fmt.Errorf("could not read index file %s: %v", indexFile, err))
	}

	vcdHeader := `# VMware vCloud Director Provider`
	expectedText := vcdHeader + ` ` + currentProviderVersion
	reExpectedVersion := regexp.MustCompile(`(?m)^` + expectedText)
	reFoundVersion := regexp.MustCompile(`(?m)^` + vcdHeader + ` \d+\.\d+`)
	if reExpectedVersion.MatchString(string(indexText)) {
		if os.Getenv(testVerbose) != "" {
			t.Logf("Found expected version <%s> in index.html.markdown", currentProviderVersion)
		}
	} else {
		foundList := reFoundVersion.FindAllStringSubmatch(string(indexText), -1)
		foundText := ""
		if foundList != nil && len(foundList) > 0 && len(foundList[0]) > 0 {
			foundText = foundList[0][0]
			t.Logf("Expected text: <%s>", expectedText)
			t.Logf("Found text   : <%s> in index.html.markdown", foundText)
		} else {
			t.Logf("No version found in index.html.markdown")
		}
		t.Fail()
	}
}

func init() {
	testingTags["unit"] = "provider_unit_test.go"
}
