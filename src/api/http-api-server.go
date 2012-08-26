
package api

import (
    "net/http"
    "strings"
    "time"
    "log"
    "fmt"
    "os"
)

type HttpServer struct {
    Logger *log.Logger
    s *http.Server
}

const defaultHandlerTag = ""

var (
    Srv *HttpServer
    RequestHandlers = map[string]func(http.ResponseWriter, *http.Request){ defaultHandlerTag: defaultHandler }
    ResponseFormats = map[string]string{ // map of api return types -> { type: Content-Type string }
        "xml":  "text/xml", 
        "json": "application/json" }
    DefaultServerReadTimeout = 30 // in seconds
)

func NewServer(port int, timeout int) { 
    Srv = &HttpServer{ 
                Logger: log.New(os.Stdout, "", log.Ldate|log.Ltime),
                s: &http.Server {
                        Addr: fmt.Sprintf(":%d", port),
                        Handler: pickHandler(defaultHandler),
                        ReadTimeout: time.Duration(timeout)*time.Second, // to prevent abuse of "keep-alive" requests by clients
                },
        }
    Srv.s.ListenAndServe()
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
    http.NotFound(w, r)
}

// Define which request handler function to use for a given request url path tag as defined by parseHandlerName()
func RegisterHandler (handler string, fn func(http.ResponseWriter, *http.Request)) {
    RequestHandlers[handler] = fn
}

// Optionally set the default request handler to be something other than http.NotFound (404)
func SetDefaultHandler (fn func(http.ResponseWriter, *http.Request)) {
    RequestHandlers[defaultHandlerTag] = fn
}

// Get the string immediately after the domain from the request url, e.g. http://127.0.0.1/edit/blah -> 'edit'
func parseHandlerName (urlPath string) (handler string) {
    handler = defaultHandlerTag
    parts := strings.Split(urlPath, "/")
	if len(parts) >= 2 {
        handler = parts[1]
    }
    return
}

// Decide which handler function to use for this request, falling back on the defaultFn if none found
func pickHandler(defaultFn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        handler := parseHandlerName(r.URL.Path)
        fn, found := RequestHandlers[handler]
        if !found {
            Srv.Logger.Printf("WARN: no handler defined for '%s'", handler)
            defaultFn(w, r)
            return
        }
        fn(w, r)
    }
}
