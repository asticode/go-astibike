package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"

	"strconv"

	"fmt"

	"github.com/asticode/go-astibike/darksky"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astiredis"
	"github.com/asticode/go-astitools/template"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/redis.v5"
)

// Constants
const (
	defaultLatitude        = 48.75
	defaultLongitude       = 2.3
	redisKeyHourlyForecast = "hourly_forecast"
)

// Server represents a server
type Server struct {
	addr        string
	channelQuit chan bool
	darkSky     *astidarksky.Client
	redis       *astiredis.Client
	router      *httprouter.Router
	templates   *template.Template
}

// NewServer creates a new server
func NewServer(addr string, darkSky *astidarksky.Client, redis *astiredis.Client) *Server {
	return &Server{
		addr:        addr,
		channelQuit: make(chan bool),
		darkSky:     darkSky,
		redis:       redis,
	}
}

// Init initializes the server
func (s *Server) Init(c Configuration) (err error) {
	// Init templates
	if s.templates, err = astitemplate.ParseDirectory(c.PathTemplates, ".html"); err != nil {
		return
	}

	// Init router
	s.router = httprouter.New()
	s.router.ServeFiles("/static/*filepath", http.Dir(c.PathStatic))
	s.router.GET("/", s.handleIndex)
	return
}

// Close closes the server properly
func (s *Server) Close() {}

// HandleSignals handles signals
func (s *Server) HandleSignals() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGABRT, syscall.SIGKILL, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		for sig := range ch {
			astilog.Debugf("Received signal %s", sig)
			s.Stop()
		}
	}()
}

// Stop stops the server
func (s *Server) Stop() {
	if s.channelQuit != nil {
		close(s.channelQuit)
		s.channelQuit = nil
	}
}

// Wait is a blocking pattern
func (s Server) Wait() {
	for {
		select {
		case <-s.channelQuit:
			return
		}
	}
}

// ListenAndServer listens and serves an addr
func (s Server) ListenAndServer() error {
	astilog.Debugf("Listening and serving on %s", s.addr)
	return http.ListenAndServe(s.addr, s.router)
}

// hourlyForecast retrieves the hourly forecast
func (s Server) hourlyForecast() (d []astidarksky.DataPoint, err error) {
	// Check in redis first
	if err = s.redis.Get(redisKeyHourlyForecast, &d); (err != nil && err != redis.Nil) || err == nil {
		return
	}

	// Retrieve hourly forecast
	if d, err = s.darkSky.HourlyForecast(defaultLatitude, defaultLongitude); err != nil {
		return
	}

	// Store in redis
	if err = s.redis.Set(redisKeyHourlyForecast, d, time.Hour); err != nil {
		return
	}
	return
}

// Data is a data sent to the template
type Data struct {
	Days         map[string]DataDay
	OrderedDays  []string
	OrderedHours []string
}

// DataDay is a data representing a day
type DataDay struct {
	Hours map[string]DataHour
	Label string
}

// DataHour is a data representing an hour
type DataHour struct {
	Grade                    int
	PrecipitationProbability string
	Temperature              string
	WindRotate               int
	WindSpeed                string
}

// handleIndex handles the /index route
func (s Server) handleIndex(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Get hourly forecast
	var dps []astidarksky.DataPoint
	var err error
	if dps, err = s.hourlyForecast(); err != nil {
		astilog.Error(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Init data
	var d = Data{
		Days:         make(map[string]DataDay),
		OrderedHours: []string{"08:00", "09:00", "10:00", "11:00", "12:00", "13:00", "14:00", "15:00", "16:00", "17:00", "18:00", "19:00", "20:00"},
	}
	var previousDay string
	for _, dp := range dps {
		// Init data day
		var dn = fmt.Sprintf("%s - %s", dp.Timestamp.Weekday(), dp.Timestamp.Format("02/01"))
		if _, ok := d.Days[dn]; !ok {
			d.Days[dn] = DataDay{
				Hours: make(map[string]DataHour),
				Label: dn,
			}
		}

		// Add data hour
		d.Days[dn].Hours[dp.Timestamp.Format("15:04")] = DataHour{
			Grade: grade(dp),
			PrecipitationProbability: strconv.Itoa(int(dp.PrecipitationProbability*100)) + "%",
			Temperature:              strconv.FormatFloat(dp.ApparentTemperature, 'f', 1, 64) + "°",
			WindRotate:               int(dp.WindBearing),
			WindSpeed:                strconv.FormatFloat(dp.WindSpeed, 'f', 0, 64) + "m/s",
		}

		// Add ordered days
		if previousDay == "" || previousDay != dn {
			previousDay = dn
			d.OrderedDays = append(d.OrderedDays, dn)
		}
	}

	// Execute template
	if err = s.templates.ExecuteTemplate(rw, "/index.html", d); err != nil {
		astilog.Error(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// grade grades a data point
// grades are between 0 and 4 included
func grade(d astidarksky.DataPoint) int {
	if d.ApparentTemperature > 15 && d.ApparentTemperature < 30 && d.PrecipitationProbability == 0 && d.WindSpeed < 6 {
		return 4
	} else if d.ApparentTemperature > 10 && d.PrecipitationProbability == 0 && d.WindSpeed < 6 {
		return 3
	} else if d.ApparentTemperature > 10 && d.PrecipitationProbability < 0.1 && d.WindSpeed < 6 {
		return 2
	} else if d.ApparentTemperature > 10 && d.PrecipitationProbability < 0.2 {
		return 1
	}
	return 0
}
