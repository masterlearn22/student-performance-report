package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "1234567890" // Password yang akan di-hash

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		panic(err)
	}
	
	fmt.Println("Password:", password)
	fmt.Println("Hash:", string(hash))
}
