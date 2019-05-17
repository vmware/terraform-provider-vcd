// IMPORTANT: DO NOT ADD build tags to this file

package vcd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

const (
	testVerbose = "TEST_VERBOSE"
)

// These variables are needed by tests running under any tags
var (
	// Collection of defined tags in the current test run
	testingTags = make(map[string]string)
	// This library major version
	currentProviderVersion string = getMajorVersion()
)

func tagsHelp(t *testing.T) {

	var helpText string = `
# -----------------------------------------------------
# Tags are required to run the tests
# -----------------------------------------------------

At least one of the following tags should be defined:

   * ALL :       Runs all the tests
   * functional: Runs all the acceptance tests
   * unit:       Runs unit tests that don't need a live vCD

   * catalog:    Runs catalog related tests (also catalog_item, media)
   * disk:       Runs disk related tests
   * network:    Runs network related tests
   * gateway:    Runs edge gateway related tests
   * org:        Runs org related tests
   * vapp:       Runs vapp related tests
   * vdc:        Runs vdc related tests
   * vm:         Runs vm related tests

Examples:

  go test -tags unit -v -timeout=45m .
  go test -tags functional -v -timeout=45m .
  go test -tags catalog -v -timeout=15m .
  go test -tags "org vdc" -v -timeout=5m .

Tagged tests can also run using make
  make testunit
  make testacc
  make testcatalog
`
	t.Logf(helpText)
}

// Tells indirectly if a tag has been set
// For every tag there is an `init` function that
// fills an item in `testingTags`
func isTagSet(tagName string) bool {
	_, ok := testingTags[tagName]
	return ok
}

// For troubleshooting:
// Shows which tags were set, and in which file.
func showTags() {
	if len(testingTags) > 0 {
		fmt.Println("# Defined tags:")
	}
	for k, v := range testingTags {
		fmt.Printf("# %s (%s)\n", k, v)
	}
}

// Checks whether any tags were defined, and raises an error if not
func TestTags(t *testing.T) {
	if len(testingTags) == 0 {
		t.Logf("# No tags were defined")
		tagsHelp(t)
		t.Fail()
		return
	}
	if os.Getenv(testVerbose) != "" {
		showTags()
	}
}

// Finds the current directory, through the path of this running test
func getCurrentDir() string {
	_, currentFilename, _, _ := runtime.Caller(0)
	return filepath.Dir(currentFilename)
}

// Reads the version from the VERSION file in the root directory
func getMajorVersion() string {

	versionFile := path.Join(getCurrentDir(), "..", "VERSION")

	// Checks whether the VERSION file exists
	_, err := os.Stat(versionFile)
	if os.IsNotExist(err) {
		panic("Could not find VERSION file")
	}

	// Reads the version from the file
	versionText, err := ioutil.ReadFile(versionFile)
	if err != nil {
		panic(fmt.Errorf("could not read VERSION file %s: %v", versionFile, err))
	}

	// The version is expected to be in the format v#.#.#
	// We only need the first two numbers
	reVersion := regexp.MustCompile(`v(\d+\.\d+)\.\d+`)
	versionList := reVersion.FindAllStringSubmatch(string(versionText), -1)
	if versionList == nil || len(versionList) == 0 {
		panic("empty or non-formatted version found in VERSION file")
	}
	if versionList[0] == nil || len(versionList[0]) < 2 {
		panic("unable to extract major version from VERSION file")
	}
	// A successful match will look like
	// [][]string{[]string{"v2.0.0", "2.0"}}
	// Where the first element is the full text matched, and the second one is the first captured text
	return versionList[0][1]
}
