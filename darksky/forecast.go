package astidarksky

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/time"
)

// Forecast represents a forecast
// https://darksky.net/dev/docs/forecast
type Forecast struct {
	Hourly    DataBlock `json:"hourly"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timezone  string    `json:"timezone"`
}

// DataBlock represents a data block
type DataBlock struct {
	Data []DataPoint `json:"data"`
}

// DataPoint represents a data point
type DataPoint struct {
	ApparentTemperature      float64            `json:"apparentTemperature"` // in degrees celsius
	CloudCover               float64            `json:"cloudCover"`          // between 0 and 1 inclusive
	DewPoint                 float64            `json:"dewPoint"`            // in degrees celsius
	Humidity                 float64            `json:"humidity"`            // between 0 and 1 inclusive
	Icon                     string             `json:"icon"`                // Free text for now
	Ozone                    float64            `json:"ozone"`               // in dobson units
	PrecipitationIntensity   float64            `json:"precipIntensity"`     // in millimeters per hour
	PrecipitationProbability float64            `json:"precipProbability"`   // between 0 and 1 inclusive
	PrecipitationType        string             `json:"precipType"`          // rain, snow or sleet
	Pressure                 float64            `json:"pressure"`            // in hectopascals
	Summary                  string             `json:"summary"`
	Timestamp                astitime.Timestamp `json:"time"`
	Visibility               float64            `json:"visibility"`  // in kilometers
	WindBearing              float64            `json:"windBearing"` // true north at 0°
	WindSpeed                float64            `json:"windSpeed"`   // meters per second
}

// HourlyForecast returns the hourly forecast for a set of latitude and longitude
func (c *Client) HourlyForecast(latitude, longitude float64) (d []DataPoint, err error) {
	// Init query parameters
	var qp = url.Values{}
	qp.Set("exclude", "currently,minutely,daily,alerts,flags")
	qp.Set("extend", "hourly")
	qp.Set("units", "si")

	// Init request
	var req *http.Request
	var u = fmt.Sprintf("%s/forecast/%s/%f,%f?%s", baseAddr, c.apiKey, latitude, longitude, qp.Encode())
	if req, err = http.NewRequest(http.MethodGet, u, nil); err != nil {
		return
	}

	// Send request
	var resp *http.Response
	astilog.Debugf("Sending GET request to %s", u)
	if resp, err = c.c.Do(req); err != nil {
		return
	}
	defer resp.Body.Close()

	// Unmarshal
	var f Forecast
	if err = json.NewDecoder(resp.Body).Decode(&f); err != nil {
		return
	}
	return f.Hourly.Data, nil
}
