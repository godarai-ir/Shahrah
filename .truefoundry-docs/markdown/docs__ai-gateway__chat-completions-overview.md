> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Chat Completions API (/chat/completions)

> Learn how to use TrueFoundry's unified Chat Completions API to interact with models from multiple providers through a consistent interface

TrueFoundry AI Gateway provides a universal API for all supported models via the standard OpenAI `/chat/completions` endpoint. This unified interface allows you to seamlessly work with models from different providers through a consistent API.

## Provider capabilities

The gateway maps the OpenAI Chat Completions contract to each provider. The table below summarizes feature support by provider for this endpoint.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider      | Stream | Non Stream | Tools                                           | JSON Mode                                       | Schema Mode                                     | Prompt Caching                                  | Reasoning                                       | Structured Output                               |
| ------------- | ------ | ---------- | ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- |
| OpenAI        | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
| Azure OpenAI  | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
| Anthropic     | ✅      | ✅          | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
| Bedrock       | ✅      | ✅          | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
| Vertex        | ✅      | ✅          | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
| Cohere        | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
| Gemini        | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
| Groq          | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
| AI21          | ✅      | ✅          | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> |
| Cerebras      | ✅      | ✅          | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
| SambaNova     | ✅      | ✅          | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |
| Perplexity-AI | ✅      | ✅          | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               |
| Together-AI   | ✅      | ✅          | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               | ✅                                               |
| xAI           | ✅      | ✅          | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               | ✅                                               |
| DeepInfra     | ✅      | ✅          | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> | ✅                                               | ✅                                               | <Icon icon="circle-minus" iconType="regular" /> |

## Contents

| Section                                         | Description                                               |
| ----------------------------------------------- | --------------------------------------------------------- |
| [Provider capabilities](#provider-capabilities) | Stream, tools, JSON mode, and related support by provider |
| [Getting Started](#getting-started)             | Basic setup and configuration                             |
| [Input Controls](#input-controls)               | System prompts and request parameters                     |
| [Working with Media](#working-with-multi-modal) | Images, audio, and video support                          |
| [Function Calling](#function-calling)           | Enabling models to invoke functions                       |
| [Thought Signatures](#thought-signatures)       | Preserving reasoning context in tool calls                |
| [Response Format](#response-format)             | Structured JSON outputs and provider support              |
| [Prompt Caching](#prompt-caching)               | Optimize API usage with caching                           |
| [Reasoning Models](#reasoning-models)           | Access model reasoning processes                          |
| [Extended Thinking](#extended-thinking)         | Thinking blocks, signatures, and multi-turn reasoning     |

## Getting Started

You can use the standard OpenAI client to send requests to the gateway:

```python lines theme={"dark"}
from openai import OpenAI

client = OpenAI(
    api_key="your_truefoundry_api_key",
    base_url="{GATEWAY_BASE_URL}"
)

response = client.chat.completions.create(
    model="openai-main/gpt-4o-mini", # this is the truefoundry model id
    messages=[{"role": "user", "content": "Hello, how are you?"}]
)

print(response.choices[0].message.content)
```

### Configuration

You will need to configure the following:

1. **base\_url**: The base URL of the TrueFoundry AI Gateway
2. **api\_key**: API key generated from [Personal Access Tokens](/docs/ai-gateway/authentication)
3. **model**: TrueFoundry model ID in the format `provider_account/model_name` (available in the LLM playground UI)

See [Integrate with code](/docs/ai-gateway/quick-start#gateway-base-url) for instructions on obtaining these values.

For using native provider SDKs (OpenAI, Google Gen AI, Anthropic, boto3), see [Native SDK Support](/docs/ai-gateway/native-sdk-support).

## Input Controls

### System Prompts

System prompts set the behavior and context for the model by defining the assistant's role, tone, and constraints:

```python lines theme={"dark"}
response = client.chat.completions.create(
    model="openai-main/gpt-4o-mini",
    messages=[
        {"role": "system", "content": "You are a helpful assistant that specializes in Python programming."},
        {"role": "user", "content": "How do I write a function to calculate factorial?"}
    ]
)
```

### Request Parameters

Fine-tune model behavior with these common parameters:

```python lines theme={"dark"}
response = client.chat.completions.create(
    model="openai-main/gpt-4o-mini",
    messages=[{"role": "user", "content": "Hello, how are you?"}],
    temperature=0.7,       # Controls randomness (0.0 to 1.0)
    max_tokens=100,        # Maximum tokens to generate
    verbosity="high",      # Constrains verbosity: low, medium, high
    top_p=0.9,             # Nucleus sampling parameter
    frequency_penalty=0.0, # Reduces repetition
    presence_penalty=0.0,  # Encourages new topics
    stop=["\n", "Human:"]  # Stop sequences
)
```

<Note>
  Some models don't support all parameters. For example, temperature is not supported by `o` series models like o3-mini.
</Note>

## Working with Multi Modal

The API supports various media types including images, audio, video and pdf.

<AccordionGroup>
  <Accordion title="Images">
    **Supported Providers:** OpenAI, Bedrock, Anthropic, Google Vertex, Google Gemini

    Send images as part of your chat completion requests using either URLs or base64 encoding:

    ### Using Image URLs

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="your_truefoundry_api_key",
        base_url="{GATEWAY_BASE_URL}"
    )

    response = client.chat.completions.create(
        model="openai-main/gpt-4o",
        messages=[
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": "What's in this image?"},
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": "https://example.com/image.jpg"
                        }
                    }
                ]
            }
        ]
    )
    ```

    ### Using Base64 Encoded Images

    ```python lines theme={"dark"}
    import base64

    def encode_image(image_path):
        with open(image_path, "rb") as image_file:
            return base64.b64encode(image_file.read()).decode('utf-8')

    response = client.chat.completions.create(
        model="openai-main/gpt-4o",
        messages=[
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": "What's in this image?"},
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": f"data:image/jpeg;base64,{encode_image('image.jpeg')}"
                        }
                    }
                ]
            }
        ]
    )
    ```
  </Accordion>

  <Accordion title="Media Resolution">
    **Supported Providers:** `OpenAI`, `Azure OpenAI`, `Google Gemini`, `Google Vertex AI`, `xAI`

    The `detail` parameter in the `image_url` object allows you to control the resolution at which images are processed. This helps balance between response quality, latency, and cost.

    **Supported Values**: `low, high, auto`

    ### Example Usage

    ```python lines theme={"dark"}
    import base64

    from openai import OpenAI

    API_KEY = "your_truefoundry_api_key"
    BASE_URL = "{GATEWAY_BASE_URL}"

    # Read and encode the image as base64
    with open("test-img.png", "rb") as image_file:
        base64_image = base64.b64encode(image_file.read()).decode('utf-8')

    client = OpenAI(
        api_key=API_KEY,
        base_url=BASE_URL
    )

    response = client.chat.completions.create(
        model="test-123/gemini-3-pro-preview",
        messages=[
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": "What is in this image?"},
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": f"data:image/png;base64,{base64_image}",
                            "detail": "low"  # Options: "low", "high", "auto"
                        }
                    }
                ]
            }
        ]
    )

    print(response.choices[0].message)
    ```

    <Note>
      For Google Gemini and Vertex AI providers, the `detail` parameter is automatically translated to the `mediaResolution` parameter:

      * `"low"` → `MEDIA_RESOLUTION_LOW` (64 tokens)
      * `"high"` → `MEDIA_RESOLUTION_HIGH` (256+ tokens with scaling)
      * `"auto"` or omitted → No explicit media resolution (model decides)
    </Note>
  </Accordion>

  <Accordion title="Audio">
    **Supported Models:** Google Gemini models (`Gemini 2.0 Flash`, etc.)

    Send audio files in supported formats (MP3, WAV, etc.). Currently supported for Google Gemini models:

    ### Using Audio URLs

    ```python lines theme={"dark"}
    response = client.chat.completions.create(
        model="internal-google/gemini-2-0-flash",
        messages=[
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": "Transcribe this audio"},
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": "https://example.com/audio.wav",
                            "mime_type": "audio/wav" # required for gemini models
                        }
                    }
                ]
            }
        ]
    )
    ```

    ### Using Base64 Encoded Audio

    ```python lines theme={"dark"}
    import base64

    def encode_audio(audio_path):
        with open(audio_path, "rb") as audio_file:
            return base64.b64encode(audio_file.read()).decode('utf-8')

    response = client.chat.completions.create(
        model="internal-google/gemini-2-0-flash",
        messages=[
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": "Transcribe this audio"},
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": f"data:audio/wav;base64,{encode_audio('audio.wav')}"
                        }
                    }
                ]
            }
        ]
    )
    ```
  </Accordion>

  <Accordion title="Video">
    **Supported Models:** Google Gemini models (`Gemini 2.0 Flash`, etc.)

    Video processing is natively supported for Google Gemini models:

    ### Using Video URLs

    ```python lines theme={"dark"}
    response = client.chat.completions.create(
        model="internal-google/gemini-2-0-flash",
        messages=[
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": "Describe what's happening in this video"},
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": "https://www.youtube.com/watch?v=example",
                            "mime_type": "video/mp4" # required for gemini models
                        }
                    }
                ]
            }
        ]
    )
    ```

    ### Using Base64 Encoded Video

    ```python lines theme={"dark"}
    import base64

    def encode_video(video_path):
        with open(video_path, "rb") as video_file:
            return base64.b64encode(video_file.read()).decode('utf-8')

    response = client.chat.completions.create(
        model="internal-google/gemini-2-0-flash",
        messages=[
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": "Describe what's happening in this video"},
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": f"data:video/mp4;base64,{encode_video('video.mp4')}",
                            "mime_type": "video/mp4" # required for gemini models
                        }
                    }
                ]
            }
        ]
    )
    ```
  </Accordion>

  <Accordion title="PDF Documents">
    **Supported Providers:** OpenAI, Bedrock, Anthropic, Google Vertex, Google Gemini

    PDF document processing allows models to analyze and extract information from PDF files:

    ### Using Base64 Encoded PDF

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="your_truefoundry_api_key",
        base_url="{GATEWAY_BASE_URL}"
    )

    import base64

    with open("sample.pdf", "rb") as file_data:
        base64_image = base64.b64encode(image_file.read()).decode('utf-8')

    response = client.chat.completions.create(
        model="tfy-ai-anthropic/claude-4-sonnet-20250514",
        messages=[
            {
                "role": "user",
                "content": [
                    {"type": "text", "text": "what's the data in the file"},
                    {
                        "type": "file",
                        "file": {
                            "filename": "sample.pdf",
                            "file_data": f"data:application/pdf;base64,{file_data}",
                        }
                    },
                ]
            }
        ]
    )

    print(response.choices[0].message.content)
    ```
  </Accordion>
