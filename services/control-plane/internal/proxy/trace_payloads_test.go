package proxy

import (
	"testing"

	contractsopenai "github.com/Delpak/shahrah/packages/contracts/openai"
)

func TestBuildTraceChatRequestPayloadRedactsImageDataURL(t *testing.T) {
	req := contractsopenai.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []contractsopenai.ChatMessage{{
			Role: "user",
			Content: []interface{}{
				map[string]interface{}{"type": "text", "text": "Describe this image."},
				map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "data:image/png;base64,AAAA"}},
			},
		}},
	}

	payload := buildTraceChatRequestPayload(req)
	messages := payload["messages"].([]map[string]interface{})
	content := messages[0]["content"].([]map[string]interface{})

	imageURL := content[1]["image_url"].(map[string]interface{})
	if got := imageURL["url"]; got != "data:image/png;base64,[redacted]" {
		t.Fatalf("image url = %#v", got)
	}
}
