> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Native SDK Support

> Use OpenAI, Google Gen AI, Anthropic, and boto3 SDKs with the TrueFoundry AI Gateway

You can use these native provider SDKs with the TrueFoundry Gateway to send requests to your configured models. **Tracing**, **cost tracking**, and **rate budget limits** are supported for all of them.

| SDK                                                       | Models supported         | Cost tracking | Rate/Budget limits | Routing configuration |
| --------------------------------------------------------- | ------------------------ | ------------- | ------------------ | --------------------- |
| [OpenAI](#openai-sdk)                                     | All models               | Yes           | Yes                | Yes                   |
| [Google Gen AI](#google-gen-ai-sdk)                       | Gemini, Vertex AI models | Yes           | Yes                | No                    |
| [Anthropic](#anthropic-sdk)                               | Anthropic models only    | Yes           | Yes                | Yes                   |
| [boto3 (Bedrock Runtime)](#boto3-sdk-aws-bedrock-runtime) | Bedrock models           | Yes           | Yes                | No                    |
| [langchain\_aws](#langchain-aws-sdk)                      | Bedrock models           | Yes           | Yes                | No                    |

## OpenAI SDK

Use the standard OpenAI client with the Gateway base URL and your TrueFoundry API key. See [Chat Completions - Getting Started](/docs/ai-gateway/chat-completions-overview#getting-started) for the full example:

```python lines theme={"dark"}
from openai import OpenAI

client = OpenAI(
    api_key="your_truefoundry_api_key",
    base_url="{GATEWAY_BASE_URL}"
)

response = client.chat.completions.create(
    model="openai-main/gpt-4o-mini", # tfy model name
    messages=[{"role": "user", "content": "Hello, how are you?"}]
)

print(response.choices[0].message.content)
```

## Google Gen AI SDK

The [Google Gen AI SDK](https://github.com/google/genai) can be used with **Gemini** and **Vertex AI** via the Gateway by setting a custom base URL.

<Note>
  **Model ID must match the model name:** When using the Google Gen AI SDK with the Gateway, the model's display name in TrueFoundry must match the model ID, because the Gateway authorizes your API key against the model. You only need to pass the **actual model ID** (e.g. `gemini-2.5-flash`) in your request.
</Note>

<Note>
  **Routing configuration** is not supported when using the Google Gen AI SDK with the Gateway.
</Note>

```python lines theme={"dark"}
from google import genai
from google.genai import types

client = genai.Client(
    api_key="TFY_API_KEY",
    http_options=types.HttpOptions(
        base_url="{GATEWAY_BASE_URL}/gemini/{providerAccountName}/proxy"
    )
)

response = client.models.generate_content(
    model="gemini-2.5-flash",  # actual model id
    contents="How does AI work?",
)

print(response.text)
```

Replace `{providerAccountName}` with your **Gemini** or **Vertex AI** provider account name from the Gateway.

## Anthropic SDK

Use the Anthropic client with the Gateway base URL and your TrueFoundry API key to call Claude models:

```python lines theme={"dark"}
from anthropic import Anthropic

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

client = Anthropic(
    api_key=API_KEY,
    base_url=BASE_URL,
    default_headers={
        "Authorization": f"Bearer {API_KEY}"
    }
)

def generate_response():
    response = client.messages.create(
        model="anthropic/claude-3-5",  # tfy model name
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

## boto3 SDK (AWS Bedrock Runtime)

Use `boto3` with the Bedrock Runtime client when you want Bedrock-native APIs (`converse`, `converse_stream`, `invoke_model`, `invoke_model_with_response_stream`) through the Gateway.

Set your endpoint to the Gateway Bedrock proxy and pass your TrueFoundry API key as `AWS_BEARER_TOKEN_BEDROCK`.

```python lines theme={"dark"}
import boto3
import os

os.environ["AWS_BEARER_TOKEN_BEDROCK"] = "your_truefoundry_api_key"

bedrock_runtime = boto3.client(
    service_name="bedrock-runtime",
    region_name=os.environ.get("AWS_REGION", "dummy"),
    endpoint_url="{GATEWAY_BASE_URL}/bedrock/proxy",
)

response = bedrock_runtime.converse(
    modelId="tfy-ai-bedrock/global-anthropic-claude-opus-4-5-20251101-v1-0",  # tfy model name
    messages=[
        {"role": "user", "content": [{"text": "Hello, how are you?"}]},
    ],
    inferenceConfig={"maxTokens": 256, "temperature": 0.7},
)

for block in response["output"]["message"]["content"]:
    if "text" in block:
        print(block["text"], end="")
```

### Anthropic models with `invoke_model*`

For Anthropic models, Bedrock `invoke_model` and `invoke_model_with_response_stream` should use Anthropic Messages API fields:

* Add `anthropic_version` (for example, `bedrock-2023-05-31`)
* Add `max_tokens`
* Use typed content blocks: `{"type": "text", "text": "..."}`

Use typed blocks for `content` to avoid validation errors, for example: `messages.0.content.0.type: Field required`.

```python lines theme={"dark"}
import json

response = bedrock_runtime.invoke_model(
    modelId="tfy-ai-bedrock/global-anthropic-claude-opus-4-5-20251101-v1-0",
    contentType="application/json",
    accept="application/json",
    body=json.dumps({
        "anthropic_version": "bedrock-2023-05-31",
        "max_tokens": 256,
        "messages": [
            {
                "role": "user",
                "content": [{"type": "text", "text": "Hello, how are you?"}],
            }
        ],
    }),
)
```

## LangChain AWS SDK

You can also use `langchain_aws` with the TFY AI Gateway Bedrock endpoint:

```python lines theme={"dark"}
from langchain_aws import ChatBedrockConverse

llm = ChatBedrockConverse(
    model="tfy-ai-bedrock/global-anthropic-claude-opus-4-5-20251101-v1-0",  # tfy model name
    region_name="dummy",
    endpoint_url="{GATEWAY_BASE_URL}/bedrock/proxy",
    api_key="your_truefoundry_api_key",
    temperature=0.2,
    max_tokens=256,
)

response = llm.invoke([("human", "Hello, how are you?")])
print(response.content)
```
