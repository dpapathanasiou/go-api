package api

import (
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

func Respond(mediaType string, charset string, fn func(w http.ResponseWriter, r *http.Request) string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=%s", mediaType, charset))
		data := fn(w, r)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
		fmt.Fprintf(w, data)
	}
}

func NewServer(port int, timeout int, handlers map[string]func(http.ResponseWriter, *http.Request)) {

	mux := http.NewServeMux()
	for pattern, handler := range handlers {
		mux.Handle(pattern, http.HandlerFunc(handler))
	}

	s := &http.Server{
		Addr:        fmt.Sprintf(":%d", port),
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
