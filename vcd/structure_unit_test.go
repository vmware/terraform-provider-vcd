//go:build unit || ALL

package vcd

import (
	"reflect"
	"testing"
)

// Test_jsonToCompactString checks that an unmarshaled JSON is correctly converted into a compact string.
func Test_jsonToCompactString(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		want    string
		wantErr bool
	}{
		{
			name: "correct JSON",
			input: map[string]interface{}{
				"foo": "bar",
			},
			want:    "{\"foo\":\"bar\"}",
			wantErr: false,
		},
		{
			name: "complex JSON",
			input: map[string]interface{}{
				"foo":  "bar",
				"foo2": 42,
				"foo3": []string{
					"a", "b",
				},
				"foo4": map[string]interface{}{
					"c": "d",
				},
			},
			want:    "{\"foo\":\"bar\",\"foo2\":42,\"foo3\":[\"a\",\"b\"],\"foo4\":{\"c\":\"d\"}}",
			wantErr: false,
		},
		{
			name:    "empty JSON",
			input:   map[string]interface{}{},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "nil value in JSON",
			input: map[string]interface{}{
				"foo": nil,
			},
			want:    "{\"foo\":null}",
			wantErr: false,
		},
		{
			name: "empty key in JSON",
			input: map[string]interface{}{
				"": "bar",
			},
			want:    "{\"\":\"bar\"}",
			wantErr: false,
		},
		{
			name: "spaces in JSON key/values",
			input: map[string]interface{}{
				" foo  ": "  bar  ",
			},
			want:    "{\" foo  \":\"  bar  \"}",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonToCompactString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("jsonToCompactString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("jsonToCompactString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_areMarshaledJsonEqual checks that two marshaled JSONs are correctly compared.
func Test_areMarshaledJsonEqual(t *testing.T) {
	tests := []struct {
		name    string
		input1  string
		input2  string
		want    bool
		wantErr bool
	}{
		{
			name:    "compared JSONs are exactly the same",
			input1:  "{\"foo\":\"bar\"}",
			input2:  "{\"foo\":\"bar\"}",
			want:    true,
			wantErr: false,
		},
		{
			name:    "compared JSON components in different order",
			input1:  `{"foo":"bar", "abc": "xyz"}`,
			input2:  `{"abc":"xyz", "foo":"bar"}`,
			want:    true,
			wantErr: false,
		},
		{
			name:    "second JSON contains trailing spaces",
			input1:  "{\"foo\":\"bar\"}",
			input2:  "{  \"foo\" :  \"bar\" }   ",
			want:    true,
			wantErr: false,
		},
		{
			name:    "first JSON is empty",
			input1:  "",
			input2:  "{\"foo\":\"bar\"}",
			wantErr: true,
		},
		{
			name:    "second JSON is empty",
			input1:  "{\"foo\":\"bar\"}",
			input2:  "",
			wantErr: true,
		},
		{
			name:    "second JSON is empty object",
			input1:  "{\"foo\":\"bar\"}",
			input2:  "{}",
			want:    false,
			wantErr: false,
		},
		{
			name:    "both JSON have null values",
			input1:  "{\"foo\": null}",
			input2:  "{\"foo\":null}",
			want:    true,
			wantErr: false,
		},
		{
			name:    "first JSON is malformed",
			input1:  "{\"foo\": }",
			input2:  "{\"foo\": \"bar\"}",
			wantErr: true,
		},
		{
			name:    "second JSON is malformed",
			input1:  "{\"foo\": \"bar\"}",
			input2:  "{\"foo\": }",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := areMarshaledJsonEqual([]byte(tt.input1), []byte(tt.input2))
			if (err != nil) != tt.wantErr {
				t.Errorf("areMarshaledJsonEqual() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("areMarshaledJsonEqual() got = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_traverseMapAndGet tests traverseMapAndGet function
func Test_traverseMapAndGet(t *testing.T) {
	type args struct {
		input interface{}
		path  string
	}
	tests := []struct {
		name     string
		args     args
		wantType string
		want     interface{}
		wantErr  string
	}{
		{
			name: "input is nil",
			args: args{
				input: nil,
			},
			wantErr: "the input is nil",
		},
		{
			name: "input is not a map",
			args: args{
				input: "error",
			},
			wantErr: "the input is a string, not a map[string]interface{}",
		},
		{
			name: "map is empty",
			args: args{
				input: map[string]interface{}{},
			},
			wantErr: "the map is empty",
		},
		{
			name: "map does not have key",
			args: args{
				input: map[string]interface{}{
					"keyA": "value",
				},
				path: "keyB",
			},
			wantErr: "key 'keyB' does not exist in input map",
		},
		{
			name: "map has a single simple key",
			args: args{
				input: map[string]interface{}{
					"keyA": "value",
				},
				path: "keyA",
			},
			wantType: "string",
			want:     "value",
		},
		{
			name: "map has a single complex key",
			args: args{
				input: map[string]interface{}{
					"keyA": map[string]interface{}{
						"keyB": "value",
					},
				},
				path: "keyA",
			},
			wantType: "map",
			want: map[string]interface{}{
				"keyB": "value",
			},
		},
		{
			name: "map has a complex structure",
			args: args{
				input: map[string]interface{}{
					"keyA": map[string]interface{}{
						"keyB": map[string]interface{}{
							"keyC": "value",
						},
					},
				},
				path: "keyA.keyB.keyC",
			},
			wantType: "string",
			want:     "value",
		},
		{
			name: "requested path is deeper than the map structure",
			args: args{
				input: map[string]interface{}{
					"keyA": map[string]interface{}{
						"keyB": map[string]interface{}{
							"keyC": "value",
						},
					},
				},
				path: "keyA.keyB.keyC.keyD",
			},
			wantErr: "key 'keyC' is a string, not a map, but there are still 1 paths to explore",
		},
		{
			name: "obtained value does not correspond to the desired type",
			args: args{
				input: map[string]interface{}{
					"keyA": map[string]interface{}{
						"keyB": map[string]interface{}{
							"keyC": map[string]interface{}{},
						},
					},
				},
				path: "keyA.keyB.keyC",
			},
			wantType: "string",
			wantErr:  "could not convert obtained type map[string]interface {} to requested string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			var err error
			if tt.wantType == "string" {
				got, err = traverseMapAndGet[string](tt.args.input, tt.args.path)
			} else if tt.wantType == "map" {
				got, err = traverseMapAndGet[map[string]interface{}](tt.args.input, tt.args.path)
			} else {
				t.Fatalf("wantType type not used in this test")
			}

			if err != nil {
				if tt.wantErr != err.Error() {
					t.Errorf("traverseMapAndGet() error = %v, wantErr = %v", err, tt.wantErr)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("traverseMapAndGet() got = %v, want %v", got, tt.want)
			}
		})
	}
}
