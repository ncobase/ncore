package types

import "time"

// Code ecode
type Code int

// Datetime Datetime type
type Datetime time.Time

// JSON JSON type
type JSON = map[string]interface{}

// Token Token Detail
type Token struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	AccessID     string `json:"access_id,omitempty"`
	RefreshID    string `json:"refresh_id,omitempty"`
	AtExpires    int64  `json:"at_expires,omitempty"`
	RtExpires    int64  `json:"rt_expires,omitempty"`
}
