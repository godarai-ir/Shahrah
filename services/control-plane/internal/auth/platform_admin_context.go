package auth

import "github.com/gofiber/fiber/v2"

type platformServiceAccountContextKey struct{}

// PlatformServiceAccountContext identifies the credential used for platform administration.
type PlatformServiceAccountContext struct {
	ID   string
	Name string
}

// PlatformServiceAccountFromContext returns the authenticated platform service account.
func PlatformServiceAccountFromContext(c *fiber.Ctx) *PlatformServiceAccountContext {
	if c == nil {
		return nil
	}

	account, ok := c.Locals(platformServiceAccountContextKey{}).(*PlatformServiceAccountContext)
	if !ok {
		return nil
	}

	return account
}

func setPlatformServiceAccountContext(c *fiber.Ctx, account PlatformServiceAccountContext) {
	c.Locals(platformServiceAccountContextKey{}, &account)
}
