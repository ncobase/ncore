package jwt

import (
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

	ErrNeedTokenProvider = TokenError("cannot sign token without token provider")
	ErrInvalidToken      = TokenError("invalid token")
	ErrTokenParsing      = TokenError("token parsing error")
)

// Token represents the token body.
type Token struct {
	JTI     string         `json:"jti"`
	Payload map[string]any `json:"payload"`
	Subject string         `json:"sub"`
	Expire  int64          `json:"exp"`
}

// generateToken generates a JWT token.
func generateToken(key string, token *Token) (string, error) {
	if key == "" {
		return "", ErrNeedTokenProvider
	}
	claims := jwtstd.MapClaims{
		"jti":     token.JTI,
		"sub":     token.Subject,
		"payload": token.Payload,
		"exp":     time.Now().Add(time.Millisecond * time.Duration(token.Expire)).Unix(),
	}
	t := jwtstd.NewWithClaims(jwtstd.SigningMethodHS256, claims)
	tokenString, err := t.SignedString([]byte(key))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateToken validates a JWT token.
func ValidateToken(key, tokenString string) (*jwtstd.Token, error) {
	if key == "" {
		return nil, ErrNeedTokenProvider
	}
	token, err := jwtstd.Parse(tokenString, func(token *jwtstd.Token) (any, error) {
		return []byte(key), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

// DecodeToken decodes a JWT token into its claims.
func DecodeToken(key, tokenString string) (map[string]any, error) {
	token, err := ValidateToken(key, tokenString)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, ErrInvalidToken
	}
	return token.Claims.(jwtstd.MapClaims), nil
}

// generateCustomToken generates a custom token with the provided subject and expiration.
func generateCustomToken(key, jti string, payload map[string]any, defaultSubject string, expireDuration time.Duration) (string, error) {
	subject := defaultSubject
	return generateToken(key, &Token{
		JTI:     jti,
		Payload: payload,
		Subject: subject,
		Expire:  expireDuration.Milliseconds(),
	})
}

// GenerateAccessToken generates an access token with a default expiration of 24 hours.
func GenerateAccessToken(key, jti string, payload map[string]any, subject ...string) (string, error) {
	return generateCustomToken(key, jti, payload, getSubject(subject, "access"), DefaultAccessTokenExpire)
}

// GenerateRegisterToken generates a register token with a default expiration of 60 minutes.
func GenerateRegisterToken(key, jti string, payload map[string]any, subject ...string) (string, error) {
	return generateCustomToken(key, jti, payload, getSubject(subject, "register"), DefaultRegisterTokenExpire)
}

// GenerateRefreshToken generates a refresh token with a default expiration of 7 days.
func GenerateRefreshToken(key, jti string, payload map[string]any, subject ...string) (string, error) {
	return generateCustomToken(key, jti, payload, getSubject(subject, "refresh"), DefaultRefreshTokenExpire)
}

// getSubject returns the subject if provided, otherwise returns the default subject.
func getSubject(subject []string, defaultSubject string) string {
	if len(subject) > 0 {
		return subject[0]
	}
	return defaultSubject
}

// GenerateAccessTokenWithExpiry generates an access token with a custom expiration duration.
func GenerateAccessTokenWithExpiry(key, jti string, payload map[string]any, expiry time.Duration, subject ...string) (string, error) {
	return generateCustomToken(key, jti, payload, getSubject(subject, "access"), expiry)
}

// GenerateRefreshTokenWithExpiry generates a refresh token with a custom expiration duration.
func GenerateRefreshTokenWithExpiry(key, jti string, payload map[string]any, expiry time.Duration, subject ...string) (string, error) {
	return generateCustomToken(key, jti, payload, getSubject(subject, "refresh"), expiry)
}

// GetTokenExpiryTime extracts the expiration time from a token.
func GetTokenExpiryTime(key, tokenString string) (time.Time, error) {
	claims, err := DecodeToken(key, tokenString)
	if err != nil {
		return time.Time{}, err
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, ErrTokenParsing
	}

	return time.Unix(int64(exp), 0), nil
}

// IsTokenExpired checks if a token is expired.
func IsTokenExpired(key, tokenString string) (bool, error) {
	expiryTime, err := GetTokenExpiryTime(key, tokenString)
	if err != nil {
		return true, err
	}

	return expiryTime.Before(time.Now()), nil
}
