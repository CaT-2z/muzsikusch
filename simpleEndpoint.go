package main

import "net/http"

func SimpleEndpoint(f func() error) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := f()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}
