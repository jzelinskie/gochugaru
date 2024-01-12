package gochugaru

import v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"

type Filter struct {
	filter *v1.RelationshipFilter
}

func NewFilter(resourceType, optionalID, optionalRelation string) *Filter {
	return &Filter{
		filter: &v1.RelationshipFilter{
			ResourceType:       resourceType,
			OptionalResourceId: optionalID,
			OptionalRelation:   optionalRelation,
		},
	}
}

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

type PreconditionedFilter struct {
	filter   *v1.RelationshipFilter
	preconds []*v1.Precondition
}

func NewPreconditionedFilter(f *Filter) *PreconditionedFilter {
	return &PreconditionedFilter{
		filter: f.filter,
	}
}

func (pf *PreconditionedFilter) MustMatch(f *Filter) {
	pf.preconds = append(pf.preconds, &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_MATCH,
		Filter:    f.filter,
	})
}

func (pf *PreconditionedFilter) MustNotMatch(f *Filter) {
	pf.preconds = append(pf.preconds, &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_NOT_MATCH,
		Filter:    f.filter,
	})
}
