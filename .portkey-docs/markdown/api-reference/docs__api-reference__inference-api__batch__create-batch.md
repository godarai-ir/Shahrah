> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Batch



## OpenAPI

````yaml post /batches
openapi: 3.0.0
info:
  title: Portkey API
  description: >-
    The Portkey REST API. Please see https://portkey.ai/docs/api-reference for
    more details.
  version: 2.0.0
  termsOfService: https://portkey.ai/terms
  contact:
    name: Portkey Developer Forum
    url: https://portkey.wiki/community
  license:
    name: MIT
    url: https://github.com/Portkey-AI/portkey-openapi/blob/master/LICENSE
servers:
  - url: https://api.portkey.ai/v1
    description: Portkey API Public Endpoint
security:
  - Portkey-Key: []
tags:
  - name: Assistants
    description: Build Assistants that can call models and use tools.
  - name: Audio
    description: Turn audio into text or text into audio.
  - name: Chat
    description: >-
      Given a list of messages comprising a conversation, the model will return
      a response.
  - name: Realtime
    description: WebSocket proxy for provider Realtime APIs
  - name: Collections
    description: Create, List, Retrieve, Update, and Delete collections of prompts.
  - name: Labels
    description: Create, List, Retrieve, Update, and Delete labels.
  - name: Prompt Collections
    description: Create, List, Retrieve, Update, and Delete prompt collections.
  - name: PromptPartials
    description: Create, List, Retrieve, Update, and Delete prompt partials.
  - name: Prompts
    description: >-
      Given a prompt template ID and variables, will run the saved prompt
      template and return a response.
  - name: Guardrails
    description: Create, List, Retrieve, Update, and Delete prompt Guardrails.
  - name: Completions
    description: >-
      Given a prompt, the model will return one or more predicted completions,
      and can also return the probabilities of alternative tokens at each
      position.
  - name: Embeddings
    description: >-
      Get a vector representation of a given input that can be easily consumed
      by machine learning models and algorithms.
  - name: Rerank
    description: >-
      Rerank a list of documents based on their relevance to a query. Supported
      providers include Cohere, Voyage, Jina, Pinecone, Bedrock, and Azure AI.
  - name: Fine-tuning
    description: Manage fine-tuning jobs to tailor a model to your specific training data.
  - name: Batch
    description: Create large batches of API requests to run asynchronously.
  - name: Files
    description: >-
      Files are used to upload documents that can be used with features like
      Assistants and Fine-tuning.
  - name: Images
    description: Given a prompt and/or an input image, the model will generate a new image.
  - name: Models
    description: List and describe the various models available in the API.
  - name: Moderations
    description: >-
      Given a input text, outputs if the model classifies it as potentially
      harmful.
  - name: Configs
    description: Create, List, Retrieve, and Update your Portkey Configs.
  - name: Feedback
    description: Send and Update any feedback.
  - name: Logs
    description: Custom Logger to add external logs to Portkey.
  - name: Integrations
    description: Create, List, Retrieve, Update, and Delete your Portkey Integrations.
  - name: Integrations > Workspaces
    description: Manage workspace access for your Portkey Integrations.
  - name: Integrations > Models
    description: Manage model access for your Portkey Integrations.
  - name: Providers
    description: Create, List, Retrieve, Update, and Delete your Portkey Providers.
  - name: Virtual-keys
    description: Create, List, Retrieve, Update, and Delete your Portkey Virtual keys.
  - name: Users
    description: Create and manage users.
  - name: User-invites
    description: Create and manage user invites.
  - name: Workspaces
    description: Create and manage workspaces.
  - name: Workspaces > Members
    description: Create and manage workspace members.
  - name: MCP Integrations
    description: Create, List, Retrieve, Update, and Delete MCP Integrations.
  - name: MCP Integrations > Workspaces
    description: Manage workspace access for MCP Integrations.
  - name: MCP Integrations > Capabilities
    description: List and manage capabilities for MCP Integrations.
  - name: MCP Integrations > Metadata
    description: Get MCP Integration metadata and sync info.
  - name: MCP Servers
    description: >-
      Create, List, Retrieve, Update, and Delete MCP Servers (workspace
      instances of MCP Integrations).
  - name: MCP Servers > Capabilities
    description: List and manage capabilities for MCP Servers.
  - name: MCP Servers > User Access
    description: List and manage user access for MCP Servers.
  - name: MCP Servers > Connections
    description: List and manage user connections for MCP Servers.
  - name: Api-Keys
    description: Create, List, Retrieve, Update, and Delete your Portkey API keys.
  - name: Logs Export
    description: Exports logs service.
  - name: Audit Logs
    description: Get audit logs for your Portkey account.
  - name: Analytics
    description: >-
      Get analytics over different data points like requests, costs, tokens,
      etc.
  - name: Analytics > Graphs
    description: Get data points for graphical representation.
  - name: Analytics > Summary
    description: Get overall summary for the selected time bucket.
  - name: Analytics > Groups
    description: Get grouped metrics for the selected time bucket.
  - name: Usage Limits Policies
    description: Manage usage limits policies to control total usage over time
  - name: Rate Limits Policies
    description: Manage rate limits policies to control request or token rates
  - name: Model Pricing
    description: Model pricing configurations for 2300+ LLMs across 40+ providers
  - name: Secret-References
    description: >-
      Create, List, Retrieve, Update, and Delete secret references to external
      secret managers.
