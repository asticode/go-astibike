package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"

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

// handleIndex handles the /index route
func (s Server) handleIndex(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Get forecast
	var d []astidarksky.DataPoint
	var err error
	if d, err = s.hourlyForecast(); err != nil {
		astilog.Error(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Execute template
	if err = s.templates.ExecuteTemplate(rw, "/index.html", d); err != nil {
		astilog.Error(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
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
