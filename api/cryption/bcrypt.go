package cryption

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

func GetHash(pwd []byte) string {
	// GenerateFromPassword returns the bcrypt hash of the password at the given cost.
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
			log.Println(err)
	}
	return string(hash)
}
