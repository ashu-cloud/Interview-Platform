package password

import "golang.org/x/crypto/bcrypt"

// Cost is the bcrypt work factor per the spec (cost 12).
const Cost = 12

// Hash generates a bcrypt hash of the plain-text password.
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), Cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Verify returns true if plain matches the bcrypt hash.
func Verify(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
