package common

import "regexp"

func IsValidEmail(email string) bool {
	pattern := "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
	matched, err := regexp.Match(pattern, []byte(email))
	if err != nil {
		return false
	}
	return matched
}