</AccordionGroup>

### Vision

TrueFoundry supports vision models from all integrated providers as they become available. These models can analyze and interpret images alongside text, enabling multimodal AI applications.

| Provider     | Models                                                                                                                                    |
| :----------- | :---------------------------------------------------------------------------------------------------------------------------------------- |
| OpenAI       | `gpt-4-vision-preview, gpt-4o, gpt-4o-mini`                                                                                               |
| Anthropic    | `claude-3-sonnet, claude-3-haiku, claude-3-opus, claude-3.5-sonnet, claude-3.5-haiku, claude-4-oppus, claude-4-sonnet, claude-3-7-sonnet` |
| Gemini       | `gemini-1.0-pro-vision, gemini-1.5-flash, gemini-1.5-flash-8b, gemini-1.5-pro, gemini-2.5-pro, gemini-2.5-flash`                          |
| AWS Bedrock  | `anthropic.claude-3-5-sonnet, anthropic.claude-3-5-haiku, anthropic.claude-3-5-sonnet-20240620-v1:0`                                      |
| Azure OpenAI | `gpt-4-vision-preview, gpt-4o, gpt-4o-mini`                                                                                               |
| xAI          | `grok-2-vision-1212`                                                                                                                      |

### Using Vision Models with OpenAI SDK

```python lines theme={"dark"}
from openai import OpenAI

client = OpenAI(
    api_key="your_truefoundry_api_key",
    base_url="{GATEWAY_BASE_URL}"
)

response = client.chat.completions.create(
    model="openai-main/gpt-4o-mini",
    messages=[
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "What is in this image?"},
                {
                    "type": "image_url",
                    "image_url": {
                        "url": "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
                    }
                }
            ]
        }
    ]
)

print(response.choices[0].message)
```

## Function Calling

