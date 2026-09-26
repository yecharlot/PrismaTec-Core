// Package security holds input validation, auth helpers and resource limits (Phase 15).
package security

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	nameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{0,62}$`)
	idRe   = regexp.MustCompile(`^[a-zA-Z0-9:._-]{1,128}$`)
)

// Limits for untrusted execution.
var DefaultExecTimeout = 5 * time.Second

// ValidateName checks organism / human names.
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name required")
	}
	if utf8.RuneCountInString(name) > 64 {
		return fmt.Errorf("name too long (max 64)")
	}
	if !nameRe.MatchString(name) {
		return fmt.Errorf("name must be alphanumeric with ._- (1–63 chars)")
	}
	return nil
}

// ValidateID checks organism or node id shape.
func ValidateID(id string) error {
	if id == "" {
		return fmt.Errorf("id required")
	}
	if !idRe.MatchString(id) {
		return fmt.Errorf("invalid id format")
	}
	return nil
}

// ValidateAction checks capability / command action strings.
func ValidateAction(action string) error {
	if action == "" {
		return fmt.Errorf("action required")
	}
	if len(action) > 64 {
		return fmt.Errorf("action too long")
	}
	for _, c := range action {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '.' || c == '_' || c == '-' || c == '*') {
			return fmt.Errorf("invalid action character")
		}
	}
	return nil
}

// ValidateMemoryKey validates working-memory keys.
func ValidateMemoryKey(key string) error {
	if key == "" {
		return fmt.Errorf("key required")
	}
	if utf8.RuneCountInString(key) > 128 {
		return fmt.Errorf("key too long")
	}
	return nil
}

// ValidateMemoryValue bounds payload size.
func ValidateMemoryValue(val string) error {
	if len(val) > 64*1024 {
		return fmt.Errorf("value too large (max 64KiB)")
	}
	return nil
}

// AIPToken returns the shared secret from env (empty = auth disabled for local dev).
func AIPToken() string {
	return strings.TrimSpace(os.Getenv("PRISMATEC_AIP_TOKEN"))
}

// CheckAIPToken validates Authorization: Bearer <token> or X-PrismaTec-Token.
func CheckAIPToken(headerAuth, headerToken string) error {
	want := AIPToken()
	if want == "" {
		return nil // open mode (local)
	}
	got := strings.TrimSpace(headerToken)
	if got == "" && strings.HasPrefix(strings.ToLower(headerAuth), "bearer ") {
		got = strings.TrimSpace(headerAuth[7:])
	}
	if got == "" {
		return fmt.Errorf("authentication required")
	}
	if got != want {
		return fmt.Errorf("invalid token")
	}
	return nil
}
