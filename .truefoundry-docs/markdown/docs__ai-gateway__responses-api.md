> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Responses API (/responses)

> Learn how to use OpenAI's Responses API through TrueFoundry Gateway for creating, retrieving, and managing text and multimodal completions.

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider     | Model Response                                              |
| ------------ | ----------------------------------------------------------- |
| OpenAI       | ✅                                                           |
| Azure OpenAI | ✅                                                           |
| Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
| Bedrock      | <Icon icon="circle-minus" iconType="regular" />             |
| Vertex       | <Icon icon="circle-minus" iconType="regular" />             |
| Cohere       | <Icon icon="circle-minus" iconType="regular" />             |
| Gemini       | <Icon icon="circle-minus" iconType="regular" />             |
| Groq         | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Cerebras     | <Icon icon="circle-minus" iconType="regular" />             |
| Together-AI  | <Icon icon="circle-minus" iconType="regular" />             |
| xAI          | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra    | <Icon icon="circle-minus" iconType="regular" />             |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

This guide explains how to use the OpenAI / Azure OpenAI client to interact with TrueFoundry's responses endpoint for making inference requests.

## Authentication

You'll need a TrueFoundry API key to authenticate your requests. You can find authentication details from [here](/docs/ai-gateway/authentication).

You'll also need to set the `x-tfy-provider-name` header to the name of the provider integration you're using.

For example, if you're using an OpenAI provider integration named `my-openai-provider`, you'll set the `x-tfy-provider-name` header to `my-openai-provider`.

```python lines theme={"dark"}
from openai import OpenAI

client = OpenAI(
    base_url="{GATEWAY_BASE_URL}",
    api_key="your_truefoundry_api_key",
    default_headers={"x-tfy-provider-name": "my-openai-provider"}
)
```

## Text Completion

To get a text completion response:

```python lines theme={"dark"}
response = client.responses.create(
    model="your_truefoundry_model_name", //  OpenAI or Azure OpenAI tfy model name
    input=[{"role": "user", "content": "Your prompt here"}]
)

print(response)
```

### Image Inputs

For tasks involving images:

```python lines theme={"dark"}
response = client.responses.create(
    model="your_truefoundry_model_name", // OpenAI or Azure OpenAI tfy model name
    input=[
        {"role": "user", "content": "Your prompt here"},
        {
            "role": "user",
            "content": [
                {
                    "type": "input_image",
                    "image_url": "your_image_url"
                }
            ]
        }
    ]
)
response_id = response.id
```

### Exprected Response

```json lines theme={"dark"}
{
  "id": "resp_6847fa2670f8819887e2d14c08bdc5a305d8dc9b22b6f0dd",
  "created_at": 1749547558,
  "error": null,
  "incomplete_details": null,
  "instructions": null,
  "metadata": {},
  "model": "gpt-4o-mini-2024-07-18",
  "object": "response",
  "output": [
    {
      "id": "msg_6847fa35e8c48198a5120ee3d38c353105d8dc9b22b6f0dd",
      "content": [
        {
          "annotations": [],
          "text": "It seems you might be looking to start a discussion or ask a question! How can I assist you today?",
          "type": "output_text"
        }
      ],
      "role": "assistant",
      "status": "completed",
      "type": "message"
    }
  ],
  "parallel_tool_calls": true,
  "temperature": 1,
  "tool_choice": "auto",
  "tools": [],
  "top_p": 1,
  "max_output_tokens": null,
  "previous_response_id": null,
  "reasoning": {
    "effort": null,
    "generate_summary": null,
    "summary": null
  },
  "status": "completed",
  "text": {
    "format": {
      "type": "text"
    }
  },
  "truncation": "disabled",
  "usage": {
    "input_tokens": 10,
    "input_tokens_details": {
      "cached_tokens": 0
    },
    "output_tokens": 23,
    "output_tokens_details": {
      "reasoning_tokens": 0
    },
    "total_tokens": 33
  },
  "user": null,
  "background": false,
  "service_tier": "default",
  "store": true,
  "provider": "openai"
}
```

## Managing Responses

### Retrieve Response

Retrieve a response by id.

```python lines theme={"dark"}
response = client.responses.retrieve(
	response_id=response_id
)

print(response)
```

### Expected Output

```json lines theme={"dark"}
{
  "id": "resp_6847fa2670f8819887e2d14c08bdc5a305d8dc9b22b6f0dd",
  "created_at": 1749547558,
  "error": null,
  "incomplete_details": null,
  "instructions": null,
  "metadata": {},
  "model": "gpt-4o-mini-2024-07-18",
  "object": "response",
  "output": [
    {
      "id": "msg_6847fa35e8c48198a5120ee3d38c353105d8dc9b22b6f0dd",
      "content": [
        {
          "annotations": [],
          "text": "It seems you might be looking to start a discussion or ask a question! How can I assist you today?",
          "type": "output_text"
        }
      ],
      "role": "assistant",
      "status": "completed",
      "type": "message"
    }
  ],
  "parallel_tool_calls": true,
  "temperature": 1,
  "tool_choice": "auto",
  "tools": [],
  "top_p": 1,
  "max_output_tokens": null,
  "previous_response_id": null,
  "reasoning": {
    "effort": null,
    "generate_summary": null,
    "summary": null
  },
  "status": "completed",
  "text": {
    "format": {
      "type": "text"
    }
  },
  "truncation": "disabled",
  "usage": {
    "input_tokens": 10,
    "input_tokens_details": {
      "cached_tokens": 0
    },
    "output_tokens": 23,
    "output_tokens_details": {
      "reasoning_tokens": 0
    },
    "total_tokens": 33
  },
  "user": null,
  "background": false,
  "service_tier": "default",
  "store": true,
  "provider": "openai"
}
```

### Delete Response

Delete a response permanently

```python lines theme={"dark"}
delete_result = client.responses.delete(
	response_id=response_id
)
```
