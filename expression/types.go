package expression

import (
	"fmt"
	"time"
)

// TokenType represents expression token type
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenNumber
	TokenString
	TokenIdentifier
	TokenOperator
	TokenLParen
	TokenRParen
	TokenDot
	TokenComma
	TokenFunction
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

// Function represents an expression function
type Function struct {
	Name      string
	Handler   any
	Validator Validator
}

// Operator represents an operator
type Operator struct {
	Name       string
	Precedence int
	Handler    func(left, right any) (any, error)
}

// Validator validates function arguments
type Validator func(args []any) error

// Config represents engine configuration
type Config struct {
	MaxDepth        int
	Timeout         int
	AllowCustom     bool
	StrictMode      bool
	CacheEnabled    bool
	CacheSize       int
	CacheTTL        time.Duration
	MaxStringLength int
	MaxArrayLength  int
}

// DefaultConfig returns default engine configuration
func DefaultConfig() *Config {
	return &Config{
		MaxDepth:        10,
		Timeout:         5000,
		AllowCustom:     true,
		StrictMode:      true,
		CacheEnabled:    true,
		CacheSize:       1024 * 1024, // 1MB
		CacheTTL:        time.Hour,   // 1 hour
		MaxStringLength: 1024 * 1024,
		MaxArrayLength:  10000,
	}
}

// ExpressionError represents an expression error
type ExpressionError struct {
	Type    string // syntax, runtime, etc.
	Message string
	Line    int
	Col     int
}

// Error returns the error message
func (e *ExpressionError) Error() string {
	if e.Line > 0 && e.Col > 0 {
		return fmt.Sprintf("%s error at line %d, col %d: %s",
			e.Type, e.Line, e.Col, e.Message)
	}
	return fmt.Sprintf("%s error: %s", e.Type, e.Message)
}
