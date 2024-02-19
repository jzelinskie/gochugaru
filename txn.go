package gochugaru

import (
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

// Txn represents an atomic modification with option preconditions.
type Txn struct {
	updates  []*v1.RelationshipUpdate
	preconds []*v1.Precondition
}

// MustMatch modifies a transaction to only apply if the provided precondition
// is met.
func (b *Txn) MustMatch(f *Filter) {
	b.preconds = append(b.preconds, &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_MATCH,
		Filter:    f.filter,
	})
}

// MustNotMatch modifies a transaction to only apply if the provided
// precondition is not met.
func (b *Txn) MustNotMatch(f *Filter) {
	b.preconds = append(b.preconds, &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_NOT_MATCH,
		Filter:    f.filter,
	})
}

// Touch idempotently creates or updates a relationship.
//
// The performance of this operation can vary based on the SpiceDB datastore.
func (b *Txn) Touch(r Relationship) {
	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_TOUCH,
		Relationship: r.v1(),
	})
}

// Create inserts a new relationship and fails if the relationship already
// exists.
func (b *Txn) Create(r Relationship) {
	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_CREATE,
		Relationship: r.v1(),
	})
}

// Delete removes a relationship.
func (b *Txn) Delete(r Relationship) {
	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_DELETE,
		Relationship: r.v1(),
	})
}
