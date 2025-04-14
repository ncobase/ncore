package validator

import (
	"net/mail"
	"regexp"
)

// IsPhoneValid verify phone number
func IsPhoneValid(phone string) bool {
	// reference: https://learnku.com/articles/31543
	phoneCN, _ := regexp.Compile(`^1(3\d|4[5-9]|5[0-35-9]|6[2567]|7[0-8]|8\d|9[0-35-9])\d{8}$`)
	if phoneCN.MatchString(phone) {
		return true
	}
	global, _ := regexp.Compile(`\+(9[976]\d|8[987530]\d|6[987]\d|5[90]\d|42\d|3[875]\d|
2[98654321]\d|9[8543210]|8[6421]|6[6543210]|5[87654321]|
4[987654310]|3[9643210]|2[70]|7|1)\d{1,14}$`)
	return global.MatchString(phone)
}

// IsEmailValid verify email
func IsEmailValid(address string) bool {
	_, err := mail.ParseAddress(address)
	return err == nil
}
