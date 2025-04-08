package consts

// AuthorizationKey Authorization header key
const AuthorizationKey string = "Authorization"

// BearerKey Bearer token prefix
const BearerKey string = "Bearer "

// GinContextKey gin context key
const GinContextKey = "gin-context"

// TraceKey global trace id
const TraceKey string = "x-md-trace"

// UserKey global user id
const UserKey string = "x-md-uid"

// UsernameKey global username
const UsernameKey string = "x-md-uname"

// TokenKey global token
const TokenKey string = "x-md-token"

// TenantKey global tenant id
// TenantKey = XMdDomainKey
const TenantKey string = "x-md-tid"

// TotalKey result total with response
const TotalKey string = "x-md-total"
