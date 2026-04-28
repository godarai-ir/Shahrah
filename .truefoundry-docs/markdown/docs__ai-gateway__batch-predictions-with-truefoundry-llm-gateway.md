> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Batch API (/batches)

> Process jobs asynchronously using TrueFoundry's batch prediction API

## Provider capabilities

The table below summarizes gateway support for this endpoint by provider.

<Info>
  Legend:

  * **✅** Supported by Provider and Truefoundry
  * <Icon icon="circle-xmark" iconType="regular" color="red" /> Provided by provider, but not by Truefoundry
  * <Icon icon="circle-minus" iconType="regular" /> Provider does not support this feature
</Info>

| Provider     | Batch                                                       |
| ------------ | ----------------------------------------------------------- |
| OpenAI       | ✅                                                           |
| Azure OpenAI | ✅                                                           |
| Anthropic    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Bedrock      | ✅                                                           |
| Vertex       | ✅                                                           |
| Cohere       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Gemini       | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| Groq         | <Icon icon="circle-minus" iconType="regular" />             |
| Cerebras     | <Icon icon="circle-minus" iconType="regular" />             |
| Together-AI  | <Icon icon="circle-xmark" iconType="regular" color="red" /> |
| xAI          | <Icon icon="circle-minus" iconType="regular" />             |
| DeepInfra    | <Icon icon="circle-xmark" iconType="regular" color="red" /> |

