package main

import (
	"flag"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astiredis"
	"github.com/asticode/go-astitools/flag"
)

func main() {
	// Parse command
	var s = astiflag.Subcommand()
	flag.Parse()

	// Init configuration
	var c = NewConfiguration()

	// Set logger
	astilog.SetLogger(astilog.New(c.Logger))

	// Init redis
	var r = astiredis.New(c.Redis)

	// Init server
	var srv = NewServer(c.ServerAddr, r)
	defer srv.Close()
	if err := srv.Init(c); err != nil {
		astilog.Fatal(err)
	}

	// Switch on subcommand
	switch s {
	default:
		// Listen and serve
		go srv.ListenAndServer()

		// Wait
		srv.Wait()
	}
}
