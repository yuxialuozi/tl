package llm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientComplete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q", got)
		}

		var request chatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.Model != "test-model" {
			t.Errorf("model = %q", request.Model)
		}
		if len(request.Messages) != 2 || request.Messages[0].Role != "system" || request.Messages[1].Role != "user" {
			t.Errorf("messages = %+v", request.Messages)
		}
		if request.Messages[1].Content != "hello" {
			t.Errorf("user content = %q", request.Messages[1].Content)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":" world "}}]}`))
	}))
	defer server.Close()

	client := New(server.URL+"/v1/", "test-key", "test-model", time.Second)
	got, err := client.Complete(context.Background(), "system", "hello")
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	if got != "world" {
		t.Fatalf("Complete() = %q, want world", got)
	}
}

func TestClientAPIErrorRedactsKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"bad key secret-key"}}`))
	}))
	defer server.Close()

	client := New(server.URL, "secret-key", "test-model", time.Second)
	_, err := client.Complete(context.Background(), "system", "hello")
	if err == nil {
		t.Fatal("Complete() error = nil")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("StatusCode = %d", apiErr.StatusCode)
	}
	if strings.Contains(err.Error(), "secret-key") {
		t.Fatalf("error leaked API key: %v", err)
	}
	if !strings.Contains(err.Error(), "<redacted>") {
		t.Fatalf("error = %v, want redacted marker", err)
	}
}

func TestClientRejectsInvalidResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed JSON", body: `{`},
		{name: "no choices", body: `{"choices":[]}`},
		{name: "empty content", body: `{"choices":[{"message":{"content":"  "}}]}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()

			client := New(server.URL, "key", "model", time.Second)
			if _, err := client.Complete(context.Background(), "system", "hello"); err == nil {
				t.Fatal("Complete() error = nil")
			}
		})
	}
}

func TestClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(150 * time.Millisecond)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"late"}}]}`))
	}))
	defer server.Close()

	client := New(server.URL, "key", "model", 20*time.Millisecond)
	_, err := client.Complete(context.Background(), "system", "hello")
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("Complete() error = %v, want timeout", err)
	}
}
