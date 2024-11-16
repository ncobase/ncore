package expression

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Expression represents an expression
type Expression struct {
	functions map[string]Function
	operators map[string]Operator
	cache     *Cache
	config    *Config
	mu        sync.RWMutex
}

// NewExpression creates a new expression engine with the given configuration.
//
// Usage:
//
//	config := &Config{
//	    MaxDepth:        10,              // Maximum expression nesting depth
//	    Timeout:         5000,            // Evaluation timeout in milliseconds
//	    AllowCustom:     true,            // Allow custom functions and operators
//	    StrictMode:      true,            // Enable strict syntax validation
//	    CacheEnabled:    true,            // Enable expression result caching
//	    CacheSize:       1024 * 1024,     // Cache size limit in bytes (1MB)
//	    CacheTTL:        time.Hour,       // Cache entry time-to-live
//	    MaxStringLength: 1024 * 1024,     // Maximum string length in bytes
//	    MaxArrayLength:  10000,           // Maximum array length
//	}
//
//	// Create engine with config
//	expr := NewExpression(config)
//
//	// Or use default configuration
//	expr := NewExpression(nil)
//
//	// Evaluate expression
//	vars := map[string]any{
//	    "x": 10,
//	    "y": 20,
//	}
//	result, err := expr.Evaluate(context.Background(), "x + y", vars)
//
//	// Register custom function
//	expr.RegisterFunction("double", func(x float64) float64 {
//	    return x * 2
//	}, validateOneNumber)
//
//	// Use built-in functions
//	result, err = expr.Evaluate(context.Background(), "abs(-10) + floor(3.7)", nil)
func NewExpression(config *Config) *Expression {
	if config == nil {
		config = DefaultConfig()
	}

	cacheConfig := &CacheConfig{
		MaxSize:         int64(config.CacheSize),
		TTL:             config.CacheTTL,
		CleanupInterval: time.Minute * 5,
	}

	e := &Expression{
		functions: make(map[string]Function),
		operators: make(map[string]Operator),
		cache:     NewCache(cacheConfig),
		config:    config,
	}

	// Register default functions
	e.registerDefaultFunctions()

	// Register default operators
	e.registerDefaultOperators()

	return e
}

// Evaluate evaluates an expression
func (e *Expression) Evaluate(ctx context.Context, expr string, variables map[string]any) (any, error) {
	// Add context timeout if configured
	if e.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(e.config.Timeout)*time.Millisecond)
		defer cancel()
	}

	// Early validation
	if err := e.validateExpression(expr); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Generate cache key based on expression and variables
	cacheKey := e.generateCacheKey(expr, variables)

	// Try cache first if enabled
	if e.config.CacheEnabled {
		if value, ok := e.cache.Get(cacheKey); ok {
			return value, nil
		}
	}

	// Tokenize expression
	tokens, err := e.tokenize(expr)
	if err != nil {
		return nil, fmt.Errorf("tokenize error: %w", err)
	}

	// Parse tokens into AST
	ast, err := e.parse(tokens)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	// Evaluate AST with recovery mechanism
	result, err := e.evaluateWithRecovery(ctx, ast, variables)
	if err != nil {
		return nil, fmt.Errorf("evaluation error: %w", err)
	}

	// Cache result if enabled
	if e.config.CacheEnabled {
		if err := e.cache.Set(cacheKey, result); err != nil {
			// Log cache error but don't fail the evaluation
			if e.config.StrictMode {
				return nil, fmt.Errorf("cache error: %w", err)
			}
		}
	}

	return result, nil
}

// validateExpression performs pre-evaluation validation
func (e *Expression) validateExpression(expr string) error {
	if expr == "" {
		return fmt.Errorf("empty expression")
	}

	if len(expr) > e.config.MaxStringLength {
		return fmt.Errorf("expression length %d exceeds maximum %d", len(expr), e.config.MaxStringLength)
	}

	depth := e.getExpressionDepth(expr)
	if depth > e.config.MaxDepth {
		return fmt.Errorf("expression depth %d exceeds maximum %d", depth, e.config.MaxDepth)
	}

	return e.validateOperators(expr)
}

