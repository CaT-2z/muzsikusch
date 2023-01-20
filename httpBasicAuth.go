package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"strings"
)

type PasswordValidator interface {
	IsValid(username, password, realm string) bool
}

type BasicAuthenticator struct {
	realm             string
	PasswordValidator PasswordValidator
}

func NewBasicAuthenticator(realm string, passwordValidator PasswordValidator) BasicAuthenticator {
	return BasicAuthenticator{
		realm:             realm,
		PasswordValidator: passwordValidator,
	}
}

func (b *BasicAuthenticator) AuthRequest(handler http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//Check for auth header
		header := r.Header.Get("Authorization")
		//If not present, send 401
		if header == "" {
			w.Header().Set("WWW-Authenticate", "Basic realm=\""+b.realm+"\"")
			w.WriteHeader(401)
			return
		}
		//Parse header
		username, password, ok := parseBasicAuthHeader(header)
		if !ok {
			w.Header().Set("WWW-Authenticate", "Basic realm=\""+b.realm+"\"")
			w.WriteHeader(401)
			return
		}

		//If present, check if credentials are correct
		valid := b.PasswordValidator.IsValid(username, password, b.realm)

		//If not, send 401
		//If yes, call handler
		if valid {
			handler.ServeHTTP(w, r)
		} else {
			w.Header().Set("WWW-Authenticate", "Basic realm=\""+b.realm+"\"")
			w.WriteHeader(401)
		}
	})
}

func parseBasicAuthHeader(header string) (username, password string, ok bool) {
	//Remove "Basic " prefix
	if !strings.HasPrefix(header, "Basic ") {
		return "", "", false
	}
	header = strings.TrimPrefix(header, "Basic ")

	data, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return "", "", false
	}

	parts := strings.Split(string(data), ":")
	if len(parts) != 2 {
		log.Println("Username or password contains ':'")
		return "", "", false
	}

	return parts[0], parts[1], true
}
