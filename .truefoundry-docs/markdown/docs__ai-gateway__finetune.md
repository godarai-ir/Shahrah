> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Finetune API (/fine_tuning)

> Fine-tune models using TrueFoundry's AI Gateway with OpenAI or Vertex AI providers

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider     | Fine Tune                                                   |
| ------------ | ----------------------------------------------------------- |
| OpenAI       | ✅                                                           |
| Azure OpenAI | <Icon icon="circle-minus" iconType="regular" />             |
| Anthropic    | <Icon icon="circle-minus" iconType="regular" />             |
| Bedrock      | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Vertex       | ✅                                                           |
| Cohere       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Gemini       | <Icon icon="circle-minus" iconType="regular" />             |
| Groq         | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Cerebras     | <Icon icon="circle-minus" iconType="regular" />             |
| Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI          | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

This guide explains how to fine-tune models using TrueFoundry's AI Gateway with OpenAI or Vertex AI providers.

## Client Setup

All providers use the OpenAI SDK with provider-specific headers. Choose your provider to get started:

<CodeGroup>
  ```python OpenAI Provider lines theme={"dark"}
  from openai import OpenAI

  client = OpenAI(
      api_key="your-truefoundry-api-key",
      base_url="{GATEWAY_BASE_URL}",
      default_headers={
          "x-tfy-provider-name": "openai-main"  # truefoundry provider integration name
      }
  )
  ```

  ```python Vertex AI Provider lines theme={"dark"}
  from openai import OpenAI

  client = OpenAI(
      api_key="your-truefoundry-api-key",
      base_url="{GATEWAY_BASE_URL}",
      default_headers={
          "x-tfy-provider-name": "vertex-main",  # truefoundry provider integration name
          "x-tfy-vertex-storage-bucket-name": "your-bucket-name",
          "x-tfy-provider-model": "gemini-2-0-flash-001",
          "x-tfy-file-purpose": "fine-tune"  # Required for Vertex AI file uploads
      }
  )
  ```
</CodeGroup>

## Training File Format

Create a JSONL file with one JSON object per line. Each line represents a conversation pair for training:

<CodeGroup>
  ```json training.jsonl lines theme={"dark"}
  {"messages":[{"role":"user","content":"What is the weather in San Francisco?"},{"role":"assistant","content":"The weather in San Francisco is currently 18°C with partly cloudy skies. It's a mild day with light winds."}]}
  {"messages":[{"role":"user","content":"What is the weather in Minneapolis?"},{"role":"assistant","content":"The weather in Minneapolis is currently 5°C with overcast skies. It's a chilly day, so you might want to dress warmly."}]}
  {"messages":[{"role":"user","content":"What is the weather in San Diego?"},{"role":"assistant","content":"The weather in San Diego is currently 22°C with sunny skies. It's a beautiful day with perfect beach weather."}]}
  ```
</CodeGroup>

## Workflow Steps

The finetuning process follows these steps for all providers:

1. **Upload**: Upload training file → Get file ID
2. **Create**: Create finetune job → Get job ID
3. **Monitor**: Check status until complete
4. **Use**: Use the fine-tuned model for inference

### Step-by-Step Examples

<AccordionGroup>
  <Accordion title="1. Upload Training File" icon="upload">
    <CodeGroup>
      ```python OpenAI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "openai-main"
          }
      )

      # Upload the training file
      file = client.files.create(
          file=open("training.jsonl", "rb"),
          purpose="fine-tune"
      )

      print(file.id)  # Example: file-PnFGrFLN5LjjcWr4eFsStK
      ```

      ```python Vertex AI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "vertex-main",
              "x-tfy-vertex-storage-bucket-name": "your-bucket-name",
              "x-tfy-provider-model": "gemini-2-0-flash-001",
              "x-tfy-file-purpose": "fine-tune"  # Required header
          }
      )

      # Upload the training file
      # File is automatically transformed to Vertex AI format
      file = client.files.create(
          file=open("training.jsonl", "rb"),
          purpose="fine-tune"
      )

      print(file.id)  # Example: gs%3A%2F%2Fbucket-name%2Ffile.jsonl
      # This is a GCS URI that you'll use in the next step
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="2. Create Finetune Job" icon="play">
    <CodeGroup>
      ```python OpenAI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "openai-main"
          }
      )

      job = client.fine_tuning.jobs.create(
          training_file=file.id,
          model="gpt-3.5-turbo",
          hyperparameters={
              "n_epochs": 2
          },
          suffix="my-custom-model"
      )

      print(job.id)  # Example: ftjob-abc123xyz
      ```

      ```python Vertex AI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "vertex-main",
              "x-tfy-vertex-storage-bucket-name": "your-bucket-name",
              "x-tfy-provider-model": "gemini-2-0-flash-001"
          }
      )

      # Use the GCS URI from file upload
      job = client.fine_tuning.jobs.create(
          training_file=file.id,  # GCS URI from step 1
          model="vertex-main/gemini-2-5-flash",  # model to finetune (truefoundry model format)
          suffix="finetune-model",  # Display name for the fine-tuned model
          extra_body={
              "hyperparameters": {
                  "n_epochs": 2
              }
          }
      )

      print(job.id)  # Example: 7170920278355083264
      print(job.status)  # Example: queued, running, succeeded, failed
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="3. Monitor Job Status" icon="eye">
    <CodeGroup>
      ```python OpenAI Provider lines theme={"dark"}
      from openai import OpenAI
      import time

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "openai-main"
          }
      )

      # Retrieve job status
      job = client.fine_tuning.jobs.retrieve(job.id)
      print(f"Status: {job.status}")  # queued, running, succeeded, failed, cancelled
      ```

      ```python Vertex AI Provider lines theme={"dark"}
      from openai import OpenAI
      import time

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "vertex-main",
              "x-tfy-vertex-storage-bucket-name": "your-bucket-name",
              "x-tfy-provider-model": "gemini-2-0-flash-001"
          }
      )

      # Retrieve job status
      job = client.fine_tuning.jobs.retrieve(job.id)
      print(f"Status: {job.status}")  # queued, running, succeeded, failed, cancelled
      ```
    </CodeGroup>
  </Accordion>
</AccordionGroup>

## Status Reference

* `queued`: Job is queued and waiting to start
* `running`: Fine-tuning is in progress
* `succeeded`: Fine-tuning completed successfully
* `failed`: Fine-tuning failed

### Hyperparameters

You can configure the following hyperparameters in the `hyperparameters` object:

* `n_epochs`: Number of training epochs
* `batch_size`: Batch size
* `learning_rate_multiplier`: Learning rate multiplier