// generateCacheKey creates a unique cache key for the expression and variables
func (e *Expression) generateCacheKey(expr string, variables map[string]any) string {
	if len(variables) == 0 {
		return expr
	}

	// Create a deterministic key that includes variables
	var builder strings.Builder
	builder.WriteString(expr)
	builder.WriteString(":")

	// Sort variable names for consistent ordering
	var keys []string
	for k := range variables {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Add variables to key
	for _, k := range keys {
		v := variables[k]
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(fmt.Sprintf("%v", v))
		builder.WriteString(";")
	}

	return builder.String()
}

// ValidateSyntax validates expression syntax without evaluating it
func (e *Expression) ValidateSyntax(expr string) error {
	// Validate depth
	if depth := e.getExpressionDepth(expr); depth > e.config.MaxDepth {
		return fmt.Errorf("expression depth %d exceeds maximum %d", depth, e.config.MaxDepth)
	}

	// Validate operators
	if err := e.validateOperators(expr); err != nil {
		return fmt.Errorf("operator validation error: %w", err)
	}

	// Tokenize expression
	tokens, err := e.tokenize(expr)
	if err != nil {
		return fmt.Errorf("tokenization error: %w", err)
	}

	// Validate syntax by attempting to parse
	_, err = e.parse(tokens)
	if err != nil {
		return fmt.Errorf("syntax error: %w", err)
	}

	// Additional syntax checks if in strict mode
	if e.config.StrictMode {
		if err := e.validateStrictSyntax(tokens); err != nil {
			return fmt.Errorf("strict syntax error: %w", err)
		}
	}

	return nil
}

// validateStrictSyntax performs additional syntax checks in strict mode
func (e *Expression) validateStrictSyntax(tokens []*Token) error {
	if len(tokens) == 0 {
		return fmt.Errorf("empty expression")
	}

	var (
		parenCount = 0
		lastToken  *Token
	)

	for i, token := range tokens {
		// Check parentheses matching
		switch token.Type {
		case TokenLParen:
			parenCount++
		case TokenRParen:
			parenCount--
			if parenCount < 0 {
				return fmt.Errorf("unmatched closing parenthesis at position %d", i)
			}
		default:
			continue
		}

		// Check for invalid token sequences
		if i > 0 {
			if err := e.validateTokenSequence(lastToken, token); err != nil {
				return fmt.Errorf("at position %d: %w", i, err)
			}
		}

		lastToken = token
	}

	// Check final parentheses count
	if parenCount != 0 {
		return fmt.Errorf("unmatched parentheses: missing %d closing parenthesis", parenCount)
	}

	// Check if expression ends with valid token
	if lastToken != nil && !isValidEndToken(lastToken.Type) {
		return fmt.Errorf("expression cannot end with token type %v", lastToken.Type)
	}

	return nil
}

// validateTokenSequence checks if two consecutive token types are valid
func (e *Expression) validateTokenSequence(prev, curr *Token) error {
	switch curr.Type {
	case TokenOperator:
		if prev.Type == TokenOperator {
			return fmt.Errorf("consecutive operators not allowed")
		}
		if prev.Type == TokenLParen {
			if !isUnaryOperator(curr.Value) {
				return fmt.Errorf("binary operator cannot follow left parenthesis")
			}
		}
	case TokenRParen:
		if prev.Type == TokenOperator {
			return fmt.Errorf("right parenthesis cannot follow operator")
		}
		if prev.Type == TokenLParen {
			return fmt.Errorf("empty parentheses not allowed")
		}
	case TokenNumber, TokenString, TokenIdentifier:
		if prev.Type == TokenNumber || prev.Type == TokenString || prev.Type == TokenIdentifier {
			return fmt.Errorf("consecutive values not allowed")
		}
		if prev.Type == TokenRParen {
			return fmt.Errorf("value cannot follow right parenthesis")
		}
	case TokenLParen:
		if prev.Type == TokenNumber || prev.Type == TokenString || prev.Type == TokenIdentifier || prev.Type == TokenRParen {
			return fmt.Errorf("left parenthesis cannot follow value or right parenthesis")
		}
	case TokenComma:
		if prev.Type == TokenOperator || prev.Type == TokenLParen || prev.Type == TokenComma {
			return fmt.Errorf("invalid token sequence near comma")
		}
	default:
		return fmt.Errorf("invalid token sequence")
	}

	return nil
}

// isValidEndToken checks if a token type can validly end an expression
func isValidEndToken(t TokenType) bool {
	switch t {
	case TokenNumber, TokenString, TokenIdentifier, TokenRParen:
		return true
	default:
		return false
	}
}

// getExpressionDepth returns the maximum depth of the expression
func (e *Expression) getExpressionDepth(expr string) int {
	depth := 0
	maxDepth := 0
	for _, char := range expr {
		if char == '(' {
			depth++
			if depth > maxDepth {
				maxDepth = depth
			}
		} else if char == ')' {
			depth--
		}
	}
	return maxDepth
}

// validateSecurity validates expression security
func (e *Expression) validateSecurity(expr string) error {
	if len(expr) > e.config.MaxStringLength {
		return fmt.Errorf("expression length exceeds maximum allowed (%d)", e.config.MaxStringLength)
	}

	depth := e.getExpressionDepth(expr)
	if depth > e.config.MaxDepth {
		return fmt.Errorf("expression depth %d exceeds maximum allowed (%d)", depth, e.config.MaxDepth)
	}

	return nil
}

// validateOperators validates expression operators
func (e *Expression) validateOperators(expr string) error {
	tokens, err := e.tokenize(expr)
	if err != nil {
		return err
	}

	var lastWasOperator bool
	parenCount := 0

	for _, token := range tokens {
		switch token.Type {
		case TokenOperator:
			if lastWasOperator {
				return fmt.Errorf("consecutive operators not allowed at line %d, col %d", token.Line, token.Col)
			}
			lastWasOperator = true
		case TokenLParen:
			parenCount++
		case TokenRParen:
			parenCount--
			if parenCount < 0 {
				return fmt.Errorf("unmatched parenthesis at line %d, col %d", token.Line, token.Col)
			}
		default:
			lastWasOperator = false
		}
	}

	if parenCount != 0 {
		return fmt.Errorf("unmatched parentheses in expression")
	}

	return nil
}

// RegisterFunction registers a new function with validation
func (e *Expression) RegisterFunction(name string, handler any, validator Validator) error {
	if _, exists := e.functions[name]; exists {
		return fmt.Errorf("function %s already registered", name)
	}
	if handler == nil {
		return fmt.Errorf("handler for function %s cannot be nil", name)
	}
	e.functions[name] = Function{Name: name, Handler: handler, Validator: validator}
	return nil
}

// RegisterOperator registers a new operator with validation
func (e *Expression) RegisterOperator(name string, precedence int, handler func(left, right any) (any, error)) error {
	if _, exists := e.operators[name]; exists {
		return fmt.Errorf("operator %s already registered", name)
	}
	if handler == nil {
		return fmt.Errorf("handler for operator %s cannot be nil", name)
	}
	e.operators[name] = Operator{Name: name, Precedence: precedence, Handler: handler}
	return nil
}

// optimizer represents an expression optimizer
type optimizer struct {
	expr string
}

// optimize removes unnecessary spaces and optimizes boolean expressions
func (o *optimizer) optimize() string {
	// Remove unnecessary spaces
	o.expr = strings.TrimSpace(o.expr)
	o.expr = strings.Join(strings.Fields(o.expr), " ")

	// Optimize boolean expressions
	o.expr = strings.Replace(o.expr, "true && true", "true", -1)
	o.expr = strings.Replace(o.expr, "false || false", "false", -1)
	o.expr = strings.Replace(o.expr, "!false", "true", -1)
	o.expr = strings.Replace(o.expr, "!true", "false", -1)

	return o.expr
}

// safeEvaluate evaluates an expression safely
func (e *Expression) safeEvaluate(ctx context.Context, expr string, variables map[string]any) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic recovered in expression evaluation: %v", r)
		}
	}()

	// Validate expression
	if err := e.validateSecurity(expr); err != nil {
		return nil, err
	}

	// Optimize expression
	opt := &optimizer{expr: expr}
	optimizedExpr := opt.optimize()

	// Check for cached result
	if e.config.CacheEnabled {
		if cached, ok := e.cache.Get(optimizedExpr); ok {
			return cached, nil
		}
	}

	// Tokenize expression
	tokens, err := e.tokenize(optimizedExpr)
	if err != nil {
		return nil, fmt.Errorf("tokenization error: %w", err)
	}

	ast, err := e.parse(tokens)
	if err != nil {
		return nil, fmt.Errorf("parsing error: %w", err)
	}

	result, err = ast.Evaluate(ctx, variables)
	if err != nil {
		return nil, fmt.Errorf("evaluation error: %w", err)
	}

	// Cache result
	if e.config.CacheEnabled {
		err := e.cache.Set(optimizedExpr, result)
		if err != nil {
			return nil, fmt.Errorf("cache error: %w", err)
		}
	}

	return result, nil
}

