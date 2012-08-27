go-api
======

About
-----

This package provides a framework for creating HTTP servers in Go (http://golang.org/) to handle API requests capable of replying in xml, json, or any other valid content type. 

Usage
-----

Clone this repo, add its path to your $GOPATH environment variable, then install the api package in your environment:

```
cd ~/[where you cloned the repo]/go-api
export GOPATH=$GOPATH:`pwd`
cd $GOPATH/src/api
go install
```

If there are no errors, you should see a /pkg folder at the same level as the /src folder under /go-api, like this (this assumes a linux environment):

```
src/
    api/
        http-api-server.go

pkg/
    linux_amd64/
        api.a
```

Now you can create API servers by importing the api package into your own code.

The api package framework is designed to register handlers based on the client request url path. The /handlerTag/ string following the domain defines which handler function is responsible for replying:

```
http://[domain/ip of server]:[port]/[handlerTag]
```

Use the api.RegisterHandler() function to assign the specific /handlerTag/ to the desired response function. Each response function must accept the following parameters, and use the http.ResponseWriter to send its reply back the client, after setting the correct Content-Type header:

```
responseFunc (w http.ResponseWriter, r *http.Request)
```

Here's an example of how to send a "Hello World" message in JSON format in reply to a client request at /hello/

```go

package main

import (
    "api"
    "net/http"
    "fmt"
    "encoding/json"
)

type Message struct {
    Text string
}

func helloWorldJSON (w http.ResponseWriter, r *http.Request) {
    // while we're not using r in this example, the http.Request object
    // has several attributes which help inform what the exact reply will be
    // (see http://golang.org/pkg/net/http/#Request for the full list of attributes,
    // as well as the example-api-server.go file in this package to get an idea of what's possible)

    w.Header().Set("Content-Type", api.ResponseFormats["json"])
    m := Message{"Hello World"}
    b, err := json.Marshal(m)
    if err != nil { 
        panic(err)
    }
    w.Header().Set("Content-Length", fmt.Sprintf("%d", len(b)))
    fmt.Fprintf(w, string(b))
}
```

With the handler function defined, the main() function of the code which has imported the api package first needs to register the handler, then start the server:

```go
func main() {
    // this line says that all requests to /hello/ will go to the helloWorldJSON(w http.ResponseWriter, r *http.Request) function:
    api.RegisterHandler("hello", helloWorldJSON) 
    // add as many other handlers as you would like/need

    // now start the server on port 9001, using the default read timeout
    api.NewServer(9001, api.DefaultServerReadTimeout)
}
```

This server, when run, will reply to requests of the form:

```
http://[domain/ip of server]:9001/hello/
```

with this JSON:

```
{"Text":"Hello World"}
```

Any undefined handlers (i.e., anything other than http://[domain/ip of server]:9001/hello/) get sent to the api.defaultHandler() function, which simply calls http.NotFound() and returns an HTTP 404.

The default handler function can be redefined by calling api.SetDefaultHandler() and passing a function which accepts (http.ResponseWriter, *http.Request) parameters as input, and replies by writing to the http.ResponseWriter object.

Usage Example
-------------

This package contains a more complete example of how to build an API server which returns current weather conditions in xml format for valid NOAA locations.

Build the example server like this:

```
cd ~/[where you cloned the repo]/go-api
go build example-api-server.go
```

Run the server by invoking the executable:

```
./example-api-server
```

You can access the page from http://localhost:9001/weather/[station id] using curl, wget, or through a web browser (see http://w1.weather.gov/xml/current_obs/ for a full list of valid station id values).

If the station id following /weather/ is valid, you will see the NOAA current conditions report in xml format, otherwise the API server will reply with an error message in xml.

While this server is trivial in that it is simple repeating the xml fed to it by the NOAA server, more complex replies are possible (e.g., fetch and return queries from a database, calculate analytics and return a summary, etc.).
