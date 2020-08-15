// Package ghost provides the binding for Ghost APIs
package ghost

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gbrlsnchs/jwt/v3"
)

type Client struct {
	URL       string
	Key       string
	Version   string
	GhostPath string
	UserAgent string
	client    *http.Client
}

// defaultHTTPTimeout is the default http.Client timeout.
const defaultHTTPTimeout = 10 * time.Second

// NewClient creates a new API client.
func NewClient(url, key string) *Client {
	httpClient := &http.Client{Timeout: defaultHTTPTimeout}
	return &Client{
		URL:       url,
		Key:       key,
		Version:   "v2",
		GhostPath: "ghost",
		client:    httpClient,
	}
}

func (c *Client) generateJWT() (string, error) {
	keyParts := strings.Split(c.Key, ":")
	if len(keyParts) != 2 {
		return "", fmt.Errorf("Invalid Client.Key format")
	}
	id := keyParts[0]
	rawSecret := []byte(keyParts[1])
	secret := make([]byte, hex.DecodedLen(len(rawSecret)))
	_, err := hex.Decode(secret, rawSecret)
	if err != nil {
		return "", err
	}

	now := time.Now()
	hs256 := jwt.NewHS256(secret)
	p := jwt.Payload{
		Audience:       jwt.Audience{"/" + c.Version + "/admin/"},
		ExpirationTime: jwt.NumericDate(now.Add(5 * time.Minute)),
		IssuedAt:       jwt.NumericDate(now),
	}
	token, err := jwt.Sign(p, hs256, jwt.KeyID(id))
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func (c *Client) Request(method, path string, data interface{}) (*http.Response, error) {
	b := &bytes.Buffer{}
	json.NewEncoder(b).Encode(data)

	url := fmt.Sprintf("%s%s", c.URL, path)
	r, err := http.NewRequest(method, url, b)
	if err != nil {
		return nil, fmt.Errorf("buildRequest: %v", err)
	}

	ua := c.UserAgent
	if ua == "" {
		ua = "go-ghost v1"
	}
	r.Header.Add("User-Agent", ua)
	r.Header.Add("Content-Type", "application/json")
	if c.Key != "" {
		token, err := c.generateJWT()
		if err != nil {
			return nil, err
		}
		r.Header.Add("Authorization", "Ghost "+token)
	}

	if err != nil {
		return nil, err
	}

	return c.client.Do(r)
}

func (c *Client) EndpointForID(api, resource, id string) string {
	return fmt.Sprintf("/%s/api/%s/%s/%s/%s/", c.GhostPath, c.Version, api, resource, id)
}

func (c *Client) EndpointForSlug(api, resource, slug string) string {
	return fmt.Sprintf("/%s/api/%s/%s/%s/slug/%s/", c.GhostPath, c.Version, api, resource, slug)
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}
