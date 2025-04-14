package mixin

import (
	"github.com/ncobase/ncore/types"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// JSONMixin defines a generic JSON field mixin.
type JSONMixin struct {
	ent.Schema
	Field    string
	Default  any
	Comment  string
	Optional bool
}

// Fields implements the ent.Mixin interface for JSONMixin.
func (m JSONMixin) Fields() []ent.Field {
	f := field.JSON(m.Field, m.Default).Default(m.Default).Comment(m.Comment)
	if m.Optional {
		f = f.Optional()
	}
	return []ent.Field{f}
}

// Implement the Mixin interface.
var _ ent.Mixin = (*JSONMixin)(nil)

// Specific mixins can be created using the generic JSONMixin.
var (
	ExtraProps = JSONMixin{Field: "extras", Default: types.JSON{}, Comment: "Extend properties", Optional: true}
	Author     = JSONMixin{Field: "author", Default: types.JSON{}, Comment: "Author information, e.g., {id: '', name: '', avatar: '', url: '', email: '', ip: ''}", Optional: true}
	Related    = JSONMixin{Field: "related", Default: types.JSON{}, Comment: "Related entity information, e.g., {id: '', name: '', type: 'user / topic /...'}", Optional: true}
	Leader     = JSONMixin{Field: "leader", Default: types.JSON{}, Comment: "Leader information, e.g., {id: '', name: '', avatar: '', url: '', email: '', ip: ''}", Optional: true}
	Links      = JSONMixin{Field: "links", Default: types.JSONArray{}, Comment: "List of social links or profile links", Optional: true}
	Payload    = JSONMixin{Field: "payload", Default: types.JSON{}, Comment: "Payload", Optional: true}
)
