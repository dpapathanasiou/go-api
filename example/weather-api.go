// weather-api.go
//
// This is a simple example of how to use the api package.
// 
// The getWeather() function takes the NOAA station id for a given location (full list at http://w1.weather.gov/xml/current_obs/)
// and returns the current weather conditions as an xml-formatted string.
//
// The weatherHandler() function ... (content-type: text/xml) format to the client.
//
// This example server runs on port 9001, and so any request in the form:
//
// http://[localhost/domain/ip of server]:9001/weather/[station id]
//
// will work, as long as [station id] corresponds to a valid NOAA value.

package main

import (
	"api"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func getWeather(w http.ResponseWriter, r *http.Request) string {
	xml := "<error>Bad Request</error>" // default response string
	parts := strings.Split(r.URL.Path, "/")

	if len(parts) < 3 {
		// there is nothing following /weather/ in the request url, so return an error message in xml
		xml = "<error>Please specify a NOAA station id</error>"
	} else {
		// Use the string immediately following /weather/ in the request url as the location;
		// the location must be a valid NOAA station id, as defined here: http://w1.weather.gov/xml/current_obs/
		location := parts[2]
		res, err := http.Get(fmt.Sprintf("http://w1.weather.gov/xml/current_obs/%s.xml", location))
		if err != nil {
			api.Srv.Logger.Print(err)
		}

		if res.StatusCode == 200 {
			// There is a NOAA station id matching the location in the request url
			// and we were able to get a valid reply from the NOAA server in xml
			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				api.Srv.Logger.Print(err)
			}
			res.Body.Close()

			// The xml returned from NOAA has an xsl definition that will confuse some clients, so remove it
			// (probably should use Go's xml parsing package, but this is good enough for a quick-and-dirty example like this
			xml = strings.Replace(string(b), "<?xml-stylesheet href=\"latest_ob.xsl\" type=\"text/xsl\"?>", "", -1)

			// also show an example of what the log output is for this request
			api.Srv.Logger.Printf("INFO: found current weather for NOAA station id %s\n", location)
		} else {
			// While the http.Get() request succeeded, something else went wrong,
			// most likely a 404 status, which means the station id is invalid
			api.Srv.Logger.Printf("WARN: problem finding weather for NOAA station id %s (NOAA server reply: %d)\n", location, res.StatusCode)
			xml = fmt.Sprintf("<error status=\"%d\">Could not get weather for NOAA station id %s</error>", res.StatusCode, location)
		}
	}
	return xml
}

func main() {
	handlers := map[string]func(http.ResponseWriter, *http.Request){}
	handlers["/weather/"] = func(w http.ResponseWriter, r *http.Request) {
		api.Respond("text/xml", "utf-8", getWeather)(w, r)
	}

	api.NewServer(9001, api.DefaultServerReadTimeout, handlers)
}
