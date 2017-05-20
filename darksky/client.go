package astidarksky

import "net/http"

// Constants
const (
	baseAddr = "https://api.darksky.net"
)

// Client represents a client capable of communicating with Dark Sky API
type Client struct {
	apiKey string
	c      *http.Client
}

// New creates a new client based on a configuration
func New(c Configuration) *Client {
	return &Client{
		apiKey: c.APIKey,
		c:      &http.Client{},
	}
}
