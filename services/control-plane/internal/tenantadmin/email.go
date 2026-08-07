package tenantadmin

import (
	"regexp"
	"strings"
)

var (
	emailLocalPartPattern = regexp.MustCompile(`^[A-Za-z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+$`)
	emailDomainLabelRegex = regexp.MustCompile(`^[A-Za-z0-9-]+$`)
	emailTLDRegex         = regexp.MustCompile(`^[A-Za-z]{2,63}$`)
)

// IsValidEmailAddress enforces the ASCII-only email format used by invite flows.
func IsValidEmailAddress(raw string) bool {
	email := strings.TrimSpace(raw)
	if email == "" || strings.ContainsAny(email, " \t\r\n") {
		return false
	}

	at := strings.Index(email, "@")
	if at <= 0 || at != strings.LastIndex(email, "@") || at == len(email)-1 {
		return false
	}

	local := email[:at]
	domain := email[at+1:]

	if !emailLocalPartPattern.MatchString(local) {
		return false
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, label := range parts {
		if label == "" || len(label) > 63 || !emailDomainLabelRegex.MatchString(label) {
			return false
		}

		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
	}

	return emailTLDRegex.MatchString(parts[len(parts)-1])
}
