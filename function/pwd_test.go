package function

import (
	"fmt"
	"testing"
)

var saltedpwd1 string
var saltedpwd2 string

func TestGetPwd(t *testing.T) {
	str1 := "123456abc"
	str2 := "123abc@#$"

	fmt.Printf("\nPassword: %v :%v\n", str1, GetPwd(str1))
	fmt.Printf("Password: %v :%v\n", str2, GetPwd(str2))
}

func TestHashAndSalt(t *testing.T) {
	pwd1 := GetPwd("123456abc")
	pwd2 := GetPwd("123abc@#$")

	saltedpwd1 = HashAndSalt(pwd1)
	saltedpwd2 = HashAndSalt(pwd2)
	fmt.Printf("\nPassword: %v :%v\n", pwd1, saltedpwd1)
	fmt.Printf("Password: %v :%v\n", pwd2, saltedpwd2)
}

func TestComparePasswords(t *testing.T) {
	str1 := "123456abc"
	str2 := "123abc@#$"
	str3 := "123456ab"

	fmt.Printf("\nPassword: %v :%v\n", str1, ComparePasswords(saltedpwd1, GetPwd(str1)))
	fmt.Printf("Password: %v :%v\n", str2, ComparePasswords(saltedpwd2, GetPwd(str2)))
	fmt.Printf("Password: %v :%v\n", str3, ComparePasswords(saltedpwd1, GetPwd(str3)))
}