// Node represents an AST node
type Node interface {
	Evaluate(ctx context.Context, variables map[string]any) (any, error)
}

// AST Node implementations

// NumberNode represents a number node
type NumberNode struct {
	Value float64
}

// Evaluate evaluates the number node
func (n *NumberNode) Evaluate(ctx context.Context, variables map[string]any) (any, error) {
	return n.Value, nil
}

// StringNode represents a string node
type StringNode struct {
	Value string
}

// Evaluate evaluates the string node
func (n *StringNode) Evaluate(ctx context.Context, variables map[string]any) (any, error) {
	return n.Value, nil
}

// IdentifierNode represents an identifier node
type IdentifierNode struct {
	Name string
}

// Evaluate evaluates the identifier node
func (n *IdentifierNode) Evaluate(ctx context.Context, variables map[string]any) (any, error) {
	value, ok := variables[n.Name]
	if !ok {
		return nil, fmt.Errorf("undefined variable: %s", n.Name)
	}
	return value, nil
}

// BinaryOpNode represents a binary operation node (e.g., +, -, *, /)
type BinaryOpNode struct {
	Left     Node
	Operator Operator
	Right    Node
}

// Evaluate evaluates the binary operation node
func (n *BinaryOpNode) Evaluate(ctx context.Context, variables map[string]any) (any, error) {
	leftVal, err := n.Left.Evaluate(ctx, variables)
	if err != nil {
		return nil, fmt.Errorf("error evaluating left operand: %w", err)
	}

	rightVal, err := n.Right.Evaluate(ctx, variables)
	if err != nil {
		return nil, fmt.Errorf("error evaluating right operand: %w", err)
	}

	result, err := n.Operator.Handler(leftVal, rightVal)
	if err != nil {
		return nil, fmt.Errorf("error applying operator %s: %w", n.Operator.Name, err)
	}
	return result, nil
}

