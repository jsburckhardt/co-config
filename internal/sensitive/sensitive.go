package sensitive

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// SensitiveFields is the list of known sensitive config field names.
var SensitiveFields = []string{
	"copilot_tokens",
	"logged_in_users",
	"last_logged_in_user",
	"staff",
}

// IsSensitive checks if a field name is classified as sensitive.
func IsSensitive(fieldName string) bool {
	for _, f := range SensitiveFields {
		if f == fieldName {
			return true
		}
	}
	return false
}

// MaskValue returns a display-safe representation of a sensitive value.
// Strings: truncated SHA-256 hash (12 hex chars) + "..."
// Maps: "[redacted — N items]"
// Slices: "[redacted — N items]"
// Other: "[redacted]"
func MaskValue(value any) string {
	switch v := value.(type) {
	case string:
		hash := sha256.Sum256([]byte(v))
		return fmt.Sprintf("%x...", hash[:6])
	case map[string]any:
		return fmt.Sprintf("[redacted — %d items]", len(v))
	case []any:
		return fmt.Sprintf("[redacted — %d items]", len(v))
	default:
		return "[redacted]"
	}
}

// LooksLikeToken checks if a string value looks like an authentication token.
func LooksLikeToken(value string) bool {
	prefixes := []string{"gho_", "ghp_", "github_pat_"}
	for _, p := range prefixes {
		if strings.HasPrefix(value, p) {
			return true
		}
	}
	return false
}