paths:
  /batches:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Batch
      summary: Creates and executes a batch from an uploaded file of requests
      operationId: createBatch
      requestBody:
        required: true
        content:
          application/json:
            schema:
              anyOf:
                - $ref: '#/components/schemas/OpenAIBatchJob'
                - $ref: '#/components/schemas/BedrockBatchJob'
                - $ref: '#/components/schemas/VertexBatchJob'
                - $ref: '#/components/schemas/PortkeyBatchJob'
      responses:
        '200':
          description: Batch created successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Batch'
      security:
        - Portkey-Key: []
          Virtual-Key: []
        - Portkey-Key: []
          Provider-Auth: []
          Provider-Name: []
        - Portkey-Key: []
          Config: []
        - Portkey-Key: []
          Provider-Auth: []
          Provider-Name: []
          Custom-Host: []
      x-code-samples:
        - lang: curl
          label: Default
          source: |
            curl https://api.portkey.ai/v1/batches \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -H "Content-Type: application/json" \
              -d '{
                "input_file_id": "file-abc123",
                "endpoint": "/v1/chat/completions",
                "completion_window": "24h"
              }'
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            client.batches.create(
              input_file_id="file-abc123",
              endpoint="/v1/chat/completions",
              completion_window="24h"
            )
        - lang: javascript
          label: Default
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const batch = await client.batches.create({
                input_file_id: "file-abc123",
                endpoint: "/v1/chat/completions",
                completion_window: "24h"
              });

              console.log(batch);
            }

            main();
        - lang: curl
          label: Self-Hosted
          source: |
            curl SELF_HOSTED_GATEWAY_URL/batches \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -H "Content-Type: application/json" \
              -d '{
                "input_file_id": "file-abc123",
                "endpoint": "/v1/chat/completions",
                "completion_window": "24h"
              }'
        - lang: python
          label: Self-Hosted
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              base_url = "SELF_HOSTED_GATEWAY_URL",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            client.batches.create(
              input_file_id="file-abc123",
              endpoint="/v1/chat/completions",
              completion_window="24h"
            )
        - lang: javascript
          label: Self-Hosted
          source: |
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              baseUrl: 'SELF_HOSTED_GATEWAY_URL',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const batch = await client.batches.create({
                input_file_id: "file-abc123",
                endpoint: "/v1/chat/completions",
                completion_window: "24h"
              });

              console.log(batch);
            }

            main();
