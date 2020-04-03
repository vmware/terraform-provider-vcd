// +build unit ALL

package vcd

import "testing"

func Test_compareDate(t *testing.T) {
	type args struct {
		wanted string
		got    string
	}
	tests := []struct {
		args args
		want bool
	}{
		// Note: "YYYY-MM-DDThh:mm:ss.µµµZ" (time.RFC3339) is the format used for catalog item
		// and vApp templates creation dates

		{args{"=2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.500Z"}, true},
		{args{">2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.500Z"}, false},
		{args{"> 2020-03-09T09:50:51.500Z", "2020-03-09T09:50:51.500Z"}, false},
		{args{">2020-03-09T09:50:50.500Z", "2020-03-10T09:50:51.500Z"}, true},
		{args{">2020-03-09T09:50:51.500Z", "2020-03-09T09:51:51.500Z"}, true},
		{args{">=2020-03-09T09:50:51.500Z", "2020-03-09T09:51:51.500Z"}, true},
		{args{">=2020-03-09T09:50:51.500Z", "2020-03-09T09:51:51.500Z"}, true},
		{args{"<=2020-03-09T09:50:51.500Z", "2020-03-08T09:50:51.500Z"}, true},
		{args{">2020-03-09T09:50:51.500Z", "2020-04-08T00:00:01.0Z"}, true},
		{args{">2020-03-09", "2020-04-08T00:00:01.0Z"}, true},
		{args{"> January 10th, 2020", "2020-04-08T00:00:01.0Z"}, true},
		{args{"<= March 1st, 2020", "2020-04-08T00:00:01.0Z"}, false},
		{args{"<= 01-mar-2020", "2020-04-08T00:00:01.0Z"}, false},
		{args{"<= 02-feb-2020", "2020-04-08T00:00:01.0Z"}, false},
		{args{">= 02-may-2020", "2020-04-08T00:00:01.0Z"}, false},
		{args{">= 02-jan-2020", "2020-04-08T00:00:01.0Z"}, true},
		{args{">= 03-jan-2020", "2020-04-08T00:00:01.0Z"}, true},
		{args{">= 02-Apr-2020", "2020-04-08T00:00:01.0Z"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.args.wanted, func(t *testing.T) {
			got, err := compareDate(tt.args.wanted, tt.args.got)
			if err != nil {
				t.Errorf("compareDate() = %v, error: %s", got, err)
			}
			if got != tt.want {
				t.Errorf("compareDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
