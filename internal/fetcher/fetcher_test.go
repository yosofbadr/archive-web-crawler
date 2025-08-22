package fetcher

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	client := &http.Client{}
	f := New(client)

	if f.client != client {
		t.Errorf("Expected client to be %v, but got %v", client, f.client)
	}
}

func TestFetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!"))
	}))
	defer server.Close()

	f := New(server.Client())
	body, err := f.Fetch(server.URL)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("Unexpected error reading body: %v", err)
	}

	if string(bodyBytes) != "Hello, world!" {
		t.Errorf("Expected body to be 'Hello, world!', but got '%s'", string(bodyBytes))
	}
}

func TestFetch_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	f := New(server.Client())
	_, err := f.Fetch(server.URL)
	if err == nil {
		t.Fatal("Expected an error, but got nil")
	}
}