// FunctionCallNode represents a function call node
type FunctionCallNode struct {
	Name string
	Args []Node
}

// Evaluate evaluates the function call node
func (f *FunctionCallNode) Evaluate(ctx context.Context, variables map[string]any) (any, error) {
	engine := ctx.Value("engine").(*Expression) // get engine from context
	engine.mu.RLock()
	defer engine.mu.RUnlock()

	function, exists := engine.functions[f.Name]
	if !exists {
		return nil, fmt.Errorf("function %s not registered", f.Name)
	}

	args := make([]any, len(f.Args))
	for i, argNode := range f.Args {
		result, err := argNode.Evaluate(ctx, variables)
		if err != nil {
			return nil, fmt.Errorf("error evaluating argument %d for function %s: %w", i, f.Name, err)
		}
		args[i] = result
	}

	// validate arguments
	if function.Validator != nil {
		if err := function.Validator(args); err != nil {
			return nil, fmt.Errorf("invalid arguments for function %s: %w", f.Name, err)
		}
	}

	// Call function
	handlerValue := reflect.ValueOf(function.Handler)
	handlerType := handlerValue.Type()

	// Check number of arguments
	if len(args) < handlerType.NumIn() {
		return nil, fmt.Errorf("not enough arguments for function %s: expected %d, got %d",
			f.Name, handlerType.NumIn(), len(args))
	}

	// Prepare arguments
	params := make([]reflect.Value, handlerType.NumIn())
	for i := 0; i < handlerType.NumIn(); i++ {
		if i < len(args) {
			params[i] = reflect.ValueOf(args[i])
		}
	}
	// Call function
	results := handlerValue.Call(params)
	if len(results) == 0 {
		return nil, fmt.Errorf("function returned no value")
	}
	// Check error
	if len(results) > 1 && !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	return results[0].Interface(), nil
}

