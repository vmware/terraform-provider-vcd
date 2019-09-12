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

func TestGetHashValuesFromKey(t *testing.T) {

	type testInfo struct {
		key      string
		parent   string
		child    string
		expected []string
	}
	var testData = []testInfo{
		{
			"first.1234.second.5678.third",
			"first",
			"second",
			[]string{"1234", "5678"},
		},
		{
			"compute_capacity.315866465.memory.508945747.limit",
			"compute_capacity",
			"memory",
			[]string{"315866465", "508945747"},
		},
		{
			"compute_capacity.315866465.cpu.798465156.limit",
			"compute_capacity",
			"cpu",
			[]string{"315866465", "798465156"},
		},
	}
	for _, td := range testData {
		testMap := map[string]string{
			td.key: "",
		}
		first, second, err := getHashValuesFromKey(testMap, td.parent, td.child)
		if err != nil {
			t.Logf("processing key '%s' got error %s", td.key, err)
			t.Fail()
		}
		if first != td.expected[0] {
			t.Logf("Expected result from key '%s' was '%s' - Got '%s' instead", td.key, td.parent, td.expected[0])
			t.Fail()
		}
		if second != td.expected[1] {
			t.Logf("Expected result from key '%s' was '%s' - Got '%s' instead", td.key, td.child, td.expected[1])
			t.Fail()
		}
	}
}

func init() {
	testingTags["unit"] = "provider_unit_test.go"
}
