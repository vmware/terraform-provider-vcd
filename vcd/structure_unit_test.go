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

func Test_removeItemsMapWithKeyPrefixes(t *testing.T) {
	type args struct {
		input    map[string]interface{}
		prefixes []string
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "empty map",
			args: args{
				input:    map[string]interface{}{},
				prefixes: []string{"_remove_"},
			},
			want: map[string]interface{}{},
		},
		{
			name: "map of plain items",
			args: args{
				input: map[string]interface{}{
					"_remove_key1":     "bar",
					"_alsoremove_key2": 1,
					"key3":             true,
					"key4":             "bar",
				},
				prefixes: []string{"_remove_", "_alsoremove_"},
			},
			want: map[string]interface{}{
				"key3": true,
				"key4": "bar",
			},
		},
		{
			name: "empty prefixes",
			args: args{
				input: map[string]interface{}{
					"_remove_key1":     "bar",
					"_alsoremove_key2": 1,
					"key3":             true,
					"key4":             "bar",
				},
				prefixes: []string{},
			},
			want: map[string]interface{}{
				"_remove_key1":     "bar",
				"_alsoremove_key2": 1,
				"key3":             true,
				"key4":             "bar",
			},
		},
		{
			name: "complex map",
			args: args{
				input: map[string]interface{}{
					"_remove_key1": "bar",
					"key2": map[string]interface{}{
						"innerkey1":             1,
						"_remove_innerkey2":     "innerbar",
						"_alsoremove_innerkey3": "innerbar",
						"innerkey3": map[string]interface{}{
							"moreinnerkey1":         1,
							"_remove_moreinnerkey2": 2,
						},
						"innerkey4": false,
					},
					"key3": "bar",
					"key4": "bar",
				},
				prefixes: []string{"_remove_", "_alsoremove_"},
			},
			want: map[string]interface{}{
				"key2": map[string]interface{}{
					"innerkey1": 1,
					"innerkey3": map[string]interface{}{
						"moreinnerkey1": 1,
					},
					"innerkey4": false,
				},
				"key3": "bar",
				"key4": "bar",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeItemsMapWithKeyPrefixes(tt.args.input, tt.args.prefixes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeItemsMapWithKeyPrefixes() = %v, want %v", got, tt.want)
			}
		})
	}
}
