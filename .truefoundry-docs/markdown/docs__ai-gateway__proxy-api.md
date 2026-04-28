> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Proxy API (/proxy)

> Route requests directly to AI provider endpoints while maintaining TrueFoundry features like logging, rate limiting, and budget management

## Provider capabilities

The Proxy API (`/proxy`) forwards provider-native requests; coverage depends on which provider account and endpoint you use. It does not use a single consolidated provider matrix like the translated APIs. See [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis) for feature matrices on those endpoints.

The Proxy API allows you to route requests directly to AI provider endpoints through TrueFoundry AI Gateway without any translation logic. This means you can use provider-native request and response formats while still benefiting from TrueFoundry's features like logging, rate limiting, and budget management.

TrueFoundry Proxy Endpoint routes your request to the appropriate provider and returns the provider's response without modifying it, preserving the native request/response format.

## Quick Start

Here's a simple example to get you started immediately:

```python lines theme={"dark"}
import openai

client = openai.OpenAI(
    api_key="your-truefoundry-api-key",
    base_url="{GATEWAY_BASE_URL}/proxy",
    default_headers={
        "x-tfy-provider-name": "openai-main"
    }
)

response = client.images.generate(
    model="dall-e-3",
    prompt="a futuristic cityscape at sunset with flying cars",
    n=1,
    size="1024x1024"
)

print("Image URL:", response.data[0].url)
```

<Note>
  **About this example:** This image generation request would work perfectly fine through TrueFoundry's standard inference API. This example specifically demonstrates how to use the **proxy route** to make requests with provider-native formats while proxying requests via TrueFoundry AI Gateway.
</Note>

That's it! The proxy API works with your existing OpenAI SDK code. For more configuration options and other providers, see the detailed examples below.

## Endpoint Structure

The proxy API supports multiple endpoint patterns for maximum flexibility:

### Standard Proxy Endpoints

```python lines theme={"dark"}
{GATEWAY_BASE_URL}/proxy/{provider-endpoint-path}
```

## Model and Provider Configuration

The proxy API supports multiple ways to specify your model and provider name:

### 1. URL Query Parameters

* **`tfyModelName`** - Complete TrueFoundry model name (e.g., `provider-name/model-name`)
* **`tfyProviderName`** - TrueFoundry provider name for routing (only required when using actual model IDs)

### 2. Request Headers

* **`x-tfy-model-name`** - Complete TrueFoundry model name
* **`x-tfy-provider-name`** - TrueFoundry provider name (only required when using actual model IDs)

### 3. Request Body

* **`model`** - Can be either TrueFoundry model name (`provider-name/model-name`) or actual model ID

<Note>
  **Model and Provider Configuration Guide:**

  **When is Provider Name Required?**

  * **Not required:** When using TrueFoundry model names (`provider-name/model-name`) - provider is auto-identified
  * **Required:** When using actual model IDs (`gpt-4o`, `claude-3-sonnet`) - specify provider via headers/query params

  **When is TrueFoundry Model Name Required?**

  * **Self-hosted models:** Always required - base URL and configuration are model-specific
  * **Full feature endpoints** (`Anthropic Messages, Claude Code, NIM Rerank`): Required for logging, rate limiting, budget tracking
  * **Other endpoints:** Provider name alone is sufficient for basic routing

  **Model Name Processing:**

  * **TrueFoundry format** (`provider-name/model-name`) → automatically converted to actual model ID
  * **Actual model ID format** → passed through unchanged to provider
</Note>

## Endpoint Support

### Endpoints with Full Feature Support

These specific endpoints have complete **logging, rate limiting, and budget management**:

