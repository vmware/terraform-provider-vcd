package vcd

import (
	"fmt"
	"regexp"

	"github.com/araddon/dateparse"
)

// This file contains general purpose functions to support queries for different entities
// Functions related to specific entities are named 'entity_type_filter.go'

// compareDate will get a date from string `got`, and will parse `wanted`
// for an expression containing an operator (>, <, >=, <=, ==) and a date
// (many formats supported, but 'YYYY-MM-DD[ hh:mm[:ss]]' preferred)
// For example:
// got:    "2020-03-09T09:50:51.500Z"
// wanted: ">= 2020-03-08"
// result: true
// got:    "2020-03-09T09:50:51.500Z"
// wanted: "< 02-mar-2020"
// result: false
// See https://github.com/araddon/dateparse for more info
func compareDate(wanted, got string) (bool, error) {

	reExpression := regexp.MustCompile(`(>=|<=|==|<|=|>)\s*(.+)`)

	expList := reExpression.FindAllStringSubmatch(wanted, -1)
	if len(expList) == 0 || len(expList[0]) == 0 {
		return false, fmt.Errorf("expression not found in '%s'", wanted)
	}

	operator := expList[0][1]
	wantedTime, err := dateparse.ParseStrict(expList[0][2])
	if err != nil {
		return false, err
	}

	gotTime, err := dateparse.ParseStrict(got)
	if err != nil {
		return false, err
	}

	wantedSeconds := wantedTime.Unix()
	gotSeconds := gotTime.Unix()

	switch operator {
	case "=", "==":
		return gotSeconds == wantedSeconds, nil
	case ">":
		return gotSeconds > wantedSeconds, nil
	case ">=":
		return gotSeconds >= wantedSeconds, nil
	case "<=":
		return gotSeconds <= wantedSeconds, nil
	case "<":
		return gotSeconds < wantedSeconds, nil
	default:
		return false, fmt.Errorf("unsupported operator '%s'", operator)
	}
}
