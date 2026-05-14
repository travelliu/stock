// Package auth provides password hashing, API-token generation, session
// store factory, and the gin middleware that resolves the calling user.
package auth

import "golang.org/x/crypto/bcrypt"

const BcryptCost = 12

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
