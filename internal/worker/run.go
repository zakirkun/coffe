package worker

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

type RunConfig struct {
	Addr             string
	Token            string
	Version          string
	TLSCert          string
	TLSKey           string
	InsecureLoopback bool
}

func Run(ctx context.Context, config RunConfig) error {
	handler, err := NewServer(ServerConfig{Token: config.Token, Version: config.Version})
	if err != nil {
		return err
	}
	if err := validateRunConfig(config); err != nil {
		return err
	}
	server := &http.Server{
		Addr: config.Addr, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 5 * time.Minute, IdleTimeout: 60 * time.Second,
	}
	serverError := make(chan error, 1)
	go func() { serverError <- listen(server, config) }()
	return waitForServer(ctx, server, serverError)
}

func validateRunConfig(config RunConfig) error {
	if config.InsecureLoopback {
		host, _, err := net.SplitHostPort(config.Addr)
		if err != nil || (host != "127.0.0.1" && host != "localhost" && host != "::1") {
			return fmt.Errorf("insecure worker must bind to loopback")
		}
		return nil
	}
	if config.TLSCert == "" || config.TLSKey == "" {
		return fmt.Errorf("worker TLS certificate and key are required")
	}
	return nil
}

func listen(server *http.Server, config RunConfig) error {
	if config.InsecureLoopback {
		return server.ListenAndServe()
	}
	return server.ListenAndServeTLS(config.TLSCert, config.TLSKey)
}

func waitForServer(ctx context.Context, server *http.Server, serverError <-chan error) error {
	select {
	case err := <-serverError:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve worker: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown worker: %w", err)
		}
		if err := <-serverError; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("stop worker: %w", err)
		}
		return nil
	}
}
