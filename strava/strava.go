package strava

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Client struct {
	client  *http.Client
	baseURL url.URL
	ifDebug bool
}

type Athlete struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	City      string `json:"city"`
}

func New(accessToken string, client *http.Client, ifDebug bool) (*Client, error) {
	if accessToken == "" {
		return nil, fmt.Errorf("accessToken is required")
	}

	if client == nil {
		client = http.DefaultClient
	}

	baseURL, err := url.Parse("https://www.strava.com/api/v3")
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	transport := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req.Header.Add("Authorization", "Bearer "+accessToken)
		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			return resp, err
		}

		if ifDebug {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return resp, err
			}

			log.Printf("Request: `%s %s`, Response body: `%s`\n", req.Method, req.URL, bodyBytes)

			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("request failed with status: %s", resp.Status)
		}

		return resp, err
	})

	client.Transport = transport
	return &Client{
		client:  client,
		baseURL: *baseURL,
		ifDebug: ifDebug,
	}, nil
}

// roundTripperFunc type is an adapter to allow the use of ordinary functions as http.RoundTripper.
type roundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the RoundTripper interface.
func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func (c *Client) Athlete(ctx context.Context) (*Athlete, error) {
	u := c.baseURL.JoinPath("athlete")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	athlete := &Athlete{}
	if err := json.NewDecoder(resp.Body).Decode(athlete); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return athlete, nil
}
