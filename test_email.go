package main

import (
	"fmt"
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func main() {
	emails := []string{"uyfagkkk@", "@example.com", "test@example", "test@@example.com", "testexample.com", "test@example.com"}
	for _, e := range emails {
		fmt.Printf("%s: %v\n", e, isValidEmail(e))
	}
}
