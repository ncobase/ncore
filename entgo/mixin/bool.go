package mixin

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// BoolMixin defines a generic boolean field mixin.
type BoolMixin struct {
	mixin.Schema
	Field     string
	Comment   string
	Default   bool
	Immutable bool
	Optional  bool
}

// Fields implements the ent.Mixin interface for BoolMixin.
func (m BoolMixin) Fields() []ent.Field {
	f := field.Bool(m.Field).Comment(m.Comment)
	if m.Default == true || m.Default == false {
		f = f.Default(m.Default)
	}

	if m.Immutable {
		f = f.Immutable()
	}
	if m.Optional {
		f = f.Optional()
	}
	return []ent.Field{f}
}

// Implement the Mixin interface.
var _ ent.Mixin = (*BoolMixin)(nil)

// Specific mixins can be created using the generic BoolMixin.
var (
	Default     = BoolMixin{Field: "default", Comment: "is default", Optional: true}
	Markdown    = BoolMixin{Field: "markdown", Comment: "is markdown", Optional: true}
	Temp        = BoolMixin{Field: "temp", Comment: "is temp", Optional: true}
	Private     = BoolMixin{Field: "private", Comment: "is private", Optional: true}
	Approved    = BoolMixin{Field: "approved", Comment: "is approved", Optional: true}
	Disabled    = BoolMixin{Field: "disabled", Comment: "is disabled", Optional: true}
	Logged      = BoolMixin{Field: "logged", Comment: "is logged", Optional: true}
	System      = BoolMixin{Field: "system", Comment: "is system", Optional: true}
	Hidden      = BoolMixin{Field: "hidden", Comment: "is hidden", Optional: true}
	IsCertified = BoolMixin{Field: "is_certified", Comment: "is certified", Optional: true}
	IsAdmin     = BoolMixin{Field: "is_admin", Comment: "is admin", Optional: true}
	Autoload    = BoolMixin{Field: "autoload", Comment: "Whether to load the option automatically", Default: true, Optional: true}
)
