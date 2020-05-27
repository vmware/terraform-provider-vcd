// +build unit ALL

package govcd

import (
	"io/ioutil"
	"os"
	"testing"
)

// goldenString is a test helper to manage Golden files. It supports `update` parameter which may be
// useful for writing such files (manual or automated way).
func goldenString(t *testing.T, goldenFile string, actual string, update bool) string {
	t.Helper()

	goldenPath := "../test-resources/golden/" + t.Name() + "_" + goldenFile + ".golden"

	f, err := os.OpenFile(goldenPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		t.Fatalf("unable to find golden file '%s': %s", goldenPath, err)
	}
	defer f.Close()

	if update {
		_, err := f.WriteString(actual)
		if err != nil {
			t.Fatalf("error writing to file %s: %s", goldenPath, err)
		}

		return actual
	}

	content, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatalf("error opening file %s: %s", goldenPath, err)
	}
	return string(content)
}

// goldenBytes wraps goldenString and returns []byte
func goldenBytes(t *testing.T, goldenFile string, actual []byte, update bool) []byte {
	return []byte(goldenString(t, goldenFile, string(actual), update))
}
