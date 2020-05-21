package govcd

import (
	"fmt"
	"regexp"

	"github.com/kr/pretty"
)

// A conditionDef is the data being carried by the filter engine when performing comparisons
type conditionDef struct {
	conditionType string      // it's one of SupportedFilters
	stored        interface{} // Any value as handled by the filter being used
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

// a parentCondition compares the entity parent name with the one stored
type parentCondition struct {
	parentName string
}

// a parentIdCondition compares the entity parent ID with the one stored
type parentIdCondition struct {
	parentId string
}

// matchParent matches the wanted parent name (passed in 'stored') to the parent of the queryItem
// Input:
//   * stored: the data of the condition (a parentCondition)
//   * item:   a QueryItem
// Returns:
//   * bool:   the result of the comparison
//   * string: a description of the operation
//   * error:  an error when the input is not as expected
func matchParent(stored, item interface{}) (bool, string, error) {
	condition, ok := stored.(parentCondition)
	if !ok {
		return false, "", fmt.Errorf("stored value is not a Parent condition (%# v)", pretty.Formatter(stored))
	}
	queryItem, ok := item.(QueryItem)
	if !ok {
		return false, "", fmt.Errorf("item is not a queryItem searchable by parent: %# v", pretty.Formatter(item))
	}
	parent := queryItem.GetParentName()

	return condition.parentName == parent, fmt.Sprintf("%s == %s", condition.parentName, queryItem.GetParentName()), nil
}

// matchParentId matches the wanted parent ID (passed in 'stored') to the parent ID of the queryItem
// The IDs being compared are filtered through extractUuid, to make them homogeneous
// Input:
//   * stored: the data of the condition (a parentCondition)
//   * item:   a QueryItem
// Returns:
//   * bool:   the result of the comparison
//   * string: a description of the operation
//   * error:  an error when the input is not as expected
func matchParentId(stored, item interface{}) (bool, string, error) {
	condition, ok := stored.(parentIdCondition)
	if !ok {
		return false, "", fmt.Errorf("stored value is not a parent ID condition (%# v)", pretty.Formatter(stored))
	}
	queryItem, ok := item.(QueryItem)
	if !ok {
		return false, "", fmt.Errorf("item is not a queryItem searchable by parent ID: %# v", pretty.Formatter(item))
	}
	parentId := queryItem.GetParentId()
	parentId = extractUuid(parentId)
	condition.parentId = extractUuid(condition.parentId)

	return condition.parentId == parentId, fmt.Sprintf("%s =~ %s", condition.parentId, parentId), nil
}

// matchName matches a name (passed in 'stored') to the name of the queryItem
// Input:
//   * stored: the data of the condition (a nameCondition)
//   * item:   a QueryItem
// Returns:
//   * bool:   the result of the comparison
//   * string: a description of the operation
//   * error:  an error when the input is not as expected
func matchName(stored, item interface{}) (bool, string, error) {
	re, ok := stored.(nameCondition)
	if !ok {
		return false, "", fmt.Errorf("stored value is not a Name Regexp (%# v)", pretty.Formatter(stored))
	}
	queryItem, ok := item.(QueryItem)
	if !ok {
		return false, "", fmt.Errorf("item is not a queryItem searchable by regex: %# v", pretty.Formatter(item))
	}
	return re.regExpression.MatchString(queryItem.GetName()), fmt.Sprintf("%s =~ %s", re.regExpression.String(), queryItem.GetName()), nil
}

// matchIp matches an IP (passed in 'stored') to the IP of the queryItem
// Input:
//   * stored: the data of the condition (an ipCondition)
//   * item:   a QueryItem
// Returns:
//   * bool:   the result of the comparison
//   * string: a description of the operation
//   * error:  an error when the input is not as expected
func matchIp(stored, item interface{}) (bool, string, error) {
	re, ok := stored.(ipCondition)
	if !ok {
		return false, "", fmt.Errorf("stored value is not a Condition Regexp (%# v)", pretty.Formatter(stored))
	}
	queryItem, ok := item.(QueryItem)
	if !ok {
		return false, "", fmt.Errorf("item is not a queryItem searchable by Ip: %# v", pretty.Formatter(item))
	}
	ip := queryItem.GetIp()
	if ip == "" {
		return false, "", fmt.Errorf("%s %s doesn't have an IP", queryItem.GetType(), queryItem.GetName())
	}
	return re.regExpression.MatchString(ip), fmt.Sprintf("%s =~ %s", re.regExpression.String(), queryItem.GetIp()), nil
}

// matchDate matches a date (passed in 'stored') to the date of the queryItem
// Input:
//   * stored: the data of the condition (a dateCondition)
//   * item:   a QueryItem
// Returns:
//   * bool:   the result of the comparison
//   * string: a description of the operation
//   * error:  an error when the input is not as expected
func matchDate(stored, item interface{}) (bool, string, error) {
	expr, ok := stored.(dateCondition)
	if !ok {
		return false, "", fmt.Errorf("stored value is not a condition date (%# v)", pretty.Formatter(stored))
	}
	queryItem, ok := item.(QueryItem)
	if !ok {
		return false, "", fmt.Errorf("item is not a queryItem searchable by date: %# v", pretty.Formatter(item))
	}
	if queryItem.GetDate() == "" {
		return false, "", nil
	}

	result, err := compareDate(expr.dateExpression, queryItem.GetDate())
	return result, fmt.Sprintf("%s %s", queryItem.GetDate(), expr.dateExpression), err
}

// matchMetadata matches a value (passed in 'stored') to the metadata value retrieved from queryItem
// Input:
//   * stored: the data of the condition (a metadataRegexpCondition)
//   * item:   a QueryItem
// Returns:
//   * bool:   the result of the comparison
//   * string: a description of the operation
//   * error:  an error when the input is not as expected
func matchMetadata(stored, item interface{}) (bool, string, error) {
	re, ok := stored.(metadataRegexpCondition)
	if !ok {
		return false, "", fmt.Errorf("stored value is not a Metadata condition (%# v)", pretty.Formatter(stored))
	}
	queryItem, ok := item.(QueryItem)
	if !ok {
		return false, "", fmt.Errorf("item is not a queryItem searchable by Metadata: %# v", pretty.Formatter(item))
	}
	return re.regExpression.MatchString(queryItem.GetMetadataValue(re.key)), fmt.Sprintf("metadata: %s -> %s", re.key, re.regExpression.String()), nil
}