Function calling allows models to invoke defined functions during conversations, enabling them to perform specific actions or retrieve external information.

### Basic Usage

Define functions that the model can call:

```python lines theme={"dark"}
from openai import OpenAI
import json

client = OpenAI(
    api_key="your_truefoundry_api_key",
    base_url="{GATEWAY_BASE_URL}"
)

# Define a function
tools = [{
    "type": "function",
    "function": {
        "name": "get_weather",
        "description": "Get the current weather in a given location",
        "parameters": {
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "The city and state, e.g. San Francisco, CA"
                },
                "unit": {
                    "type": "string",
                    "enum": ["celsius", "fahrenheit"]
                }
            },
            "required": ["location"]
        }
    }
}]

# Make the request
response = client.chat.completions.create(
    model="openai-main/gpt-4o-mini",
    messages=[{"role": "user", "content": "What's the weather in New York?"}],
    tools=tools
)

# Check if the model wants to call a function
if response.choices[0].message.tool_calls:
    tool_call = response.choices[0].message.tool_calls[0]
    function_name = tool_call.function.name
    function_args = json.loads(tool_call.function.arguments)

    print(f"Function called: {function_name}")
    print(f"Arguments: {function_args}")
```

### Function Definition Reference

<AccordionGroup>
  <Accordion title="Creating Well-Structured Function Definitions">
    When defining functions, you need to provide:

    * **name**: The function name
    * **description**: What the function does
    * **parameters**: JSON Schema object describing the parameters

    ```python lines theme={"dark"}
    function_schema = {
        "name": "get_weather",
        "description": "Get current weather information",
        "parameters": {
            "type": "object",
            "properties": {
                "location": {
                    "type": "string",
                    "description": "City name"
                }
            },
            "required": ["location"]
        }
    }
    ```
  </Accordion>

  <Accordion title="Supported Parameter Types for Function Arguments">
    Functions support various parameter types:

    ```python lines theme={"dark"}
    function_schema = {
        "name": "process_data",
        "description": "Process data with various parameters",
        "parameters": {
            "type": "object",
            "properties": {
                "text": {
                    "type": "string",
                    "description": "Text to process"
                },
                "count": {
                    "type": "integer",
                    "description": "Number of items"
                },
                "confidence": {
                    "type": "number",
                    "description": "Threshold (0.0 to 1.0)"
                },
                "enabled": {
                    "type": "boolean",
                    "description": "Whether processing is enabled"
                },
                "categories": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "List of categories"
                }
            },
            "required": ["text"]
        }
    }
    ```
  </Accordion>
</AccordionGroup>

### Implementation Workflows

<AccordionGroup>
  <Accordion title="Working with Multiple Function Definitions">
    Define multiple functions for the model to choose from:

    ```python lines theme={"dark"}
    tools = [
        {
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "Get current weather information",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "location": {"type": "string"},
                        "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
                    },
                    "required": ["location"]
                }
            }
        },
        {
            "type": "function",
            "function": {
                "name": "search_web",
                "description": "Search the web for information",
                "parameters": {
                    "type": "object",
                    "properties": {
                        "query": {"type": "string"},
                        "max_results": {"type": "integer", "default": 5}
                    },
                    "required": ["query"]
                }
            }
        }
    ]
    ```
  </Accordion>

  <Accordion title="Processing and Responding to Function Calls">
    Process function calls and continue the conversation:

    ```python lines theme={"dark"}
    # Initial request
    messages = [{"role": "user", "content": "What's the weather in Tokyo?"}]
    response = client.chat.completions.create(
        model="openai-main/gpt-4o-mini",
        messages=messages,
        tools=tools
    )

    # Handle function call
    if response.choices[0].message.tool_calls:
        messages.append(response.choices[0].message)

        for tool_call in response.choices[0].message.tool_calls:
            function_name = tool_call.function.name
            function_args = json.loads(tool_call.function.arguments)

            # Execute your function (simulated here)
            if function_name == "get_weather":
                result = f"The weather in {function_args['location']} is 22°C and sunny"

            # Add the function result to the conversation
            messages.append({
                "role": "tool",
                "tool_call_id": tool_call.id,
                "content": result
            })

        # Continue the conversation
        final_response = client.chat.completions.create(
            model="openai-main/gpt-4o-mini",
            messages=messages
        )

        print(final_response.choices[0].message.content)
    ```
  </Accordion>

  <Accordion title="Controlling When and How Functions Are Called">
    Control when and how functions are called:

    ```python lines theme={"dark"}
    # Force a specific function call
    response = client.chat.completions.create(
        model="openai-main/gpt-4o-mini",
        messages=[{"role": "user", "content": "What's the weather?"}],
        tools=tools,
        tool_choice={"type": "function", "function": {"name": "get_weather"}}
    )

    # Allow automatic function calling (default)
    response = client.chat.completions.create(
        model="openai-main/gpt-4o-mini",
        messages=[{"role": "user", "content": "What's the weather?"}],
        tools=tools,
        tool_choice="auto"
    )

    # Prevent function calling
    response = client.chat.completions.create(
        model="openai-main/gpt-4o-mini",
        messages=[{"role": "user", "content": "What's the weather?"}],
        tools=tools,
        tool_choice="none"
    )

    # Force any function call
    response = client.chat.completions.create(
        model="openai-main/gpt-4o-mini",
        messages=[{"role": "user", "content": "What's the weather?"}],
        tools=tools,
        tool_choice="required"
    )
    ```
  </Accordion>
</AccordionGroup>

## Thought Signatures

Thought signatures are encrypted representations of a model's internal reasoning process that help maintain context and coherence across multi-turn interactions, particularly during function calling. When using certain Gemini 3 preview models, the API includes a `thought_signature` field in tool call responses.

