package gochugaru

import (
	"errors"
	"strings"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type Relationshipper interface{ Relationship() Relationship }

type Relationship struct {
	ResourceType     string
	ResourceID       string
	ResourceRelation string
	SubjectType      string
	SubjectID        string
	SubjectRelation  string
	CaveatName       string
	CaveatContext    map[string]any
}

func (r Relationship) Relationship() Relationship { return r }
func (r Relationship) Permission() string         { return r.ResourceRelation }
func (r Relationship) HasCaveat() bool            { return r.CaveatName != "" }

func (r Relationship) Caveat() (name string, context map[string]any, exists bool) {
	return r.CaveatName, r.CaveatContext, r.HasCaveat()
}

func (r Relationship) WithCaveat(name string, context map[string]any) Relationship {
	return Relationship{
		ResourceType:     r.ResourceType,
		ResourceID:       r.ResourceID,
		ResourceRelation: r.ResourceRelation,
		SubjectType:      r.SubjectType,
		SubjectID:        r.SubjectID,
		SubjectRelation:  r.SubjectRelation,
		CaveatName:       name,
		CaveatContext:    context,
	}
}

func (r Relationship) Filter() *Filter {
	f := NewFilter(r.ResourceType, r.ResourceID, r.ResourceRelation)
	f.WithSubjectFilter(r.SubjectType, r.SubjectID, r.SubjectRelation)
	return f
}

func (r Relationship) v1() *v1.Relationship {
	return &v1.Relationship{
		Resource: &v1.ObjectReference{
			ObjectType: r.ResourceType,
			ObjectId:   r.ResourceID,
		},
		Relation: r.ResourceRelation,
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: r.SubjectType,
				ObjectId:   r.SubjectID,
			},
			OptionalRelation: r.SubjectRelation,
		},
		OptionalCaveat: r.caveat(),
	}
}

func fromV1(r *v1.Relationship) Relationship {
	var caveatName string
	var caveatContext map[string]any
	if r.OptionalCaveat != nil {
		caveatName = r.OptionalCaveat.CaveatName
		caveatContext = r.OptionalCaveat.Context.AsMap()
	}

	return Relationship{
		ResourceType:     r.Resource.ObjectType,
		ResourceID:       r.Resource.ObjectId,
		ResourceRelation: r.Relation,
		SubjectType:      r.Subject.Object.ObjectType,
		SubjectID:        r.Subject.Object.ObjectId,
		SubjectRelation:  r.Subject.OptionalRelation,
		CaveatName:       caveatName,
		CaveatContext:    caveatContext,
	}
}

func (r Relationship) caveat() *v1.ContextualizedCaveat {
	if name, context, ok := r.Caveat(); ok {
		cc, err := structpb.NewStruct(context)
		if err != nil {
			panic("caveat created with non-utf8 context key")
		}

		return &v1.ContextualizedCaveat{
			CaveatName: name,
			Context:    cc,
		}
	}

	return nil
}

type Object struct {
	Typ      string
	ID       string
	Relation string
}

func (o Object) Object() Object { return o }

type Objecter interface{ Object() Object }

func RelationshipFromObjects(resource, subject Objecter) Relationship {
	r, s := resource.Object(), subject.Object()
	return Relationship{
		ResourceType:     r.Typ,
		ResourceID:       r.ID,
		ResourceRelation: r.Relation,
		SubjectType:      s.Typ,
		SubjectID:        s.ID,
		SubjectRelation:  s.Relation,
	}
}

func MustRelationshipFromTriple(resource, relation, subject string) Relationship {
	r, err := RelationshipFromTriple(resource, relation, subject)
	if err != nil {
		panic(err)
	}
	return r
}

func RelationshipFromTriple(resource, relation, subject string) (Relationship, error) {
	return RelationshipFromTuple(resource+"#"+relation, subject)
}

func MustRelationshipFromTuple(resource, subject string) Relationship {
	r, err := RelationshipFromTuple(resource, subject)
	if err != nil {
		panic(err)
	}
	return r
}

func RelationshipFromTuple(resource, subject string) (Relationship, error) {
	var (
		r     Relationship
		found bool
	)

	resource, r.ResourceRelation, found = strings.Cut(resource, "#")
	if !found {
		return r, errors.New("missing resource relation")
	}

	r.ResourceType, r.ResourceID, found = strings.Cut(resource, ":")
	if !found {
		return r, errors.New("missing resource id")
	}

	// Optional
	subject, r.SubjectRelation, _ = strings.Cut(subject, "#")

	r.SubjectType, r.SubjectID, found = strings.Cut(subject, ":")
	if !found {
		return r, errors.New("missing subject id")
	}

	return r, nil
}

type RelationshipFunc func(r Relationship) error
