package gochugaru

import (
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

type Txn struct {
	updates  []*v1.RelationshipUpdate
	preconds []*v1.Precondition
}

func (b *Txn) MustMatch(f *Filter) {
	b.preconds = append(b.preconds, &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_MATCH,
		Filter:    f.filter,
	})
}

func (b *Txn) MustNotMatch(f *Filter) {
	b.preconds = append(b.preconds, &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_NOT_MATCH,
		Filter:    f.filter,
	})
}

func (b *Txn) Touch(r Relationship) {
	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_TOUCH,
		Relationship: r.v1(),
	})
}

func (b *Txn) Create(r Relationship) {
	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_CREATE,
		Relationship: r.v1(),
	})
}

func (b *Txn) Delete(r Relationship) {
	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_DELETE,
		Relationship: r.v1(),
	})
}
