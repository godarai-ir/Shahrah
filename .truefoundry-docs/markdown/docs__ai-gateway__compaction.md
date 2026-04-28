> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Compaction API (/responses/compact)

> Learn how to use the Compaction API through TrueFoundry Gateway to manage long-running conversations by reducing context size while preserving state.

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider | Compaction API |
| -------- | -------------- |
| OpenAI   | ✅              |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

Compaction reduces context size while preserving state for the next turn, so you can balance quality, cost, and latency as conversations grow.

The gateway supports **standalone** `POST /responses/compact` and **server-side** compaction via `context_management` in `POST /responses`.

<Note>
  Compaction is only supported by OpenAI. Use the gateway’s OpenAI inference base URL.
</Note>

## Standalone: `POST /responses/compact`

Send a full context window; the API returns a compacted window (including an opaque, encrypted compaction item) to pass as input to your next `/responses` call. Body: `model`, `input`, and optionally `instructions`, `previous_response_id`.

<Warning>
  Do not prune the compact response. Pass the full `output` into your next `/responses` call as-is.
</Warning>

### Code

```python theme={"dark"}
from openai import OpenAI

client = OpenAI(
    api_key="your-tfy-api-key",
    base_url="{GATEWAY_BASE_URL}",
)

compacted = client.responses.compact(
    model="openai-main/gpt-4o",
    input=long_input_items_array,
)

next_input = [
    *compacted.output,
    {"type": "message", "role": "user", "content": user_input_message()},
]

next_response = client.responses.create(
    model="openai-main/gpt-4o",
    input=next_input,
    store=False,
)
```

### Response shape

```json theme={"dark"}
{
  "id": "resp_compact_123",
  "object": "response.compaction",
  "created_at": 1234567890,
  "output": [
    { "type": "compaction", "encrypted_content": "..." }
  ],
  "usage": {
    "input_tokens": 15000,
    "output_tokens": 1200,
    "total_tokens": 16200
  }
}
```

## Server-side: `POST /responses` with `context_management`

Set `context_management: [{"type": "compaction", "compact_threshold": 200000}]` on create. When the rendered token count crosses the threshold, the server compacts and emits a compaction item in the stream. No separate `/responses/compact` call needed.

* **Stateless:** Append response output (including compaction items) to your input each turn.
* **Stateful:** Use `previous_response_id` and send only the new user message; do not manually prune.

<Tip>
  With stateless chaining, you can drop items that came *before* the most recent compaction item to keep requests smaller. With `previous_response_id`, do not prune.
</Tip>

### Code

```python theme={"dark"}
conversation = [
    {"type": "message", "role": "user", "content": "Let's begin a long coding task."},
]

while keep_going:
    response = client.responses.create(
        model="openai-main/gpt-4o",
        input=conversation,
        store=False,
        context_management=[{"type": "compaction", "compact_threshold": 200000}],
    )
    conversation.extend(response.output)
    conversation.append(
        {"type": "message", "role": "user", "content": get_next_user_input()},
    )
```

## References

* [OpenAI Compaction guide](https://developers.openai.com/api/docs/guides/compaction)
* [OpenAI Responses API reference](https://developers.openai.com/docs/api-reference/responses)
