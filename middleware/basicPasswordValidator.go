package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
)

type BasicPasswordValidator struct {
	passwords map[string]string
}

func NewBasicPasswordValidator(passwordFile io.Reader) (*BasicPasswordValidator, error) {
	var passwords map[string]string
	err := json.NewDecoder(passwordFile).Decode(&passwords)

	if err != nil {
		return nil, err
	}

	return &BasicPasswordValidator{
		passwords: passwords,
	}, nil
}

func (v *BasicPasswordValidator) IsValid(username, password, realm string) bool {
	//Hash password
	hash := sha256.Sum256([]byte(password))
	//Compare with stored hash
	return v.passwords[username] == hex.EncodeToString(hash[:])
}
