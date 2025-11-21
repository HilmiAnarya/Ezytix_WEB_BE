package hash

import "golang.org/x/crypto/bcrypt"

func HashPassword(pw string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pw), 12)
	return string(bytes), err
}

func CheckPassword(pw, hashed string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pw))
	return err == nil
}
