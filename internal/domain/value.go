package domain

import (
	"errors"
	"fmt"
	"net"
	"path"
	"regexp"
	"strings"
	"unicode"
)

var (
	ErrInvalidName        = errors.New("domain: invalid name")
	ErrInvalidRemotePath  = errors.New("domain: invalid remote path")
	ErrInvalidBranch      = errors.New("domain: invalid branch")
	ErrInvalidComposeFile = errors.New("domain: invalid compose file")
	ErrInvalidHost        = errors.New("domain: invalid host")
	ErrInvalidSSHUser     = errors.New("domain: invalid ssh user")
	ErrInvalidPort        = errors.New("domain: invalid port")
)

var (
	branchPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]{0,127}$`)
	hostPattern   = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9.-]{0,251}[A-Za-z0-9])?$`)
	userPattern   = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]{0,31}$`)
)

type Name struct{ raw string }

func NewName(raw string) (Name, error) {
	clean := strings.TrimSpace(raw)
	if clean == "" || len(clean) > 80 || hasControl(clean) {
		return Name{}, fmt.Errorf("name %q: %w", raw, ErrInvalidName)
	}
	return Name{raw: clean}, nil
}

func (n Name) String() string { return n.raw }

type RemotePath struct{ raw string }

func NewRemotePath(raw string) (RemotePath, error) {
	clean := path.Clean(strings.TrimSpace(raw))
	if clean == "/" || !path.IsAbs(clean) || hasControl(clean) {
		return RemotePath{}, fmt.Errorf("remote path %q: %w", raw, ErrInvalidRemotePath)
	}
	return RemotePath{raw: clean}, nil
}

func (p RemotePath) String() string { return p.raw }

type Branch struct{ raw string }

func NewBranch(raw string) (Branch, error) {
	clean := strings.TrimSpace(raw)
	if !branchPattern.MatchString(clean) || strings.Contains(clean, "..") || strings.HasSuffix(clean, "/") {
		return Branch{}, fmt.Errorf("branch %q: %w", raw, ErrInvalidBranch)
	}
	return Branch{raw: clean}, nil
}

func (b Branch) String() string { return b.raw }

type ComposeFile struct{ raw string }

func NewComposeFile(raw string) (ComposeFile, error) {
	trimmed := strings.TrimSpace(raw)
	clean := path.Clean(trimmed)
	if clean == "." || clean != trimmed || path.IsAbs(clean) || strings.HasPrefix(clean, "../") || hasControl(clean) {
		return ComposeFile{}, fmt.Errorf("compose file %q: %w", raw, ErrInvalidComposeFile)
	}
	return ComposeFile{raw: clean}, nil
}

func (f ComposeFile) String() string { return f.raw }

type Host struct{ raw string }

func NewHost(raw string) (Host, error) {
	clean := strings.TrimSpace(raw)
	isIP := net.ParseIP(clean) != nil
	if clean == "" || (!isIP && !hostPattern.MatchString(clean)) {
		return Host{}, fmt.Errorf("host %q: %w", raw, ErrInvalidHost)
	}
	return Host{raw: clean}, nil
}

func (h Host) String() string { return h.raw }

type SSHUser struct{ raw string }

func NewSSHUser(raw string) (SSHUser, error) {
	clean := strings.TrimSpace(raw)
	if !userPattern.MatchString(clean) {
		return SSHUser{}, fmt.Errorf("ssh user %q: %w", raw, ErrInvalidSSHUser)
	}
	return SSHUser{raw: clean}, nil
}

func (u SSHUser) String() string { return u.raw }

type Port uint16

func NewPort(raw int) (Port, error) {
	if raw < 1 || raw > 65535 {
		return 0, fmt.Errorf("port %d: %w", raw, ErrInvalidPort)
	}
	return Port(raw), nil
}

func (p Port) Int() int { return int(p) }

func hasControl(value string) bool {
	return strings.IndexFunc(value, unicode.IsControl) >= 0
}
