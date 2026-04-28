> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Anthropic

> Add and configure Anthropic's Claude models in TrueFoundry's AI Gateway

### Adding Models

This section explains the steps to add Anthropic models and configure the required access controls.

<Steps>
  <Step title="Navigate to Anthropic Models in AI Gateway">
    From the TrueFoundry dashboard, navigate to `AI Gateway` > `Models` and select `Anthropic`.

    <Frame caption="Navigate to Anthropic Models">
      <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/anthropic-0.jpeg?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=8d3a3394140dad8492d2e866d2425c23" alt="Navigating to Anthropic Provider Account in AI Gateway" width="1280" height="650" data-path="images/anthropic-0.jpeg" />
    </Frame>
  </Step>

  <Step title="Add Anthropic Account Details">
    Click `Add Anthropic Account`. Give a unique name to your Anthropic account and complete the form with your Anthropic authentication details (API Key). Add collaborators to your account, this will give access to the account to other users/teams. Learn more about access control [here](/docs/ai-gateway/gateway-access-control). For [Claude Code Max](/docs/ai-gateway/claude-code-max), leave the API key empty.

    <Frame caption="Anthropic Model Account Form">
      <img src="https://mintcdn.com/truefoundry/JFTbQOWMkMfvFjDC/images/anthropic-1.jpeg?fit=max&auto=format&n=JFTbQOWMkMfvFjDC&q=85&s=48c11072974cd6a16f3393683fcf6539" alt="Anthropic account configuration form with fields for API key and collaborators" width="1280" height="1122" data-path="images/anthropic-1.jpeg" />
    </Frame>
  </Step>

  <Step title="Add Models">
    Select the model from the list. If you see the model you want to add in the list of checkboxes, we support public model cost for these models.

    <Note>
      (Optional) If the model you are looking for is not present in the options, you can add it using `+ Add Model` at the end of list (scroll down to see the option) by filling the form.
    </Note>

    <Info>
      TrueFoundry AI Gateway supports all text and image models in Anthropic.The complete list of models supported by Bedrock can be found [here](https://docs.anthropic.com/en/docs/models-overview).
    </Info>
  </Step>
</Steps>

### Inference

After adding the models, you can perform inference using an Anthropic-compatible API via the Playground or by integrating with your own application.

<Info>
  For Anthropic streaming requests, AI Gateway supports fallback on `overloaded_error` before generation begins. The gateway waits for the first non-empty stream chunk; if it receives an `overloaded_error` before that first chunk, it automatically falls back to the next configured model. Learn more in [Anthropic Stream Overload Fallback](/docs/ai-gateway/anthropic-stream-overload-fallback).
</Info>

<img src="https://mintcdn.com/truefoundry/FeKcq2n1MMm83Par/images/Screenshot2025-07-22at2.41.00PM-min.png?fit=max&auto=format&n=FeKcq2n1MMm83Par&q=85&s=db92c1fde225cb1370fc9d0e39711629" alt="Code Snippet and Try in Playgroud Buttons for each model" width="3588" height="2004" data-path="images/Screenshot2025-07-22at2.41.00PM-min.png" />

### Supported APIs

Once your Anthropic provider account is configured, the following API surfaces are available through the gateway. The table below summarizes each endpoint alongside platform feature support (tracing, cost tracking).

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Supported by Provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| API                                   | Endpoint            | Tracing | Cost Tracking |
| ------------------------------------- | ------------------- | ------- | ------------- |
| [Chat Completions](#chat-completions) | `/chat/completions` | **✅**   | **✅**         |
| [Messages API](#messages-api)         | `/messages`         | **✅**   | **✅**         |
| [Files API](#files-api)               | `/files`            | **✅**   | **✅**         |

<AccordionGroup>
  <Accordion title="Chat Completions">
    The chat completions endpoint is the most widely used — it supports streaming, tools, multimodal input (images, PDF), structured JSON outputs, prompt caching and extended thinking.
    Full provider capability matrix: [Chat Completions API](/docs/ai-gateway/chat-completions-overview).

    ```python Python lines theme={"dark"}
    from openai import OpenAI

    client = OpenAI(
        api_key="your-truefoundry-api-key",
        base_url="{GATEWAY_BASE_URL}",
    )

    response = client.chat.completions.create(
        model="anthropic-main/claude-sonnet-4-20250514",
        messages=[
            {"role": "user", "content": "What is TrueFoundry in one line?"},
        ],
    )
    print(response.choices[0].message.content)
    ```

    <AccordionGroup>
      <Accordion title="Streaming">
        Set `stream=True` to start streaming responses and iterate over delta chunks. You may defensively check that `chunk.choices` is non-empty and `delta.content` is not `None`.

        ```python Python lines theme={"dark"}
        stream = client.chat.completions.create(
            model="openai-main/gpt-4o-mini",
            messages=[{"role": "user", "content": "Count from 1 to 5."}],
            stream=True,
        )
        for chunk in stream:
            if (
                chunk.choices
                and len(chunk.choices) > 0
                and chunk.choices[0].delta.content is not None
            ):
                print(chunk.choices[0].delta.content, end="", flush=True)
        ```
      </Accordion>

      <Accordion title="Function calling / tools">
        Advertise a tool, hand the model's `tool_calls` back as a `tool` role message, then request the final response. Use `tool_choice` to force the model to call a specific tool when you need deterministic behaviour.

        ```python Python lines theme={"dark"}
        import json

        tools = [{
            "type": "function",
            "function": {
                "name": "get_weather",
                "description": "Get the current weather for a city.",
                "parameters": {
                    "type": "object",
                    "properties": {"city": {"type": "string"}},
                    "required": ["city"],
                },
            },
        }]

        messages = [{"role": "user", "content": "Weather in Bengaluru?"}]
        first = client.chat.completions.create(
            model="anthropic-main/claude-sonnet-4-20250514",
            messages=messages,
            tools=tools,
            tool_choice={"type": "function", "function": {"name": "get_weather"}},
        )

        assistant_msg = first.choices[0].message
        tool_calls = assistant_msg.tool_calls or []
        if tool_calls:
            tool_call = tool_calls[0]
            messages.append(assistant_msg)
            messages.append({
                "role": "tool",
                "tool_call_id": tool_call.id,
                "content": json.dumps({"city": "Bengaluru", "temp_c": 28, "summary": "partly cloudy"}),
            })
            second = client.chat.completions.create(
                model="anthropic-main/claude-sonnet-4-20250514",
                messages=messages,
            )
            print(second.choices[0].message.content)
        ```
      </Accordion>

      <Accordion title="Vision (multimodal images)">
        Claude 3+ models support image inputs via the `image_url` content part.The URL can be a public HTTP URL or an inline `data:image/...;base64,...` URI. For self-contained examples we recommend the inline form.

        ```python Python lines theme={"dark"}
        image_url = (
            "https://upload.wikimedia.org/wikipedia/commons/thumb/d/dd/Gfp-wisconsin-madison-the-nature-boardwalk.jpg/2560px-Gfp-wisconsin-madison-the-nature-boardwalk.jpg"
        )

        response = client.chat.completions.create(
            model="anthropic-main/claude-sonnet-4-20250514",
            messages=[{
                "role": "user",
                "content": [
                    {"type": "text", "text": "Describe this image in one sentence."},
                    {"type": "image_url", "image_url": {"url": image_url}},
                ],
            }],
        )
        print(response.choices[0].message.content)
        ```
      </Accordion>

      <Accordion title="PDF document input">
        Claude models support PDF documents via the `file` content type with base64 encoding.

        ```python Python lines theme={"dark"}
        import base64

        with open("sample.pdf", "rb") as f:
            pdf_b64 = base64.b64encode(f.read()).decode("ascii")

        response = client.chat.completions.create(
            model="anthropic-main/claude-sonnet-4-20250514",
            messages=[{
                "role": "user",
                "content": [
                    {"type": "text", "text": "What text is in this PDF?"},
                    {
                        "type": "file",
                        "file": {
                            "filename": "sample.pdf",
                            "file_data": f"data:application/pdf;base64,{pdf_b64}",
                        },
                    },
                ],
            }],
        )
        print(response.choices[0].message.content)
        ```
      </Accordion>

      <Accordion title="Structured outputs (JSON schema)">
        Use `response_format={"type": "json_schema", ...}` to force the model to return data matching a JSON schema. Claude 4.5/4.6 models use native JSON schema support; older models use a tool-conversion fallback.

        <Warning>
          Anthropic does not support numeric constraint parameters (`ge`, `le`,`minimum`, `maximum`) in schemas. If you use Pydantic-generated schemas,
          strip these constraints before passing them through.
        </Warning>

        ```python Python lines theme={"dark"}
        import json

        schema = {
            "name": "person",
            "schema": {
                "type": "object",
                "properties": {
                    "name": {"type": "string"},
                    "age": {"type": "integer"},
                    "hobbies": {"type": "array", "items": {"type": "string"}},
                },
                "required": ["name", "age", "hobbies"],
                "additionalProperties": False,
            },
            "strict": True,
        }

        response = client.chat.completions.create(
            model="anthropic-main/claude-sonnet-4-20250514",
            messages=[{"role": "user", "content": "Invent a fictional person with name, age, and three hobbies."}],
            response_format={"type": "json_schema", "json_schema": schema},
        )

        message = response.choices[0].message
        if getattr(message, "refusal", None):
            print("model refused:", message.refusal)
        elif not message.content:
            print("model returned empty content")
        else:
            print(json.dumps(json.loads(message.content), indent=2))
        ```
      </Accordion>

      <Accordion title="Prompt caching">
        Anthropic requires **explicit** `cache_control` on content blocks you want cached (unlike OpenAI's automatic caching). Cached tokens appear as `cache_creation_input_tokens` (first call) and `cache_read_input_tokens` (subsequent calls) in the usage response.

        <Warning>
          Minimum cacheable prefix: **1024 tokens** for Claude Sonnet/Opus, **2048 tokens** for Claude Haiku. Prompts shorter than this will accept the `cache_control` hint but won't actually be cached.
        </Warning>

        ```python Python lines theme={"dark"}
        response = client.chat.completions.create(
            model="anthropic-main/claude-sonnet-4-20250514",
            messages=[
                {
                    "role": "system",
                    "content": [
                        {
                            "type": "text",
                            "text": "<LONG_SYSTEM_PROMPT_OVER_1024_TOKENS>",
                            "cache_control": {"type": "ephemeral", "ttl": "5m"},
                        },
                    ],
                },
                {"role": "user", "content": "What should I check in a Helm chart review?"},
            ],
        )
        usage = response.usage
        extra = getattr(usage, "model_extra", {}) or {}
        print("cache_creation:", extra.get("cache_creation_input_tokens", 0))
        print("cache_read    :", extra.get("cache_read_input_tokens", 0))
        ```
      </Accordion>

      <Accordion title="Extended thinking (reasoning)">
        Claude Sonnet 3.7, Claude 4, and Claude 4.5 series models support extended thinking. Use the `reasoning_effort` parameter — the gateway translates it into Anthropic's native `thinking` parameter format.

        <Info>
          The gateway maps `reasoning_effort` to a `thinking.budget_tokens` ratio of the request's `max_tokens`: `none` = 0%, `low` = 30%, `medium` = 60%, `high` = 90%.
        </Info>

        The response includes `reasoning_content` (plain text) and `thinking_blocks` (structured blocks with cryptographic signatures required for multi-turn reasoning continuity).

        ```python Python lines theme={"dark"}
        response = client.chat.completions.create(
            model="anthropic-main/claude-sonnet-4-20250514",
            messages=[{"role": "user", "content": "A bat and ball cost $1.10.The bat costs $1.00 more than the ball. How much is the ball?"}],
            reasoning_effort="high",
            max_tokens=8000,
        )

        msg = response.choices[0].message
        print("answer:", msg.content)
        print("reasoning:", getattr(msg, "reasoning_content", None))
        # thinking_blocks carry signatures for multi-turn continuity
        for block in getattr(msg, "thinking_blocks", []) or []:
            print("  block:", block.get("type"), "signature:", block.get("signature", "")[:30])
        ```

        <Warning>
          Always echo `thinking_blocks` exactly as returned when continuing a conversation. Blocks with missing or modified `signature` fields are rejected by Anthropic.
        </Warning>
      </Accordion>
    </AccordionGroup>
  </Accordion>

  <Accordion title="Messages API">
    Anthropic's native Messages API (`/messages`) is also exposed through the gateway, letting you use the official `anthropic` Python SDK directly.
    You get the same gateway features — routing, logging, rate-limiting, budget management — as with the OpenAI-compatible interface.
    Full docs:  [Messages API](/docs/ai-gateway/messages-overview), [Native SDK Support](/docs/ai-gateway/native-sdk-support)

    <Note>
      The gateway accepts both Anthropic SDK auth patterns and translates
      internally:

      * `api_key=TFY_API_KEY` - SDK sends the `x-api-key` header
      * `auth_token=TFY_API_KEY` — SDK sends the `Authorization: Bearer` header

      Either works; the request body is identical. `api_key` is the idiomatic Anthropic SDK pattern - use it unless you have a reason to send a Bearer token.
    </Note>

    ```python Python lines theme={"dark"}
    import os
    from anthropic import Anthropic

    client = Anthropic(
        api_key="your-truefoundry-api-key",
        base_url="{GATEWAY_BASE_URL}",
    )

    message = client.messages.create(
        model="anthropic-main/claude-sonnet-4-20250514",
        max_tokens=256,
        # `system` is a top-level parameter in Anthropic's native API, not a message role. Pass it here — not inside `messages`.
        system="You answer in one short sentence.",
        messages=[
            {"role": "user", "content": "What is TrueFoundry in one line?"}
        ],
    )

    print(message.content[0].text)
    print(message.usage)
    ```

    <AccordionGroup>
      <Accordion title="Streaming">
        Use `.messages.stream()` and iterate over `text_stream` for incremental
        output.

        ```python Python lines theme={"dark"}
        with client.messages.stream(
            model="anthropic-main/claude-sonnet-4-20250514",
            max_tokens=256,
            messages=[{"role": "user", "content": "Count from 1 to 5, one per line."}],
        ) as stream:
            for text in stream.text_stream:
                print(text, end="", flush=True)
        ```
      </Accordion>
    </AccordionGroup>
  </Accordion>

  <Accordion title="Files API">
    Upload, list, retrieve, and delete files held by the gateway. The gateway translates the OpenAI-compatible Files API into Anthropic's native Files API automatically.
    Full docs: [Files API](/docs/ai-gateway/file-endpoints).

    <Warning>
      The Files API requires the `x-tfy-provider-name` header on the client so the gateway can route the request to the right Anthropic provider account.
    </Warning>

    <Note>
      File content retrieval (`files.content`) only works for files created by skills or the code execution tool. User-uploaded files cannot be downloaded back — you can only list metadata and delete them.
    </Note>

    ```python Python lines theme={"dark"}
    from openai import OpenAI

    files_client = OpenAI(
        api_key="your-truefoundry-api-key",
        base_url="{GATEWAY_BASE_URL}",
        default_headers={"x-tfy-provider-name": "anthropic-main"},
    )

    # Upload
    with open("document.txt", "rb") as f:
        uploaded = files_client.files.create(file=f, purpose="assistants")
    print(uploaded.id, uploaded.filename, uploaded.bytes)

    # List (slice client-side; the gateway may not honour `limit`)
    listed = files_client.files.list()
    for f in listed.data[:5]:
        print(f.id, f.purpose, f.bytes)

    # Retrieve metadata
    meta = files_client.files.retrieve(uploaded.id)

    # Delete
    deleted = files_client.files.delete(uploaded.id)
    print(deleted.deleted)
    ```
  </Accordion>
</AccordionGroup>
