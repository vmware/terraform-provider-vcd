// +build unit ALL

package vcd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"testing"

	semver "github.com/hashicorp/go-version"
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
		if len(foundList) > 0 && len(foundList[0]) > 0 {
			foundText = foundList[0][0]
			t.Logf("Expected text: <%s>", expectedText)
			t.Logf("Found text   : <%s> in index.html.markdown", foundText)
		} else {
			t.Logf("No version found in index.html.markdown")
		}
		t.Fail()
	}
}

// Checks that a PREVIOUS_VERSION file exists, and it contains a version lower than the one in VERSION
func TestProviderUpgradeVersion(t *testing.T) {
	currentVersionText, err := getVersionFromFile("VERSION")
	if err != nil {
		t.Logf("error retrieving version from VERSION file: %s", err)
		t.Fail()
		return
	}
	previousVersionText, err := getVersionFromFile("PREVIOUS_VERSION")
	if err != nil {
		t.Logf("error retrieving version from PREVIOUS_VERSION file: %s", err)
		t.Fail()
		return
	}

	currentVersion, err := semver.NewVersion(currentVersionText)
	if err != nil {
		t.Logf("error converting current version to Hashicorp version: %s", err)
		t.Fail()
		return
	}
	previousVersion, err := semver.NewVersion(previousVersionText)
	if err != nil {
		t.Logf("error converting previous version to Hashicorp version: %s", err)
		t.Fail()
		return
	}
	result := currentVersion.Compare(previousVersion)
	// result < 0 means current version is lower than previous version
	// result == 0 means current version is the same as previous version
	// result == 0 means current version is higher than previous version
	if result < 0 {
		t.Logf("current version (%s) is lower than previous version (%s)", currentVersionText, previousVersionText)
		t.Fail()
	}
	if result == 0 {
		t.Logf("current version (%s) is the same as previous version (%s)", currentVersionText, previousVersionText)
		t.Fail()
	}
}

func TestGetMajorVersion(t *testing.T) {
	version := getMajorVersion()

	reVersion := regexp.MustCompile(`^\d+\.\d+$`)
	if !reVersion.MatchString(version) {
		t.Fail()
	}
	t.Logf("%s", version)
}

func init() {
	testingTags["unit"] = "provider_unit_test.go"
}
