package secrets

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

var ErrNotFound = errors.New("secrets: not found")

type Keyring struct{ service string }

func NewKeyring(service string) (*Keyring, error) {
	clean := strings.TrimSpace(service)
	if clean == "" {
		return nil, fmt.Errorf("keyring service: empty")
	}
	return &Keyring{service: clean}, nil
}

func (k *Keyring) Get(ctx context.Context, name string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("get secret: %w", err)
	}
	value, err := keyring.Get(k.service, name)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", fmt.Errorf("secret %q: %w", name, ErrNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("get secret %q: %w", name, err)
	}
	return value, nil
}

func (k *Keyring) Set(ctx context.Context, name, value string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("set secret: %w", err)
	}
	if strings.TrimSpace(name) == "" || value == "" {
		return fmt.Errorf("secret name or value: empty")
	}
	if err := keyring.Set(k.service, name, value); err != nil {
		return fmt.Errorf("set secret %q: %w", name, err)
	}
	return nil
}

func (k *Keyring) Delete(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("delete secret: %w", err)
	}
	if err := keyring.Delete(k.service, name); err != nil {
		return fmt.Errorf("delete secret %q: %w", name, err)
	}
	return nil
}
