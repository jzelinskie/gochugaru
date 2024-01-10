package gochugaru

import (
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
)

type Txn struct {
	updates  []*v1.RelationshipUpdate
	preconds []*v1.Precondition
}

func (b *Txn) MustMatch(filter string)
func (b *Txn) MustNotMatch(filter string)

func (b *Txn) Touch(object, relation, subject string) {
	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_TOUCH,
		Relationship: strToRel(object, relation, subject),
	})
}

func (b *Txn) Create(object, relation, subject string) {
	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_CREATE,
		Relationship: strToRel(object, relation, subject),
	})
}

func (b *Txn) Delete(object, relation, subject string) {
	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_DELETE,
		Relationship: strToRel(object, relation, subject),
	})
}

func (b *Txn) CaveatedTouch(object, relation, subject, caveatName string, caveatCtx map[string]any) {
	r := strToRel(object, relation, subject)
	r.OptionalCaveat = mustCaveat(caveatName, caveatCtx)

	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_TOUCH,
		Relationship: r,
	})
}

func (b *Txn) CaveatedCreate(object, relation, subject, caveatName string, caveatCtx map[string]any) {
	r := strToRel(object, relation, subject)
	r.OptionalCaveat = mustCaveat(caveatName, caveatCtx)

	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_CREATE,
		Relationship: r,
	})
}

func (b *Txn) CaveatedDelete(object, relation, subject, caveatName string, caveatCtx map[string]any) {
	r := strToRel(object, relation, subject)
	r.OptionalCaveat = mustCaveat(caveatName, caveatCtx)

	b.updates = append(b.updates, &v1.RelationshipUpdate{
		Operation:    v1.RelationshipUpdate_OPERATION_DELETE,
		Relationship: r,
	})
}
