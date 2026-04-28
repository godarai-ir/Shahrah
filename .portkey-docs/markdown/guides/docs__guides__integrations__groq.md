> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Groq

[<img src="https://mintcdn.com/portkey-docs/izapyWTWQvJmiZ2Q/images/guides/colab-badge.svg?fit=max&auto=format&n=izapyWTWQvJmiZ2Q&q=85&s=cadfc29fd7966d28a3c852d79a015cce" alt="" width="117" height="20" data-path="images/guides/colab-badge.svg" />](https://colab.research.google.com/drive/1USSOBS3uWrpgZirIIAJmlyC9XHnFCVSQ?usp=sharing)

## Portkey + Groq

[Portkey](https://app.portkey.ai/) is the Control Panel for AI apps. With it's popular AI Gateway and Observability Suite, hundreds of teams ship reliable, cost-efficient, and fast apps.

With Portkey:

* Connect to 1,600+ models through a unified API,
* View 40+ metrics & logs for all requests,
* Enable semantic cache to reduce latency & costs,
* Implement automatic retries & fallbacks for failed requests,
* Add custom tags to requests for better tracking and analysis and more.

### Use Groq API with OpenAI Compatibility

Portkey is fully compatible with the OpenAI signature. Connect to the Portkey AI Gateway through the OpenAI Client:

* Set `base_url` to `PORTKEY_GATEWAY_URL`
* Add `default_headers` using the `createHeaders` helper method

**Prerequisites:**

* [Portkey API key](https://app.portkey.ai/)
* [Groq API key](https://console.groq.com/keys)

```bash icon="square-terminal" theme={"system"}
pip install -qU portkey-ai openai
```

### With OpenAI Client

```python OpenAI Python icon="openai" theme={"system"}
from openai import OpenAI
from portkey_ai import PORTKEY_GATEWAY_URL, createHeaders
from google.colab import userdata

client = OpenAI(
    api_key=userdata.get('GROQ_API_KEY'),  # replace with your Groq API key
    base_url=PORTKEY_GATEWAY_URL,
    default_headers=createHeaders(
        provider="groq",
        api_key=userdata.get('PORTKEY_API_KEY')  # replace with your Portkey API key
    )
)

chat_complete = client.chat.completions.create(
    model="llama3-70b-8192",
    messages=[{"role": "user", "content": "What's the purpose of Generative AI?"}]
)

print(chat_complete.choices[0].message.content)
```

```
The primary purpose of generative AI is to create new, original, and often 
realistic data or content, such as images, videos, music, text, or speeches...
```

### With Portkey Client

Note: Add your Groq API key in [Model Catalog](https://app.portkey.ai/model-catalog) and access models using your provider slug

```python Python icon="python" theme={"system"}
from portkey_ai import Portkey

portkey = Portkey(api_key="PORTKEY_API_KEY")

completion = portkey.chat.completions.create(
    model="@groq-prod/llama3-70b-8192",  # @provider-slug/model
    messages=[{"role": "user", "content": "Who are you?"}],
    max_tokens=250
)
```

```python theme={"system"}
print(completion)
```

```json Output theme={"system"}
{
    "id": "chatcmpl-8cec08e0-910e-4331-9c4b-f675d9923371",
    "choices": [{
            "finish_reason": "stop",
            "index": 0,
            "logprobs": null,
            "message": {
            "content": "I am LLaMA, an AI assistant developed by Meta AI...",
                "role": "assistant",
                "function_call": null,
                "tool_calls": null
        }
    }],
    "created": 1714136032,
    "model": "llama3-70b-8192",
    "object": "chat.completion",
    "system_fingerprint": null,
    "usage": {"prompt_tokens": 14, "completion_tokens": 147, "total_tokens": 161}
}
```

### Advanced Routing - Load Balancing

Load balancing distributes traffic across multiple API keys or providers based on custom weights for high availability and optimal performance.

**Example:** Split traffic between Groq's `llama-3-70b` (70%) and OpenAI's `gpt-3.5` (30%):

```python Python icon="python" theme={"system"}
config = {
    "strategy": {"mode": "loadbalance"},
  "targets": [
        {"override_params": {"model": "@groq-prod/llama3-70b-8192"}, "weight": 0.7},
        {"override_params": {"model": "@openai-prod/gpt-4o"}, "weight": 0.3}
    ]
}
```

```python OpenAI Python icon="openai" theme={"system"}
from openai import OpenAI
from portkey_ai import PORTKEY_GATEWAY_URL, createHeaders
from google.colab import userdata

client = OpenAI(
    api_key=userdata.get("PORTKEY_API_KEY"),
    base_url=PORTKEY_GATEWAY_URL,
    default_headers=createHeaders(
        config=config
    )
)

chat_complete = client.chat.completions.create(
    model="X",
    messages=[{"role": "user", "content": "Just say hi!"}]
)

print(chat_complete.model)
print(chat_complete.choices[0].message.content)
```

```text Output theme={"system"}
gpt-3.5-turbo-0125
Hi! How can I assist you today?
```

### Observability with Portkey

Route requests through Portkey to track metrics like tokens used, latency, and cost.

<Frame>
  <img src="https://mintcdn.com/portkey-docs/T0lFtdapIPX8YtCI/images/guides/vercel-2.png?fit=max&auto=format&n=T0lFtdapIPX8YtCI&q=85&s=70ebf98e998e78cbb59e611a26e3fc4c" width="2304" height="1119" data-path="images/guides/vercel-2.png" />
</Frame>
