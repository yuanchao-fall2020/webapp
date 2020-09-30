package function

import (
	"fmt"
	"testing"
)

func TestCheckEmail(t *testing.T) {
	str1 := "ç$€§/az@gmail.com"
	str2 := "abcd@gmail_yahoo.com"
	str3 := "123456@gmail-yahoo.com"
	str4 := "abcafqcd@gmailyahoo"
	str5 := "aadf2@#$sdfbcd@gmail.yahoo"

	fmt.Printf("\nEmail: %v :%v\n", str1, CheckEmail(&str1))
	fmt.Printf("Email: %v :%v\n", str2, CheckEmail(&str2))
	fmt.Printf("Email: %v :%v\n", str3, CheckEmail(&str3))
	fmt.Printf("Email: %v :%v\n", str4, CheckEmail(&str4))
	fmt.Printf("Email: %v :%v\n", str5, CheckEmail(&str5))
}

func TestCheckPassword(t *testing.T) {
	str1 := "123456"
	str2 := "12345678"
	str3 := "ç$€§12345"
	str4 := "123456abc"
	str5 := "123abc@#$"

	fmt.Printf("\nPassword: %v(2) :%v\n", str1, CheckPassword(str1, 2))
	fmt.Printf("Password: %v(2) :%v\n", str2, CheckPassword(str2, 2))
	fmt.Printf("Password: %v(2) :%v\n", str3, CheckPassword(str3, 2))
	fmt.Printf("Password: %v(3) :%v\n", str4, CheckPassword(str4, 3))
	fmt.Printf("Password: %v(3) :%v\n", str5, CheckPassword(str5,3))
}