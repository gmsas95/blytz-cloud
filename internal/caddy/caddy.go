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
	// First, get all routes
	url := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes", c.adminURL)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("get routes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("get routes failed: %s", resp.Status)
	}

	// Parse routes
	var routes []Route
	if err := json.NewDecoder(resp.Body).Decode(&routes); err != nil {
		return fmt.Errorf("decode routes: %w", err)
	}

	// Find route index by subdomain
	var routeIndex *int
	for i, route := range routes {
		for _, match := range route.Match {
			for _, host := range match.Host {
				if host == subdomain {
					idx := i
					routeIndex = &idx
					break
				}
			}
			if routeIndex != nil {
				break
			}
		}
		if routeIndex != nil {
			break
		}
	}

	if routeIndex == nil {
		return fmt.Errorf("route not found for subdomain: %s", subdomain)
	}

	// Delete the specific route using index
	deleteURL := fmt.Sprintf("%s/config/apps/http/servers/srv0/routes/%d", c.adminURL, *routeIndex)

	req, err := http.NewRequest(http.MethodDelete, deleteURL, nil)
	if err != nil {
		return fmt.Errorf("create delete request: %w", err)
	}

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete route: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete route failed: %s", resp.Status)
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
