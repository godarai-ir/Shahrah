package bootstrap

import (
	"errors"
	"fmt"

	"github.com/Delpak/shahrah/services/control-plane/internal/chatprovisioning"
)

// buildChatProvisioning starts the reconciler that registers every tenant with GodarChat. It is
// off unless CHAT_PROVISIONING_ENABLED is set, and it needs the OIDC provider to create clients.
func (b *builder) buildChatProvisioning() error {
	settings := b.cfg.ChatProvisioning
	if !settings.Enabled {
		return nil
	}

	if err := settings.Validate(); err != nil {
		return err
	}

	if b.oidcApplication == nil {
		return errors.New("CHAT_PROVISIONING_ENABLED requires OIDC_ENABLED")
	}

	chat, err := chatprovisioning.NewHTTPClient(settings.ProvisioningURL, settings.Token, settings.RequestTimeout)
	if err != nil {
		return err
	}

	provisioner, err := chatprovisioning.New(chatprovisioning.Dependencies{
		Store:           b.store.ChatProvisioningRepository,
		Clients:         b.oidcApplication,
		Chat:            chat,
		ChatURLTemplate: settings.ChatURLTemplate,
		Interval:        settings.Interval,
		BatchSize:       settings.BatchSize,
		Logger:          b.logger,
	})
	if err != nil {
		return fmt.Errorf("initialize chat provisioning: %w", err)
	}

	return b.lifecycle.AddLoop("chat provisioning", provisioner.Run)
}
