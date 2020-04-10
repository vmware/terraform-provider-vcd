package govcd

import (
	"fmt"
	"regexp"
)

// A conditionDef is the data being carried by the filter engine when performing comparisons
type conditionDef struct {
	conditionType string        // it's one of SupportedFilters
	stored        interface{}   // Any value as handled by the filter being used
}

// A dateCondition can evaluate a date expression
type dateCondition struct {
	dateExpression string
}

// A regexpCondition is a generic filter that is the basis for other filters that require a regular expression
type regexpCondition struct {
	regExpression *regexp.Regexp
}

// an ipCondition is a condition that compares an IP using a regexp
type ipCondition regexpCondition

// a nameCondition is a condition that compares a name using a regexp
type nameCondition regexpCondition

// a metadataRegexpCondition compares the values corresponding to the given key using a regexp
type metadataRegexpCondition struct {
	key           string
	regExpression *regexp.Regexp
}

// matchName matches a name (passed in 'stored') to the name of the queryItem
func matchName(stored, item interface{}) (bool, error) {
	re, ok := stored.(nameCondition)
	if !ok {
		return false, fmt.Errorf("stored value is not a Name Regexp")
	}
	queryItem, ok := item.(QueryItem)
	if !ok {
		return false, fmt.Errorf("item is not a queryItem searchable by regex")
	}
	return re.regExpression.MatchString(queryItem.GetName()), nil
}

// matchName matches an IP (passed in 'stored') to the IP of the queryItem
func matchIp(stored, item interface{}) (bool, error) {
	re, ok := stored.(ipCondition)
	if !ok {
		return false, fmt.Errorf("stored value is not a Condition Regexp")
	}
	queryItem, ok := item.(QueryItem)
	if !ok {
		return false, fmt.Errorf("item is not a queryItem searchable by Ip")
	}
	ip := queryItem.GetIp()
	if ip == "" {
		return false, fmt.Errorf("%s %s doesn't have an IP", queryItem.GetType(), queryItem.GetName())
	}
	return re.regExpression.MatchString(ip), nil
}

// matchName matches a date (passed in 'stored') to the date of the queryItem
func matchDate(stored, item interface{}) (bool, error) {
	expr, ok := stored.(dateCondition)
	if !ok {
		return false, fmt.Errorf("stored value is not a condition date")
	}
	queryItem, ok := item.(QueryItem)
	if !ok {
		return false, fmt.Errorf("item is not a queryItem searchable by date")
	}
	return compareDate(expr.dateExpression, queryItem.GetDate())
}

// matchMetadata matches a value (passed in 'stored') to the metadata value retrieved from queryItem
func matchMetadata(stored, item interface{}) (bool, error) {
	re, ok := stored.(metadataRegexpCondition)
	if !ok {
		return false, fmt.Errorf("stored value is not a Metadata condition")
	}
	queryItem, ok := item.(QueryItem)
	if !ok {
		return false, fmt.Errorf("item is not a queryItem searchable by Metadata")
	}
	return re.regExpression.MatchString(queryItem.GetMetadataValue(re.key)), nil
}
