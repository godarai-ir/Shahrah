package guardrails

import (
	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

type messageRef struct {
	MessageIndex int
	PartIndex    int
	PartContent  bool
}

type choiceRef struct {
	ChoiceIndex int
}

func inputTexts(req contractsopenai.ChatCompletionRequest, scope Scope) ([]string, []messageRef) {
	start := 0
	if scope == ScopeLast {
		start = len(req.Messages) - 1
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if req.Messages[i].Role == "user" {
				start = i
				break
			}
		}
	}
	texts := []string{}
	refs := []messageRef{}
	for i := start; i < len(req.Messages); i++ {
		msg := req.Messages[i]
		switch content := msg.Content.(type) {
		case string:
			texts = append(texts, content)
			refs = append(refs, messageRef{MessageIndex: i})
		default:
			parts, ok, err := contractsopenai.ChatContentParts(content)
			if !ok || err != nil {
				continue
			}
			for partIndex, part := range parts {
				if part.Type != "text" {
					continue
				}
				texts = append(texts, part.Text)
				refs = append(refs, messageRef{MessageIndex: i, PartIndex: partIndex, PartContent: true})
			}
		}
		if scope == ScopeLast {
			break
		}
	}
	return texts, refs
}

func applyInputTexts(req *contractsopenai.ChatCompletionRequest, refs []messageRef, texts []string) {
	if req == nil {
		return
	}
	for i, ref := range refs {
		if i >= len(texts) || ref.MessageIndex < 0 || ref.MessageIndex >= len(req.Messages) {
			continue
		}
		if !ref.PartContent {
			req.Messages[ref.MessageIndex].Content = texts[i]
			continue
		}
		parts, ok, err := contractsopenai.ChatContentParts(req.Messages[ref.MessageIndex].Content)
		if !ok || err != nil || ref.PartIndex < 0 || ref.PartIndex >= len(parts) {
			continue
		}
		parts[ref.PartIndex].Text = texts[i]
		req.Messages[ref.MessageIndex].Content = parts
	}
}

func outputTexts(resp contractsopenai.ChatCompletionResponse) ([]string, []choiceRef) {
	texts := []string{}
	refs := []choiceRef{}
	for i, choice := range resp.Choices {
		if choice.Message == nil {
			continue
		}
		text := contractsopenai.ChatContentText(choice.Message.Content)
		if text == "" {
			continue
		}
		texts = append(texts, text)
		refs = append(refs, choiceRef{ChoiceIndex: i})
	}
	return texts, refs
}

func applyOutputTexts(resp *contractsopenai.ChatCompletionResponse, refs []choiceRef, texts []string) {
	if resp == nil {
		return
	}
	for i, ref := range refs {
		if i >= len(texts) || ref.ChoiceIndex < 0 || ref.ChoiceIndex >= len(resp.Choices) || resp.Choices[ref.ChoiceIndex].Message == nil {
			continue
		}
		resp.Choices[ref.ChoiceIndex].Message.Content = texts[i]
	}
}
