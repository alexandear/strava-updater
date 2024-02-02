package strava

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/time/rate"
)

// The default rate limit is 100 requests every 15 minutes.
// See https://developers.strava.com/docs/rate-limits/.
//
// rateLimiter is global to override it for tests.
var rateLimiter = rate.NewLimiter(rate.Every(9*time.Second), 1)

type Client struct {
	baseURL    url.URL
	httpClient http.Client

	ifDebug bool
}

func New(accessToken string, client *http.Client, ifDebug bool) (Client, error) {
	if accessToken == "" {
		return Client{}, fmt.Errorf("accessToken is required")
	}

	if client == nil {
		client = http.DefaultClient
	}

	baseURL, _ := url.Parse("https://www.strava.com/api/v3")

	client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if rateLimiter != nil {
			err := rateLimiter.Wait(req.Context())
			if err != nil {
				return nil, fmt.Errorf("rate limiter: %w", err)
			}
		}

		req = req.Clone(req.Context())

		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if ifDebug {
			bodyBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return resp, err
			}

			log.Printf("Request: `%s %s`, Response: Headers: `%#v`, Body: `%s`\n", req.Method, req.URL, resp.Header, bodyBytes)

			resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("request failed with status: %s", resp.Status)
		}

		return resp, err
	})

	return Client{
		httpClient: *client,
		baseURL:    *baseURL,
		ifDebug:    ifDebug,
	}, nil
}

// roundTripperFunc type is an adapter to allow the use of ordinary functions as http.RoundTripper.
type roundTripperFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the RoundTripper interface.
func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type Athlete struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	City      string `json:"city"`
}

// Athlete returns the current authenticated athlete.
// https://developers.strava.com/docs/reference/#api-Athletes-getLoggedInAthlete
func (c *Client) Athlete(ctx context.Context) (Athlete, error) {
	u := c.baseURL.JoinPath("athlete")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return Athlete{}, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Athlete{}, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var athlete Athlete
	if err := json.NewDecoder(resp.Body).Decode(&athlete); err != nil {
		return Athlete{}, fmt.Errorf("decode response: %w", err)
	}

	return athlete, nil
}

type SummaryActivity struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
}

// Activities returns the activities of the current authenticated athlete.
// hasNext is true if there are more activities to fetch.
// https://developers.strava.com/docs/reference/#api-Activities-getLoggedInAthleteActivities
func (c *Client) Activities(ctx context.Context, from, to time.Time, page int) (activities []SummaryActivity, hasNext bool, _ error) {
	if to.Before(from) || to.Equal(from) {
		return nil, false, errors.New("to date must be after from date")
	}

	u := c.baseURL.JoinPath("athlete", "activities")
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	q.Set("before", strconv.FormatInt(to.Unix(), 10))
	q.Set("after", strconv.FormatInt(from.Unix(), 10))
	const perPage = 100
	q.Set("per_page", strconv.Itoa(perPage))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
	if err != nil {
		return nil, false, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return nil, false, fmt.Errorf("decode response: %w", err)
	}

	return activities, len(activities) == perPage, nil
}
