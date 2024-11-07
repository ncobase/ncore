package mixin

import (
	"ncobase/common/consts"
	"ncobase/common/nanoid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// // PrimaryKey adds primary key field.
// type PrimaryKey struct{ ent.Schema }
//
// // Fields of the primary key mixin.
// func (PrimaryKey) Fields() []ent.Field {
// 	return []ent.Field{
// 		field.String("id").Comment("primary key").Immutable().Unique().DefaultFunc(nanoid.PrimaryKey()), // primary key
// 	}
// }
//
// // Indexes of the PrimaryKey.
// func (PrimaryKey) Indexes() []ent.Index {
// 	return []ent.Index{
// 		index.Fields("id"),
// 	}
// }
//
// // primary key mixin must implement `Mixin` interface.
// var _ ent.Mixin = (*PrimaryKey)(nil)

var PrimaryKey = StringMixin{
	Field:       "id",
	Comment:     "primary key",
	Immutable:   true,
	Unique:      true,
	MaxLen:      consts.PrimaryKeySize,
	DefaultFunc: nanoid.PrimaryKey(),
}

// PrimaryKeyAlias adds a primary key alias field.
type PrimaryKeyAlias struct {
	ent.Schema
	AliasName string
	AliasKey  string
}

// Fields of the primary key alias mixin.
func (m PrimaryKeyAlias) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey(m.AliasKey).
			Comment(m.AliasName + " primary key alias").
			Unique().
			DefaultFunc(nanoid.PrimaryKey()), // primary key alias
	}
}

// Indexes of the PrimaryKeyAlias.
func (m PrimaryKeyAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id"),
	}
}

// primary key alias mixin must implement `Mixin` interface.
var _ ent.Mixin = (*PrimaryKeyAlias)(nil)

// NewPrimaryKeyAlias creates a new PrimaryKeyAlias mixin with the given alias name and key.
func NewPrimaryKeyAlias(aliasName, aliasKey string) PrimaryKeyAlias {
	return PrimaryKeyAlias{
		AliasName: aliasName,
		AliasKey:  aliasKey,
	}
}
