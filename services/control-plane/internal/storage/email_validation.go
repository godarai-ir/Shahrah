package storage

import "github.com/Delpak/shahrah/services/control-plane/internal/tenantadmin"

// IsValidEmailAddress enforces a strict, ASCII-only email format used by invite flows.
func IsValidEmailAddress(raw string) bool {
	return tenantadmin.IsValidEmailAddress(raw)
}