components:
  schemas:
    OpenAIBatchJob:
      type: object
      required:
        - input_file_id
        - completion_window
        - endpoint
      properties:
        input_file_id:
          type: string
          description: The input file to use for the batch job
        completion_window:
          type: string
          enum:
            - immediate
            - 24h
          description: >-
            Completion window for the batch job, `immediate` is only supported
            with Portkey Managed Batching.
        endpoint:
          type: string
          enum:
            - /v1/chat/completions
            - /v1/completions
            - /v1/embeddings
          description: Inference endpoint
        metadata:
          description: metadata related for the batch job
          type: object
          additionalProperties: true
          nullable: true
      description: Gateway supported body params for OpenAI, Azure OpenAI and VertexAI.
      title: OpenAI Params
    BedrockBatchJob:
      type: object
      required:
        - model
        - role_arn
      properties:
        job_name:
          type: string
          description: Job name for the batch job
        output_data_config:
          type: string
          description: >-
            Batch job's output storage location, will be constructed based on
            `input_file_id` if not provided
        model:
          type: string
          description: Model to start batch job with
        role_arn:
          type: string
          description: Role ARN for the bedrock batch job
      allOf:
        - $ref: '#/components/schemas/OpenAIBatchJob'
      description: Gateway supported body params for bedrock fine-tuning.
      title: Bedrock Params
    VertexBatchJob:
      type: object
      required:
        - model
      properties:
        job_name:
          type: string
          description: Job name for the batch job
        output_data_config:
          type: string
          description: >-
            Batch job's output storage location, will be constructed based on
            `input_file_id` if not provided
        model:
          type: string
          description: Model to start batch job with
      allOf:
        - $ref: '#/components/schemas/OpenAIBatchJob'
      description: Gateway supported body params for Vertext fine-tuning.
      title: Vertex Params
    PortkeyBatchJob:
      type: object
      required:
        - model
      properties:
        job_name:
          type: string
          description: Job name for the batch job
        output_data_config:
          type: string
          description: >-
            Batch job's output storage location, will be constructed based on
            `input_file_id` if not provided
        model:
          type: string
          description: Model to start batch job with
        role_arn:
          type: string
          description: Role ARN for the bedrock batch job
        portkey_options:
          allOf:
            - $ref: '#/components/schemas/PortkeyBatchOptions'
          description: >-
            Portkey Gateway Provider specific headers to be passed to the
            provider, if portkey is used as a provider
        provider_options:
          anyOf:
            - type: object
              title: Bedrock Options
              properties:
                job_name:
                  type: string
                  description: Job name for the batch job
                output_data_config:
                  type: string
                  description: >-
                    Batch job's output storage location, will be constructed
                    based on `input_file_id` if not provided
                model:
                  type: string
                  description: Model to start batch job with
                role_arn:
                  type: string
                  description: Role ARN for the bedrock batch job
              required:
                - model
                - role_arn
            - type: object
              title: Vertex Options
              properties:
                job_name:
                  type: string
                  description: Job name for the batch job
                output_data_config:
                  type: string
                  description: >-
                    Batch job's output storage location, will be constructed
                    based on `input_file_id` if not provided
                model:
                  type: string
                  description: Model to start batch job with
              required:
                - model
          description: >-
            Provider specific options to be passed to the provider, optional can
            be passed directly as well.
      allOf:
        - $ref: '#/components/schemas/OpenAIBatchJob'
      description: Gateway supported body params for portkey managed batching.
      title: Portkey Params
    Batch:
      type: object
      properties:
        id:
          type: string
        object:
          type: string
          enum:
            - batch
          description: The object type, which is always `batch`.
        endpoint:
          type: string
          description: The Portkey API endpoint used by the batch.
        errors:
          type: object
          properties:
            object:
              type: string
              description: The object type, which is always `list`.
            data:
              type: array
              items:
                type: object
                properties:
                  code:
                    type: string
                    description: An error code identifying the error type.
                  message:
                    type: string
                    description: >-
                      A human-readable message providing more details about the
                      error.
                  param:
                    type: string
                    description: >-
                      The name of the parameter that caused the error, if
                      applicable.
                    nullable: true
                  line:
                    type: integer
                    description: >-
                      The line number of the input file where the error
                      occurred, if applicable.
                    nullable: true
        input_file_id:
          type: string
          description: The ID of the input file for the batch.
        completion_window:
          type: string
          description: The time frame within which the batch should be processed.
        status:
          type: string
          description: The current status of the batch.
          enum:
            - validating
            - failed
            - in_progress
            - finalizing
            - completed
            - expired
            - cancelling
            - cancelled
        output_file_id:
          type: string
          description: >-
            The ID of the file containing the outputs of successfully executed
            requests.
        error_file_id:
          type: string
          description: The ID of the file containing the outputs of requests with errors.
        created_at:
          type: integer
          description: The Unix timestamp (in seconds) for when the batch was created.
        in_progress_at:
          type: integer
          description: >-
            The Unix timestamp (in seconds) for when the batch started
            processing.
        expires_at:
          type: integer
          description: The Unix timestamp (in seconds) for when the batch will expire.
        finalizing_at:
          type: integer
          description: >-
            The Unix timestamp (in seconds) for when the batch started
            finalizing.
        completed_at:
          type: integer
          description: The Unix timestamp (in seconds) for when the batch was completed.
        failed_at:
          type: integer
          description: The Unix timestamp (in seconds) for when the batch failed.
        expired_at:
          type: integer
          description: The Unix timestamp (in seconds) for when the batch expired.
        cancelling_at:
          type: integer
          description: >-
            The Unix timestamp (in seconds) for when the batch started
            cancelling.
        cancelled_at:
          type: integer
          description: The Unix timestamp (in seconds) for when the batch was cancelled.
        request_counts:
          type: object
          properties:
            total:
              type: integer
              description: Total number of requests in the batch.
            completed:
              type: integer
              description: Number of requests that have been completed successfully.
            failed:
              type: integer
              description: Number of requests that have failed.
          required:
            - total
            - completed
            - failed
          description: The request counts for different statuses within the batch.
        metadata:
          description: >
            Set of 16 key-value pairs that can be attached to an object. This
            can be useful for storing additional information about the object in
            a structured format. Keys can be a maximum of 64 characters long and
            values can be a maxium of 512 characters long.
          type: object
          x-oaiTypeLabel: map
          nullable: true
      required:
        - id
        - object
        - endpoint
        - input_file_id
        - completion_window
        - status
        - created_at
    PortkeyBatchOptions:
      type: object
      required:
        - x-portkey-virtual-key
      properties:
        x-portkey-virtual-key:
          type: string
          description: The virtual key to communicate with the provider
        x-portkey-aws-s3-bucket:
          type: string
          description: The AWS S3 bucket to use for file upload during finetune
        x-portkey-vertex-storage-bucket-name:
          type: string
          description: Google Storage bucket to use for file upload during finetune
        x-portkey-provider-model:
          type: string
          description: >-
            Model to use for the batch job also for file transformation for
            model specific inference input.
      example:
        x-portkey-virtual-key: vkey-1234567890
        x-portkey-aws-s3-bucket: my-bucket
        x-portkey-provider-model: meta.llama3-1-8b-instruct-v1:0
        x-portkey-vertex-storage-bucket-name: my-bucket
      description: >-
        Options to be passed to the provider, supports all options supported by
        the provider from gateway.
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key
    Virtual-Key:
      type: apiKey
      in: header
      name: x-portkey-virtual-key
    Provider-Auth:
      type: http
      scheme: bearer
    Provider-Name:
      type: apiKey
      in: header
      name: x-portkey-provider
    Config:
      type: apiKey
      in: header
      name: x-portkey-config
    Custom-Host:
      type: apiKey
      in: header
      name: x-portkey-custom-host

````