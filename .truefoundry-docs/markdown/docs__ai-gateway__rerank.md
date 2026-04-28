> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Rerank API (/rerank)

> Learn how to use the Rerank API through TrueFoundry Gateway to enhance search relevance by scoring and reordering document results.

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider     | Rerank                                                      |
| ------------ | ----------------------------------------------------------- |
| OpenAI       | <Icon icon="circle-minus" iconType="regular" />             |
| Azure OpenAI | <Icon icon="circle-minus" iconType="regular" />             |
| Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
| Bedrock      | ✅                                                           |
| Vertex       | <Icon icon="circle-minus" iconType="regular" />             |
| Cohere       | ✅                                                           |
| Gemini       | <Icon icon="circle-minus" iconType="regular" />             |
| Groq         | <Icon icon="circle-minus" iconType="regular" />             |
| Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI          | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

**Reranking** is a powerful technique that enhances the relevance of search results. It works by taking a **query** and a list of **documents**, then scoring each document based on how relevant it is to the query. This is especially useful in multi-step search pipelines where you:

1. Use a traditional or vector-based search system to retrieve candidate documents.
2. Use the **Rerank API** to reorder those candidates based on relevance.

Reranking helps boost precision in **semantic search**, **chatbots**, **RAG systems**, and **knowledge retrieval systems**.

The reranker takes:

* A **query** (user input or search question)
* A list of **documents** (initially retrieved search results)
* Returns a ranked list with **relevance scores**

### Code Snippet

```python lines theme={"dark"}
import cohere

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

# Configure Cohere client with TrueFoundry settings
client = cohere.ClientV2(
	api_key=API_KEY,
	base_url=BASE_URL
)

response = client.rerank(
    model="test-ss-cohere/rerank-multilingual-v3-0",
    query="Where is New Delhi?",
    documents=["New Delhi is capital of India","The sky is blue"],
    top_n=1,
)

print(response)
```

### Expected Output

```json lines theme={"dark"}
{
  "id": "a9279c7a-1e34-4bee-9892-db1d38ca9be0",
  "results": [
    {
      "index": 0,
      "relevance_score": 0.783542
    }
  ]
}
```

In this example, the document *"New Delhi is the capital of India."* has a higher relevance score to the query compared to *"The sky is blue."*, and is therefore ranked higher.
