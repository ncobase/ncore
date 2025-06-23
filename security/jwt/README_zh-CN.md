# JWT 包

一个简单高效的 Go 应用 JWT 令牌管理库。

## 基本用法

```go
import "your-project/security/jwt"

// 创建令牌管理器
tm := jwt.NewTokenManager("your-secret-key")

// 自定义配置
config := &jwt.TokenConfig{
    AccessTokenExpiry:   2 * time.Hour,
    RefreshTokenExpiry:  7 * 24 * time.Hour,
    RegisterTokenExpiry: 30 * time.Minute,
}
tm := jwt.NewTokenManager("secret", config)
```

## 令牌生成

```go
payload := map[string]any{"username": "john", "role": "admin"}

// 生成令牌
accessToken, err := tm.GenerateAccessToken("user-123", payload)
refreshToken, err := tm.GenerateRefreshToken("user-123", payload)
registerToken, err := tm.GenerateRegisterToken("user-123", payload, "register")

// 自定义过期时间
customConfig := &jwt.TokenConfig{Expiry: 1 * time.Hour}
token, err := tm.GenerateAccessToken("user-123", payload, customConfig)
```

## 令牌验证与解码

```go
// 验证令牌
token, err := tm.ValidateToken(tokenString)

// 解码声明
claims, err := tm.DecodeToken(tokenString)

// 仅获取载荷
payload, err := tm.GetPayload(tokenString)

// 检查过期
expired := tm.IsTokenExpired(tokenString)
expiryTime, err := tm.GetTokenExpiry(tokenString)
```

## 令牌刷新

```go
// 按需刷新（剩余时间 < 30 分钟时刷新）
newToken, refreshed, err := tm.RefreshTokenIfNeeded(tokenString, 30*time.Minute)
if refreshed {
    // 使用新令牌
}
```

## 配置方法

```go
tm.SetSecret("new-secret")
secret := tm.GetSecret()
tm.SetAccessTokenExpiry(3 * time.Hour)
tm.SetRefreshTokenExpiry(14 * 24 * time.Hour)
tm.SetRegisterTokenExpiry(1 * time.Hour)
```

## 声明提取工具

```go
// 标准声明
tokenID := jwt.GetTokenID(claims)
subject := jwt.GetSubject(claims)
issuer := jwt.GetIssuer(claims)
audience := jwt.GetAudience(claims)
expiry := jwt.GetExpiration(claims)
issuedAt := jwt.GetIssuedAt(claims)
notBefore := jwt.GetNotBefore(claims)

// 载荷提取
payload := jwt.GetPayload(claims)
username := jwt.GetPayloadString(claims, "username")
isAdmin := jwt.GetPayloadBool(claims, "admin")
level := jwt.GetPayloadInt(claims, "level")
roles := jwt.GetPayloadStringSlice(claims, "roles")
hasKey := jwt.HasPayloadValue(claims, "key")

// 安全类型提取
str := jwt.GetString(data, "key")
num := jwt.GetInt(data, "key")
flag := jwt.GetBool(data, "key")
slice := jwt.GetStringSlice(data, "key")
nested := jwt.GetMap(data, "key")
```

## 令牌类型验证

```go
isAccess := jwt.IsAccessToken(claims)
isRefresh := jwt.IsRefreshToken(claims)

// 验证特定类型
err := jwt.ValidateTokenType(claims, "access")
```

## 令牌时序验证

```go
// 检查所有时序约束
err := jwt.ValidateTokenTiming(claims)

// 单独检查
expired := jwt.IsTokenExpired(claims)
active := jwt.IsTokenActive(claims)
stale := jwt.IsTokenStale(claims, 24*time.Hour)
```

## 工具函数

```go
// 检查切片成员
contains := jwt.ContainsValue(slice, "value")
containsAny := jwt.ContainsAnyValue(slice, "val1", "val2")
```
