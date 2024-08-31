package consts

// Character sets
const (
	Number        = "0123456789"                    // Numbers
	Lowercase     = "abcdefghijklmnopqrstuvwxyz"    // Lowercase letters
	Uppercase     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"    // Uppercase letters
	Symbol        = "!#$%&()*+,-./:;<=>?@[]^_`{|}~" // Symbols
	NumLower      = Number + Lowercase              // Numbers + Lowercase letters
	NumUpper      = Number + Uppercase              // Numbers + Uppercase letters
	LowerUpper    = Lowercase + Uppercase           // Lowercase + Uppercase letters
	NumLowerUpper = Number + Lowercase + Uppercase  // Numbers + Lowercase + Uppercase letters
	All           = NumLowerUpper + Symbol          // Combination of all
)

// PrimaryKey is the character set used for primary keys
const PrimaryKey = NumLowerUpper
