go-api
======

About
-----

This package provides a framework for creating HTTP servers in [Go](http://golang.org/) under common [network transport layers](http://golang.org/pkg/net/#Dial) (tcp, udp, ip, unix) capable of handling API requests and replying in xml, json, or any other valid content type. 

Usage
-----

Install the package in your environment:

```
go get github.com/dpapathanasiou/go-api
```

To use it within your own code, import "github.com/dpapathanasiou/go-api" and define a map of type:

```go
{ string: func(http.ResponseWriter, *http.Request) }
```

Where the string represents the url pattern string to match, and the corresponding function calls the api.Respond function (which defines both the media type and the charset), which in turn calls the function which actually processes the client request, returning a string in the expected format.

Here's an example of how to send a "Hello World" message in JSON format in reply to a client request at /hello/ 

```go

package main

import (
    "net/http"
    "encoding/json"
    "github.com/dpapathanasiou/go-api"
)

type Message struct {
    Text string
}

func helloWorldJSON (w http.ResponseWriter, r *http.Request) string {
    // while we're not using r in this example, the http.Request object
    // has several attributes which help inform what the exact reply will be
    // (see http://golang.org/pkg/net/http/#Request for the full list of attributes,
    // as well as the weather-api.go example file in this package to get an idea of what's possible)
    m := Message{"Hello World"}
    b, err := json.Marshal(m)
    if err != nil { 
        panic(err) // no, not really
    }

    return string(b)
}
```

With the handler function defined, the main() function needs to associate it to the right pattern and response type (this example defines just one pattern, "/hello/", which returns a greeting in JSON as utf-8, but several other patterns and response function combinations can be added to the multiplexer as needed):

```go
func main() {
	handlers := map[string]func(http.ResponseWriter, *http.Request){}
	handlers["/hello/"] = func(w http.ResponseWriter, r *http.Request) {
		api.Respond("application/json", "utf-8", helloWorldJSON)(w, r)
	}

	api.NewServer("192.168.1.1", api.DefaultServerTransport, 9001, api.DefaultServerReadTimeout, false, handlers)
}
```

When the server is running, responses for any defined pattern can be access by calling:

```
http://[domain/ip of server]:[port]/[pattern]
```

This particular server will reply to requests of the form:

```
http://192.168.1.1:9001/hello/
```

with this JSON:

```
{"Text":"Hello World"}
```

If the server will run on the localhost and there is no ambiguity about the IP address or the hostname, then the <tt>NewLocalServer()</tt> function can be invoked instead, without the need to specify the IP address/hostname string: 

```go
func main() {
	// [ handlers defined same as above ... ]

	api.NewLocalServer(api.DefaultServerTransport, 9001, api.DefaultServerReadTimeout, false, handlers)
}
```

Any undefined handlers (i.e., anything other than http://[domain/ip of server]:9001/hello/) get sent to the default handler, http.NotFound, and returns an HTTP 404.

The full listing for this example is in [examples/hello-world-json.go](https://github.com/dpapathanasiou/go-api/blob/master/examples/hello-world-json.go).

Other Usage Examples
--------------------

### [examples/get-request.go](https://github.com/dpapathanasiou/go-api/blob/master/examples/get-request.go)

This example shows how to handle client API requests where data is sent to the server via an [HTTP GET](http://www.w3.org/Protocols/rfc2616/rfc2616-sec9.html#sec9.3) request, i.e., as name/value pairs encoded into the query string.

### [examples/post-request.go](https://github.com/dpapathanasiou/go-api/blob/master/examples/post-request.go)

This example shows how to handle client API requests where data is sent to the server via an [HTTP POST](http://www.w3.org/Protocols/rfc2616/rfc2616-sec9.html#sec9.5) request.

We can simulate the client POST request using [curl](http://curl.haxx.se/) on a command line prompt.

Assuming <tt>post-request.go</tt> has been built and running on port 9001, we can test it like this:

```
$ curl -d "id=1&w=0&v=1199&a=381&wh=0&pcy=0&frq=599&va=458&rnc=0&sr=20" http://127.0.0.1:9001/logger
```

The server will respond with this JSON:

```
{"Status":"ok","Data":["v=1199","pcy=0","id=1","w=0","a=381","wh=0","frq=599","va=458","rnc=0","sr=20"]}
```

### [examples/weather-api.go](https://github.com/dpapathanasiou/go-api/blob/master/examples/weather-api.go)

This is a more complete example, which creates an API server that returns current weather conditions in xml format for valid NOAA locations.

This server follows the HTTP GET request model from the <tt>get-request.go</tt> example, above.

Build the example server like this:

```
go build weather-api.go
```

Run the server by invoking the executable:

```
./weather-api
```

You can access the page from http://localhost:9001/weather?q=[station id] using curl, wget, or through a web browser (see http://w1.weather.gov/xml/current_obs/ for a full list of valid station id values).

If the station id is valid, you will see the NOAA current conditions report in xml format, otherwise the API server will reply with an error message in xml.

Optionally, you can add an hmac digest for security:

```
http://[localhost/domain/ip of server]:9001/weather?q=[station id]&d=[hmac digest of "q" in sha1 with a shared private key]
```

The "d" parameter is a sha1 digest of the station id using "secret" as the shared private key in this example (in practice, the private key is known only by the authorized api client and the server -- see http://en.wikipedia.org/wiki/Hmac for more details on how it works).

While this server is trivial in that it is simple repeating the xml fed to it by the NOAA server, more complex replies are possible (e.g., fetch and return queries from a database, calculate analytics and return a summary, etc.).
