package gochugaru_test

import (
	"fmt"
	"testing"

	gg "github.com/jzelinskie/gochugaru"
)

func TestRelationshipParsingTriples(t *testing.T) {
	cases := []struct {
		name, resource, relation, subject string
		expectedErr                       error
	}{
		{"valid rel", "document:example", "viewer", "user:jzelinskie", nil},
		{"missing resource", "", "viewer", "user:jzelinskie", gg.ErrInvalidResource},
		{"missing relation", "document:example", "", "user:jzelinskie", gg.ErrInvalidRelation},
		{"missing subject", "document:example", "viewer", "", gg.ErrInvalidSubject},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := gg.RelationshipFromTriple(c.resource, c.relation, c.subject); err != c.expectedErr {
				t.Fatal(err)
			}
		})
	}
}

func ExampleMustRelationshipFromTriple() {
	// If you're using this in tests, you might want to try this
	// to abbreivate the function name because it's long by default:
	rel := gg.MustRelationshipFromTriple

	r := rel("document:example", "viewer", "user:jzelinskie")
	fmt.Println(r)
	// Output:
	// {document example viewer user jzelinskie   map[]}
}

func ExampleRelationship_WithCaveat() {
	rel := gg.
		MustRelationshipFromTriple("document:example", "viewer", "user:jzelinskie").
		WithCaveat("only_on_tuesday", map[string]any{"day_of_the_week": "wednesday"})
	fmt.Println(rel)
	// Output:
	// {document example viewer user jzelinskie  only_on_tuesday map[day_of_the_week:wednesday]}
}
