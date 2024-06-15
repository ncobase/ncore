package jwt

import (
	"ncobase/common/validator"
	"time"

	jwtstd "github.com/golang-jwt/jwt/v5"
)

type TokenError string

func (e TokenError) Error() string {
	return string(e)
}

const (
	DefaultAccessTokenExpire   = time.Hour * 24
	DefaultRegisterTokenExpire = time.Minute * 60
	DefaultRefreshTokenExpire  = time.Hour * 24 * 7

	ErrNeedTokenProvider = TokenError("can not sign token without token provider")
	ErrInvalidToken      = TokenError("invalid token")
	ErrTokenParsing      = TokenError("token parsing error")
)

// Token token body
type Token struct {
	JTI     string         `json:"jti"`
	Payload map[string]any `json:"payload"`
	Subject string         `json:"sub"`
	Expire  int64          `json:"exp"`
}

// generateToken Generate token
func generateToken(key string, token *Token) (string, error) {
	if validator.IsEmpty(key) {
		return "", ErrNeedTokenProvider
	}
	claims := &jwtstd.MapClaims{
		"jti":     token.JTI,
		"sub":     token.Subject,
		"payload": token.Payload,
		"exp":     time.Now().UnixMilli() + token.Expire,
	}
	t := jwtstd.NewWithClaims(jwtstd.SigningMethodHS256, claims)
	tokenString, err := t.SignedString([]byte(key))
	if validator.IsNotNil(err) {
		return "", ErrNeedTokenProvider
	}
	return tokenString, nil
}

// ValidateToken Validate Token
func ValidateToken(key, token string) (*jwtstd.Token, error) {
	if validator.IsEmpty(key) {
		return nil, ErrNeedTokenProvider
	}
	t, err := jwtstd.Parse(token, func(t *jwtstd.Token) (any, error) {
		return []byte(key), nil
	})
	if validator.IsNotNil(err) {
		return nil, ErrTokenParsing
	}
	return t, nil
}

// DecodeToken Decode token
func DecodeToken(key, token string) (map[string]any, error) {
	t, err := ValidateToken(key, token)
	if validator.IsNotNil(err) {
		return nil, err
	}
	if !t.Valid {
		return nil, ErrInvalidToken
	}
	return t.Claims.(jwtstd.MapClaims), nil
}

// GenerateAccessToken Generate access token, default expire time is 24 hours
func GenerateAccessToken(key, jti string, payload map[string]any, subject ...string) (string, error) {
	defaultSubject := "access"
	if len(subject) > 0 {
		defaultSubject = subject[0]
	}
	return generateToken(key, &Token{
		JTI:     jti,
		Payload: payload,
		Subject: defaultSubject,
		Expire:  DefaultAccessTokenExpire.Milliseconds(),
	})
}

// GenerateRegisterToken Generate register token, default expire time is 60 minutes
func GenerateRegisterToken(key, jti string, payload map[string]any, subject ...string) (string, error) {
	defaultSubject := "register"
	if len(subject) > 0 {
		defaultSubject = subject[0]
	}
	return generateToken(key, &Token{
		JTI:     jti,
		Payload: payload,
		Subject: defaultSubject,
		Expire:  DefaultRegisterTokenExpire.Milliseconds(),
	})
}

// GenerateRefreshToken Generate refresh token, default expire time is 7 days
func GenerateRefreshToken(key, jti string, payload map[string]any, subject ...string) (string, error) {
	defaultSubject := "refresh"
	if len(subject) > 0 {
		defaultSubject = subject[0]
	}
	return generateToken(key, &Token{
		JTI:     jti,
		Payload: payload,
		Subject: defaultSubject,
		Expire:  DefaultRefreshTokenExpire.Milliseconds(),
	})
}
