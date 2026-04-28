> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Messages API (/messages)

> Learn how to use Messages API through TrueFoundry's AI Gateway for interacting with Claude models.

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider  | Messages API |
| --------- | ------------ |
| Anthropic | ✅            |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

[Anthropic's Messages API](https://docs.anthropic.com/en/api/messages) is a powerful interface for interacting with Claude models. When using TruefFundry as your model gateway, you can access this API through a proxy endpoint that handles authentication and routing to the appropriate model.

## **Prerequisites**

To use the Anthropic Messages API through Truefoundry, you'll need:

1. TrueFoundry API Key
2. Provider account configured in TrueFoundry (Anthropic)
3. Python environment with `anthropic sdk` library installed

## Using the Anthropic SDK

The Anthropic Python SDK provides a convenient way to interact with Claude models. Here's how to configure it to work with the Truefoundry proxy:

<Note>
  The gateway accepts both Anthropic SDK auth patterns and translates internally:

  * `api_key=TFY_API_KEY` - SDK sends the `x-api-key` header
  * `auth_token=TFY_API_KEY` — SDK sends the `Authorization: Bearer` header

  Either works; the request body is identical. `api_key` is the idiomatic Anthropic SDK pattern - use it unless you have a reason to send a Bearer token.
</Note>

```python lines theme={"dark"}
from anthropic import Anthropic

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

# Configure the Anthropic client to use TrueFoundry's Gateway
client = Anthropic(
    api_key=API_KEY,
    base_url=BASE_URL
)

# Make a request to the Messages API
def generate_response():
    response = client.messages.create(
        model="anthropic/claude-3-5",  # The model name configured in Truefoundry
        max_tokens=1024,
        messages=[
            {
                "role": "user",
                "content": "Hello, Claude! Please explain quantum computing in simple terms."
            }
        ]
    )

    print(response.content)

generate_response()
```

## Request Format

When using the Messages API through Truefoundry, your request should follow this format:

```json lines theme={"dark"}
{
  "model": "tfy-model-name",
  "max_tokens": 1024,
  "messages": [
    {
      "role": "user",
      "content": "Hello, Claude! Please explain quantum computing in simple terms."
    }
  ]
}
```

## Response Format

The response from the Messages API will have this structure:

```json lines theme={"dark"}
{
  "id": "msg_01XB89YSAA2VGMCF3ZS8ATTA1B",
  "type": "message",
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "Quantum computing is like traditional computing but it uses quantum bits or 'qubits' instead of regular bits. While traditional bits can only be in a state of 0 or 1, qubits can exist in multiple states simultaneously thanks to a quantum property called 'superposition.' This allows quantum computers to process certain types of information much faster than regular computers.\n\nAnother key quantum property is 'entanglement,' where qubits become connected and the state of one instantly affects the other, no matter the distance between them.\n\nThese properties give quantum computers the potential to solve certain complex problems much faster than traditional computers, like factoring large numbers (important for encryption) or simulating molecular structures (useful for drug development).\n\nHowever, quantum computers are still in early development stages. They're extremely sensitive to their environment and require special conditions like ultra-cold temperatures to operate. They're not replacements for regular computers but specialized tools for specific types of problems."
    }
  ],
  "model": "claude-3-opus-20240229",
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 14,
    "output_tokens": 178
  }
}
```

## Advanced Features

The Messages API supports several advanced features:

### System Prompts

You can include a system prompt to guide Claude's behavior:

```python lines theme={"dark"}
client.messages.create(
    model="anthropic/claude-3-5",
    system="You are a helpful AI assistant that specializes in explaining complex topics simply.",
    messages=[
        {"role": "user", "content": "Explain quantum entanglement."}
    ]
)
```

### Multi-turn Conversations

For multi-turn conversations, include previous messages:

```python lines theme={"dark"}
client.messages.create(
    model="anthropic/claude-3-5",
    messages=[
        {"role": "user", "content": "What is machine learning?"},
        {"role": "assistant", "content": "Machine learning is a subset of artificial intelligence..."},
        {"role": "user", "content": "Can you explain supervised vs unsupervised learning?"}
    ]
)
```

### Streaming Responses

For streaming responses, use the streaming parameter:

```python lines theme={"dark"}
with client.messages.stream(
    model="anthropic/claude-3-5",
    messages=[{"role": "user", "content": "Write a short poem about AI."}]
) as stream:
    for text in stream.text_stream:
        print(text, end="", flush=True)

    # Access the final message at the end
    print("\nFinal message:", stream.get_final_message())
```
