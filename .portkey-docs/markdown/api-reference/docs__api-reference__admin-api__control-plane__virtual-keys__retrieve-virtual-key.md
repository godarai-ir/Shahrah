> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Retrieve Virtual Key

<Warning>
  **Deprecated.** Use the [Providers API](/api-reference/admin-api/control-plane/providers/retrieve-provider) instead. Existing virtual keys continue to work — no code changes needed.
</Warning>


## OpenAPI

````yaml get /virtual-keys/{slug}
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
  /virtual-keys/{slug}:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_CONTROL_PLANE_URL
        description: Self-Hosted Control Plane URL
    get:
      tags:
        - Virtual-keys
      summary: Get a Virtual Key
      parameters:
        - name: slug
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/VirtualKeys'
        '401':
          description: Unauthorized response
          content:
            application/json:
              schema:
                type: object
                properties:
                  success:
                    type: boolean
                  data:
                    type: object
                    properties:
                      message:
                        type: string
              example:
                success: false
                data:
                  message: Unauthorised Request
      x-code-samples:
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            # Initialize the Portkey client
            portkey = Portkey(
                api_key="PORTKEY_API_KEY",
            )

            # Get a specific virtual key
            virtual_key = portkey.virtual_keys.retrieve(
                slug='VIRTUAL_KEY_SLUG'
            )

            print(virtual_key)
        - lang: javascript
          label: Default
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
            })

            const vKey=await portkey.virtualKeys.retrieve({
                slug:'VIRTUAL_KEY_SLUG'
            })
            console.log(vKey);
        - lang: curl
          label: Default
          source: >
            curl -X GET https://api.portkey.ai/v1/virtual-keys/VIRTUAL_KEY_SLUG
            \

            -H "x-portkey-api-key: PORTKEY_API_KEY"
        - lang: curl
          label: Self-Hosted
          source: >
            curl -X GET
            SELF_HOSTED_CONTROL_PLANE_URL/virtual-keys/VIRTUAL_KEY_SLUG \

            -H "x-portkey-api-key: PORTKEY_API_KEY"
        - lang: python
          label: Self-Hosted
          source: |
            from portkey_ai import Portkey

            # Initialize the Portkey client
            portkey = Portkey(
                api_key="PORTKEY_API_KEY",
                base_url="SELF_HOSTED_CONTROL_PLANE_URL"
            )

            # Get a specific virtual key
            virtual_key = portkey.virtual_keys.retrieve(
                slug='VIRTUAL_KEY_SLUG'
            )

            print(virtual_key)
        - lang: javascript
          label: Self-Hosted
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
                baseUrl: "SELF_HOSTED_CONTROL_PLANE_URL"
            })

            const vKey=await portkey.virtualKeys.retrieve({
                slug:'VIRTUAL_KEY_SLUG'
            })
            console.log(vKey);
components:
  schemas:
    VirtualKeys:
      type: object
      properties:
        name:
          type: string
          example: Open AI Workspace
        note:
          type: string
          nullable: true
          example: randomness
        status:
          type: string
          enum:
            - active
            - exhausted
        usage_limits:
          $ref: '#/components/schemas/UsageLimits'
        reset_usage:
          type: number
          nullable: true
          example: 0
        created_at:
          type: string
          format: date-time
        slug:
          type: string
        model_config:
          type: object
        rate_limits:
          type: array
          nullable: true
          items:
            $ref: '#/components/schemas/RateLimits'
        expires_at:
          type: string
          format: date-time
        secret_mappings:
          type: array
          items:
            $ref: '#/components/schemas/SecretMapping'
          description: >-
            Secret reference mappings for this virtual key. Valid target_field
            values are "key" or "model_config.<field>".
        object:
          type: string
          enum:
            - virtual-key
    UsageLimits:
      type: object
      properties:
        credit_limit:
          type: integer
          description: Credit Limit. Used for tracking usage
          minimum: 1
          default: null
        type:
          type: string
          description: Type of credit limit
          enum:
            - cost
            - tokens
        alert_threshold:
          type: integer
          description: Alert Threshold. Used for alerting when usage reaches more than this
          minimum: 1
          default: null
        periodic_reset:
          type: string
          description: Reset the usage periodically.
          enum:
            - monthly
            - weekly
          nullable: true
          example: monthly
        periodic_reset_days:
          type: integer
          description: >-
            Reset the usage counter every N days (1-365). Mutually exclusive
            with periodic_reset.
          minimum: 1
          maximum: 365
          nullable: true
          example: 30
        next_usage_reset_at:
          type: string
          format: date-time
          description: >-
            ISO 8601 datetime for the next scheduled usage reset. Auto-computed
            from periodic_reset or periodic_reset_days if not provided.
          nullable: true
          example: '2026-05-01T00:00:00Z'
      example:
        credit_limit: 10
        periodic_reset: monthly
        alert_threshold: 8
    RateLimits:
      type: object
      properties:
        type:
          type: string
          enum:
            - requests
            - tokens
        unit:
          type: string
          enum:
            - rpd
            - rph
            - rpm
        value:
          type: integer
    SecretMapping:
      type: object
      required:
        - target_field
        - secret_reference_id
      properties:
        target_field:
          type: string
          description: >
            The field on the entity to populate from the secret reference. Must
            be unique within the array.

            - **Integrations**: `key` or `configurations.<field>` (e.g.
            `configurations.aws_secret_access_key`)

            - **Virtual Keys**: `key` or `model_config.<field>` (e.g.
            `model_config.awsSecretAccessKey`)
          example: key
        secret_reference_id:
          type: string
          description: >-
            UUID or slug of the secret reference. Must belong to the same
            organisation and be accessible by the workspace.
          example: my-aws-secret
        secret_key:
          type: string
          nullable: true
          description: >-
            Override the secret_key defined on the secret reference. Use to pick
            a specific key from a multi-value secret.
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````