> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Audio Translation API

> Convert audio files to text using translations models through TrueFoundry's AI Gateway

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider         | Translation                                                 |
| ---------------- | ----------------------------------------------------------- |
| OpenAI           | ✅                                                           |
| Azure OpenAI     | ✅                                                           |
| Azure AI Foundry | ✅                                                           |
| Anthropic        | <Icon icon="circle-minus" iconType="regular" />             |
| Bedrock          | <Icon icon="circle-minus" iconType="regular" />             |
| Vertex           | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Cohere           | <Icon icon="circle-minus" iconType="regular" />             |
| Gemini           | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Groq             | ✅                                                           |
| Together-AI      | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI              | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra        | <Icon icon="circle-xmark" iconType="regular" color="red" /> |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

### Supported Providers

* OpenAI
* Azure OpenAI
* Groq

The Translations API converts spoken audio into English text using the `whisper-1` model. You can provide audio in a variety of formats, and the API will return an accurate English translation.

You can customize the input and output as follows:

* **Model**: `whisper-1`
* **Input formats**: `mp3`, `mp4`, `mpeg`, `mpga`, `m4a`, `wav`, `webm`
* **Output**: Translated English text

## Code Snippet

```python lines theme={"dark"}
from openai import OpenAI

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

# Configure OpenAI client with TrueFoundry settings
client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
)

audio_file= open("/path/to/file/audio.mp3", "rb")

translation = client.audio.translations.create(
    model="openai-main/whisper-1",
    file=audio_file
)

print(translation.text)
```

By default, the response type will be json with the raw text included.

```json lines theme={"dark"}
{
  "text": "Hello, how are you?"
}
```
