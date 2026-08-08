package service

import "errors"

var (
	ErrNotFound          = errors.New("service: not found")
	ErrInvalidInput      = errors.New("service: invalid input")
	ErrUnsupportedIntent = errors.New("service: unsupported intent")
	ErrNotRunning        = errors.New("service: plan is not running")
	ErrReadOnlyQuery     = errors.New("service: only read queries are allowed")
	ErrPolicyDenied      = errors.New("service: operation denied by policy")
)
