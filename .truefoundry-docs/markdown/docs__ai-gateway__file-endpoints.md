> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Files API (/files)

> Upload, manage, and retrieve files through TrueFoundry's AI Gateway

## Provider capabilities

The summary below reflects high-level Files API availability per provider. For per-operation details (upload, list, delete, etc.), see **Provider Support Matrix** below.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider     | Files                                                       |
| ------------ | ----------------------------------------------------------- |
| OpenAI       | ✅                                                           |
| Azure OpenAI | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Anthropic    | ✅                                                           |
| Bedrock      | ✅                                                           |
| Vertex       | ✅                                                           |
| Cohere       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Gemini       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Groq         | ✅                                                           |
| Cerebras     | <Icon icon="circle-minus" iconType="regular" />             |
| Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI          | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

### Provider Support Matrix

TrueFoundry's Files API provides a unified OpenAI-compatible interface across multiple providers. Not all providers support every operation — refer to the matrix below.

| Provider             | Upload | List | Retrieve | Content | Delete | Integration  | Storage Backend      |
| -------------------- | ------ | ---- | -------- | ------- | ------ | ------------ | -------------------- |
| **OpenAI**           | ✅      | ✅    | ✅        | ✅       | ✅      | Pass-through | OpenAI API           |
| **Azure OpenAI**     | ✅      | ✅    | ✅        | ✅       | ✅      | Pass-through | Azure OpenAI API     |
| **Anthropic**        | ✅      | ✅    | ✅        | ✅       | ✅      | Transformed  | Anthropic Files API  |
| **AWS Bedrock**      | ✅      | ❌    | ✅        | ✅       | ❌      | Transformed  | AWS S3               |
| **Google Vertex AI** | ✅      | ❌    | ✅        | ✅       | ❌      | Transformed  | Google Cloud Storage |
| **Groq**             | ✅      | ✅    | ✅        | ✅       | ✅      | Transformed  | Groq API             |

* **Pass-through** — Requests are proxied directly to the provider's native File API without modification.
* **Transformed** — The gateway translates requests and responses between OpenAI's format and the provider's native format. For Bedrock and Vertex AI, the gateway also manages file storage (S3 and GCS respectively) on behalf of the provider.

This guide demonstrates how to set up the client and use file-related endpoints.

Setup OpenAI client first -
Provider Support Matrix

```python lines theme={"dark"}
from openai import OpenAI

BASE_URL = "{GATEWAY_BASE_URL}"
API_KEY = "your-truefoundry-api-key"

# Configure OpenAI client with TrueFoundry settings
client = OpenAI(
    api_key=API_KEY,
    base_url=BASE_URL,
)
```

### **Provider Specific Extra Headers**

When making requests, you'll need to specify provider-specific headers based on which LLM provider you're using. Choose your provider:

<CodeGroup>
  ```python OpenAI Provider lines theme={"dark"}
  extra_headers = {
      "x-tfy-provider-name": "openai-provider-name"  # name of tfy provider integration
  }
  ```

  ```python Vertex AI Provider lines theme={"dark"}
  extra_headers = {
      "x-tfy-provider-name": "google-provider-name",  # name of tfy provider integration
      "x-tfy-vertex-storage-bucket-name": "your-bucket-name",
      "x-tfy-vertex-region": "your-region",  # e.g., "europe-west4"
      "x-tfy-provider-model": "gemini-2-0-flash"  # or any other supported model
  }
  ```

  ```python Anthropic Provider lines theme={"dark"}
  extra_headers = {
      "x-tfy-provider-name": "anthropic-provider-name"  # name of tfy provider integration
  }
  ```
</CodeGroup>

### 1. Upload File

Use this to upload files for usage with Batch or Assistants APIs.

```python lines theme={"dark"}
file = client.files.create(
    file=open("request.jsonl", "rb"),
    purpose="batch",
    extra_headers=extra_headers
)

print(file.id)
```

**Notes**:

* Max individual file size: **512 MB**
* Max org-wide storage: **100 GB**
* For Batch API: only `.jsonl`, max 50 MB
* For Assistants API: up to 2M tokens

**Expected Output**

```json lines theme={"dark"}
{
  "id": "file-abc123",
  "object": "file",
  "bytes": 120000,
  "created_at": 1677610602,
  "filename": "mydata.jsonl",
  "purpose": "fine-tune",
}
```

### 2. List Files

Returns a list of files

```python lines theme={"dark"}
files = client.files.list(
	extra_headers=extra_headers
)
```

**Expected Output**

```json lines theme={"dark"}
{
  "data": [
    {
      "id": "file-abc123",
      "object": "file",
      "bytes": 175,
      "created_at": 1613677385,
      "filename": "salesOverview.pdf",
      "purpose": "assistants",
    },
    {
      "id": "file-abc123",
      "object": "file",
      "bytes": 140,
      "created_at": 1613779121,
      "filename": "puppy.jsonl",
      "purpose": "fine-tune",
    }
  ],
  "object": "list"
}
```

### 3. Retrieve File

Returns information about a specific file.

```python lines theme={"dark"}
file = client.files.retrieve(
	file.id,
	extra_headers=extra_headers
)

print(file.json())
```

**Expected Output**

```json lines theme={"dark"}
{
  "id": "file-abc123",
  "object": "file",
  "bytes": 120000,
  "created_at": 1677610602,
  "filename": "mydata.jsonl",
  "purpose": "fine-tune",
}
```

### 4. Delete File

Delete a file permanently

```python lines theme={"dark"}
deleted_file = client.files.delete(
	file.id,
	extra_headers=extra_headers
)

print(deleted_file.json())
```

**Expected Output**

```json lines theme={"dark"}
{
  "id": "file-abc123",
  "object": "file",
  "deleted": true
}
```

### 5. Retrieve File Content

Returns the contents of the specified file.

```python lines theme={"dark"}
output_content = client.files.content(
    file.id,
    extra_headers=extra_headers
)

print(output_content.json())
```

<Note>
  **Anthropic:** Retrieve File Content works only for files created by [skills](https://docs.anthropic.com/en/docs/build-with-claude/skills-guide) or the [code execution tool](https://docs.anthropic.com/en/docs/agents-and-tools/tool-use/code-execution-tool). Files you upload via the Files API cannot be downloaded.
</Note>
