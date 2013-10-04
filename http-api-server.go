// Package api provides a framework for creating HTTP servers in Go
// (http://golang.org/) to handle API requests capable of replying in xml,
// json, or any other valid content type.

package api

import (
	"crypto/hmac"
	"crypto/sha1"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type Server struct {
	mux    *http.ServeMux
	s      *http.Server
	Logger *log.Logger
}

var (
	Srv                      *Server
	DefaultServerReadTimeout = 30 // in seconds
)

// DigestMatches is an optional hmac check that can be applied to any or
// all api queries. DigestMatches takes three strings: a private key,
// a query term, and a sha1 digest of the query term string using the
// shared private key known only by authorized api clients and this server
// (see http://en.wikipedia.org/wiki/Hmac for more details on how it
// works). DigestMatches returns a boolean if the hmac digest is correct
// or not.
func DigestMatches(privateKey string, queryTerm string, queryTermDigest string) bool {
	h := hmac.New(sha1.New, []byte(privateKey))
	h.Write([]byte(queryTerm))
	hashed := fmt.Sprintf("%x", h.Sum(nil))
	return (hashed == queryTermDigest)
}

// Respond accepts an HTTP media type, charset, and a response function
// which returns a string. Respond wraps the server reply in the correct
// Content-type, charset, and Content-length, returning an http.HandlerFunc
// invoked by the HTTP multiplexer in reponse to the particular url pattern
// associated with this response function.
func Respond(mediaType string, charset string, fn func(w http.ResponseWriter, r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=%s", mediaType, charset))
		data := fn(w, r)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		fmt.Fprintf(w, data)
	}
}

// NewServer takes a host or ip address string, a port number, read timeout
// (in secords), along with a map defining url string patterns, and their
// corresponding response functions. NewServer sets each map entry into
// the HTTP multiplexer, then starts the HTTP server on the given host/ip
// address and port. The api.Server struct also provides a Logger for each
// response function to use, to log warnings, errors, and other information.
func NewServer(host string, port int, timeout int, handlers map[string]func(http.ResponseWriter, *http.Request)) {

	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.Handle(pattern, http.HandlerFunc(handler))
	}

	s := &http.Server{
		Addr:        fmt.Sprintf("%s:%d", host, port),
		Handler:     mux,
		ReadTimeout: time.Duration(timeout) * time.Second, // to prevent abuse of "keep-alive" requests by clients
	}

	Srv = &Server{
		mux:    mux,
		s:      s,
		Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
	}
	Srv.s.ListenAndServe()
}

// NewLocalServer takes a port number, read timeout (in secords), along with
// a map defining url string patterns, and their corresponding response
// functions. This function is a simpler alternative to NewServer, used
// when the api server will be running on the localhost.
func NewLocalServer(port int, timeout int, handlers map[string]func(http.ResponseWriter, *http.Request)) {
	NewServer("", port, timeout, handlers)
}