* [**Anthropic Messages API**](#anthropic-messages-api) - `/messages`
* [**Claude Code**](#anthropic-messages-api) - Available from Vertex, Bedrock, and Anthropic providers
* [**NIM Rerank API**](#nim-rerank-api) - `/ranking`

### Universal Proxy Support

You can use **any** provider endpoint through the TrueFoundry Proxy API, even if it doesn't have specific feature support.

## Code Examples

### Endpoints with Full Feature Support

These endpoints provide complete TrueFoundry feature integration including logging, rate limiting, and budget management.

#### Anthropic Messages API

<CodeGroup>
  ```python Using Request Body Model Name lines theme={"dark"}
  from anthropic import Anthropic

  BASE_URL = "{GATEWAY_BASE_URL}/proxy"
  API_KEY = "your-truefoundry-api-key"

  # Configure Anthropic client with TrueFoundry settings
  client = Anthropic(
      api_key=API_KEY,
      base_url=BASE_URL
  )

  # Make requests with your TrueFoundry model name in the request body
  response = client.messages.create(
      model="anthropic-main/claude-3-5-sonnet-20241022",  # TrueFoundry model name
      max_tokens=1024,
      messages=[
          {"role": "user", "content": "Tell me a joke about programming"}
      ]
  )

  print(response.content)
  ```

  ```python Using Headers lines theme={"dark"}
  import requests

  url = "{GATEWAY_BASE_URL}/proxy/v1/messages"

  headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer your-truefoundry-api-key",
      "X-TFY-LOGGING-CONFIG": '{"enabled": true}',
      "x-tfy-model-name": "anthropic-main/claude-3-5-sonnet-20241022"
  }

  payload = {
      "max_tokens": 1024,
      "messages": [
          {"role": "user", "content": "What are the benefits of renewable energy?"}
      ]
  }

  response = requests.post(url, headers=headers, json=payload)
  print(response.json())
  ```

  ```python Using URL Query Parameters lines theme={"dark"}
  import requests

  url = "{GATEWAY_BASE_URL}/proxy/v1/messages"
  params = {
      "tfyModelName": "anthropic-main/claude-3-5-sonnet-20241022"
  }

  headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer your-truefoundry-api-key",
      "X-TFY-LOGGING-CONFIG": '{"enabled": true}'
  }

  payload = {
      "max_tokens": 1024,
      "messages": [
          {"role": "user", "content": "Explain quantum computing"}
      ]
  }

  response = requests.post(url, headers=headers, json=payload, params=params)
  print(response.json())
  ```
</CodeGroup>

#### NIM Rerank API

<CodeGroup>
  ```python Using Request Body Model Name lines theme={"dark"}
  import requests

  url = "{GATEWAY_BASE_URL}/proxy/v1/ranking"

  headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer your-truefoundry-api-key",
      "X-TFY-LOGGING-CONFIG": '{"enabled": true}'
  }

  payload = {
      "model": "custom-provider-main/llama-3.2-nemoretriever-500m-rerank-v2",
      "query": {
          "text": "What is machine learning?"
      },
      "passages": [
          {"text": "Machine learning is a subset of artificial intelligence"},
          {"text": "The weather is nice today"},
          {"text": "Python is a programming language"}
      ]
  }

  response = requests.post(url, headers=headers, json=payload)
  print(response.json())
  ```

  ```python Using Headers lines theme={"dark"}
  import requests

  url = "{GATEWAY_BASE_URL}/proxy/v1/ranking"

  headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer your-truefoundry-api-key",
      "X-TFY-LOGGING-CONFIG": '{"enabled": true}',
      "x-tfy-model-name": "custom-provider-main/llama-3.2-nemoretriever-500m-rerank-v2"
  }

  payload = {
      "query": {
          "text": "What is artificial intelligence?"
      },
      "passages": [
          {"text": "AI is a branch of computer science"},
          {"text": "Machine learning is a subset of AI"},
          {"text": "Today is a sunny day"}
      ]
  }

  response = requests.post(url, headers=headers, json=payload)
  print(response.json())
  ```

  ```python Using URL Query Parameters lines theme={"dark"}
  import requests

  url = "{GATEWAY_BASE_URL}/proxy/v1/ranking"
  params = {
      "tfyModelName": "custom-provider-main/llama-3.2-nemoretriever-500m-rerank-v2"
  }

  headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer your-truefoundry-api-key",
      "X-TFY-LOGGING-CONFIG": '{"enabled": true}'
  }

  payload = {
      "query": {
          "text": "Where is New Delhi?"
      },
      "passages": [
          {"text": "New Delhi is capital of India"},
          {"text": "The sky is blue"},
          {"text": "Paris is the capital of France"}
      ]
  }

  response = requests.post(url, headers=headers, json=payload, params=params)
  print(response.json())
  ```
</CodeGroup>

### Standard Proxy Examples

All provider endpoints can be proxied through TrueFoundry AI Gateway for basic routing, but won't include advanced features like detailed logging, rate limiting, and budget management.

#### Custom Provider Endpoints

<Note>
  For self-hosted models, always provide the complete TrueFoundry model name since the base URL and configuration are model-specific.
</Note>

<CodeGroup>
  ```python Using Request Body Model Name lines theme={"dark"}
  import requests

  # Using a self-hosted model endpoint
  url = "{GATEWAY_BASE_URL}/proxy/{endpoint}"

  headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer your-truefoundry-api-key",
      "X-TFY-LOGGING-CONFIG": '{"enabled": true}'
  }

  payload = {
      "model": "my-custom-provider/my-model",  # TrueFoundry model name (will be auto-converted)
      "input": "Your custom payload here",
      "parameters": {
          "temperature": 0.7
      }
  }

  response = requests.post(url, headers=headers, json=payload)
  print(response.json())
  ```

  ```python Using Headers lines theme={"dark"}
  import requests

  # Using a self-hosted model endpoint
  url = "{GATEWAY_BASE_URL}/proxy/{endpoint}"

  headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer your-truefoundry-api-key",
      "X-TFY-LOGGING-CONFIG": '{"enabled": true}',
      "x-tfy-model-name": "my-custom-provider/my-model"
  }

  payload = {
      "input": "Your custom payload here",
      "parameters": {
          "temperature": 0.7
      }
  }

  response = requests.post(url, headers=headers, json=payload)
  print(response.json())
  ```

  ```python Using URL Query Parameters lines theme={"dark"}
  import requests

  # Using a self-hosted model endpoint
  url = "{GATEWAY_BASE_URL}/proxy/{endpoint}"
  params = {
      "tfyModelName": "my-custom-provider/my-model"
  }

  headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer your-truefoundry-api-key",
      "X-TFY-LOGGING-CONFIG": '{"enabled": true}'
  }

  payload = {
      "input": "Your custom payload here",
      "parameters": {
          "temperature": 0.7
      }
  }

  response = requests.post(url, headers=headers, json=payload, params=params)
  print(response.json())
  ```
</CodeGroup>

#### Image Generation

<Note>
  **For all other proxy endpoints:** You can simply pass the TrueFoundry provider name via headers (`x-tfy-provider-name`) or query parameters (`tfyProviderName`) along with the actual model ID (if required), and the proxy will route your request correctly without requiring the full TrueFoundry model name format.
</Note>

<CodeGroup>
  ```python Using Provider Name Only lines theme={"dark"}
  import openai

  client = openai.OpenAI(
      api_key="your-truefoundry-api-key",
      base_url="{GATEWAY_BASE_URL}/proxy",
      default_headers={
          "x-tfy-provider-name": "openai-main"
      }
  )

  response = client.images.generate(
      model="dall-e-3",  # Actual model ID
      prompt="a serene mountain landscape with aurora borealis",
      n=1,
      size="1024x1024"
  )

  print("Image URL:", response.data[0].url)
  ```

  ```python Using URL Query Parameters lines theme={"dark"}
  import requests

  url = "{GATEWAY_BASE_URL}/proxy/images/generations"
  params = {
      "tfyProviderName": "openai-main"
  }

  headers = {
      "Content-Type": "application/json",
      "Authorization": "Bearer your-truefoundry-api-key",
      "X-TFY-LOGGING-CONFIG": '{"enabled": true}'
  }

  payload = {
      "model": "dall-e-3",
      "prompt": "a magical forest with glowing mushrooms and fairy lights",
      "n": 1,
      "size": "1024x1024"
  }

  response = requests.post(url, headers=headers, json=payload, params=params)
  print(response.json())
  ```
</CodeGroup>

## Use Cases

This section demonstrates real-world applications and specific use cases for the Proxy API.

### Mistral AI OCR - Document Processing

Extract text from documents while maintaining structure and formatting (headers, paragraphs, lists, tables). Returns markdown format and supports multiple formats including PDFs, images (png, jpeg, avif), and office documents (pptx, docx).

Learn more about [Mistral Document AI OCR capabilities](https://docs.mistral.ai/capabilities/document_ai/basic_ocr).

```python lines theme={"dark"}
import os
import base64
import requests
import json

def encode_pdf(pdf_path):
    with open(pdf_path, "rb") as pdf_file:
        return base64.b64encode(pdf_file.read()).decode("utf-8")

pdf_path = "path-to-your-pdf"
base64_pdf = encode_pdf(pdf_path)

url = "{GATEWAY_BASE_URL}/proxy/ocr"

headers = {
    "Content-Type": "application/json",
    "Authorization": "Bearer your-truefoundry-api-key",
    "x-tfy-provider-name": "mistral-provider-name"  # TrueFoundry Mistral provider name
}

payload = {
    "model": "mistral-ocr-latest",
    "document": {
        "type": "document_url",
        "document_url": f"data:application/pdf;base64,{base64_pdf}"
    },
    "include_image_base64": True
}

response = requests.post(url, headers=headers, json=payload)

if response.status_code == 200:
    # Save output to file
    with open("ocr_output.json", "w") as f:
        json.dump(response.json(), f, indent=2)
    print("OCR output saved to ocr_output.json")
else:
    print(f"Error: {response.status_code}")
    print(response.text)
```
