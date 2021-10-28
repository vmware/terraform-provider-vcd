//go:build unit || ALL
// +build unit ALL

package vcd

import "testing"

func Test_isScalar(t *testing.T) {
	type args struct {
		t interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isScalar(tt.args.t); got != tt.want {
				t.Errorf("isScalar() = %v, want %v", got, tt.want)
			}
		})
	}
}
