package mixin

import (
	"regexp"

	"github.com/ncobase/ncore/consts"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// StringMixin is a generic mixin for adding various fields.
type StringMixin struct {
	mixin.Schema
	Field       string
	Comment     string
	StorageKey  string
	Nillable    bool
	Optional    bool
	Unique      bool
	NotEmpty    bool
	Immutable   bool
	MaxLen      int
	MatchRegex  *regexp.Regexp
	Sensitive   bool
	DefaultFunc any
	Default     string
}

// Fields of the StringMixin.
func (m StringMixin) Fields() []ent.Field {
	f := field.String(m.Field).Comment(m.Comment)
	if m.StorageKey != "" {
		f = f.StorageKey(m.StorageKey)
	}
	if m.Nillable {
		f = f.Nillable()
	}
	if m.Optional {
		f = f.Optional()
	}
	if m.Unique {
		f = f.Unique()
	}
	if m.NotEmpty {
		f = f.NotEmpty()
	}
	if m.Immutable {
		f = f.Immutable()
	}
	if m.MaxLen > 0 {
		f = f.MaxLen(m.MaxLen)
	}
	if m.MatchRegex != nil {
		f = f.Match(m.MatchRegex)
	}
	if m.Sensitive {
		f = f.Sensitive()
	}
	if m.DefaultFunc != nil {
		f = f.DefaultFunc(m.DefaultFunc)
	}
	if m.Default != "" {
		f = f.Default(m.Default)
	}
	return []ent.Field{f}
}

// Indexes of the StringMixin.
func (m StringMixin) Indexes() []ent.Index {
	if m.Unique {
		return []ent.Index{
			index.Fields(m.Field).Unique(),
		}
	}
	return nil
}

// Implement the Mixin interface.
var _ ent.Mixin = (*StringMixin)(nil)

// Specific mixins can be created using the generic StringMixin.
var (
	TextStatus     = StringMixin{Field: "status", Comment: "Status, text status", Optional: true}
	Email          = StringMixin{Field: "email", Comment: "email", Optional: true}
	Username       = StringMixin{Field: "username", Comment: "username", Optional: true}
	UsernameUnique = StringMixin{Field: "username", Comment: "username", Unique: true, NotEmpty: true, Optional: true, MaxLen: 50, MatchRegex: regexp.MustCompile("^[a-zA-Z0-9._-]{3,20}$")}
	Password       = StringMixin{Field: "password", Comment: "password", Sensitive: true, Optional: true}
	Secret         = StringMixin{Field: "secret", Comment: "secret key", Optional: true}
	Phone          = StringMixin{Field: "phone", Comment: "phone", Optional: true}
	BankName       = StringMixin{Field: "bank_name", Comment: "bank name", Optional: true}
	CardNo         = StringMixin{Field: "card_no", Comment: "card no", Optional: true}
	CCV            = StringMixin{Field: "ccv", Comment: "ccv", Optional: true}
	Province       = StringMixin{Field: "province", Comment: "province", Optional: true}
	ZipCode        = StringMixin{Field: "zip_code", Comment: "zip code", Optional: true}
	City           = StringMixin{Field: "city", Comment: "city", Optional: true}
	District       = StringMixin{Field: "district", Comment: "district", Optional: true}
	Address        = StringMixin{Field: "address", Comment: "address", Optional: true}
	FirstName      = StringMixin{Field: "first_name", Comment: "first name", Optional: true}
	LastName       = StringMixin{Field: "last_name", Comment: "last name", Optional: true}
	DisplayName    = StringMixin{Field: "display_name", Comment: "display name", Optional: true}
	Language       = StringMixin{Field: "language", Comment: "language", Optional: true}
	About          = StringMixin{Field: "about", Comment: "about", Optional: true}
	Identifier     = StringMixin{Field: "identifier", Comment: "Identifier", Optional: true, NotEmpty: true}
	Name           = StringMixin{Field: "name", Comment: "name", Optional: true}
	NameUnique     = StringMixin{Field: "name", Comment: "name", Unique: true, NotEmpty: true, Optional: true}
	Prefix         = StringMixin{Field: "prefix", Comment: "prefix", Optional: true}
	Suffix         = StringMixin{Field: "suffix", Comment: "suffix", Optional: true}
	Label          = StringMixin{Field: "label", Comment: "label", Optional: true}
	Code           = StringMixin{Field: "code", Comment: "code", Optional: true}
	Slug           = StringMixin{Field: "slug", Comment: "slug / alias", Optional: true}
	SlugUnique     = StringMixin{Field: "slug", Comment: "slug / alias", Unique: true, Optional: true}
	Cover          = StringMixin{Field: "cover", Comment: "cover", Optional: true}
	Thumbnail      = StringMixin{Field: "thumbnail", Comment: "thumbnail", Optional: true}
	Path           = StringMixin{Field: "path", Comment: "path", Optional: true}
	Target         = StringMixin{Field: "target", Comment: "target", Optional: true}
	URL            = StringMixin{Field: "url", Comment: "url, website / link...", Optional: true}
	Icon           = StringMixin{Field: "icon", Comment: "icon", Optional: true}
	Perms          = StringMixin{Field: "perms", Comment: "perms", Optional: true}
	Color          = StringMixin{Field: "color", Comment: "color", Optional: true}
	Content        = StringMixin{Field: "content", Comment: "content, big text", Optional: true}
	Keywords       = StringMixin{Field: "keywords", Comment: "keywords", Optional: true}
	Copyright      = StringMixin{Field: "copyright", Comment: "copyright", Optional: true}
	Logo           = StringMixin{Field: "logo", Comment: "logo", Optional: true}
	LogoAlt        = StringMixin{Field: "logo_alt", Comment: "logo alt", Optional: true}
	Type           = StringMixin{Field: "type", Comment: "type", Optional: true}
	Storage        = StringMixin{Field: "storage", Comment: "storage type", Optional: true}
	Bucket         = StringMixin{Field: "bucket", Comment: "bucket", Optional: true}
	Endpoint       = StringMixin{Field: "endpoint", Comment: "endpoint", Optional: true}
	Action         = StringMixin{Field: "action", Comment: "action", Optional: true}
	Subject        = StringMixin{Field: "subject", Comment: "subject", Optional: true}
	Provider       = StringMixin{Field: "provider", Comment: "provider", Optional: true}
	AccessToken    = StringMixin{Field: "access_token", Comment: "access token", NotEmpty: true}
	RefreshToken   = StringMixin{Field: "refresh_token", Comment: "refresh token", NotEmpty: true}
	SessionID      = StringMixin{Field: "session_id", Comment: "session id", Optional: true}
	ShortBio       = StringMixin{Field: "short_bio", Comment: "short bio", Optional: true}
	Bio            = StringMixin{Field: "bio", Comment: "bio", Optional: true}
	Hash           = StringMixin{Field: "hash", Comment: "hash", Optional: true}
	Title          = StringMixin{Field: "title", Comment: "title", Optional: true}
	Caption        = StringMixin{Field: "caption", Comment: "caption", Optional: true}
	MediaType      = StringMixin{Field: "mime", Comment: "resource type", Optional: true}
	ExtensionName  = StringMixin{Field: "ext", Comment: "extension name", Optional: true}
	Memo           = StringMixin{Field: "memo", Comment: "Memo, big text", Optional: true}
	Remark         = StringMixin{Field: "remark", Comment: "Remark, big text", Optional: true}
	PType          = StringMixin{Field: "p_type", Comment: "permission type", Optional: true}
	Version        = StringMixin{Field: "version", Comment: "Version", Optional: true}
	V0             = StringMixin{Field: "v0", Comment: "version 0", Optional: true}
	V1             = StringMixin{Field: "v1", Comment: "version 1", Optional: true}
	V2             = StringMixin{Field: "v2", Comment: "version 2", Optional: true}
	V3             = StringMixin{Field: "v3", Comment: "version 3", Optional: true}
	V4             = StringMixin{Field: "v4", Comment: "version 4", Optional: true}
	V5             = StringMixin{Field: "v5", Comment: "version 5", Optional: true}
	V6             = StringMixin{Field: "v6", Comment: "version 6", Optional: true}
	V7             = StringMixin{Field: "v7", Comment: "version 7", Optional: true}
	CreatedBy      = StringMixin{Field: "created_by", Comment: "id of the creator", Optional: true, MaxLen: consts.PrimaryKeySize}
	UpdatedBy      = StringMixin{Field: "updated_by", Comment: "id of the last updater", Optional: true, MaxLen: consts.PrimaryKeySize}
	DeletedBy      = StringMixin{Field: "deleted_by", Comment: "id of the deleter", Optional: true, MaxLen: consts.PrimaryKeySize}
)

// DateFormat default date format
func DateFormat(f ...string) StringMixin {
	format := "20060102"
	if len(f) > 0 {
		format = f[0]
	}
	return StringMixin{Field: "date_format", Comment: "Date format, default YYYYMMDD", Optional: true, Default: format}
}

// OperatorBy combines CreatedBy, UpdatedBy, and DeletedBy fields into a single mixin.
type OperatorBy struct{ mixin.Schema }

// Fields of the OperatorBy mixin.
func (OperatorBy) Fields() []ent.Field {
	return append(
		CreatedBy.Fields(),
		UpdatedBy.Fields()...,
	)
}

// Ensure OperatorBy implements the Mixin interface.
var _ ent.Mixin = (*OperatorBy)(nil)

// Operator Specific mixins can be created using the generic OperatorBy.
var (
	Operator = OperatorBy{}
)

// TextMixin is a generic mixin for adding various fields.
type TextMixin struct {
	mixin.Schema
	Field       string
	Comment     string
	StorageKey  string
	Nillable    bool
	Optional    bool
	Unique      bool
	NotEmpty    bool
	Immutable   bool
	MaxLen      int
	MatchRegex  *regexp.Regexp
	Sensitive   bool
	DefaultFunc any
}

// Fields of the TextMixin.
func (m TextMixin) Fields() []ent.Field {
	f := field.Text(m.Field).Comment(m.Comment)
	if m.StorageKey != "" {
		f = f.StorageKey(m.StorageKey)
	}
	if m.Nillable {
		f = f.Nillable()
	}
	if m.Optional {
		f = f.Optional()
	}
	if m.Unique {
		f = f.Unique()
	}
	if m.NotEmpty {
		f = f.NotEmpty()
	}
	if m.Immutable {
		f = f.Immutable()
	}
	if m.MaxLen > 0 {
		f = f.MaxLen(m.MaxLen)
	}
	if m.MatchRegex != nil {
		f = f.Match(m.MatchRegex)
	}
	if m.Sensitive {
		f = f.Sensitive()
	}
	if m.DefaultFunc != nil {
		f = f.DefaultFunc(m.DefaultFunc)
	}
	return []ent.Field{f}
}

// Indexes of the TextMixin.
func (m TextMixin) Indexes() []ent.Index {
	if m.Unique {
		return []ent.Index{
			index.Fields(m.Field).Unique(),
		}
	}
	return nil
}

// Implement the Mixin interface.
var _ ent.Mixin = (*TextMixin)(nil)

// Specific mixins can be created using the generic TextMixin.
var (
	Value       = TextMixin{Field: "value", Comment: "value", Optional: true}
	Description = TextMixin{Field: "description", Comment: "description", Optional: true}
)
