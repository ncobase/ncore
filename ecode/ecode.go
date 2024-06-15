package ecode

// Define all error codes
var (
	OK                      = 0    // Success
	AppKeyInvalid           = -1   // Application does not exist or has been banned
	AccessKeyErr            = -2   // Access Key error
	SignCheckErr            = -3   // API signature check error
	MethodNoPermission      = -4   // Caller has no permission for this method
	NoLogin                 = -101 // Account not logged in
	UserDisabled            = -102 // Account suspended
	LackOfScores            = -103 // Lack of scores
	LackOfCoins             = -104 // Lack of coins
	CaptchaErr              = -105 // Captcha error
	UserInactive            = -106 // Account inactive
	UserNoMember            = -107 // Account not a formal member or in adaptation period
	AppDenied               = -108 // Application does not exist or has been banned
	MobileNoVerfiy          = -110 // Mobile phone not bound
	CsrfNotMatchErr         = -111 // CSRF check failed
	ServiceUpdate           = -112 // System is upgrading
	UserIDCheckInvalid      = -113 // Account not yet verified
	UserIDCheckInvalidPhone = -114 // Please bind your mobile phone first
	UserIDCheckInvalidCard  = -115 // Please complete real-name authentication first

	NotModified           = -304 // Not modified
	TemporaryRedirect     = -307 // Temporary redirect
	RequestErr            = -400 // Request error
	Unauthorized          = -401 // Unauthorized
	AccessDenied          = -403 // Access denied
	NothingFound          = -404 // Nothing found
	MethodNotAllowed      = -405 // Method not allowed
	Conflict              = -409 // Conflict
	ServerErr             = -500 // Server error
	ServiceUnavailable    = -503 // Service unavailable
	Deadline              = -504 // Service call timeout
	LimitExceed           = -509 // Exceed limit
	FileNotExists         = -616 // File does not exist
	FileTooLarge          = -617 // File too large
	FailedTooManyTimes    = -625 // Too many login failures
	UserNotExist          = -626 // User does not exist
	PasswordTooLeak       = -628 // Password too weak
	UsernameOrPasswordErr = -629 // Incorrect username or password
	TargetNumberLimit     = -632 // Operation target number limit
	TargetBlocked         = -643 // Blocked
	UserLevelLow          = -650 // User level too low
	UserDuplicate         = -652 // Duplicate user
	AccessTokenExpires    = -658 // Token expired
	PasswordHashExpires   = -662 // Password timestamp expired
	AreaLimit             = -688 // Geographic area limit
	CopyrightLimit        = -689 // Copyright limit
	FailToAddMoral        = -701 // Failed to deduct moral

	Degrade     = -1200 // Filtered request due to degradation
	RPCNoClient = -1201 // No available client for rpc service
	RPCNoAuth   = -1202 // Client for rpc service not authorized
)

// ecodeText defines the mapping of error codes to their corresponding texts
var ecodeText = map[int]string{
	OK:                      "Success",
	AppKeyInvalid:           "Application does not exist or has been banned",
	AccessKeyErr:            "Access Key error",
	SignCheckErr:            "API signature check error",
	MethodNoPermission:      "Caller has no permission for this method",
	NoLogin:                 "Account not logged in",
	UserDisabled:            "Account suspended",
	LackOfScores:            "Lack of scores",
	LackOfCoins:             "Lack of coins",
	CaptchaErr:              "Captcha error",
	UserInactive:            "Account inactive",
	UserNoMember:            "Account not a formal member or in adaptation period",
	AppDenied:               "Application does not exist or has been banned",
	MobileNoVerfiy:          "Mobile phone not bound",
	CsrfNotMatchErr:         "CSRF check failed",
	ServiceUpdate:           "System is upgrading",
	UserIDCheckInvalid:      "Account not yet verified",
	UserIDCheckInvalidPhone: "Please bind your mobile phone first",
	UserIDCheckInvalidCard:  "Please complete real-name authentication first",

	NotModified:           "Not modified",
	TemporaryRedirect:     "Temporary redirect",
	RequestErr:            "Request error",
	Unauthorized:          "Unauthorized",
	AccessDenied:          "Access denied",
	NothingFound:          "Nothing found",
	MethodNotAllowed:      "Method not allowed",
	Conflict:              "Conflict",
	ServerErr:             "Server error",
	ServiceUnavailable:    "Service unavailable",
	Deadline:              "Service call timeout",
	LimitExceed:           "Exceed limit",
	FileNotExists:         "File does not exist",
	FileTooLarge:          "File too large",
	FailedTooManyTimes:    "Too many login failures",
	UserNotExist:          "User does not exist",
	PasswordTooLeak:       "Password too weak",
	UsernameOrPasswordErr: "Incorrect username or password",
	TargetNumberLimit:     "Operation target number limit",
	TargetBlocked:         "Blocked",
	UserLevelLow:          "User level too low",
	UserDuplicate:         "Duplicate user",
	AccessTokenExpires:    "Token expired",
	PasswordHashExpires:   "Password timestamp expired",
	AreaLimit:             "Geographic area limit",
	CopyrightLimit:        "Copyright limit",
	FailToAddMoral:        "Failed to deduct moral",
	Degrade:               "Filtered request due to degradation",
	RPCNoClient:           "No available client for rpc service",
	RPCNoAuth:             "Client for rpc service not authorized",
}

// Text returns the text corresponding to the error code. If the error code does not exist, it returns a default message.
func Text(code int) string {
	if code > 0 {
		code = -code
	}
	if text, ok := ecodeText[code]; ok {
		return text
	}
	return "Unknown error"
}
