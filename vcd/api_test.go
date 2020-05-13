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
	"strings"
	"testing"
)

// These variables are needed by tests running under any tags
var (
	vcdTestVerbose = os.Getenv("VCD_TEST_VERBOSE") != ""
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
   * user:       Runs user related tests
   * vapp:       Runs vapp related tests
   * vdc:        Runs vdc related tests
   * vm:         Runs vm related tests
   * lb:         Runs load balancer related tests

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
	if vcdTestVerbose {
		showTags()
	}
}

// Checks if a file exists
func fileExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	fileMode := f.Mode()
	return fileMode.IsRegular()
}

// Checks if a directory exists
func dirExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	fileMode := f.Mode()
	return fileMode.IsDir()
}

// Finds the current directory, through the path of this running test
func getCurrentDir() string {
	_, currentFilename, _, _ := runtime.Caller(0)
	return filepath.Dir(currentFilename)
}

// Reads the version from the VERSION file in the root directory
func getMajorVersion() string {
	return getMajorVersionFromFile("VERSION")
}

// Reads the version from a given file in the root directory
func getMajorVersionFromFile(fileName string) string {

	versionText, err := getVersionFromFile(fileName)
	if err != nil {
		panic(fmt.Sprintf("error retrieving version from %s: %s", fileName, err))
	}

	// The version is expected to be in the format v#.#.#
	// We only need the first two numbers
	reVersion := regexp.MustCompile(`v(\d+\.\d+)\.\d+`)
	versionList := reVersion.FindAllStringSubmatch(string(versionText), -1)
	if len(versionList) == 0 {
		panic(fmt.Sprintf("empty or non-formatted version found in file %s", fileName))
	}
	if versionList[0] == nil || len(versionList[0]) < 2 {
		panic(fmt.Sprintf("unable to extract major version from file %s", fileName))
	}
	// A successful match will look like
	// [][]string{[]string{"v2.0.0", "2.0"}}
	// Where the first element is the full text matched, and the second one is the first captured text
	return versionList[0][1]
}

// Reads the version from a given file in the root directory
func getVersionFromFile(fileName string) (string, error) {

	versionFile := path.Join(getCurrentDir(), "..", fileName)

	// Checks whether the wanted file exists
	_, err := os.Stat(versionFile)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("could not find file %s: %v", versionFile, err)
	}

	// Reads the version from the file
	versionText, err := ioutil.ReadFile(versionFile)
	if err != nil {
		return "", fmt.Errorf("could not read file %s: %v", versionFile, err)
	}

	return strings.TrimSpace(string(versionText)), nil
}

// firstNonEmpty returns the first non empty string from a list
// If all arguments are empty, returns an empty string
func firstNonEmpty(args ...string) string {
	for _, s := range args {
		if s != "" {
			return s
		}
	}
	return ""
}
