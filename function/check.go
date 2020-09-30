package function

import "regexp"

func CheckEmail(email *string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return re.MatchString(*email)
}

// This function can check the complexity of the password and return the status code
func CheckPassword(password string, level int)(statusCode int) {
	if len(password) < 8 {
		// -1 means the length of the password is less than 8
		return -1
	}

	// password = "12345678" - 1 week
	// "123456abc" - 2 strong
	count := 0
	patternList := []string{`[0-9]+`, `[a-z]+`, `[A-Z]+`, `[~!@#$%^&*?_-]+`}
	for _, pattern := range patternList {
		match, _ := regexp.MatchString(pattern, password)
		if match {
			count++
		}
	}

	if count < level {
		// 0 means the password is too week
		return 0
	} else {
		// 1 means the password is strong enough
		return 1
	}
}