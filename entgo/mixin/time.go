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
	Default       func() int64
	UpdateDefault func() int64
	Optional      bool
	Immutable     bool
}

// Fields implements the ent.Mixin interface for TimeMixin.
func (m TimeMixin) Fields() []ent.Field {
	f := field.Int64(m.Field).Comment(m.Comment)
	if m.Default != nil {
		f = f.DefaultFunc(m.Default)
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
	CreatedAt = TimeMixin{
		Field:     "created_at",
		Comment:   "created at",
		Default:   func() int64 { return time.Now().UnixMilli() },
		Immutable: true,
		Optional:  true,
	}
	UpdatedAt = TimeMixin{
		Field:         "updated_at",
		Comment:       "updated at",
		Default:       func() int64 { return time.Now().UnixMilli() },
		UpdateDefault: func() int64 { return time.Now().UnixMilli() },
		Optional:      true,
	}
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
