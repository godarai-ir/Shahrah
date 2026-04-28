> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Image Generation API (/images/generations)

> Generate images from text prompts through TrueFoundry's AI Gateway

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider     | Generate                                                    |
| ------------ | ----------------------------------------------------------- |
| OpenAI       | ✅                                                           |
| Azure OpenAI | ✅                                                           |
| Bedrock      | ✅                                                           |
| Vertex       | ✅                                                           |
| Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
| Cohere       | <Icon icon="circle-minus" iconType="regular" />             |
| Gemini       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Groq         | <Icon icon="circle-minus" iconType="regular" />             |
| Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI          | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

TrueFoundry AI Gateway supports text-to-image generation capabilities from multiple providers. Users can send text prompts to generate corresponding images.

## Supported Providers

* **OpenAI**: `gpt-image-1`, `dall-e-2`, `dall-e-3`
* **Vertex AI**:
  * **Imagen 3.0/4.0 models**: `imagen-3.0-generate-002`, `imagen-3.0-generate-001`, `imagen-3.0-fast-generate-001`, `imagen-4.0-generate`, `imagen-4.0-fast-generate`, `imagen-4.0-ultra-generate`
  * **Gemini models**: `gemini-2.5-flash-image`, `gemini-3-pro-image-preview`
* **Bedrock**: Titan and Nova models (`amazon.titan-image-generator-v1`, `amazon.titan-image-generator-v2`, `amazon.nova-canvas-v1-0`) + Stability AI models
* **Azure OpenAI**: `gpt-image-1`, `dall-e-2`, `dall-e-3`

## Code Example

```python lines theme={"dark"}
from openai import OpenAI
import base64

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

# Configure OpenAI client with TrueFoundry settings
client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
)

try:
    response = client.images.generate(
        model="openai-main/dalle-e-2",
        prompt="A beautiful sunset over mountains",
        n=1,
        size="1024x1024"
    )

    # Handle URL response
    if hasattr(response.data[0], 'url') and response.data[0].url:
        print(f"Image URL: {response.data[0].url}")

    # Handle base64 response
    if hasattr(response.data[0], 'b64_json') and response.data[0].b64_json:
        image_bytes = base64.b64decode(response.data[0].b64_json)
        with open("generated_image.png", "wb") as f:
            f.write(image_bytes)
        print("Image saved as generated_image.png")

except Exception as e:
    print(f"Error: {e}")
```

## Response Format

```json lines theme={"dark"}
{
  "created": 1713833628,
  "data": [
    {
      "url": "...",
      "b64_json": "..."
    }
  ]
}
```
