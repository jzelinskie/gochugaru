package gochugaru

import (
	"strings"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func mustCaveat(name string, context map[string]any) *v1.ContextualizedCaveat {
	cc, err := structpb.NewStruct(context)
	if err != nil {
		panic("caveat created with non-utf8 context key")
	}

	return &v1.ContextualizedCaveat{
		CaveatName: name,
		Context:    cc,
	}
}

func strToRel(object, relation, subject string) *v1.Relationship {
	objectType, objectID, found := strings.Cut(object, ":")
	if !found {
		panic("invalid object: " + object)
	}

	var subjRel string
	subject, subjRel, _ = strings.Cut(subject, "#")

	subjType, subjID, found := strings.Cut(subject, ":")
	if !found {
		panic("invalid subject: " + subject)
	}

	return &v1.Relationship{
		Resource: &v1.ObjectReference{
			ObjectType: objectType,
			ObjectId:   objectID,
		},
		Relation: relation,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: subjType,
				ObjectId:   subjID,
			},
			OptionalRelation: subjRel,
		},
	}
}
