package billing

import (
	"strings"

	"github.com/Delpak/shahrah/packages/contracts/openai"
)

type pricing struct {
	Input  float64
	Output float64
}

func modelPrice(model string) pricing {
	switch {
	case strings.Contains(model, "gpt-4"):
		return pricing{Input: 0.00003, Output: 0.00006}
	default:
		return pricing{Input: 0.0000005, Output: 0.0000015}
	}
}

func estimatePromptTokens(req openai.ChatCompletionRequest) int {
	totalChars := 0
	for _, msg := range req.Messages {
		totalChars += len(msg.Content) + len(msg.Role) + len(msg.Name)
	}
	if totalChars == 0 {
		return 1
	}
	return (totalChars / 4) + len(req.Messages)*4
}