```python theme={"dark"}
from openai import OpenAI
import json

client = OpenAI(
    api_key="your_truefoundry_api_key",
    base_url="{GATEWAY_BASE_URL}"
)

tools = [{
    "type": "function",
    "function": {
        "name": "get_weather",
        "description": "Get weather for a location",
        "parameters": {
            "type": "object",
            "properties": {"location": {"type": "string"}},
            "required": ["location"]
        }
    }
}]

# First call - model requests tool
response = client.chat.completions.create(
    model="vertex-main/gemini-3-pro-preview",
    messages=[{"role": "user", "content": "What's the weather in San Francisco?"}],
    tools=tools
)

message = response.choices[0].message

if message.tool_calls:
    tool_call = message.tool_calls[0]
    args = json.loads(tool_call.function.arguments)
    result = f"The weather in {args['location']} is 18°C and cloudy."

    # Convert message to dict (preserves thought_signature)
    assistant_message = message.model_dump(exclude_none=True)

    # Second call - send tool result back
    final_response = client.chat.completions.create(
        model="vertex-main/gemini-3-pro-preview",
        messages=[
            {"role": "user", "content": "What's the weather in San Francisco?"},
            assistant_message,  # Includes thought_signature
            {
                "role": "tool",
                "content": json.dumps(result),
                "tool_call_id": tool_call.id
            }
        ]
    )

    print(final_response.choices[0].message.content)
```

## Response Format

The chat completions API supports structured response formats, enabling you to receive consistent, predictable outputs in JSON format. This is useful for parsing responses programmatically.

### JSON Response Options

