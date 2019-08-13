package router

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/sebest/xff"
	"github.com/sirupsen/logrus"
)

type chiWrapper struct {
	chi chi.Router
}

// Router wraps the chi router to make it slightly more accessible
type Router interface {
	// Use appends one middleware onto the Router stack.
	Use(fn Middleware)

	// With adds an inline middleware for an endpoint handler.
	With(fn Middleware) Router

	// Route mounts a sub-Router along a `pattern`` string.
	Route(pattern string, fn func(r Router))

	// Method adds a routes for a `pattern` that matches the `method` HTTP method.
	Method(method, pattern string, h APIHandler)

	// HTTP-method routing along `pattern`
	Delete(pattern string, h APIHandler)
	Get(pattern string, h APIHandler)
	Post(pattern string, h APIHandler)
	Put(pattern string, h APIHandler)

	ServeHTTP(http.ResponseWriter, *http.Request)
}

//  creates a router with sensible defaults (xff, request id, cors)
func New(log logrus.FieldLogger, options ...Option) Router {
	r := &chiWrapper{chi.NewRouter()}

	xffmw, _ := xff.Default()
	r.Use(xffmw.Handler)
	for _, opt := range options {
		opt(r)
	}

	return r
}

// Route allows creating a generic route
func (r *chiWrapper) Route(pattern string, fn func(Router)) {
	r.chi.Route(pattern, func(c chi.Router) {
		fn(&chiWrapper{c})
	})
}

// Method adds a routes for a `pattern` that matches the `method` HTTP method.
func (r *chiWrapper) Method(method, pattern string, h APIHandler) {
	r.chi.Method(method, pattern, HandlerFunc(h))
}

// Get adds a GET route
func (r *chiWrapper) Get(pattern string, fn APIHandler) {
	r.chi.Get(pattern, HandlerFunc(fn))
}

// Post adds a POST route
func (r *chiWrapper) Post(pattern string, fn APIHandler) {
	r.chi.Post(pattern, HandlerFunc(fn))
}

// Put adds a PUT route
func (r *chiWrapper) Put(pattern string, fn APIHandler) {
	r.chi.Put(pattern, HandlerFunc(fn))
}

// Delete adds a DELETE route
func (r *chiWrapper) Delete(pattern string, fn APIHandler) {
	r.chi.Delete(pattern, HandlerFunc(fn))
}

// WithBypass adds an inline chi middleware for an endpoint handler
func (r *chiWrapper) With(fn Middleware) Router {
	r.chi = r.chi.With(fn)
	return r
}

// UseBypass appends one chi middleware onto the Router stack
func (r *chiWrapper) Use(fn Middleware) {
	r.chi.Use(fn)
}

// ServeHTTP will serve a request
func (r *chiWrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.chi.ServeHTTP(w, req)
}

// ======================================
// Custom error handler
// ======================================
type APIHandler func(w http.ResponseWriter, r *http.Request) *HTTPError

func HandlerFunc(fn APIHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			HandleError(err, w, r)
		}
	}
}
