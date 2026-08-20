package httpapi

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Delpak/shahrah/packages/contracts/platform"
)

func validateChatMessage(message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("message is required")
	}

	if utf8.RuneCountInString(message) > platform.RAGMaxQueryCharacters {
		return fmt.Errorf("message exceeds %d characters", platform.RAGMaxQueryCharacters)
	}

	return nil
}

func validateTextLength(name, value string, minimum, maximum int) error {
	length := utf8.RuneCountInString(strings.TrimSpace(value))
	if length < minimum || length > maximum {
		return fmt.Errorf("%s must contain between %d and %d characters", name, minimum, maximum)
	}

	return nil
}
