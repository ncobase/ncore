package mixin

import (
	"ncobase/common/nanoid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// IDMixin is a generic mixin for adding an ID field.
type IDMixin struct {
	ent.Schema
	Field      string
	Comment    string
	StorageKey string
	MaxLen     int
}

// Fields of the IDMixin.
func (m IDMixin) Fields() []ent.Field {
	f := field.String(m.Field).Comment(m.Comment).Optional()
	if m.MaxLen > 0 {
		f = f.MaxLen(m.MaxLen)
	}
	if m.StorageKey != "" {
		f = f.StorageKey(m.StorageKey)
	}
	return []ent.Field{
		f,
	}
}

// Indexes of the IDMixin.
func (m IDMixin) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields(m.Field),
	}
}

// Implement the Mixin interface.
var _ ent.Mixin = (*IDMixin)(nil)

// Specific mixins can be created using the generic IDMixin.
var (
	UserID         = IDMixin{Field: "user_id", Comment: "user id", MaxLen: nanoid.PrimaryKeySize}
	RoleID         = IDMixin{Field: "role_id", Comment: "role id", MaxLen: nanoid.PrimaryKeySize}
	PermissionID   = IDMixin{Field: "permission_id", Comment: "permission id", MaxLen: nanoid.PrimaryKeySize}
	GroupID        = IDMixin{Field: "group_id", Comment: "group id", MaxLen: nanoid.PrimaryKeySize}
	TenantID       = IDMixin{Field: "tenant_id", Comment: "tenant id", MaxLen: nanoid.PrimaryKeySize}
	OrganizationID = IDMixin{Field: "organization_id", Comment: "organization id", MaxLen: nanoid.PrimaryKeySize}
	ParentID       = IDMixin{Field: "parent_id", Comment: "parent id", MaxLen: nanoid.PrimaryKeySize}
	TopicID        = IDMixin{Field: "topic_id", Comment: "topic id", MaxLen: nanoid.PrimaryKeySize}
	ReplyToMixin   = IDMixin{Field: "reply_to", Comment: "reply to object id", MaxLen: nanoid.PrimaryKeySize}
	TaxonomyID     = IDMixin{Field: "taxonomy_id", Comment: "taxonomy id", MaxLen: nanoid.PrimaryKeySize}
	StoreID        = IDMixin{Field: "store_id", Comment: "store id", MaxLen: nanoid.PrimaryKeySize}
	CatalogID      = IDMixin{Field: "catalog_id", Comment: "catalog id", MaxLen: nanoid.PrimaryKeySize}
	ObjectID       = IDMixin{Field: "object_id", Comment: "object id", MaxLen: nanoid.PrimaryKeySize}
	OAuthID        = IDMixin{Field: "oauth_id", Comment: "oauth id"}
)
