//go:build unit || ALL

package vcd

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// TestJsonToCompactString checks that an unmarshaled JSON is correctly converted into a compact string.
func TestJsonToCompactString(t *testing.T) {
	givenJson := map[string]interface{}{
		"foo": "bar",
	}
	compactedJson, err := jsonToCompactString(givenJson)
	if err != nil {
		t.Errorf("did not expect an error but got: %s", err)
	}
	var obtainedJson map[string]interface{}
	err = json.Unmarshal([]byte(compactedJson), &obtainedJson)
	if err != nil {
		t.Errorf("did not expect an error but got: %s", err)
	}
	if !reflect.DeepEqual(givenJson, obtainedJson) {
		t.Errorf("compacted string %s is not equal to the original JSON: %s", obtainedJson, givenJson)
	}
	if strings.Contains(compactedJson, " ") {
		t.Errorf("compacted string %s contains spaces", compactedJson)
	}
}

// TestAreMarshaledJsonEqual checks that two marshaled JSONs are correctly compared.
func TestAreMarshaledJsonEqual(t *testing.T) {
	givenJson1 := []byte("{\"foo\":\"bar\"}")
	givenJson2 := []byte("{  \"foo\" :  \"bar\" }   ")

	areEqual, err := areMarshaledJsonEqual(givenJson1, givenJson2)
	if err != nil {
		t.Errorf("did not expect an error but got: %s", err)
	}
	if !areEqual {
		t.Errorf("%s and %s should be equal", givenJson1, givenJson2)
	}
}
