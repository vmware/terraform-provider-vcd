// IMPORTANT: DO NOT ADD build tags to this file

package vcd

import (
	"fmt"
	"os"
	"testing"
)

var testingTags = make(map[string]string)

func tagsHelp(t *testing.T) {

	var helpText string = `
# -----------------------------------------------------
# Tags are required to run the tests
# -----------------------------------------------------

At least one of the following tags should be defined:

   * ALL :       Runs all the tests
   * functional: Runs all the acceptance tests
   * unit:       Runs unit tests that don't need a live vCD (currently unused, but we plan to)

   * catalog:    Runs catalog related tests (also catalog_item, media)
   * disk:       Runs disk related tests
   * network:    Runs network related tests
   * gateway:    Runs edge gateway related tests
   * org:        Runs org related tests
   * vapp:       Runs vapp related tests
   * vdc:        Runs vdc related tests
   * vm:         Runs vm related tests

Examples:

go test -tags functional -v -timeout=45m .
go test -tags catalog -v -timeout=15m .
go test -tags "org vdc" -v -timeout=5m .
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
		fmt.Printf("# %s (%s)", k, v)
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
	if os.Getenv("SHOW_TAGS") != "" {
		showTags()
	}
}
