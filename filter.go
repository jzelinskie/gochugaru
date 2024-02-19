package gochugaru

import v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"

// Filter represents a filter to match against relationships.
type Filter struct {
	filter *v1.RelationshipFilter
}

// NewFilter creates a Filter used to match against the Resource of
// relationships.
//
// Filters must provide a Resource Type, but empty string can be used to forego
// any filtering on the Resource ID or Relation.
func NewFilter(resourceType, optionalID, optionalRelation string) *Filter {
	return &Filter{
		filter: &v1.RelationshipFilter{
			ResourceType:       resourceType,
			OptionalResourceId: optionalID,
			OptionalRelation:   optionalRelation,
		},
	}
}

// WithSubjectFilter modifies a Filter to also include matching against the
// Subject of relationships.
func (f *Filter) WithSubjectFilter(subjectType, optionalID, optionalRelation string) {
	f.filter.OptionalSubjectFilter = &v1.SubjectFilter{
		SubjectType:       subjectType,
		OptionalSubjectId: optionalID,
	}
	if optionalRelation != "" {
		f.filter.OptionalSubjectFilter.OptionalRelation = &v1.SubjectFilter_RelationFilter{
			Relation: optionalRelation,
		}
	}
}

// PreconditionedFilter represents a filter used to match or not match against
// relationships used as a precondition to performing another action.
type PreconditionedFilter struct {
	filter   *v1.RelationshipFilter
	preconds []*v1.Precondition
}

// NewPreconditionedFilter creates a PreconditionedFilter from an existing
// Filter that will only apply an action if all the preconditions are met.
func NewPreconditionedFilter(f *Filter) *PreconditionedFilter {
	return &PreconditionedFilter{
		filter: f.filter,
	}
}

// MustMatch modifies a PreconditionedFilter to only apply if the provided
// precondition is met.
func (pf *PreconditionedFilter) MustMatch(f *Filter) {
	pf.preconds = append(pf.preconds, &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_MATCH,
		Filter:    f.filter,
	})
}

// MustNotMatch modifies a PreconditionedFilter to only apply if the provided
// precondition is not met.
func (pf *PreconditionedFilter) MustNotMatch(f *Filter) {
	pf.preconds = append(pf.preconds, &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_NOT_MATCH,
		Filter:    f.filter,
	})
}
