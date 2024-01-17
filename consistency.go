package gochugaru

import v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"

type Consistency struct {
	v1c *v1.Consistency
}

// ConsistencyFull configures the request to evaluate at the most recent
// revision of the database.
//
// This is the least performant, but guarantees read consistency.
func ConsistencyFull() *Consistency {
	return &Consistency{
		v1c: &v1.Consistency{
			Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true},
		},
	}
}

// ConsistencyMinLatency configures the request to evaluate at the database's
// preferred revision.
//
// This provides the optimal performance and is the default if no consistency
// is specified.
func ConsistencyMinLatency() *Consistency {
	return &Consistency{
		v1c: &v1.Consistency{
			Requirement: &v1.Consistency_MinimizeLatency{MinimizeLatency: true},
		},
	}
}

// ConsistencyAtLeast configures the request to evaluate at the provided
// database revision or newer.
//
// This should be used to avoid read-after-write inconsistencies.
func ConsistencyAtLeast(revision string) *Consistency {
	return &Consistency{
		v1c: &v1.Consistency{
			Requirement: &v1.Consistency_AtLeastAsFresh{
				AtLeastAsFresh: &v1.ZedToken{Token: revision},
			},
		},
	}
}

// ConsistencySnapshot configures the request to evaluate at the provided
// database revision.
//
// This should be very rarely used because SpiceDB is designed to pick the
// optimal revision for a request.
func ConsistencySnapshot(revision string) *Consistency {
	return &Consistency{
		v1c: &v1.Consistency{
			Requirement: &v1.Consistency_AtExactSnapshot{
				AtExactSnapshot: &v1.ZedToken{Token: revision},
			},
		},
	}
}
