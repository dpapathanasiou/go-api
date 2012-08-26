// example-api-server.go
//
// This is a simple example of how to use the api package.
// 
// The weatherHandler() function takes the NOAA station id for a given location (full list at http://w1.weather.gov/xml/current_obs/)
// and returns the current weather conditions in xml (content-type: text/xml) format to the client.
//
// This example server runs on port 9001, and so any request in the form:
//
// http://[localhost/domain/ip of server]:9001/weather/[station id]
//
// will work, as long as [station id] corresponds to a valid NOAA value.

package main

import (
    "fmt"
    "strings"
    "net/http"
    "io/ioutil"
    "api"
)

const (
    weatherHandlerTag = "weather"
    weatherHandlerLen = len(weatherHandlerTag) + 2 // account for the leading and trailing slash in the url
)

// Lookup the current weather for the location specified in the request url and return it as xml
func weatherHandler (w http.ResponseWriter, r *http.Request) {
    /* Define the API response format for the reply back to the requesting client (xml in this case).
       We could do a lot of interesting things here, but for the sake of the example, 
       it will just reiterate the xml from NOAA, unless there is an error of some kind.
    */
    w.Header().Set("Content-Type", api.ResponseFormats["xml"])
    
    xml := ""
    if len(r.URL.Path) <= weatherHandlerLen {
        // there is nothing following /weather in the request url, so return an error message in xml
        xml = "<error>Please specify a NOAA station id</error>"
    } else {
        /* Use the string immediately following /weather/ in the request url as the location;
           the location must be a valid NOAA station id, as defined here: http://w1.weather.gov/xml/current_obs/
        */
        location := r.URL.Path[weatherHandlerLen:]
        res, err := http.Get(fmt.Sprintf("http://w1.weather.gov/xml/current_obs/%s.xml", location))
        if err != nil {
            api.Srv.Logger.Print(err)
        }
        
        if res.StatusCode == 200 {
            /* There is a NOAA station id matching the location in the request url
               and we were able to get a valid reply from the NOAA server in xml
            */
            b, err := ioutil.ReadAll(res.Body)
            if err != nil {
                api.Srv.Logger.Print(err)
            }
            res.Body.Close()
            
            /* The xml returned from NOAA has an xsl definition that will confuse some clients, so remove it
               (probably should use Go's xml parsing package, but this is good enough for a quick-and-dirty example like this)
            */
            xml = strings.Replace(string(b), "<?xml-stylesheet href=\"latest_ob.xsl\" type=\"text/xsl\"?>", "", -1)
            
            // also show an example of what the log output is for this request
            api.Srv.Logger.Printf("INFO: found current weather for NOAA station id %s", location)
        } else {
            /* While the http.Get() request succeeded, something else went wrong,
               most likely a 404 status, which means the station id is invalid
            */
            api.Srv.Logger.Printf("WARN: problem finding weather for NOAA station id %s (NOAA server reply: %d)", location, res.StatusCode)
            xml = fmt.Sprintf("<error status=\"%d\">Could not get weather for NOAA station id %s</error>", res.StatusCode, location)
        }
    }

    // send the reply to the client
    w.Header().Set("Content-Length", fmt.Sprintf("%d", len(xml)))
    fmt.Fprintf(w, xml)
}

func main() {
    api.RegisterHandler(weatherHandlerTag, weatherHandler)
    api.NewServer(9001, api.DefaultServerReadTimeout)
}

