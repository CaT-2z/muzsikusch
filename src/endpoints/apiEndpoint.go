package endpoints

import (
	"net/http"
	"strings"
)

type Endpoint struct {
	Get     func(w http.ResponseWriter, r *http.Request)
	Post    func(w http.ResponseWriter, r *http.Request)
	Put     func(w http.ResponseWriter, r *http.Request)
	Delete  func(w http.ResponseWriter, r *http.Request)
	Patch   func(w http.ResponseWriter, r *http.Request)
	Options func(w http.ResponseWriter, r *http.Request)
}

func EmptyEndpoint() *Endpoint {
	return &Endpoint{}
}

func GetEndpoint(f http.HandlerFunc) *Endpoint {
	return &Endpoint{Get: f}
}
func PostEndpoint(f http.HandlerFunc) *Endpoint {
	return &Endpoint{Post: f}
}
func PutEndpoint(f http.HandlerFunc) *Endpoint {
	return &Endpoint{Put: f}
}
func DeleteEndpoint(f http.HandlerFunc) *Endpoint {
	return &Endpoint{Delete: f}
}
func PatchEndpoint(f http.HandlerFunc) *Endpoint {
	return &Endpoint{Patch: f}
}
func OptionsEndpoint(f http.HandlerFunc) *Endpoint {
	return &Endpoint{Options: f}
}

func (e *Endpoint) WithGet(f func(w http.ResponseWriter, r *http.Request)) *Endpoint {
	e.Get = f
	return e
}

func (e *Endpoint) WithPost(f func(w http.ResponseWriter, r *http.Request)) *Endpoint {
	e.Post = f
	return e
}

func (e *Endpoint) WithDelete(f func(w http.ResponseWriter, r *http.Request)) *Endpoint {
	e.Delete = f
	return e
}

func (e *Endpoint) WithPatch(f func(w http.ResponseWriter, r *http.Request)) *Endpoint {
	e.Patch = f
	return e
}

func (e *Endpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET" && e.Get != nil:
		e.Get(w, r)
	case r.Method == "POST" && e.Post != nil:
		e.Post(w, r)
	case r.Method == "DELETE" && e.Delete != nil:
		e.Delete(w, r)
	case r.Method == "PATCH" && e.Patch != nil:
		e.Patch(w, r)
	case r.Method == "OPTIONS":
		if e.Options != nil {
			e.Options(w, r)
		} else {
			allows := []string{}
			if e.Get != nil {
				allows = append(allows, "GET")
			}
			if e.Post != nil {
				allows = append(allows, "POST")
			}
			if e.Put != nil {
				allows = append(allows, "PUT")
			}
			if e.Delete != nil {
				allows = append(allows, "DELETE")
			}
			if e.Patch != nil {
				allows = append(allows, "PATCH")
			}
			w.Header().Set("Allow", strings.Join(allows, " "))
			w.WriteHeader(http.StatusNoContent)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type FailableHandler func(http.ResponseWriter, *http.Request) error

type FailableEndpoint struct {
	Get     FailableHandler
	Post    FailableHandler
	Put     FailableHandler
	Delete  FailableHandler
	Patch   FailableHandler
	Options FailableHandler
}

func EmptyFEndpoint() *FailableEndpoint {
	return &FailableEndpoint{}
}

func GetFEndpoint(f FailableHandler) FailableEndpoint {
	return FailableEndpoint{Get: f}
}
func PostFEndpoint(f FailableHandler) FailableEndpoint {
	return FailableEndpoint{Post: f}
}
func PutFEndpoint(f FailableHandler) FailableEndpoint {
	return FailableEndpoint{Put: f}
}
func DeleteFEndpoint(f FailableHandler) FailableEndpoint {
	return FailableEndpoint{Delete: f}
}
func PatchFEndpoint(f FailableHandler) FailableEndpoint {
	return FailableEndpoint{Patch: f}
}
func OptionsFEndpoint(f FailableHandler) FailableEndpoint {
	return FailableEndpoint{Options: f}
}

func (e *FailableEndpoint) WithGet(f FailableHandler) *FailableEndpoint {
	e.Get = f
	return e
}

func (e *FailableEndpoint) WithPost(f FailableHandler) *FailableEndpoint {
	e.Post = f
	return e
}

func (e *FailableEndpoint) WithDelete(f FailableHandler) *FailableEndpoint {
	e.Delete = f
	return e
}

func (e *FailableEndpoint) WithPatch(f FailableHandler) *FailableEndpoint {
	e.Patch = f
	return e
}

func (e *FailableEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var err error
	switch {
	case r.Method == "GET" && e.Get != nil:
		err = e.Get(w, r)
	case r.Method == "POST" && e.Post != nil:
		err = e.Post(w, r)
	case r.Method == "DELETE" && e.Delete != nil:
		err = e.Delete(w, r)
	case r.Method == "PATCH" && e.Patch != nil:
		err = e.Patch(w, r)
	case r.Method == "OPTIONS":
		if e.Options != nil {
			err = e.Options(w, r)
		} else {
			allows := []string{}
			if e.Get != nil {
				allows = append(allows, "GET")
			}
			if e.Post != nil {
				allows = append(allows, "POST")
			}
			if e.Put != nil {
				allows = append(allows, "PUT")
			}
			if e.Delete != nil {
				allows = append(allows, "DELETE")
			}
			if e.Patch != nil {
				allows = append(allows, "PATCH")
			}
			w.Header().Set("Allow", strings.Join(allows, " "))
			w.WriteHeader(http.StatusNoContent)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