For every gateway endpoint and provider, see [Supported APIs](/docs/ai-gateway/intro-to-llm-gateway#supported-apis).

This guide explains how to perform batch predictions using TrueFoundry's AI Gateway with OpenAI, Azure OpenAI, Vertex AI, or AWS Bedrock providers.

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

  ```python Azure OpenAI Provider lines theme={"dark"}
  from openai import OpenAI

  client = OpenAI(
      api_key="your-truefoundry-api-key",
      base_url="{GATEWAY_BASE_URL}",
      default_headers={
          "x-tfy-provider-name": "azure-batch",  # truefoundry provider integration name
          "x-tfy-azure-api-version": "2024-12-01-preview"  # api version for the model being used
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
          "x-tfy-vertex-region": "your-region",  # e.g., "europe-west4"
          "x-tfy-provider-model": "gemini-2-0-flash"
      }
  )
  ```

  ```python AWS Bedrock Provider lines theme={"dark"}
  from openai import OpenAI

  client = OpenAI(
      api_key="your-truefoundry-api-key",
      base_url="{GATEWAY_BASE_URL}",
      default_headers={
          "x-tfy-provider-name": "bedrock-main",  # truefoundry provider integration name
          "x-tfy-aws-s3-bucket": "your-s3-bucket-name",
          "x-tfy-aws-bedrock-model": "amazon.nova-lite-v1:0"
      }
  )
  ```
</CodeGroup>

## Input File Format

Create a JSONL file with one JSON object per line. Each line represents a single request:

<CodeGroup>
  ```json request.jsonl lines theme={"dark"}
  {"custom_id": "request-1", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "gpt-3.5-turbo-0125", "messages": [{"role": "system", "content": "You are a helpful assistant."},{"role": "user", "content": "Hello world!"}],"max_tokens": 1000}}
  {"custom_id": "request-2", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "gpt-3.5-turbo-0125", "messages": [{"role": "system", "content": "You are an unhelpful assistant."},{"role": "user", "content": "Hello world!"}],"max_tokens": 1000}}
  {"custom_id": "request-3", "method": "POST", "url": "/v1/chat/completions", "body": {"model": "gpt-4-vision-preview", "messages": [{"role": "user", "content": [{"type": "text", "text": "What's in this image?"}, {"type": "image_url", "image_url": {"url": "https://example.com/image.jpg"}}]}], "max_tokens": 1000}}
  ```
</CodeGroup>

**Requirements:** Valid JSON per line, meaningful `custom_id` values. See provider notes below for any provider-specific limits (e.g. minimum records, URL/body format).

<Note>
  Before using AWS Bedrock batch processing, ensure you have:

  * **S3 Bucket**: For storing input and output files
  * **IAM Execution Role**: With permissions for S3 access and Bedrock model invocation
  * **User Permissions**: Including `iam:PassRole` to pass the execution role to Bedrock
</Note>

<Note>
  Before using Vertex AI batch processing, ensure you have:

  * **Cloud Storage Bucket**: The bucket must be in the same region as the model, and the service account must have read/write access to the bucket
  * **Batch Prediction Permissions**: The service account must have permission to create batch prediction jobs
</Note>

<Note>
  Before using Azure OpenAI batch processing, ensure you have:

  * **Deployment type**: A model deployed as **Global Batch** or **Data Zone Batch** (standard/online deployments cannot be used for batch).
  * **Headers**: Set `x-tfy-azure-api-version` (e.g. `2024-12-01-preview`) to match the API version used by your batch deployment.
  * **Endpoint**: When creating the job, use `endpoint="/chat/completions"` for chat (Azure API uses this form). For the Responses API, use `url: "/v1/responses"` and `body.input` in each JSONL line.

  See [Azure OpenAI batch documentation](https://learn.microsoft.com/en-us/azure/ai-services/openai/how-to/batch) for details.
</Note>

## Workflow Steps

The batch process follows these steps for all providers:

1. **Upload**: Upload JSONL file → Get file ID
2. **Create**: Create batch job → Get batch ID
3. **Monitor**: Check status until complete
4. **Fetch**: Download results

### Step-by-Step Examples

<AccordionGroup>
  <Accordion title="1. Upload Input File" icon="upload">
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

      # Upload the input file
      file = client.files.create(
          file=open("request.jsonl", "rb"),
          purpose="batch"
      )

      print(file.id)  # Example: file-PnFGrFLN5LjjcWr4eFsStK
      ```

      ```python Azure OpenAI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "azure-batch",
              "x-tfy-azure-api-version": "2024-12-01-preview"
          }
      )

      # Upload the input file (purpose must be "batch")
      file = client.files.create(
          file=open("request.jsonl", "rb"),
          purpose="batch"
      )

      print(file.id)  # Example: file-xxxxx
      ```

      ```python Vertex AI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "vertex-main",  # truefoundry provider integration name
              "x-tfy-vertex-storage-bucket-name": "your-bucket-name",
              "x-tfy-vertex-region": "your-region",
              "x-tfy-provider-model": "gemini-2-0-flash"
          }
      )

      # Upload the input file
      file = client.files.create(
          file=open("request.jsonl", "rb"),
          purpose="batch"
      )

      print(file.id)  # Example: file-PnFGrFLN5LjjcWr4eFsStK
      ```

      ```python AWS Bedrock Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "bedrock-main",  # truefoundry provider integration name
              "x-tfy-aws-s3-bucket": "your-s3-bucket-name",
              "x-tfy-aws-bedrock-model": "amazon.nova-lite-v1:0"
          }
      )

      # Upload the input file
      file = client.files.create(
          file=open("bedrock-requests.jsonl", "rb"),
          purpose="batch"
      )

      print(file.id)  # Example: s3://bucket/uuid.jsonl
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="2. Create Batch Job" icon="play">
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

      batch_job = client.batches.create(
          input_file_id=file.id,
          endpoint="/v1/chat/completions",
          completion_window="24h"
      )

      print(batch_job.id)  # Example: batch_67f7bfc50b288190893f242d9fa47c52
      ```

      ```python Azure OpenAI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "azure-batch",
              "x-tfy-azure-api-version": "2024-12-01-preview"
          }
      )

      # Azure uses endpoint="/chat/completions"; completion_window must be "24h"
      batch_job = client.batches.create(
          input_file_id=file.id,
          endpoint="/chat/completions",
          completion_window="24h"
      )

      print(batch_job.id)  # Example: batch_xxxxx
      ```

      ```python Vertex AI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "vertex-main",  # truefoundry provider integration name
              "x-tfy-vertex-storage-bucket-name": "your-bucket-name",
              "x-tfy-vertex-region": "your-region",
              "x-tfy-provider-model": "gemini-2-0-flash"
          }
      )

      batch_job = client.batches.create(
          input_file_id=file.id,
          endpoint="/v1/chat/completions",
          completion_window="24h"
      )

      print(batch_job.id)  # Example: batch_67f7bfc50b288190893f242d9fa47c52
      ```

      ```python AWS Bedrock Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "bedrock-main",  # truefoundry provider integration name
              "x-tfy-aws-s3-bucket": "your-s3-bucket-name",
              "x-tfy-aws-bedrock-model": "amazon.nova-lite-v1:0"
          }
      )

      # Bedrock requires extra_body with additional parameters
      role_arn = "arn:aws:iam::YOUR_ACCOUNT_ID:role/BedrockBatchExecutionRole"

      batch_job = client.batches.create(
          input_file_id=file.id,  # S3 URL from file upload
          endpoint="/v1/chat/completions",
          completion_window="24h",
          extra_body={
              "model": "bedrock-main/amazon-nova-lite-v1-0",
              "role_arn": role_arn,
              "job_name": "my-custom-batch-job"  # Optional
          }
      )

      print(batch_job.id)  # Example: arn:aws:bedrock:us-east-1:ACCOUNT:model-invocation-job/JOB_ID
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="3. Check Batch Status" icon="eye">
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

      batch_status = client.batches.retrieve(batch_job.id)
      print(batch_status.status)  # Example: completed, validating, in_progress, etc.
      ```

      ```python Azure OpenAI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "azure-batch",
              "x-tfy-azure-api-version": "2024-12-01-preview"
          }
      )

      batch_status = client.batches.retrieve(batch_job.id)
      print(batch_status.status)  # Example: completed, validating, in_progress, finalizing, etc.
      ```

      ```python Vertex AI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "vertex-main",  # truefoundry provider integration name
              "x-tfy-vertex-storage-bucket-name": "your-bucket-name",
              "x-tfy-vertex-region": "your-region",
              "x-tfy-provider-model": "gemini-2-0-flash"
          }
      )

      batch_status = client.batches.retrieve(batch_job.id)
      print(batch_status.status)  # Example: completed, validating, in_progress, etc.
      ```

      ```python AWS Bedrock Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "bedrock-main",  # truefoundry provider integration name
              "x-tfy-aws-s3-bucket": "your-s3-bucket-name",
              "x-tfy-aws-bedrock-model": "amazon.nova-lite-v1:0"
          }
      )

      batch_status = client.batches.retrieve(batch_job.id)
      print(batch_status.status)  # Example: completed, validating, in_progress, etc.
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="4. Fetch Results" icon="download">
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

      if batch_status.status == "completed":
          output_content = client.files.content(batch_status.output_file_id)
          print(output_content.content.decode('utf-8'))
      ```

      ```python Azure OpenAI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "azure-batch",
              "x-tfy-azure-api-version": "2024-12-01-preview"
          }
      )

      if batch_status.status == "completed":
          output_content = client.files.content(batch_status.output_file_id)
          print(output_content.content.decode('utf-8'))
      ```

      ```python Vertex AI Provider lines theme={"dark"}
      from openai import OpenAI

      client = OpenAI(
          api_key="your-truefoundry-api-key",
          base_url="{GATEWAY_BASE_URL}",
          default_headers={
              "x-tfy-provider-name": "vertex-main",  # truefoundry provider integration name
              "x-tfy-vertex-storage-bucket-name": "your-bucket-name",
              "x-tfy-vertex-region": "your-region",
              "x-tfy-provider-model": "gemini-2-0-flash"
          }
      )

      if batch_status.status == "completed":
          output_content = client.files.content(batch_status.output_file_id)
          print(output_content.content.decode('utf-8'))
      ```

      ```python AWS Bedrock Provider lines theme={"dark"}
      import requests

      # For Bedrock, use REST API to fetch results
      if batch_status.status == "completed":

      response = requests.get(
          f"{{GATEWAY_BASE_URL}}/batches/{batch_job.id}/output",
          headers={
              "Authorization": "Bearer your-truefoundry-api-key",
              "x-tfy-provider-name": "bedrock-main"
          }
      )

      print("Batch results:")
      print(response.json())
      ```
    </CodeGroup>
  </Accordion>
</AccordionGroup>

## Batch Status Reference

* `validating`: Initial validation of the batch
* `in_progress`: Processing the requests
* `completed`: All requests processed successfully
* `failed`: Batch processing failed (use `error_file_id` from the batch response to fetch error details)
* `finalizing`: Results being prepared (OpenAI / Azure OpenAI)
* `expired`: Did not complete within the time window (OpenAI / Azure OpenAI)
* `cancelling` / `cancelled`: Batch cancelled; partial results may be available via `output_file_id`

## Best Practices

1. **File Format**: Use meaningful `custom_id` values and valid JSONL format
2. **Error Handling**: Implement proper error handling and status monitoring
3. **Security**: Store API keys securely, use minimal IAM permissions
4. **Provider-specific**: Check the prerequisites above (e.g. Azure: batch deployment type and endpoint; Bedrock: minimum 100 records, IAM and S3)

## Vertex AI Permissions

The following permissions are required for Vertex AI batch prediction:

<AccordionGroup>
  <Accordion title="Storage and service account requirements" icon="folder">
    **Cloud Storage bucket**

    * The Cloud Storage bucket used for batch input and output must be in the **same region** as the Vertex AI model you are using.
    * The service account that runs the batch job must have **read and write** access to this bucket (e.g. `roles/storage.objectAdmin` or equivalent object-level read/write permissions on the bucket).
  </Accordion>

  <Accordion title="Batch prediction job permissions" icon="play">
    **Create batch prediction jobs**

    * The service account must have permission to **create and manage batch prediction jobs** in Vertex AI (e.g. `roles/aiplatform.user` or the `aiplatform.batchPredictionJobs.create` permission).
  </Accordion>
</AccordionGroup>

## AWS Bedrock Permissions

<AccordionGroup>
  <Accordion title="User Permissions (for API calls)" icon="user">
    These are the minimum permissions required to use the Bedrock Batch APIs. For complete official guidance, see [AWS Bedrock Batch Inference Permissions](https://docs.aws.amazon.com/bedrock/latest/userguide/batch-inference-permissions.html).

    <CodeGroup>
      ```json User Policy lines theme={"dark"}
      {
        "Version": "2012-10-17",
        "Statement": [
          {
            "Effect": "Allow",
            "Action": [
              "bedrock:ListFoundationModels",
              "bedrock:GetFoundationModel",
              "bedrock:ListInferenceProfiles",
              "bedrock:GetInferenceProfile",
              "bedrock:ListCustomModels",
              "bedrock:GetCustomModel",
              "bedrock:TagResource",
              "bedrock:UntagResource",
              "bedrock:ListTagsForResource",
              "bedrock:CreateModelInvocationJob",
              "bedrock:GetModelInvocationJob",
              "bedrock:ListModelInvocationJobs",
              "bedrock:StopModelInvocationJob"
            ],
            "Resource": [
              "arn:aws:bedrock:<region>:<account_id>:model-customization-job/*",
              "arn:aws:bedrock:<region>:<account_id>:custom-model/*",
              "arn:aws:bedrock:<region>::foundation-model/*"
            ]
          },
          {
            "Effect": "Allow",
            "Action": ["s3:ListBucket", "s3:PutObject", "s3:GetObject", "s3:GetObjectAttributes"],
            "Resource": ["arn:aws:s3:::<bucket>", "arn:aws:s3:::<bucket>/*"]
          },
          {
            "Action": ["iam:PassRole"],
            "Effect": "Allow",
            "Resource": "arn:aws:iam::<account_id>:role/<service_role_name>",
            "Condition": {
              "StringEquals": {
                "iam:PassedToService": ["bedrock.amazonaws.com"]
              }
            }
          }
        ]
      }
      ```
    </CodeGroup>
  </Accordion>

  <Accordion title="Service Role Permissions (for batch execution)" icon="shield">
    The service role (role\_arn) used for creating and executing the batch job requires:

    **Trust Relationship:**

    <CodeGroup>
      ```json Trust Policy lines theme={"dark"}
      {
        "Version": "2012-10-17",
        "Statement": [
          {
            "Effect": "Allow",
            "Principal": {
              "Service": "bedrock.amazonaws.com"
            },
            "Action": "sts:AssumeRole",
            "Condition": {
              "StringEquals": {
                "aws:SourceAccount": "<account_id>"
              },
              "ArnEquals": {
                "aws:SourceArn": "arn:aws:bedrock:<region>:<account_id>:model-invocation-job/*"
              }
            }
          }
        ]
      }
      ```
    </CodeGroup>

    **Permission Policy:**

    <CodeGroup>
      ```json Permission Policy lines theme={"dark"}
      {
        "Version": "2012-10-17",
        "Statement": [
          {
            "Effect": "Allow",
            "Action": ["s3:GetObject", "s3:PutObject", "s3:ListBucket"],
            "Resource": ["arn:aws:s3:::<bucket>", "arn:aws:s3:::<bucket>/*"]
          }
        ]
      }
      ```
    </CodeGroup>
  </Accordion>
</AccordionGroup>
