//go:build wailsapp

package main

import (
	"strings"

	"coffe/internal/secrets"
)

type SecretInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (a *App) StoreSecret(input SecretInput) error {
	keyring, err := secrets.NewKeyring("coffe")
	if err != nil {
		return err
	}
	return keyring.Set(a.ctx, input.Name, strings.TrimSpace(input.Value))
}
