> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Update Prompt

> Update a prompt's metadata and/or create a new version with updated template content.

**Partial version updates:** Set `patch: true` to perform a partial update of version fields (`string`, `parameters`, `model`, `virtual_key`, `version_description`, `functions`, `tools`, `tool_choice`, `is_raw_template`, `prompt_metadata`). When enabled, any version fields omitted from the request are backfilled from the current latest version, allowing you to update only the fields you need. When `patch` is omitted or `false`, all version fields must be provided together (original strict validation).

**Metadata-only updates:** Fields like `name`, `collection_id`, `version_description`, and `virtual_key` can always be updated independently without affecting versioning.




## OpenAPI

````yaml put /prompts/{promptId}
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
  /prompts/{promptId}:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_CONTROL_PLANE_URL
        description: Self-Hosted Control Plane URL
    put:
      tags:
        - Prompts
      summary: Update a prompt
      description: >
        Update a prompt's metadata and/or create a new version with updated
        template content.


        **Partial version updates:** Set `patch: true` to perform a partial
        update of version fields (`string`, `parameters`, `model`,
        `virtual_key`, `version_description`, `functions`, `tools`,
        `tool_choice`, `is_raw_template`, `prompt_metadata`). When enabled, any
        version fields omitted from the request are backfilled from the current
        latest version, allowing you to update only the fields you need. When
        `patch` is omitted or `false`, all version fields must be provided
        together (original strict validation).


        **Metadata-only updates:** Fields like `name`, `collection_id`,
        `version_description`, and `virtual_key` can always be updated
        independently without affecting versioning.
      operationId: updatePrompt
      parameters:
        - name: promptId
          in: path
          required: true
          schema:
            type: string
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                patch:
                  type: boolean
                  description: >
                    When `true`, enables partial version updates. Missing
                    version fields (`string`, `parameters`, `model`) are
                    backfilled from the current latest version, so you only need
                    to provide the fields you want to change. When `false` or
                    omitted, the original strict validation is preserved for
                    backward compatibility.
                name:
                  type: string
                collection_id:
                  type: string
                string:
                  type: string
                  description: >-
                    The prompt template string. When `patch` is `true`, this
                    field is optional and will be inherited from the current
                    latest version if omitted.
                parameters:
                  type: object
                  description: >-
                    Model parameters (e.g. temperature, max_tokens). When
                    `patch` is `true`, this field is optional and will be
                    inherited from the current latest version if omitted.
                model:
                  type: string
                  description: >-
                    The model identifier. When `patch` is `true`, this field is
                    optional and will be inherited from the current latest
                    version if omitted.
                virtual_key:
                  type: string
                  description: >-
                    The virtual key to associate with this version. When `patch`
                    is `true`, this field is optional and will be inherited from
                    the current latest version if omitted.
                version_description:
                  type: string
                  description: >-
                    A human-readable description for this version. When `patch`
                    is `true`, this field is optional and will be inherited from
                    the current latest version if omitted.
                functions:
                  type: array
                  items:
                    type: object
                  description: >-
                    Function definitions available to the model. When `patch` is
                    `true`, this field is optional and will be inherited from
                    the current latest version if omitted.
                tools:
                  type: array
                  items:
                    type: object
                  description: >-
                    Tool definitions available to the model. When `patch` is
                    `true`, this field is optional and will be inherited from
                    the current latest version if omitted.
                tool_choice:
                  type: object
                  description: >-
                    Controls which tool the model uses. When `patch` is `true`,
                    this field is optional and will be inherited from the
                    current latest version if omitted.
                is_raw_template:
                  type: integer
                  enum:
                    - 0
                    - 1
                  description: >-
                    Whether the template string is raw (1) or processed (0).
                    When `patch` is `true`, this field is optional and will be
                    inherited from the current latest version if omitted.
                prompt_metadata:
                  type: object
                  description: >-
                    Additional metadata for the prompt version. When `patch` is
                    `true`, this field is optional and will be inherited from
                    the current latest version if omitted.
      responses:
        '200':
          description: Prompt updated successfully
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    format: uuid
                  slug:
                    type: string
                  prompt_version_id:
                    type: string
                    format: uuid
        '400':
          description: Bad request
        '401':
          description: Unauthorized
        '403':
          description: Forbidden
        '404':
          description: Prompt not found
        '500':
          description: Server error
      security:
        - Portkey-Key: []
components:
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````