// Parser implementation

// parser represents an expression parser
type parser struct {
	tokens    []*Token
	operators map[string]Operator
	current   int
	depth     int
	config    *Config
}

// parse parses an expression
func (e *Expression) parse(tokens []*Token) (Node, error) {
	p := &parser{
		tokens:    tokens,
		operators: e.operators,
		config:    e.config,
	}
	return p.parseExpression()
}

func (p *parser) parseExpression() (Node, error) {
	// Check expression depth
	p.depth++
	if p.config.MaxDepth > 0 && p.depth > p.config.MaxDepth {
		return nil, fmt.Errorf("max expression depth exceeded")
	}
	defer func() { p.depth-- }()

	// Parse expression
	return p.parseBinaryExpression(0)
}

// parseBinaryExpression parses a binary expression
func (p *parser) parseBinaryExpression(precedence int) (Node, error) {
	left, err := p.parsePrimaryExpression()
	if err != nil {
		return nil, err
	}

	for {
		if p.current >= len(p.tokens) {
			break
		}

		token := p.tokens[p.current]
		if token.Type != TokenOperator {
			break
		}

		op, ok := p.operators[token.Value]
		if !ok || op.Precedence < precedence {
			break
		}

		p.current++
		right, err := p.parseBinaryExpression(op.Precedence + 1)
		if err != nil {
			return nil, err
		}

		left = &BinaryOpNode{
			Operator: op,
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

// parsePrimaryExpression parses a primary expression
func (p *parser) parsePrimaryExpression() (Node, error) {
	token := p.tokens[p.current]
	p.current++

	switch token.Type {
	case TokenNumber:
		value, err := strconv.ParseFloat(token.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", token.Value)
		}
		return &NumberNode{Value: value}, nil

	case TokenString:
		return &StringNode{Value: token.Value}, nil

	case TokenIdentifier:
		// Look ahead for function call
		if p.current < len(p.tokens) && p.tokens[p.current].Type == TokenLParen {
			return p.parseFunctionCall(token.Value)
		}
		return &IdentifierNode{Name: token.Value}, nil

	case TokenLParen:
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		if p.current >= len(p.tokens) || p.tokens[p.current].Type != TokenRParen {
			return nil, fmt.Errorf("expected )")
		}
		p.current++
		return expr, nil

	default:
		return nil, fmt.Errorf("unexpected token: %v", token)
	}
}

// parasFunctionCall parses a function call
func (p *parser) parseFunctionCall(name string) (Node, error) {
	p.current++ // Skip (
	var args []Node

	// Parse arguments
	for {
		if p.current >= len(p.tokens) {
			return nil, fmt.Errorf("unexpected end of input")
		}

		if p.tokens[p.current].Type == TokenRParen {
			p.current++
			break
		}

		if len(args) > 0 {
			if p.tokens[p.current].Type != TokenComma {
				return nil, fmt.Errorf("expected `,`")
			}
			p.current++
		}

		arg, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	return &FunctionCallNode{
		Name: name,
		Args: args,
	}, nil
}

// Tokenizer methods

// tokenize splits the expression into tokens
func (e *Expression) tokenize(expr string) ([]*Token, error) {
	var tokens []*Token
	var current int
	line := 1
	col := 1

	for current < len(expr) {
		char := expr[current]

		switch {
		case isWhitespace(char):
			if char == '\n' {
				line++
				col = 1
			} else {
				col++
			}
			current++
			continue

		case isDigit(char):
			token, newPos, err := readNumber(expr, current, line, col)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			col += newPos - current
			current = newPos

		case isLetter(char):
			token, newPos := readIdentifier(expr, current, line, col)
			tokens = append(tokens, token)
			col += newPos - current
			current = newPos

		case char == '"' || char == '\'':
			token, newPos, err := readString(expr, current, line, col)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token)
			col += newPos - current
			current = newPos

		case isOperator(char):
			token, newPos := readOperator(expr, current, line, col)
			tokens = append(tokens, token)
			col += newPos - current
			current = newPos

		case char == '(' || char == ')' || char == '.' || char == ',':
			tokens = append(tokens, &Token{
				Type:  tokenType(char),
				Value: string(char),
				Line:  line,
				Col:   col,
			})
			current++
			col++

		default:
			return nil, fmt.Errorf("unexpected character at line %d, col %d: %c", line, col, char)
		}
	}

	tokens = append(tokens, &Token{Type: TokenEOF, Line: line, Col: col})
	return tokens, nil
}

// Cache implementations

// AddToCache adds a value to the cache and enforces LRU policy
func (e *Expression) AddToCache(key string, value any) {
	if e.config.CacheEnabled {
		err := e.cache.Set(key, value)
		if err != nil {
			fmt.Printf("failed to add to cache: %v", err)
			return
		}
	}
}

// enforceCacheSize enforces the cache size limit by removing least recently used items
func (e *Expression) enforceCacheSize() {
	if e.config.CacheEnabled {
		for e.cache.Size() > int64(e.config.CacheSize) {
			e.cache.evictOldest()
		}
	}
}

// cacheSize calculates the current cache size
func (e *Expression) cacheSize() int {
	if e.config.CacheEnabled {
		return e.cache.Len()
	}
	return 0
}

// evaluateWithRecovery evaluates AST with panic recovery
func (e *Expression) evaluateWithRecovery(ctx context.Context, node Node, variables map[string]any) (result any, err error) {
	// Set up recovery
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during evaluation: %v", r)
		}
	}()

	// Create context with timeout if needed
	if e.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(e.config.Timeout)*time.Millisecond)
		defer cancel()
	}

	// Add engine to context for function evaluation
	ctx = context.WithValue(ctx, "engine", e)

	// Evaluate with context checks
	type evalResult struct {
		result any
		err    error
	}
	resultCh := make(chan evalResult, 1)
	done := make(chan struct{})

	go func() {
		defer close(done)
		r, e := node.Evaluate(ctx, variables)
		select {
		case resultCh <- evalResult{r, e}:
		case <-ctx.Done():
		}
	}()

	select {
	case <-ctx.Done():
		<-done
		return nil, ctx.Err()
	case res := <-resultCh:
		<-done
		return res.result, res.err
	}
}

