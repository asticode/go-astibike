package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astiredis"
	"github.com/asticode/go-astitools/template"
	"github.com/julienschmidt/httprouter"
)

// Server represents a server
type Server struct {
	addr        string
	channelQuit chan bool
	redis       *astiredis.Client
	router      *httprouter.Router
	templates   *template.Template
}

// NewServer creates a new server
func NewServer(addr string, redis *astiredis.Client) *Server {
	return &Server{
		addr:        addr,
		channelQuit: make(chan bool),
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
	// TODO Fetch info from dark sky

	// Execute template
	if err := s.templates.ExecuteTemplate(rw, "/index.html", nil); err != nil {
		astilog.Error(err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
}
