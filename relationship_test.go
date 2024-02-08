package gochugaru_test

import (
	"fmt"

	gg "github.com/jzelinskie/gochugaru"
)

func ExampleRelationship_WithCaveat() {
	rel := gg.
		MustRelationshipFromTriple("document:example", "viewer", "user:jzelinskie").
		WithCaveat("only_on_tuesday", map[string]any{"day_of_the_week": "wednesday"})
	fmt.Println(rel)
	// Output:
	// {document example viewer user jzelinskie  only_on_tuesday map[day_of_the_week:wednesday]}
}