// EvaluateParallel evaluates multiple expressions in parallel
func (e *Expression) EvaluateParallel(ctx context.Context, exprs []string, variables map[string]any) ([]any, error) {
	results := make([]any, len(exprs))
	errors := make([]error, len(exprs))
	var wg sync.WaitGroup

	for i, expr := range exprs {
		wg.Add(1)
		go func(i int, expr string) {
			defer wg.Done()
			results[i], errors[i] = e.Evaluate(ctx, expr, variables)
		}(i, expr)
	}

	wg.Wait()

	// Check errors
	for _, err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

type Metrics struct {
	ParseTime   time.Duration
	EvalTime    time.Duration
	CacheHits   int64
	CacheMisses int64
}

// EvaluateWithMetrics evaluates an expression with metrics
func (e *Expression) EvaluateWithMetrics(ctx context.Context, expr string, variables map[string]any) (any, *Metrics, error) {
	metrics := &Metrics{}
	// Parse
	parseStart := time.Now()
	tokens, err := e.tokenize(expr)
	if err != nil {
		return nil, metrics, err
	}
	ast, err := e.parse(tokens)
	if err != nil {
		return nil, metrics, err
	}
	metrics.ParseTime = time.Since(parseStart)

	// Evaluate
	evalStart := time.Now()
	result, err := ast.Evaluate(ctx, variables)
	metrics.EvalTime = time.Since(evalStart)

	return result, metrics, err
}

// Register default functions and operators
func (e *Expression) registerDefaultFunctions() {
	// Math functions
	e.functions["abs"] = Function{
		Name: "abs",
		Handler: func(x float64) float64 {
			return math.Abs(x)
		},
		Validator: validateOneNumber,
	}

	e.functions["round"] = Function{
		Name: "round",
		Handler: func(x float64) float64 {
			return math.Round(x)
		},
		Validator: validateOneNumber,
	}

	e.functions["ceil"] = Function{
		Name: "ceil",
		Handler: func(x float64) float64 {
			return math.Ceil(x)
		},
		Validator: validateOneNumber,
	}

	e.functions["floor"] = Function{
		Name: "floor",
		Handler: func(x float64) float64 {
			return math.Floor(x)
		},
		Validator: validateOneNumber,
	}

	// String functions
	e.functions["len"] = Function{
		Name: "len",
		Handler: func(s string) int {
			return len(s)
		},
		Validator: validateOneString,
	}

	e.functions["lower"] = Function{
		Name: "lower",
		Handler: func(s string) string {
			return strings.ToLower(s)
		},
		Validator: validateOneString,
	}

	e.functions["upper"] = Function{
		Name: "upper",
		Handler: func(s string) string {
			return strings.ToUpper(s)
		},
		Validator: validateOneString,
	}

	e.functions["trim"] = Function{
		Name: "trim",
		Handler: func(s string) string {
			return strings.TrimSpace(s)
		},
		Validator: validateOneString,
	}

	// Array functions
	e.functions["count"] = Function{
		Name: "count",
		Handler: func(arr []any) int {
			return len(arr)
		},
		Validator: validateOneArray,
	}

	e.functions["sum"] = Function{
		Name: "sum",
		Handler: func(arr []any) float64 {
			var sum float64
			for _, v := range arr {
				if num, ok := toNumber(v); ok {
					sum += num
				}
			}
			return sum
		},
		Validator: validateOneArray,
	}

	// Date functions
	e.functions["now"] = Function{
		Name: "now",
		Handler: func() time.Time {
			return time.Now()
		},
	}

	e.functions["date"] = Function{
		Name: "date",
		Handler: func(s string) (time.Time, error) {
			return time.Parse(time.RFC3339, s)
		},
		Validator: validateOneString,
	}

	// Logical functions
	e.functions["if"] = Function{
		Name: "if",
		Handler: func(cond bool, t, f any) any {
			if cond {
				return t
			}
			return f
		},
		Validator: validateThreeArgs,
	}

	e.functions["coalesce"] = Function{
		Name: "coalesce",
		Handler: func(args ...any) any {
			for _, arg := range args {
				if arg != nil {
					return arg
				}
			}
			return nil
		},
	}
}

// Register default operators
func (e *Expression) registerDefaultOperators() {
	// Arithmetic operators
	e.operators["+"] = Operator{
		Name:       "+",
		Precedence: 10,
		Handler:    add,
	}

	e.operators["-"] = Operator{
		Name:       "-",
		Precedence: 10,
		Handler:    subtract,
	}

	e.operators["*"] = Operator{
		Name:       "*",
		Precedence: 20,
		Handler:    multiply,
	}

	e.operators["/"] = Operator{
		Name:       "/",
		Precedence: 20,
		Handler:    divide,
	}

	e.operators["%"] = Operator{
		Name:       "%",
		Precedence: 20,
		Handler:    modulo,
	}

	// Comparison operators
	e.operators["=="] = Operator{
		Name:       "==",
		Precedence: 7,
		Handler:    equal,
	}

	e.operators["!="] = Operator{
		Name:       "!=",
		Precedence: 7,
		Handler:    notEqual,
	}

	e.operators[">"] = Operator{
		Name:       ">",
		Precedence: 8,
		Handler:    greater,
	}

	e.operators[">="] = Operator{
		Name:       ">=",
		Precedence: 8,
		Handler:    greaterEqual,
	}

	e.operators["<"] = Operator{
		Name:       "<",
		Precedence: 8,
		Handler:    less,
	}

	e.operators["<="] = Operator{
		Name:       "<=",
		Precedence: 8,
		Handler:    lessEqual,
	}

	// Logical operators
	e.operators["&&"] = Operator{
		Name:       "&&",
		Precedence: 5,
		Handler:    and,
	}

	e.operators["||"] = Operator{
		Name:       "||",
		Precedence: 4,
		Handler:    or,
	}

	e.operators["!"] = Operator{
		Name:       "!",
		Precedence: 3,
		Handler:    not,
	}
}

// readNumber reads a number token from the input
func readNumber(input string, start int, line int, col int) (*Token, int, error) {
	var value strings.Builder
	current := start
	hasDot := false

	for current < len(input) {
		char := input[current]
		if isDigit(char) {
			value.WriteByte(char)
		} else if char == '.' && !hasDot {
			value.WriteByte(char)
			hasDot = true
		} else {
			break
		}
		current++
	}

	return &Token{
		Type:  TokenNumber,
		Value: value.String(),
		Line:  line,
		Col:   col,
	}, current, nil
}

// readIdentifier reads an identifier token from the input
func readIdentifier(input string, start int, line int, col int) (*Token, int) {
	var value strings.Builder
	current := start

	for current < len(input) {
		char := input[current]
		if isLetter(char) || isDigit(char) || char == '_' {
			value.WriteByte(char)
			current++
		} else {
			break
		}
	}

	return &Token{
		Type:  TokenIdentifier,
		Value: value.String(),
		Line:  line,
		Col:   col,
	}, current
}

// readString reads a string token from the input
func readString(input string, start int, line int, col int) (*Token, int, error) {
	var value strings.Builder
	quote := input[start]
	current := start + 1

	for current < len(input) {
		char := input[current]
		if char == quote {
			current++
			return &Token{
				Type:  TokenString,
				Value: value.String(),
				Line:  line,
				Col:   col,
			}, current, nil
		}
		if char == '\\' && current+1 < len(input) {
			// Handle escape sequences
			current++
			char = input[current]
			switch char {
			case 'n':
				value.WriteByte('\n')
			case 't':
				value.WriteByte('\t')
			case 'r':
				value.WriteByte('\r')
			case '\\', '"', '\'':
				value.WriteByte(char)
			default:
				return nil, 0, fmt.Errorf("invalid escape sequence at line %d, col %d: \\%c", line, col+current-start, char)
			}
		} else {
			value.WriteByte(char)
		}
		current++
	}

	return nil, 0, fmt.Errorf("unterminated string at line %d, col %d", line, col)
}

// readOperator reads an operator token from the input
func readOperator(input string, start int, line int, col int) (*Token, int) {
	var value strings.Builder
	current := start

	// Check for compound operators first
	if current+1 < len(input) {
		compound := input[current : current+2]
		switch compound {
		case "==", "!=", ">=", "<=", "&&", "||":
			return &Token{
				Type:  TokenOperator,
				Value: compound,
				Line:  line,
				Col:   col,
			}, current + 2
		}
	}

	// Read operator(s)
	maxOperatorLen := 2 // Maximum operator length
	operatorCount := 0

	for current < len(input) && operatorCount < maxOperatorLen {
		char := input[current]
		if !isOperator(char) {
			break
		}

		value.WriteByte(char)
		current++
		operatorCount++
	}

	// No operator found
	if value.Len() == 0 {
		return nil, current
	}

	// Validate operator
	op := value.String()
	switch op {
	case "+", "-", "*", "/", "%", "=", "!", "<", ">", "&", "|":
		return &Token{
			Type:  TokenOperator,
			Value: op,
			Line:  line,
			Col:   col,
		}, current
	default:
		// Invalid operator combination
		return nil, start
	}
}

// tokenType returns the token type for the given character
func tokenType(c byte) TokenType {
	switch c {
	case '(':
		return TokenLParen
	case ')':
		return TokenRParen
	case '.':
		return TokenDot
	case ',':
		return TokenComma
	default:
		return TokenOperator
	}
}
