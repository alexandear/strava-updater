package strava

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	// disable rate limiter for tests
	rateLimiter = nil
	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	t.Run("returns error when access token is empty", func(t *testing.T) {
		_, err := New("", nil, false)

		if err == nil {
			t.Error("expected an error but didn't get one")
		}
	})

	t.Run("sets Authorization header correctly", func(t *testing.T) {
		var gotAuthHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuthHeader = r.Header.Get("Authorization")
		}))
		t.Cleanup(server.Close)

		client, err := New("access_token", nil, false)
		if err != nil {
			t.Fatalf("expected no error but got %q", err)
		}

		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, http.NoBody)
		resp, err := client.httpClient.Do(req)
		if err != nil {
			t.Fatalf("expected no error but got '%s'", err)
		}
		defer resp.Body.Close()

		wantAuthHeader := "Bearer access_token"
		if gotAuthHeader != wantAuthHeader {
			t.Errorf(`expected Authorization header to be %q but got %q`, wantAuthHeader, gotAuthHeader)
		}
	})
}

func TestAthlete(t *testing.T) {
	t.Run("returns athlete and no error on successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/athlete" {
				return
			}
			athlete := Athlete{
				ID:        123,
				FirstName: "John",
				LastName:  "Doe",
				City:      "Gotham",
			}
			_ = json.NewEncoder(w).Encode(athlete)
		}))
		t.Cleanup(server.Close)

		baseURL, _ := url.Parse(server.URL)
		client := &Client{
			baseURL:    *baseURL,
			httpClient: *http.DefaultClient,
		}

		athlete, err := client.Athlete(context.Background())
		if err != nil {
			t.Fatalf("expected no error but got %q", err)
		}

		wantAthlete := Athlete{
			ID:        123,
			FirstName: "John",
			LastName:  "Doe",
			City:      "Gotham",
		}
		if athlete != wantAthlete {
			t.Errorf("expected athlete to be %v but got %v", wantAthlete, athlete)
		}
	})
}

func TestClient_Activities(t *testing.T) {
	t.Run("returns error when to date is not after from date", func(t *testing.T) {
		client := &Client{}

		_, _, err := client.Activities(context.Background(), time.Now(), time.Now(), 1)

		if err == nil {
			t.Error("expected an error but didn't get one")
		}
	})

	t.Run("returns activities and no error on successful request", func(t *testing.T) {
		var gotURL string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotURL = r.URL.String()
			w.Write([]byte(`[{"id": 1, "name": "Morning Run", "start_date": "2022-01-01T00:00:00Z"}]`))
		}))
		t.Cleanup(server.Close)

		baseURL, _ := url.Parse(server.URL)
		client := &Client{
			baseURL:    *baseURL,
			httpClient: *http.DefaultClient,
		}

		from := time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 0, 1)
		activities, hasNext, err := client.Activities(context.Background(), from, to, 2)
		if err != nil {
			t.Fatalf("expected no error but got %q", err)
		}

		wantURL := "/athlete/activities?after=1640995200&before=1641081600&page=2&per_page=100"
		if gotURL != wantURL {
			t.Errorf(`expected URL to be %q but got %q`, wantURL, gotURL)
		}
		if hasNext {
			t.Error("expected hasNext to be false but got true")
		}
		if diff := cmp.Diff(activities, []SummaryActivity{{
			ID:        1,
			Name:      "Morning Run",
			StartDate: "2022-01-01T00:00:00Z",
		}}); diff != "" {
			t.Errorf("unexpected activity (-got +want):\n%s", diff)
		}
	})
}

func TestClient_UpdateActivity(t *testing.T) {
	t.Run("encodes request correctly", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPut {
				t.Errorf("expected method to be PUT but got %s", r.Method)
			}

			if r.URL.Path != "/activities/1" {
				t.Errorf("expected path to be /activities/1 but got %s", r.URL.Path)
			}

			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type to be application/json but got %s", contentType)
			}

			var activity struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
				t.Fatalf("expected no error but got %q", err)
			}

			wantName := "Morning Workout"
			if activity.Name != wantName {
				t.Errorf("expected name to be %q but got %q", wantName, activity.Name)
			}
		}))
		defer server.Close()

		baseURL, _ := url.Parse(server.URL)
		client := &Client{
			baseURL:    *baseURL,
			httpClient: *http.DefaultClient,
		}
		err := client.UpdateActivity(context.Background(), 1, "Morning Workout")
		if err != nil {
			t.Fatalf("expected no error but got %q", err)
		}
	})
}