<AccordionGroup>
  <Accordion title="Basic JSON Mode: Getting Valid JSON Without Structure Constraints">
    JSON mode ensures the model's output is valid JSON without enforcing a specific structure:

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="your_truefoundry_api_key",
        base_url="{GATEWAY_BASE_URL}"
    )

    response = client.chat.completions.create(
        model="openai-main/gpt-4o",
        messages=[
            {"role": "system", "content": "You are a helpful assistant designed to output JSON."},
            {"role": "user", "content": "Extract information about the 2020 World Series winner"}
        ],
        response_format={"type": "json_object"}
    )

    print(response.choices[0].message.content)
    ```

    **Output:**

    ```json lines theme={"dark"}
    {
      "team": "Los Angeles Dodgers",
      "year": 2020,
      "opponent": "Tampa Bay Rays",
      "games_played": 6,
      "series_result": "4-2"
    }
    ```
  </Accordion>

  <Accordion title="JSON Schema Mode: Enforcing Specific Data Structures">
    JSON Schema mode provides strict structure validation using predefined schemas:

    ```python lines theme={"dark"}
    from openai import OpenAI
    import json

    client = OpenAI(
        api_key="your_truefoundry_api_key",
        base_url="{GATEWAY_BASE_URL}"
    )

    # Define JSON schema
    user_info_schema = {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer", "minimum": 0},
            "occupation": {"type": "string"},
            "location": {"type": "string"},
            "skills": {
                "type": "array",
                "items": {"type": "string"}
            }
        },
        "required": ["name", "age", "occupation", "location", "skills"],
        "additionalProperties": False
    }

    response = client.chat.completions.create(
        model="openai-main/gpt-4o",
        messages=[
            {
                "role": "system",
                "content": "Extract user information and respond according to the provided JSON schema."
            },
            {
                "role": "user",
                "content": "My name is Sarah Johnson, I'm 28 years old, and I work as a data scientist in New York. I'm skilled in Python, SQL, and machine learning."
            }
        ],
        response_format={
            "type": "json_schema",
            "json_schema": {
                "name": "user_info",
                "schema": user_info_schema,
                "strict": True
            }
        }
    )

    # Parse response
    result = json.loads(response.choices[0].message.content)
    ```

    <Note>
      When using JSON schema with strict mode set to true, all properties defined in the schema must be included in the required array. If any property is defined but not marked as required, the API will return a 400 Bad Request Error.
    </Note>
  </Accordion>
</AccordionGroup>

### Advanced Schema Integration

<AccordionGroup>
  <Accordion title="Python Type Validation with Pydantic Models">
    Pydantic provides automatic validation, serialization, and type hints for structured data:

    ```python lines theme={"dark"}
    from openai import OpenAI
    from pydantic import BaseModel, Field
    from typing import List

    client = OpenAI(
        api_key="your_truefoundry_api_key",
        base_url="{GATEWAY_BASE_URL}"
    )

    # Define Pydantic model
    class UserInfo(BaseModel):
        name: str = Field(description="Full name of the user")
        age: int = Field(ge=0, description="Age in years")
        occupation: str = Field(description="Job title or profession")
        location: str = Field(description="City or location")
        skills: List[str] = Field(description="List of professional skills")

        class Config:
            extra = "forbid"  # Prevent additional fields

    response = client.chat.completions.create(
        model="openai-main/gpt-4o",
        messages=[
            {
                "role": "system",
                "content": "Extract user information and respond according to the provided schema."
            },
            {
                "role": "user",
                "content": "Hi, I'm Mike Chen, a 32-year-old software architect from Seattle. I specialize in cloud computing, microservices, and Kubernetes."
            }
        ],
        response_format={
            "type": "json_schema",
            "json_schema": {
                "name": "user_info",
                "schema": UserInfo.model_json_schema(),
                "strict": True
            }
        }
    )

    # Parse and validate with Pydantic
    user_data = UserInfo.model_validate_json(response.choices[0].message.content)
    ```

    <Note>
      When using OpenAI models with Pydantic Models, there should not be any optional fields in the pydantic model when strict mode is true. This is because the corresponding JSON schema will have missing fields in the "required" section.
    </Note>
  </Accordion>

  <Accordion title="Streamlined Pydantic Integration with OpenAI's Beta Parse API">
    The beta parse client provides the most streamlined approach for Pydantic integration:

    ```python lines theme={"dark"}
    from openai import OpenAI
    from pydantic import BaseModel, Field
    from typing import List, Optional

    class UserInfo(BaseModel):
        name: str = Field(description="Full name of the user")
        age: int = Field(ge=0, description="Age in years")
        occupation: str = Field(description="Job title or profession")
        location: Optional[str] = Field(None, description="City or location")
        skills: List[str] = Field(default=[], description="List of professional skills")

    client = OpenAI(
        api_key="your_truefoundry_api_key",
        base_url="{GATEWAY_BASE_URL}"
    )

    completion = client.beta.chat.completions.parse(
        model="openai-main/gpt-4o",
        messages=[
            {
                "role": "system",
                "content": "Extract user information from the provided text."
            },
            {
                "role": "user",
                "content": "Hello, I'm Alex Rodriguez, a 29-year-old product manager from Austin. I have experience in agile methodologies, data analysis, and team leadership."
            }
        ],
        response_format=UserInfo,
    )

    user_result = completion.choices[0].message.parsed
    ```

    This approach allows for optional fields in your Pydantic model and provides a cleaner API for structured responses.
  </Accordion>
</AccordionGroup>

### Provider support for response schema

You can use `response_format` with any provider. The Gateway either uses the provider’s native structured output or converts your schema into a tool the model must call and then puts the result in `message.content`.

| Provider                                                                     | Support                                                                                                                                     |
| ---------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| **OpenAI**                                                                   | Native for `json_object` and for `json_schema` on supported models (e.g. gpt-4o, gpt-5, gpt-4.1, o3, o4). Other models use tool conversion. |
| **Azure OpenAI**                                                             | Same as OpenAI.                                                                                                                             |
| **Anthropic**                                                                | Native for Claude 4.5/4.6 with `json_schema`. Other models use tool conversion.                                                             |
| **Google Gemini, Google Vertex**                                             | Native when the request has no tools; otherwise tool conversion.                                                                            |
| **All others** (Bedrock, Cohere, Mistral, OpenRouter, Groq, xAI, vLLM, etc.) | Tool conversion only. The Gateway turns your schema into a required tool and extracts the result into `message.content`.                    |

<Note>
  **Anthropic and JSON schema constraints:** The code examples in this doc use Pydantic's `ge=0` for fields such as `age`. Anthropic's API does not support these constraint parameters in the schema. If you use structured output with Anthropic models, omit `ge`, `le` and similar numeric/string constraints from your schema (or use a schema without them). The code will work with Anthropic once those constraints are removed.
</Note>

## Prompt Caching

Prompt caching reduces processing time and costs by reusing previously computed prefixes. Some providers handle this automatically, while others require you to mark which content to cache. The gateway handles both: you can include `cache_control` in your requests and the gateway will forward it to providers that support it and **automatically strip it** for providers that don't.

<Info>
  You can safely include `cache_control` in requests regardless of the target provider. The gateway ensures it never causes errors.
</Info>

### Provider Support

**Explicit caching** (you mark content with `cache_control`):

* **Anthropic** (direct, Vertex AI, Azure AI Foundry): `cache_control` is forwarded to the provider
* **AWS Bedrock**: `cache_control` is translated into Bedrock's native `cachePoint` format

**Automatic caching** (no markup needed, `cache_control` is stripped):

* **OpenAI / Azure OpenAI**: also supports `prompt_cache_key` for better cache hit rates
* **Google Gemini / Vertex AI (Gemini)**
* **Groq**, **xAI**, and all other providers

### Usage

Add `cache_control` to any content block, system block, or tool definition you want to cache. The format is:

```json theme={"dark"}
{
  "cache_control": {"type": "ephemeral", "ttl": "5m"}
}
```

The `ttl` field is optional (only Anthropic supports it). The `type` must be `"ephemeral"`.

<AccordionGroup>
  <Accordion title="Anthropic / Bedrock">
    For Anthropic (direct, Vertex AI, Azure AI Foundry) and AWS Bedrock, you must explicitly mark content to cache. The gateway forwards `cache_control` to Anthropic as is, and translates it into Bedrock's native `cachePoint` format automatically.

    <AccordionGroup>
      <Accordion title="Caching System Prompts">
        ```python lines theme={"dark"}
        from openai import OpenAI

        client = OpenAI(
            api_key="your_truefoundry_api_key",
            base_url="{GATEWAY_BASE_URL}"
        )

        response = client.chat.completions.create(
            model="anthropic-main/claude-3-7-sonnet-latest",
            messages=[
                {
                    "role": "system",
                    "content": [
                        {
                            "type": "text",
                            "text": "<LARGE_SYSTEM_PROMPT_TO_CACHE>",
                            "cache_control": {"type": "ephemeral"},
                        },
                    ],
                },
                {
                    "role": "user",
                    "content": "Your question here",
                },
            ]
        )
        ```
      </Accordion>

      <Accordion title="Caching Tool Definitions">
        ```python lines theme={"dark"}
        response = client.chat.completions.create(
            model="anthropic-main/claude-4-sonnet",
            messages=[{"role": "user", "content": "What's the weather?"}],
            tools=[
                {
                    "type": "function",
                    "function": {
                        "name": "get_weather",
                        "description": "Get current weather",
                        "parameters": {
                            "type": "object",
                            "properties": {"location": {"type": "string"}},
                            "required": ["location"]
                        }
                    },
                    "cache_control": {"type": "ephemeral"}
                }
            ]
        )
        ```
      </Accordion>
    </AccordionGroup>

    #### Minimum Cacheable Length

    **4096 tokens:** Claude Mythos Preview, Opus 4.7, Opus 4.6, Opus 4.5, Haiku 4.5

    **2048 tokens:** Sonnet 4.6, Haiku 3.5, Haiku 3

    **1024 tokens:** Sonnet 4.5, Opus 4.1, Opus 4, Sonnet 4, Sonnet 3.7

    <Note>
      For Amazon Titan models on Bedrock, `cache_control` on tool definitions is automatically skipped since these models do not support cache points on tools.
    </Note>
  </Accordion>

  <Accordion title="OpenAI / Azure OpenAI">
    OpenAI and Azure OpenAI cache prompts automatically. No `cache_control` markup is needed. Any `cache_control` in the request is stripped by the gateway.

    You can optionally pass `prompt_cache_key` to improve cache hit rates across requests with shared prefixes:

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="your_truefoundry_api_key",
        base_url="{GATEWAY_BASE_URL}"
    )

    response = client.chat.completions.create(
        model="openai-main/gpt-4o",
        messages=[{"role": "user", "content": "Your prompt here"}],
        prompt_cache_key="optional-custom-key"
    )
    ```

    <Info>
      `prompt_cache_key` is only available for OpenAI and Azure OpenAI.
    </Info>
  </Accordion>

  <Accordion title="Gemini / Groq / xAI">
    These providers handle caching automatically. No configuration is needed. Any `cache_control` in the request is stripped by the gateway before forwarding.

    Cached token counts are still reported in the response usage when the provider returns them.
  </Accordion>
</AccordionGroup>

### Cache Usage in Responses

When caching is active, the response `usage` object includes cached token counts:

