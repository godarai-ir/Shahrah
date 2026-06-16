package storage

import (
	"strings"

	"github.com/Delpak/shahrah/services/control-plane/internal/modelid"
)

func normalizeModelFQN(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}
	provider, accountAlias, officialModel, ok := modelid.SplitModelFQN(model)
	if !ok {
		return model
	}
	return modelid.BuildModelFQN(provider, accountAlias, officialModel)
}
