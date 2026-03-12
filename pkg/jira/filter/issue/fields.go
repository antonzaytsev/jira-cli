package issue

import (
	"github.com/ankitpokhrel/jira-cli/pkg/jira/filter"
)

// KeyIssueFields is a filter key for selecting specific issue fields.
const KeyIssueFields = filter.Key("issue-fields")

// FieldsFilter limits which fields are returned by the API.
type FieldsFilter struct {
	key   filter.Key
	value []string
}

// NewFieldsFilter constructs a filter to select specific fields.
func NewFieldsFilter(fields []string) FieldsFilter {
	return FieldsFilter{
		key:   KeyIssueFields,
		value: fields,
	}
}

// Key returns key of this filter.
func (f FieldsFilter) Key() filter.Key {
	return f.key
}

// Val returns value of this filter.
func (f FieldsFilter) Val() interface{} {
	return f.value
}
