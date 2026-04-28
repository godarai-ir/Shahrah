> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Update Provider



## OpenAPI

````yaml put /providers/{slug}
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
  /providers/{slug}:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_CONTROL_PLANE_URL
        description: Self-Hosted Control Plane URL
    put:
      tags:
        - Providers
      summary: Update a Provider
      parameters:
        - in: query
          name: workspace_id
          schema:
            type: string
          description: optional, needed when using organisation admin keys
        - in: path
          name: slug
          schema:
            type: string
          required: true
          description: Provider slug
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                note:
                  type: string
                usage_limits:
                  $ref: '#/components/schemas/UsageLimits'
                  nullable: true
                rate_limits:
                  $ref: '#/components/schemas/RateLimits'
                expires_at:
                  type: string
                  format: date-time
                  nullable: true
                reset_usage:
                  type: boolean
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    format: UUID
                  slug:
                    type: string
      x-code-samples:
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            # Initialize the Portkey client
            portkey = Portkey(
                api_key="PORTKEY_API_KEY",
            )

            # Update a specific provider
            updated_provider = portkey.providers.update(
                slug="PROVIDER_SLUG",
                name="updated-name",
                note="updated-note",
                workspace_id=""
            )

            print(updated_provider)
        - lang: javascript
          label: Default
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
            })

            const updatedProvider = await portkey.providers.update({
                slug: "PROVIDER_SLUG",
                name:"updated-name",
                note: "updated-note",
                workspace_id: ""
            })
            console.log(updatedProvider);
        - lang: curl
          label: Default
          source: >
            curl -X PUT
            "https://api.portkey.ai/v1/providers/PROVIDER_SLUG?workspace_id=" \

            -H "x-portkey-api-key: PORTKEY_API_KEY" \

            -H "Content-Type: application/json" \

            -d '{
                "name": "updated-name",
                "note": "updated-note"
            }'
        - lang: curl
          label: Self-Hosted
          source: >
            curl -X PUT
            "SELF_HOSTED_CONTROL_PLANE_URL/providers/PROVIDER_SLUG?workspace_id="
            \

            -H "x-portkey-api-key: PORTKEY_API_KEY" \

            -H "Content-Type: application/json" \

            -d '{
                "name": "updated-name",
                "note": "updated-note"
            }'
        - lang: python
          label: Self-Hosted
          source: |
            from portkey_ai import Portkey

            # Initialize the Portkey client
            portkey = Portkey(
                api_key="PORTKEY_API_KEY",
                base_url="SELF_HOSTED_CONTROL_PLANE_URL"
            )

            # Update a specific provider
            updated_provider = portkey.providers.update(
                slug="PROVIDER_SLUG",
                name="updated-name",
                note="updated-note",
                workspace_id=""
            )

            print(updated_provider)
        - lang: javascript
          label: Self-Hosted
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
                baseUrl: "SELF_HOSTED_CONTROL_PLANE_URL"
            })

            const updatedProvider = await portkey.providers.update({
                slug: "PROVIDER_SLUG",
                name:"updated-name",
                note: "updated-note",
                workspace_id: ""
            })
            console.log(updatedProvider);
components:
  schemas:
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
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````