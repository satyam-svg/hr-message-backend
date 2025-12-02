package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a password using bcrypt with concurrency support
func HashPassword(password string) (string, error) {
	// Channel for result
	resultChan := make(chan struct {
		hash string
		err  error
	})

	// Hash password in goroutine for concurrency
	go func() {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		resultChan <- struct {
			hash string
			err  error
		}{string(hash), err}
	}()

	// Wait for result
	result := <-resultChan
	return result.hash, result.err
}

// ComparePassword compares a hashed password with plain text password
func ComparePassword(hashedPassword, password string) error {
	// Channel for result
	resultChan := make(chan error)

	// Compare password in goroutine for concurrency
	go func() {
		err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
		resultChan <- err
	}()

	// Wait for result
	return <-resultChan
}