```json theme={"dark"}
{
  "usage": {
    "prompt_tokens": 1500,
    "completion_tokens": 200,
    "total_tokens": 1700,
    "prompt_tokens_details": {
      "cached_tokens": 1200
    },
    "cache_read_input_tokens": 1200,
    "cache_creation_input_tokens": 300
  }
}
```

* `prompt_tokens_details.cached_tokens`: tokens served from cache. Available across all providers that report cache usage.
* `cache_read_input_tokens` / `cache_creation_input_tokens`: Anthropic style fields, present for Anthropic, Bedrock, and Groq when values are non zero.

## Reasoning Models

TrueFoundry AI Gateway provides access to model reasoning processes through thinking/reasoning tokens, available for models from multiple providers including `Anthropic`,`OpenAI`,`Azure OpenAI`,`Groq`, `xAI` and `Vertex`.

These models expose their internal reasoning process, allowing you to see how they arrive at conclusions. The thinking/reasoning tokens provide step-by-step insights into the model's cognitive process.

### Supported Reasoning Models

<AccordionGroup>
  <Accordion title="OpenAI">
    Supported models: `o4-mini`, `o4-preview`, `o3` model family, `o1` model family, `gpt-5-mini`, `gpt-5-nano`, `gpt-5`

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="TFY_API_KEY",
        base_url="{GATEWAY_BASE_URL}"
    )

    response = client.chat.completions.create(
        model="openai-main/o4-mini",
        messages=[
            {"role": "user", "content": "How to compute 3^3^3?"}
        ],
        reasoning_effort="high",  # Options: "high", "medium", "low"
        max_tokens=8000
    )

    print(response.choices[0].message.content)
    ```
  </Accordion>

  <Accordion title="Azure OpenAI">
    Supported models: `gpt-5`, `gpt-5-mini`, `gpt-5-nano`, `o3-pro`, `codex-mini`, `o4-mini`, `o3`, `o3-mini`, `o1`, `o1-mini`

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="TFY_API_KEY",
        base_url="{GATEWAY_BASE_URL}"
    )

    response = client.chat.completions.create(
        model="azure-openai-main/o3-mini",
        messages=[
            {"role": "user", "content": "How to compute 3^3^3?"}
        ],
        reasoning_effort="high",  # Options: "high", "medium", "low"
        max_tokens=8000
    )

    print(response.choices[0].message.content)
    ```
  </Accordion>

  <Accordion title="Anthropic">
    Supported models: `Claude Opus 4.1` (claude-opus-4-1-20250805), `Claude Opus 4` (claude-opus-4-20250514), `Claude Sonnet 4` (claude-sonnet-4-20250514), `Claude Sonnet 3.7` (claude-3-7-sonnet-20250219) \
    via `Anthropic`, `AWS Bedrock`, and `Google Vertex AI`

    ### Using OpenAI SDK

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="TFY_API_KEY",
        base_url="{GATEWAY_BASE_URL}"
    )

    response = client.chat.completions.create(
        model="anthropic-main/claude-3-7-sonnet",
        messages=[
            {"role": "user", "content": "How to compute 3^3^3?"}
        ],
        reasoning_effort="high",  # Options: "high", "medium", "low", "none"
        max_tokens=8000
    )

    print(response.choices[0].message.content)
    ```

    <Note>
      For Anthropic models (from Anthropic, Google Vertex AI, AWS Bedrock), TrueFoundry automatically translates the `reasoning_effort` parameter into Anthropic's native `thinking` parameter format since Anthropic doesn't support the `reasoning_effort` parameter directly.

      The translation uses the `max_tokens` parameter with the following ratios:

      * `none`: 0% of max\_tokens
      * `low`: 30% of max\_tokens
      * `medium`: 60% of max\_tokens
      * `high`: 90% of max\_tokens
    </Note>

    ### Using Direct API Calls with Native `thinking` Parameter

    For more precise control with Anthropic models, you can use the native `thinking` parameter directly:

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="TFY_API_KEY",
        base_url="{GATEWAY_BASE_URL}"
    )

    response = client.chat.completions.create(
        model="anthropic-main/claude-3-7-sonnet",
        messages=[
            {"role": "user", "content": "How to compute 3^3^3?"}
        ],
        max_tokens=8000,
        extra_body={
            "thinking": {
                "type": "enabled",
                "budget_tokens": 1000
            }
        }
    )

    print(response.choices[0].message.content)
    ```
  </Accordion>

  <Accordion title="Groq">
    Supported models: `OpenAI GPT-OSS 20B` (openai/gpt-oss-20b), `OpenAI GPT-OSS 120B` (openai/gpt-oss-120b), `Qwen 3 32B` (qwen/qwen3-32b), `DeepSeek R1 Distil Llama 70B` (deepseek-r1-distill-llama-70b)

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="TFY_API_KEY",
        base_url="{GATEWAY_BASE_URL}"
    )

    response = client.chat.completions.create(
        model="groq-main/deepseek-r1-distill-llama-70b",
        messages=[
            {"role": "user", "content": "How to compute 3^3^3?"}
        ],
        reasoning_effort="high",  # Options: "high", "medium", "low"
        max_tokens=8000
    )

    print(response.choices[0].message.content)
    ```
  </Accordion>

  <Accordion title="xAI">
    Supported models: `grok-3-mini` (with `reasoning_effort` parameter), `grok-4-0709`, `grok-4-1-fast-reasoning`, `grok-4-fast-reasoning` (reasoning built-in)

    For `grok-3-mini`, you can use the `reasoning_effort` parameter to control reasoning depth. Other Grok models like `grok-4-0709` have reasoning capabilities built-in but do not support the `reasoning_effort` parameter.

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="TFY_API_KEY",
        base_url="{GATEWAY_BASE_URL}"
    )

    # For grok-3-mini with reasoning_effort parameter
    response = client.chat.completions.create(
        model="xai-main/grok-3-mini",
        messages=[
            {"role": "user", "content": "How to compute 3^3^3?"}
        ],
        reasoning_effort="high",  # Options: "high", "low" (only for grok-3-mini)
        max_tokens=8000
    )

    print(response.choices[0].message.content)
    ```

    <Note>
      The `reasoning_effort` parameter is only supported for `grok-3-mini`. For other Grok models like `grok-4-0709` and `grok-4-1-fast-reasoning`, reasoning is built-in and the `reasoning_effort` parameter should not be used. Reasoning tokens are included in the usage metrics for all reasoning-capable models.

      **Parameter Restrictions**: Reasoning models (like `grok-4-0709` and `grok-4-1-fast-reasoning`) do not support `presence_penalty`, `frequency_penalty`, or `stop` parameters. Using these parameters with reasoning models will result in an error.
    </Note>
  </Accordion>

  <Accordion title="Gemini">
    Supported models: All `Gemini 2.5 Series` Models.

    These models can be accessed from `Google Vertex` or `Google Gemini` Providers

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="TFY_API_KEY",
        base_url="{GATEWAY_BASE_URL}"
    )

    response = client.chat.completions.create(
        model="vertex-main/gemini-2-5-pro",
        messages=[
            {"role": "user", "content": "How to compute 3^3^3?"}
        ],
        reasoning_effort="high",  # Options: "high", "medium", "low", "none"
        max_tokens=8000
    )

    print(response.choices[0].message.content)
    ```

    <Note>
      For Gemini models (from Anthropic, Google Vertex AI, AWS Bedrock), TrueFoundry automatically translates the `reasoning_effort` parameter into Gemini's native `thinking` parameter format since Gemini doesn't support the `reasoning_effort` parameter directly.

      The translation uses the `max_tokens` parameter with the following ratios:

      * `none`: 0% of max\_tokens
      * `low`: 30% of max\_tokens
      * `medium`: 60% of max\_tokens
      * `high`: 90% of max\_tokens

      Note: Gemini 2.5 Pro and 2.5 Flash comes with reasoning on by default.
    </Note>

    ### Using Direct API Calls with Native `thinking` Parameter

    For more precise control with Gemini models, you can use the native `thinking` parameter directly:

    ```python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="TFY_API_KEY",
        base_url="{GATEWAY_BASE_URL}"
    )

    response = client.chat.completions.create(
        model="vertex-main/gemini-2-5-pro",
        messages=[
            {"role": "user", "content": "How to compute 3^3^3?"}
        ],
        max_tokens=8000,
        extra_body={
            "thinking": {
                "type": "enabled",
                "budget_tokens": 1000
            }
        }
    )

    print(response.choices[0].message.content)
    ```
  </Accordion>
</AccordionGroup>

### Response Format

When reasoning tokens are enabled, the response includes both thinking and content sections:

```json lines theme={"dark"}
{
  "id": "1742890579083",
  "object": "chat.completion",
  "created": 1742890579,
  "model": "",
  "provider": "aws",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "To compute 3^3^3, you need to understand that exponentiation is evaluated from right to left (right-associative). So:\n\n3^3^3 = 3^(3^3), not (3^3)^3\n\nHere's how to calculate it step by step:\n\n**Step 1:** Calculate the rightmost exponent first\n- 3^3 = 27\n\n**Step 2:** Now calculate 3^27\n- 3^27 = 7,625,597,484,987\n\nTherefore: **3^3^3 = 7,625,597,484,987**\n\n---\n\n**Note:** If you meant (3^3)^3 instead, that would be:\n- (3^3)^3 = 27^3 = 19,683\n\nThis is much smaller than 3^3^3 because:\n- (3^3)^3 = 3^(3×3) = 3^9\n- While 3^3^3 = 3^27\n\nThe difference between 3^9 and 3^27 is enormous!",
        "reasoning_content": "The user is asking how to compute 3^3^3. This is a question about exponentiation and specifically about the order of operations when dealing with repeated exponentiation (also known as tetration or power towers).\n\nThe key thing to understand here is that exponentiation is right-associative, meaning 3^3^3 = 3^(3^3), not (3^3)^3.\n\nSo we need to:\n1. First compute 3^3 = 27\n2. Then compute 3^27\n\nLet me work through this:\n- 3^3 = 27\n- 3^27 = 3^27\n\nNow 3^27 is a large number. Let me think about how to compute it:\n3^27 = 3^(3×9) = (3^3)^9 = 27^9\n\nOr we could compute it directly:\n3^1 = 3\n3^2 = 9\n3^3 = 27\n3^4 = 81\n3^5 = 243\n3^6 = 729\n3^7 = 2,187\n3^8 = 6,561\n3^9 = 19,683\n3^10 = 59,049\n...\n\nActually, let me just state that 3^27 = 7,625,597,484,987\n\nSo 3^3^3 = 3^(3^3) = 3^27 = 7,625,597,484,987"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 45,
    "completion_tokens": 180,
    "total_tokens": 225
  }
}
```

### Streaming with Reasoning Tokens

For streaming responses, the thinking section is always sent before the content section.

## Extended Thinking

Extended thinking exposes the model's step-by-step reasoning as structured `thinking_blocks`. Unlike the plain-text `reasoning_content` field, thinking blocks carry cryptographic **signatures** — required to continue a reasoning chain across multiple turns.

### Provider Support

| Provider                  | Models                                  | Notes                                                |
| ------------------------- | --------------------------------------- | ---------------------------------------------------- |
| Anthropic                 | Claude 3.7, Claude 4, Claude 4.5 series | Claude 4.5+ uses adaptive thinking (no fixed budget) |
| AWS Bedrock               | All Claude models on Bedrock            | Minimum 1024 budget tokens enforced                  |
| Google Gemini / Vertex AI | Gemini 2.5, Gemini 3 series             | Thinking on by default for Gemini 2.5 Pro and Flash  |

### Response Fields

| Field                       | Description                                    |
| --------------------------- | ---------------------------------------------- |
| `message.reasoning_content` | Plain-text summary of the model's reasoning    |
| `message.thinking_blocks`   | Array of `{type, thinking, signature}` objects |

```json theme={"dark"}
{
  "message": {
    "content": "3^3^3 = 7,625,597,484,987",
    "reasoning_content": "Exponentiation is right-associative: 3^3 = 27, then 3^27...",
    "thinking_blocks": [
      {
        "type": "thinking",
        "thinking": "Exponentiation is right-associative: 3^3 = 27, then 3^27...",
        "signature": "ErUBCkQIARgCIkC..."
      }
    ]
  }
}
```

### Basic Usage

```python lines theme={"dark"}
from openai import OpenAI

