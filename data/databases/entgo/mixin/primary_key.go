package mixin

import (
	"github.com/ncobase/ncore/consts"
	"github.com/ncobase/ncore/utils/nanoid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

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

// CustomPrimaryKey allows customizing the primary key's length and default function
type CustomPrimaryKey struct {
	ent.Schema
	Length      int
	DefaultFunc func() string
}

// Fields of the custom primary key mixin
func (m CustomPrimaryKey) Fields() []ent.Field {
	f := field.String("id").
		Comment("primary key").
		Immutable().
		Unique()

	if m.Length > 0 {
		f = f.MaxLen(m.Length)
	}

	if m.DefaultFunc != nil {
		f = f.DefaultFunc(m.DefaultFunc)
	}

	return []ent.Field{f}
}

// Indexes of the CustomPrimaryKey
func (m CustomPrimaryKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id"),
	}
}

// NewCustomPrimaryKey creates a new CustomPrimaryKey mixin with specified length and default function
func NewCustomPrimaryKey(length int, defaultFunc func() string) CustomPrimaryKey {
	return CustomPrimaryKey{
		Length:      length,
		DefaultFunc: defaultFunc,
	}
}
