package mixin

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// TimeMixin defines a generic time field mixin.
type TimeMixin struct {
	mixin.Schema
	Field         string
	Comment       string
	Default       func() time.Time
	UpdateDefault func() time.Time
	Optional      bool
	Immutable     bool
}

// Fields implements the ent.Mixin interface for TimeMixin.
func (m TimeMixin) Fields() []ent.Field {
	f := field.Time(m.Field).Comment(m.Comment)
	if m.Default != nil {
		f = f.Default(m.Default)
	}
	if m.UpdateDefault != nil {
		f = f.UpdateDefault(m.UpdateDefault)
	}
	if m.Optional {
		f = f.Optional()
	}
	if m.Immutable {
		f = f.Immutable()
	}
	return []ent.Field{f}
}

// Implement the Mixin interface.
var _ ent.Mixin = (*TimeMixin)(nil)

// Specific mixins can be created using the generic TimeMixin.
var (
	CreatedAt = TimeMixin{Field: "created_at", Comment: "created at", Default: time.Now, Immutable: true, Optional: true}
	UpdatedAt = TimeMixin{Field: "updated_at", Comment: "updated at", Default: time.Now, UpdateDefault: time.Now, Optional: true}
	DeletedAt = TimeMixin{Field: "deleted_at", Comment: "deleted at", Optional: true}
	ExpiredAt = TimeMixin{Field: "expired_at", Comment: "expired at", Optional: true}
	Expires   = TimeMixin{Field: "expires", Comment: "expires", Optional: true}
	Released  = TimeMixin{Field: "released", Comment: "released", Optional: true}
)

// TimeAt composes created_at and updated_at time fields.
type TimeAt struct{ mixin.Schema }

// Fields of the TimeAt mixin.
func (TimeAt) Fields() []ent.Field {
	return append(
		CreatedAt.Fields(),
		UpdatedAt.Fields()...,
	)
}

// Ensure TimeAt implements the Mixin interface.
var _ ent.Mixin = (*TimeAt)(nil)
