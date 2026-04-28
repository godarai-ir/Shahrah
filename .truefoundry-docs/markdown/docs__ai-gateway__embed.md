> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Embeddings API (/embeddings)

> Generate vector embeddings through TrueFoundry's AI Gateway

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider     | String                                                      | List of String                                              |
| ------------ | ----------------------------------------------------------- | ----------------------------------------------------------- |
| OpenAI       | ✅                                                           | ✅                                                           |
| Azure OpenAI | ✅                                                           | ✅                                                           |
| Anthropic    | <Icon icon="circle-minus" iconType="regular" />             | <Icon icon="circle-minus" iconType="regular" />             |
| Bedrock      | ✅                                                           | ✅                                                           |
| Vertex       | ✅                                                           | ✅                                                           |
| Cohere       | ✅                                                           | ✅                                                           |
| Gemini       | <Icon icon="circle-minus" iconType="regular" />             | <Icon icon="circle-minus" iconType="regular" />             |
| Groq         | <Icon icon="circle-minus" iconType="regular" />             | <Icon icon="circle-minus" iconType="regular" />             |
| SambaNova    | <Icon icon="circle-xmark" iconType="regular" color="red" /> | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Together-AI  | ✅                                                           | ✅                                                           |
| xAI          | <Icon icon="circle-minus" iconType="regular" />             | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> | <Icon icon="circle-xmark" iconType="regular" color="red" /> |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

An **embedding** is a sequence of numbers (a vector) that represents the semantic meaning of content such as **natural language**, **code**, or other structured data. They are widely used in Clustering, Semantic search and retrieval, Recommendation engines, Retrieval-Augmented Generation (RAG)

## Basic Usage

Generate embeddings using the OpenAI-compatible API:

```python lines theme={"dark"}
from openai import OpenAI

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

# Configure OpenAI client with TrueFoundry settings
client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
)

response = client.embeddings.create(
    model="openai-main/text-embedding-3-small",
    input="TrueFoundry is amazing!"
)

print(response.data[0].embedding)
```

**Expected Output:**

```json lines theme={"dark"}
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "embedding": [-0.006929283495992422, -0.005336422007530928, -4.547132266452536e-5, -0.024047505110502243]
    }
  ],
  "model": "text-embedding-3-small",
  "usage": {
    "prompt_tokens": 5,
    "total_tokens": 5
  }
}
```

## Provider-Specific Features

### Input Type - Cohere

When using Cohere models via the embeddings API, you must include an additional field called `input_type` in the request. This field indicates the purpose of the embedding and must be one of the following:

* `search_query`
* `search_document`
* `classification`
* `clustering`

**Example:**

```python lines theme={"dark"}
from openai import OpenAI

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
)

# Embed a search query
response = client.embeddings.create(
    model="cohere-main/embed-english-v3.0",
    input="Find similar documents about AI.",
    input_type="search_query"
)

print(response.data[0].embedding)
```

**Expected Output:**

```json lines theme={"dark"}
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "embedding": [-0.0123, 0.0456, -0.0789, ...]
    }
  ],
  "model": "embed-english-v3.0",
  "usage": {
    "prompt_tokens": 6,
    "total_tokens": 6
  }
}
```

### Task Types - Vertex AI & Gemini

Vertex AI and Gemini embedding models support `task_type` that optimizes embeddings for specific use cases. Specify the task type using the `extra_body` parameter.

**Available task types:** `RETRIEVAL_DOCUMENT`, `RETRIEVAL_QUERY`, `QUESTION_ANSWERING`, `FACT_VERIFICATION`, `SEMANTIC_SIMILARITY`, `CLASSIFICATION`, `CLUSTERING`, `CODE_RETRIEVAL_QUERY`

For detailed information about task types and when to use them, see:

* [Vertex AI Task Types Documentation](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/embeddings/task-types)
* [Gemini Task Types Documentation](https://ai.google.dev/gemini-api/docs/embeddings#task-types)

**Example:**

```python lines theme={"dark"}
from openai import OpenAI

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
)

# Generate embedding with task type
response = client.embeddings.create(
    model="tfy-ai-gemini/gemini-embedding-001",
    input="Enter your input here",
    extra_body={
        "task_type": "SEMANTIC_SIMILARITY",
    }
)

print(response.data[0].embedding)
```

**Expected Output:**

```python lines theme={"dark"}
CreateEmbeddingResponse(
    data=[
        Embedding(
            embedding=[-0.03504209965467453, 0.012172757647931576, ...],
            index=0,
            object='embedding'
        )
    ],
    model='tfy-ai-gemini/gemini-embedding-001',
    object='list',
    usage=Usage(prompt_tokens=0, total_tokens=0)
)
```

### Multimodal Embeddings - Vertex AI

> **Note:** Multimodal embeddings are only available for the Vertex AI `multimodalembedding@001` model.

Vertex AI supports multimodal embeddings that can encode **images** and **videos** along with text. This enables applications to search across multiple modalities or find semantic similarity between text and visual content.

#### Image Embeddings

Generate embeddings from images with optional text captions:

```python lines theme={"dark"}
from openai import OpenAI

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
)

response = client.embeddings.create(
    model="vertex-ai-main/multimodalembedding@001",
    input=[
        {
            "text": "A red sports car on a mountain road",
            "image": {
                "base64": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
                # "url": "gs://your-bucket/path/to/image.jpg"  # Alternative: use GCS URL
            }
        }
    ]
)

# Access text embedding
print(response.data[0].embedding)

# Access image embedding
print(response.data[0].image_embedding)
```

**Expected Output:**

```json lines theme={"dark"}
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "embedding": [0.0234, -0.0145, 0.0567, ...],
      "image_embedding": [-0.0347, 0.0336, -0.0160, ...]
    }
  ],
  "model": "multimodalembedding@001",
  "usage": {
    "prompt_tokens": 0,
    "total_tokens": 0
  }
}
```

#### Video Embeddings

Generate embeddings from video segments with temporal control:

```python lines theme={"dark"}
from openai import OpenAI

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
)

response = client.embeddings.create(
    model="vertex-ai-main/multimodalembedding@001",
    input=[
        {
            "text": "A cooking tutorial showing pasta preparation",
            "video": {
                "base64": "AAAAGGZ0eXBpc29tAAACAGlzb21pc28yYXZjMW1wNDEAAAAIZnJlZQAACKBtZGF0AAAA...",
                "start_offset": 0,      # Start time in seconds
                "end_offset": 30,       # End time in seconds
                "interval": 10,         # Sample interval in seconds
                # "url": "gs://your-bucket/path/to/video.mp4"  # Alternative: use GCS URL
            }
        }
    ]
)

# Access text embedding
print(response.data[0].embedding)

# Access video embeddings (array of embeddings for each segment)
print(response.data[0].video_embeddings)
```

**Expected Output:**

```json lines theme={"dark"}
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "index": 0,
      "embedding": [0.0234, -0.0145, 0.0567, ...],
      "video_embeddings": [
        {
          "object": "embedding",
          "embedding": [0.0189, -0.0201, 0.0423, ...],
          "index": 0,
          "start_offset": 0,
          "end_offset": 10
        },
        {
          "object": "embedding",
          "embedding": [0.0278, -0.0156, 0.0512, ...],
          "index": 1,
          "start_offset": 10,
          "end_offset": 20
        }
      ]
    }
  ],
  "model": "multimodalembedding@001",
  "usage": {
    "prompt_tokens": 0,
    "total_tokens": 0
  }
}
```
