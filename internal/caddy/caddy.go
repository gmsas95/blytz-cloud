package caddy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	adminURL string
}

func NewClient(adminURL string) *Client {
	return &Client{
		adminURL: adminURL,
	}
}

type Route struct {
	Handle []Handler `json:"handle"`
	Match  []Match   `json:"match,omitempty"`
}

type Handler struct {
	Handler   string              `json:"handler"`
	Upstreams []Upstream          `json:"upstreams,omitempty"`
	Headers   map[string][]string `json:"headers,omitempty"`
}

type Upstream struct {
	Dial string `json:"dial"`
}

type Match struct {
	Host []string `json:"host"`
}

func (c *Client) AddSubdomain(subdomain, target string) error {
	route := Route{
		Handle: []Handler{
			{
				Handler: "reverse_proxy",
				Upstreams: []Upstream{
					{Dial: target},
				},
			},
		},
		Match: []Match{
			{
				Host: []string{subdomain},
			},
		},
	}

	data, err := json.Marshal(route)
	if err != nil {
		return fmt.Errorf("marshal route: %w", err)
	}

	url := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes", c.adminURL)
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("add route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("add route failed: %s", resp.Status)
	}

	return nil
}

func (c *Client) RemoveSubdomain(subdomain string) error {
	url := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes", c.adminURL)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("get routes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get routes failed: %s", resp.Status)
	}

	return nil
}

func (c *Client) Reload() error {
	url := fmt.Sprintf("%s/load", c.adminURL)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("create reload request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("reload caddy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reload failed: %s", resp.Status)
	}

	return nil
}
