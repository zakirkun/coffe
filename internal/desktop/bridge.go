package desktop

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

var ErrInvalidRequest = errors.New("invalid desktop request")

type Request struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Body   string `json:"body"`
}

type Response struct {
	Status      int    `json:"status"`
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
}

type Bridge struct {
	handler http.Handler
}

func NewBridge(handler http.Handler) *Bridge {
	return &Bridge{handler: handler}
}

func (b *Bridge) Request(ctx context.Context, input Request) (Response, error) {
	if err := validateRequest(input); err != nil {
		return Response{}, err
	}
	request, err := http.NewRequestWithContext(
		ctx,
		input.Method,
		"http://coffe.local"+input.Path,
		strings.NewReader(input.Body),
	)
	if err != nil {
		return Response{}, fmt.Errorf("create desktop request: %w", err)
	}
	if input.Body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	b.handler.ServeHTTP(recorder, request)
	result := recorder.Result()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return Response{}, fmt.Errorf("read desktop response: %w", err)
	}
	if err := result.Body.Close(); err != nil {
		return Response{}, fmt.Errorf("close desktop response: %w", err)
	}
	return Response{
		Status:      result.StatusCode,
		ContentType: result.Header.Get("Content-Type"),
		Body:        string(body),
	}, nil
}

func validateRequest(input Request) error {
	switch input.Method {
	case http.MethodGet, http.MethodPost, http.MethodPut:
	default:
		return fmt.Errorf("method %q: %w", input.Method, ErrInvalidRequest)
	}
	parsed, err := url.ParseRequestURI(input.Path)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return fmt.Errorf("path %q: %w", input.Path, ErrInvalidRequest)
	}
	if parsed.Path != "/healthz" && !strings.HasPrefix(parsed.Path, "/api/v1/") {
		return fmt.Errorf("path %q: %w", input.Path, ErrInvalidRequest)
	}
	return nil
}