client = OpenAI(api_key="TFY_API_KEY", base_url="{GATEWAY_BASE_URL}")

response = client.chat.completions.create(
    model="anthropic-main/claude-opus-4-1-20250805",
    messages=[{"role": "user", "content": "How to compute 3^3^3?"}],
    reasoning_effort="high",  # "high" | "medium" | "low" | "none"
    max_tokens=8000
)

msg = response.choices[0].message
print(msg.content)            # Final answer
print(msg.reasoning_content)  # Plain text reasoning
print(msg.thinking_blocks)    # Structured blocks with signatures
```

### Streaming

Use the helper below to accumulate a stream into `content`, `thinking_blocks`, and `tool_calls` — ready to pass directly into the next turn.

```python lines theme={"dark"}
def accumulate_stream(stream):
    content = ""
    reasoning_content = ""
    reasoning_signature = None
    thinking_blocks_direct = []
    tool_calls_buf: dict[int, dict] = {}

    for chunk in stream:
        delta = chunk.choices[0].delta
        extra = delta.model_extra or {}

        if delta.content:
            content += delta.content

        # Anthropic / Bedrock
        if extra.get("reasoning_content"):
            reasoning_content += extra["reasoning_content"]
        if extra.get("reasoning_signature"):
            reasoning_signature = extra["reasoning_signature"]

        # Gemini / Vertex AI
        for tb in (extra.get("thinking_blocks") or []):
            thinking_blocks_direct.append(tb)

        # Tool calls — accumulated by index
        for tc in (delta.tool_calls or []):
            slot = tool_calls_buf.setdefault(tc.index, {
                "id": "", "type": "function",
                "function": {"name": "", "arguments": ""},
            })
            if tc.id:
                slot["id"] += tc.id
            if tc.function:
                slot["function"]["name"]      += tc.function.name      or ""
                slot["function"]["arguments"] += tc.function.arguments or ""

    # Normalize both provider formats into a unified thinking_blocks shape
    if thinking_blocks_direct:
        thinking_blocks = thinking_blocks_direct
    elif reasoning_content or reasoning_signature:
        thinking_blocks = [{
            "type": "thinking",
            "thinking": reasoning_content,
            **({"signature": reasoning_signature} if reasoning_signature else {}),
        }]
    else:
        thinking_blocks = []

    return content, thinking_blocks, list(tool_calls_buf.values())


