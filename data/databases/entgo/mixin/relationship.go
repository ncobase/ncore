package mixin

import (
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
	UserID         = IDMixin{Field: "user_id", Comment: "user id"}
	RoleID         = IDMixin{Field: "role_id", Comment: "role id"}
	PermissionID   = IDMixin{Field: "permission_id", Comment: "permission id"}
	GroupID        = IDMixin{Field: "group_id", Comment: "group id"}
	TenantID       = IDMixin{Field: "tenant_id", Comment: "tenant id"}
	DictionaryID   = IDMixin{Field: "dictionary_id", Comment: "dictionary id"}
	OptionID       = IDMixin{Field: "option_id", Comment: "option id"}
	MenuID         = IDMixin{Field: "menu_id", Comment: "menu id"}
	OrganizationID = IDMixin{Field: "organization_id", Comment: "organization id"}
	ParentID       = IDMixin{Field: "parent_id", Comment: "parent id"}
	TopicID        = IDMixin{Field: "topic_id", Comment: "topic id"}
	ReplyToMixin   = IDMixin{Field: "reply_to", Comment: "reply to object id"}
	TaxonomyID     = IDMixin{Field: "taxonomy_id", Comment: "taxonomy id"}
	StoreID        = IDMixin{Field: "store_id", Comment: "store id"}
	CatalogID      = IDMixin{Field: "catalog_id", Comment: "catalog id"}
	ObjectID       = IDMixin{Field: "object_id", Comment: "object id"}
	OAuthID        = IDMixin{Field: "oauth_id", Comment: "oauth id"}
	ChannelID      = IDMixin{Field: "channel_id", Comment: "channel id"}
)
