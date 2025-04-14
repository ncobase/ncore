package expression

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// estimateSize estimates the memory size of a value
func estimateSize(v any) int64 {
	switch val := v.(type) {
	case string:
		return int64(len(val))
	case []byte:
		return int64(len(val))
	case []any:
		var size int64
		for _, item := range val {
			size += estimateSize(item)
		}
		return size
	case map[string]any:
		var size int64
		for k, v := range val {
			size += int64(len(k))
			size += estimateSize(v)
		}
		return size
	default:
		// Default size for basic types
		return 8
	}
}

// add returns the sum or concatenation of two operands
func add(left, right any) (any, error) {
	// Handle string concatenation
	if ls, ok := left.(string); ok {
		if rs, ok := right.(string); ok {
			return ls + rs, nil
		}
	}

	// Handle numeric addition
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln + rn, nil
	}

	return nil, fmt.Errorf("invalid operands for +: %T and %T", left, right)
}

// subtract returns the difference of two operands
func subtract(left, right any) (any, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln - rn, nil
	}
	return nil, fmt.Errorf("invalid operands for -: %T and %T", left, right)
}

// multiply returns the product of two operands
func multiply(left, right any) (any, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln * rn, nil
	}
	return nil, fmt.Errorf("invalid operands for *: %T and %T", left, right)
}

// divide returns the quotient of two operands
func divide(left, right any) (any, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if !lok || !rok {
		return nil, fmt.Errorf("invalid operands for /: %T and %T", left, right)
	}
	if rn == 0 {
		return nil, fmt.Errorf("division by zero")
	}
	return ln / rn, nil
}

// modulo returns the remainder of two operands
func modulo(left, right any) (any, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if !lok || !rok {
		return nil, fmt.Errorf("invalid operands for %%: %T and %T", left, right)
	}
	if rn == 0 {
		return nil, fmt.Errorf("modulo by zero")
	}
	return math.Mod(ln, rn), nil
}

// equal returns true if two operands are equal
func equal(left, right any) (any, error) {
	return reflect.DeepEqual(left, right), nil
}

// notEqual returns true if two operands are not equal
func notEqual(left, right any) (any, error) {
	return !reflect.DeepEqual(left, right), nil
}

// greater returns true if left is greater than right
func greater(left, right any) (any, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln > rn, nil
	}
	return nil, fmt.Errorf("invalid operands for >: %T and %T", left, right)
}

// greaterEqual returns true if left is greater than or equal to right
func greaterEqual(left, right any) (any, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln >= rn, nil
	}
	return nil, fmt.Errorf("invalid operands for >=: %T and %T", left, right)
}

// less returns true if left is less than right
func less(left, right any) (any, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln < rn, nil
	}
	return nil, fmt.Errorf("invalid operands for <: %T and %T", left, right)
}

// lessEqual returns true if left is less than or equal to right
func lessEqual(left, right any) (any, error) {
	ln, lok := toNumber(left)
	rn, rok := toNumber(right)
	if lok && rok {
		return ln <= rn, nil
	}
	return nil, fmt.Errorf("invalid operands for <=: %T and %T", left, right)
}

// and returns true if both operands are true
func and(left, right any) (any, error) {
	lb, lok := left.(bool)
	rb, rok := right.(bool)
	if lok && rok {
		return lb && rb, nil
	}
	return nil, fmt.Errorf("invalid operands for &&: %T and %T", left, right)
}

// or returns true if either operand is true
func or(left, right any) (any, error) {
	lb, lok := left.(bool)
	rb, rok := right.(bool)
	if lok && rok {
		return lb || rb, nil
	}
	return nil, fmt.Errorf("invalid operands for ||: %T and %T", left, right)
}

// not returns true if the operand is false
func not(_, right any) (any, error) {
	if b, ok := right.(bool); ok {
		return !b, nil
	}
	return nil, fmt.Errorf("invalid operand for !: %T", right)
}

// validateOneNumber validates that there is exactly one number
func validateOneNumber(args []any) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	if _, ok := toNumber(args[0]); !ok {
		return fmt.Errorf("expected number, got %T", args[0])
	}
	return nil
}

// validateOneString validates that there is exactly one string
func validateOneString(args []any) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	if _, ok := args[0].(string); !ok {
		return fmt.Errorf("expected string, got %T", args[0])
	}
	return nil
}

// validateOneArray validates that there is exactly one array
func validateOneArray(args []any) error {
	if len(args) != 1 {
		return fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	if reflect.TypeOf(args[0]).Kind() != reflect.Slice {
		return fmt.Errorf("expected array, got %T", args[0])
	}
	return nil
}

// validateThreeArgs validates that there are exactly three arguments
func validateThreeArgs(args []any) error {
	if len(args) != 3 {
		return fmt.Errorf("expected 3 arguments, got %d", len(args))
	}
	if _, ok := args[0].(bool); !ok {
		return fmt.Errorf("first argument must be boolean, got %T", args[0])
	}
	return nil
}

// isWhitespace returns true if the given character is whitespace
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// isDigit returns true if the given character is a digit
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// isLetter returns true if the given character is a letter
func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isOperator returns true if the given character is an operator
func isOperator(c byte) bool {
	return strings.ContainsRune("+-*/%=!<>|&", rune(c))
}

// isUnaryOperator checks if the operator is unary
func isUnaryOperator(operator string) bool {
	switch operator {
	case "!", "+", "-":
		return true
	default:
		return false
	}
}

// toNumber attempts to convert a value to a float64
func toNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case string:
		if f, err := strconv.ParseFloat(n, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// toString converts any value to its string representation
func toString(v any) (string, error) {
	switch val := v.(type) {
	case string:
		return val, nil
	case int, int32, int64, float32, float64:
		return fmt.Sprintf("%v", val), nil
	case bool:
		return strconv.FormatBool(val), nil
	case time.Time:
		return val.Format(time.RFC3339), nil
	default:
		return "", fmt.Errorf("cannot convert type %T to string", v)
	}
}

// toBoolean converts any value to its boolean representation
func toBoolean(v any) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case string:
		return strconv.ParseBool(val)
	case int:
		return val != 0, nil
	case float64:
		return val != 0, nil
	default:
		return false, fmt.Errorf("cannot convert type %T to boolean", v)
	}
}
