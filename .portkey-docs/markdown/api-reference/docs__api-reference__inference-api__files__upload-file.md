> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Upload File



## OpenAPI

````yaml post /files
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
  /files:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_GATEWAY_URL
        description: Self-Hosted Gateway URL
    post:
      tags:
        - Files
      summary: >
        Upload a file to be used across various endpoints, such as Assistant
        (<2M tokens), Fine-Tuning, and Batch (<100 MB). Total size of your
        bucket is 100 GB.
      operationId: createFile
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              $ref: '#/components/schemas/CreateFileRequest'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OpenAIFile'
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
          source: |
            curl https://api.portkey.ai/v1/files \
              -H "x-portkey-api-key: $PORTKEY_API_KEY" \
              -H "x-portkey-virtual-key: $PORTKEY_PROVIDER_VIRTUAL_KEY" \
              -F purpose="fine-tune" \
              -F file="@mydata.jsonl"
        - lang: python
          source: |
            from portkey_ai import Portkey

            client = Portkey(
              api_key = "PORTKEY_API_KEY",
              virtual_key = "PROVIDER_VIRTUAL_KEY"
            )

            client.files.create(
              file=open("mydata.jsonl", "rb"),
              purpose="fine-tune"
            )
        - lang: javascript
          source: |
            import fs from "fs";
            import Portkey from 'portkey-ai';

            const client = new Portkey({
              apiKey: 'PORTKEY_API_KEY',
              virtualKey: 'PROVIDER_VIRTUAL_KEY'
            });

            async function main() {
              const file = await client.files.create({
                file: fs.createReadStream("mydata.jsonl"),
                purpose: "fine-tune",
              });

              console.log(file);
            }

            main();
components:
  schemas:
    CreateFileRequest:
      type: object
      additionalProperties: false
      properties:
        file:
          description: |
            The File object (not file name) to be uploaded.
          type: string
          format: binary
        purpose:
          description: >
            The intended purpose of the uploaded file.


            Use "assistants" for
            [Assistants](https://platform.openai.com/docs/api-reference/assistants)
            and
            [Message](https://platform.openai.com/docs/api-reference/messages)
            files, "vision" for Assistants image file inputs, "batch" for [Batch
            API](https://platform.openai.com/docs/guides/batch), and "fine-tune"
            for
            [Fine-tuning](https://platform.openai.com/docs/api-reference/fine-tuning).
          type: string
          enum:
            - assistants
            - batch
            - fine-tune
            - vision
      required:
        - file
        - purpose
    OpenAIFile:
      title: OpenAIFile
      description: >-
        The `File` object represents a document that has been uploaded to
        OpenAI.
      properties:
        id:
          type: string
          description: The file identifier, which can be referenced in the API endpoints.
        bytes:
          type: integer
          description: The size of the file, in bytes.
        created_at:
          type: integer
          description: The Unix timestamp (in seconds) for when the file was created.
        filename:
          type: string
          description: The name of the file.
        object:
          type: string
          description: The object type, which is always `file`.
          enum:
            - file
        purpose:
          type: string
          description: >-
            The intended purpose of the file. Supported values are `assistants`,
            `assistants_output`, `batch`, `batch_output`, `fine-tune`,
            `fine-tune-results` and `vision`.
          enum:
            - assistants
            - assistants_output
            - batch
            - batch_output
            - fine-tune
            - fine-tune-results
            - vision
        status:
          type: string
          deprecated: true
          description: >-
            Deprecated. The current status of the file, which can be either
            `uploaded`, `processed`, or `error`.
          enum:
            - uploaded
            - processed
            - error
        status_details:
          type: string
          deprecated: true
          description: >-
            Deprecated. For details on why a fine-tuning training file failed
            validation, see the `error` field on `fine_tuning.job`.
      required:
        - id
        - object
        - bytes
        - created_at
        - filename
        - purpose
        - status
      x-code-samples:
        name: The file object
        example: |
          {
            "id": "file-abc123",
            "object": "file",
            "bytes": 120000,
            "created_at": 1677610602,
            "filename": "salesOverview.pdf",
            "purpose": "assistants",
          }
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