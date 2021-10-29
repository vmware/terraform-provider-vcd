//go:build unit || ALL
// +build unit ALL

package vcd

import (
	"testing"
	"time"
)

func Test_isScalar(t *testing.T) {
	type args struct {
		t interface{}
	}
	var stringSlice interface{} = []string{"a", "b"}
	var interfaceScalar1 interface{} = "a"
	var interfaceScalar2 interface{} = 42
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "scalar-nil",
			args: args{nil},
			want: true,
		},
		{
			name: "scalar-string",
			args: args{"hello"},
			want: true,
		},
		{
			name: "scalar-int",
			args: args{int(1)},
			want: true,
		},
		{
			name: "scalar-int32",
			args: args{int32(1)},
			want: true,
		},
		{
			name: "scalar-int64",
			args: args{int64(1)},
			want: true,
		},
		{
			name: "scalar-bool",
			args: args{true},
			want: true,
		},
		{
			name: "scalar-float32",
			args: args{float32(1)},
			want: true,
		},
		{
			name: "scalar-float64",
			args: args{float64(1)},
			want: true,
		},
		{
			name: "scalar-interface1",
			args: args{interfaceScalar1},
			want: true,
		},
		{
			name: "scalar-interface2",
			args: args{interfaceScalar2},
			want: true,
		},
		{
			name: "struct",
			args: args{struct{}{}},
			want: false,
		},
		{
			name: "map",
			args: args{map[string]string{}},
			want: false,
		},
		{
			name: "string slice",
			args: args{[]string{}},
			want: false,
		},
		{
			name: "int slice",
			args: args{[]int{}},
			want: false,
		},
		{
			name: "array",
			args: args{[2]string{"", ""}},
			want: false,
		},
		{
			name: "array of int",
			args: args{[2]int{0, 0}},
			want: false,
		},
		{
			name: "array of structs",
			args: args{[2]struct{}{{}, {}}},
			want: false,
		},
		{
			name: "array of interface",
			args: args{[3]interface{}{1, "aaa", true}},
			want: false,
		},
		{
			name: "interface as slice",
			args: args{stringSlice},
			want: false,
		},
		{
			name: "library object1",
			args: args{time.Now()},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isScalar(tt.args.t); got != tt.want {
				t.Errorf("isScalar() = %v, want %v", got, tt.want)
			}
		})
	}
}
