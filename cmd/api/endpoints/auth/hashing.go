package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes the given password using bcrypt.
func HashPassword(password string) (string, error) {
	// Use bcrypt's GenerateFromPassword to hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CompareHashAndPassword checks if the provided password matches the hashed password.
func CompareHashAndPassword(hashedPassword, password string) error {
	// Use bcrypt's CompareHashAndPassword to compare the hashed password with the plain-text password
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
