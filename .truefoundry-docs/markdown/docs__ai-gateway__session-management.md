> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Session Management

> Track and manage multi-turn conversations in TrueFoundry AI Gateway.

## The Problem: Stateless Chat Completions

The Chat Completions API is inherently stateless. Each API call is independent and doesn't maintain any memory of previous interactions. When building conversational applications, this creates challenges:

* **No automatic conversation tracking**: Multiple requests from the same user conversation are treated as isolated calls
* **Difficult debugging**: When users report issues, it's hard to trace all requests belonging to a specific conversation
* **Poor observability**: Logs don't naturally group related requests together, making it hard to understand the full context as you will see multiple request logs for same conversation.

## The Solution: Conversation ID Header

TrueFoundry AI Gateway solves this by allowing you to tag requests with a conversation identifier. By passing the `X-TFY-CONVERSATION-ID` header with each request, you can:

* Group multiple API calls into a logical conversation session
* Filter and view all requests belonging to a specific conversation in the UI

## Implementation

Add the `X-TFY-CONVERSATION-ID` header to your chat completion requests. The value should be a unique identifier for each conversation (e.g., UUID, user session ID, or any unique string).

```python theme={"dark"}
from openai import OpenAI

client = OpenAI(
    api_key="Your TFY API key",
    base_url="<Your Gateway Base URL>"
)

stream = client.chat.completions.create(
    messages=[
        {"role": "system", "content": "You are an AI bot."},
        {"role": "user", "content": "Enter your prompt here"},
    ],
    model="test-groq-prod-key/deepseek-r1-distill-llama-70b",
    stream=True,
    extra_headers={
        "X-TFY-CONVERSATION-ID": "conv-12345",  # Your unique conversation ID
        "X-TFY-LOGGING-CONFIG": '{"enabled": true}',
    },
)

for chunk in stream:
    if (
        chunk.choices
        and len(chunk.choices) > 0
        and chunk.choices[0].delta.content is not None
    ):
        print(chunk.choices[0].delta.content, end="", flush=True)
```

<Note>
  Use a consistent conversation ID across all requests in the same conversation. Generate a new unique ID when starting a new conversation.
  Generally it is recommended to use UUID5 for the conversation id.
</Note>

## Viewing Sessions in the UI

Once you've tagged requests with conversation IDs, you can filter them in the TrueFoundry UI's request logs section. This allows you to:

* View all API calls within a specific conversation
* Analyze the complete request-response history

<Frame caption="Find conversation ID in the TrueFoundry UI">
  <img src="https://mintcdn.com/truefoundry/eHCVL_OgWdTpr4Oz/images/conversation-id-see-min.png?fit=max&auto=format&n=eHCVL_OgWdTpr4Oz&q=85&s=b4ec855e2821712d7428642eba816ac2" width="3010" height="1468" data-path="images/conversation-id-see-min.png" />
</Frame>

<Frame caption="Filter requests by conversation ID in the TrueFoundry UI">
  <img src="https://mintcdn.com/truefoundry/eHCVL_OgWdTpr4Oz/images/conversation-id-filter-min.png?fit=max&auto=format&n=eHCVL_OgWdTpr4Oz&q=85&s=1f288aa361b1fab8ea30ed5fd11bb32f" width="3024" height="766" data-path="images/conversation-id-filter-min.png" />
</Frame>