stream = client.chat.completions.create(
    model="anthropic-main/claude-opus-4-1-20250805",
    messages=[{"role": "user", "content": "How to compute 3^3^3?"}],
    reasoning_effort="high",
    max_tokens=8000,
    stream=True
)

content, thinking_blocks, tool_calls = accumulate_stream(stream)

assistant_message = {
    "role": "assistant",
    "content": content,
    "thinking_blocks": thinking_blocks,
    **({"tool_calls": tool_calls} if tool_calls else {}),
}
```

### Multi-Turn Conversations

Use `model_dump(exclude_none=True)` on the assistant message — it captures `content`, `tool_calls`, and `thinking_blocks` in one shot, so you don't need to construct the dict manually.

```python lines theme={"dark"}
# Turn 1
response = client.chat.completions.create(
    model="anthropic-main/claude-opus-4-1-20250805",
    messages=[{"role": "user", "content": "What is 3^3^3?"}],
    reasoning_effort="high",
    max_tokens=8000
)

# Serialize the full assistant message (preserves thinking_blocks + signatures)
assistant_message = response.choices[0].message.model_dump(exclude_none=True)

# Turn 2 — pass the serialized message back as-is
response2 = client.chat.completions.create(
    model="anthropic-main/claude-opus-4-1-20250805",
    messages=[
        {"role": "user", "content": "What is 3^3^3?"},
        assistant_message,
        {"role": "user", "content": "Now explain why exponentiation is right-associative."}
    ],
    reasoning_effort="high",
    max_tokens=8000
)
```

<Warning>
  Always echo `thinking_blocks` exactly as returned. Blocks with missing or modified `signature` fields are rejected by the provider.
</Warning>

### Multi-Turn with Tool Calls

When thinking is enabled, Anthropic and Bedrock require the assistant message to include `thinking_blocks` alongside `tool_calls`. Use `model_dump(exclude_none=True)` — it captures both in one step.

```python lines theme={"dark"}
import json
from openai import OpenAI

client = OpenAI(api_key="TFY_API_KEY", base_url="{GATEWAY_BASE_URL}")

tools = [{
    "type": "function",
    "function": {
        "name": "get_weather",
        "description": "Get weather for a city",
        "parameters": {
            "type": "object",
            "properties": {"city": {"type": "string"}},
            "required": ["city"]
        }
    }
}]

messages = [{"role": "user", "content": "What's the weather in Tokyo?"}]

# Turn 1 — model responds with thinking + tool call
response = client.chat.completions.create(
    model="anthropic-main/claude-opus-4-1-20250805",
    messages=messages,
    tools=tools,
    reasoning_effort="high",
    max_tokens=8000
)

msg = response.choices[0].message

# Append assistant message — model_dump captures tool_calls + thinking_blocks together
messages.append(msg.model_dump(exclude_none=True))

# Execute the tool and append the result
for tool_call in msg.tool_calls:
    args = json.loads(tool_call.function.arguments)
    result = f"Sunny, 24°C in {args['city']}"
    messages.append({
        "role": "tool",
        "tool_call_id": tool_call.id,
        "content": result
    })

# Turn 2 — model summarizes with full context
response2 = client.chat.completions.create(
    model="anthropic-main/claude-opus-4-1-20250805",
    messages=messages,
    tools=tools,
    reasoning_effort="high",
    max_tokens=8000
)

print(response2.choices[0].message.content)
```

## Grounding with Google Search

Google Gemini support **grounding with Google Search**, which allows the model to augment its responses with real-time web results.

When grounding is enabled, the model can call a search tool during generation to retrieve up-to-date information and incorporate it into the final answer.

```python theme={"dark"}
from openai import OpenAI

client = OpenAI(
    api_key="TFY_API_KEY",
    base_url="{GATEWAY_BASE_URL}"
)

response = client.chat.completions.create(
    model="tfy-ai-gemini/gemini-2-5-pro", # tfy gemini model name
    messages=[{"role": "user", "content": "what date and time is right now?"}],
    tools=[{
        "type": "function",
        "function": {
            "name": "google_search",
        }
    }]
    
)

print(response.choices[0].message)
```

## API Reference

For detailed API specifications, parameters, and response schemas, see the Chat Completions API Reference in the API Reference tab.
