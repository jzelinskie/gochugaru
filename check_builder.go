package gochugaru

import v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"

// CheckBuilder is used to create a set of permissions to be checked.
type CheckBuilder struct {
	items       []*v1.BulkCheckPermissionRequestItem
	consistency *v1.Consistency
}

// WithConsistencyFull configures checks to evaluate at the most recent revision
// of the database.
//
// This is the least performant, but guarantees read consistency.
func (b *CheckBuilder) WithConsistencyFull() {
	b.consistency = &v1.Consistency{Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true}}
}

// WithConsistencyMinLatency configures checks to evaluate at the database's
// preferred revision.
//
// This provides the optimal performance and is the default if no consistency
// is specified.
func (b *CheckBuilder) WithConsistencyMinLatency() {
	b.consistency = &v1.Consistency{Requirement: &v1.Consistency_MinimizeLatency{MinimizeLatency: true}}
}

// WithConsistencyAtLeast configures checks to evaluate at the provided database
// revision or newer.
//
// This should be used to avoid read-after-write inconsistencies.
func (b *CheckBuilder) WithConsistencyAtLeast(revision string) {
	b.consistency = &v1.Consistency{Requirement: &v1.Consistency_AtLeastAsFresh{
		AtLeastAsFresh: &v1.ZedToken{Token: revision},
	}}
}

// WithConsistencySnapshot configures the checks to evaluate at the provided
// database revision.
//
// This should be very rarely used because SpiceDB is designed to pick the
// optimal revision for a request.
func (b *CheckBuilder) WithConsistencySnapshot(revision string) {
	b.consistency = &v1.Consistency{Requirement: &v1.Consistency_AtExactSnapshot{
		AtExactSnapshot: &v1.ZedToken{Token: revision},
	}}
}

// AddRelationship appends a relationship to be checked for permissionship.
func (b *CheckBuilder) AddRelationship(r Relationship) {
	v1Rel := r.v1()
	b.items = append(b.items, &v1.BulkCheckPermissionRequestItem{
		Resource:   v1Rel.Subject.Object,
		Permission: v1Rel.Relation,
		Subject:    v1Rel.Subject,
	})
}
