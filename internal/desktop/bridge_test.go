package desktop

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
)

func TestBridgeRequest_whenLocalAPIRequest_thenForwardsResponse(t *testing.T) {
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", request.Method)
		}
		if request.URL.RequestURI() != "/api/v1/plans/preview?source=desktop" {
			t.Errorf("request URI = %q", request.URL.RequestURI())
		}
		if request.Header.Get("Content-Type") != "application/json" {
			t.Errorf("content type = %q", request.Header.Get("Content-Type"))
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if string(body) != `{"prompt":"deploy"}` {
			t.Errorf("body = %q", body)
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusCreated)
		if _, err := writer.Write([]byte(`{"status":"draft"}`)); err != nil {
			t.Errorf("write response: %v", err)
		}
	})
	bridge := NewBridge(handler)

	response, err := bridge.Request(context.Background(), Request{
		Method: http.MethodPost,
		Path:   "/api/v1/plans/preview?source=desktop",
		Body:   `{"prompt":"deploy"}`,
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if response.Status != http.StatusCreated {
		t.Errorf("status = %d, want %d", response.Status, http.StatusCreated)
	}
	if response.ContentType != "application/json" {
		t.Errorf("content type = %q", response.ContentType)
	}
	if response.Body != `{"status":"draft"}` {
		t.Errorf("body = %q", response.Body)
	}
}

func TestBridgeRequest_whenRequestEscapesLocalAPI_thenRejects(t *testing.T) {
	bridge := NewBridge(http.NotFoundHandler())
	tests := []Request{
		{Method: http.MethodDelete, Path: "/api/v1/projects"},
		{Method: http.MethodGet, Path: "https://example.com/api/v1/projects"},
		{Method: http.MethodGet, Path: "/debug"},
	}
	for _, request := range tests {
		_, err := bridge.Request(context.Background(), request)
		if !errors.Is(err, ErrInvalidRequest) {
			t.Errorf("Request(%+v) error = %v, want ErrInvalidRequest", request, err)
		}
	}
}
