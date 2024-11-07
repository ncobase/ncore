package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// IntField defines a generic boolean field mixin.
type IntField struct {
	mixin.Schema
	Field    string
	Comment  string
	Default  int
	Positive bool
}

// Fields implements the ent.Mixin interface for IntField.
func (m IntField) Fields() []ent.Field {
	f := field.Int(m.Field).
		Default(m.Default).
		Comment(m.Comment)

	if m.Positive {
		f = f.Positive()
	}
	return []ent.Field{f}
}

// Implement the Mixin interface.
var _ ent.Mixin = (*IntField)(nil)

// Specific mixins can be created using the generic BoolMixin.
var (
	Status        = IntField{Field: "status", Comment: "status: 0 activated, 1 unactivated, 2 disabled", Default: 0}
	Order         = IntField{Field: "order", Comment: "display order", Default: 0}
	Size          = IntField{Field: "size", Comment: "size in bytes", Default: 0}
	IncrementStep = IntField{Field: "increment_step", Comment: "Increment step", Default: 1}
	StartValue    = IntField{Field: "start_value", Comment: "Start value", Default: 1}
	CurrentValue  = IntField{Field: "current_value", Comment: "Current value", Default: 0}
)
