> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create API Key

> Creates a new API key.




## OpenAPI

````yaml post /api-keys/{type}/{sub-type}
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
  /api-keys/{type}/{sub-type}:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_CONTROL_PLANE_URL
        description: Self-Hosted Control Plane URL
    post:
      tags:
        - Api-Keys
      summary: Create API Keys
      description: |
        Creates a new API key.
      parameters:
        - name: type
          in: path
          schema:
            type: string
            enum:
              - organisation
              - workspace
          required: true
        - name: sub-type
          in: path
          schema:
            type: string
            enum:
              - user
              - service
          required: true
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateApiKeyObject'
            examples:
              create_basic:
                summary: Create a basic service API key
                value:
                  name: API_KEY_NAME_0909
                  scopes:
                    - completions.write
              create_with_rotation:
                summary: Create an API key with automatic monthly rotation
                value:
                  name: Auto-rotating Key
                  scopes:
                    - completions.write
                  rotation_policy:
                    rotation_period: monthly
                    key_transition_period_ms: 3600000
              create_with_usage_limits:
                summary: Create an API key with usage limits and alert emails
                value:
                  name: Budget-limited Key
                  scopes:
                    - completions.write
                  usage_limits:
                    type: cost
                    credit_limit: 100
                    alert_threshold: 80
                    periodic_reset: monthly
                  alert_emails:
                    - admin@example.com
      responses:
        '200':
          description: OK
          headers:
            Content-Type:
              schema:
                type: string
                example: application/json
          content:
            application/json:
              schema:
                type: object
                properties:
                  id:
                    type: string
                    format: uuid
                    example: 183f497a-2a7f-4f47-992e-26213fa863we
                  key:
                    type: string
                    example: abssofjosfjs
                  object:
                    type: string
                    enum:
                      - api-key
                    example: api-key
      x-code-samples:
        - lang: python
          label: Default
          source: |
            from portkey_ai import Portkey

            # Initialize the Portkey client
            portkey = Portkey(
                api_key="PORTKEY_API_KEY",
            )

            # Create a new API key
            api_key = portkey.api_keys.create(
                name="API_KEY_NAME_0909",
                type="organisation",
                sub_type="service",
                workspace_id="WORKSPACE_ID",
                scopes=[
                    "logs.export",
                    "logs.list",
                    "logs.view",
                    "configs.create",
                    "configs.update",
                    "configs.delete",
                    "configs.read",
                    "configs.list",
                    "virtual_keys.create",
                    "virtual_keys.update",
                    "virtual_keys.delete",
                    "virtual_keys.read",
                    "virtual_keys.list",
                    "virtual_keys.copy"
                ]
            )

            print(api_key)
        - lang: javascript
          label: Default
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
            })

            const apiKey=await portkey.apiKeys.create({
                name:"API_KEY_NAME_0909",
                type:"organisation",
                "sub-type":"service",
                workspace_id:"WORKSPACE_ID",
                "scopes": [
                    "logs.export",
                    "logs.list",
                    "logs.view",
                    "configs.create",
                    "configs.update",
                    "configs.delete",
                    "configs.read",
                    "configs.list",
                    "virtual_keys.create",
                    "virtual_keys.update",
                    "virtual_keys.delete",
                    "virtual_keys.read",
                    "virtual_keys.list",
                    "virtual_keys.copy"
                ]
            })
            console.log(apiKey);
        - lang: curl
          label: Default
          source: |
            curl -X POST https://api.portkey.ai/v1/api-keys/organisation/service
              -H "x-portkey-api-key: PORTKEY_API_KEY" \
              -H "Content-Type: application/json" \
              -d '{
                "name":"API_KEY_NAME_0909",
                "type":"organisation",
                "sub-type":"service",
                "workspace_id":"WORKSPACE_ID",
                "scopes":[
                  "logs.export",
                  "logs.list",
                  "logs.view",
                  "configs.create",
                  "configs.update",
                  "configs.delete",
                  "configs.read",
                  "configs.list",
                  "virtual_keys.create",
                  "virtual_keys.update",
                  "virtual_keys.delete",
                  "virtual_keys.read",
                  "virtual_keys.list",
                  "virtual_keys.copy"
                ]
              }'
        - lang: curl
          label: Self-Hosted
          source: >
            curl -X POST
            SELF_HOSTED_CONTROL_PLANE_URL/api-keys/organisation/service
              -H "x-portkey-api-key: PORTKEY_API_KEY" \
              -H "Content-Type: application/json" \
              -d '{
                "name":"API_KEY_NAME_0909",
                "type":"organisation",
                "sub-type":"service",
                "workspace_id":"WORKSPACE_ID",
                "scopes":[
                  "logs.export",
                  "logs.list",
                  "logs.view",
                  "configs.create",
                  "configs.update",
                  "configs.delete",
                  "configs.read",
                  "configs.list",
                  "virtual_keys.create",
                  "virtual_keys.update",
                  "virtual_keys.delete",
                  "virtual_keys.read",
                  "virtual_keys.list",
                  "virtual_keys.copy"
                ]
              }'
        - lang: javascript
          label: Self-Hosted
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
                baseUrl: "SELF_HOSTED_CONTROL_PLANE_URL"
            })

            const apiKey = await portkey.admin.apiKeys.create({
                name: "API_KEY_NAME_0909",
                type: "organisation",
                subType: "service",
                workspaceId: "WORKSPACE_ID",
                scopes: [
                    "logs.export",
                    "logs.list",
                    "logs.view",
                    "configs.create",
                    "configs.update",
                    "configs.delete",
                    "configs.read",
                    "configs.list",
                    "virtual_keys.create",
                    "virtual_keys.update",
                    "virtual_keys.delete",
                    "virtual_keys.read",
                    "virtual_keys.list",
                    "virtual_keys.copy"
                ]
            })
            console.log(apiKey);
        - lang: python
          label: Self-Hosted
          source: |
            from portkey import Portkey

            portkey = Portkey(
                api_key="PORTKEY_API_KEY",
                base_url="SELF_HOSTED_CONTROL_PLANE_URL"
            )

            api_key = portkey.admin.api_keys.create(
                name="API_KEY_NAME_0909",
                type="organisation",
                sub_type="service",
                workspace_id="WORKSPACE_ID",
                scopes=[
                    "logs.export",
                    "logs.list",
                    "logs.view",
                    "configs.create",
                    "configs.update",
                    "configs.delete",
                    "configs.read",
                    "configs.list",
                    "virtual_keys.create",
                    "virtual_keys.update",
                    "virtual_keys.delete",
                    "virtual_keys.read",
                    "virtual_keys.list",
                    "virtual_keys.copy"
                ]
            )
            print(api_key)
        - lang: python
          label: User API Key
          source: |
            from portkey_ai import Portkey

            # Initialize the Portkey client
            portkey = Portkey(
                api_key="PORTKEY_API_KEY",
            )

            # Create a user API key (requires user_id)
            api_key = portkey.api_keys.create(
                name="User API Key",
                type="workspace",
                sub_type="user",
                workspace_id="WORKSPACE_ID",
                user_id="USER_ID",  # Required for user API keys
                scopes=[
                    "completions.write",
                    "logs.view"
                ]
            )

            print(api_key)
        - lang: javascript
          label: User API Key
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
            })

            const apiKey = await portkey.apiKeys.create({
                name: "User API Key",
                type: "workspace",
                "sub-type": "user",
                workspace_id: "WORKSPACE_ID",
                user_id: "USER_ID",  // Required for user API keys
                "scopes": [
                    "completions.write",
                    "logs.view"
                ]
            })
            console.log(apiKey);
        - lang: curl
          label: User API Key
          source: |
            curl -X POST https://api.portkey.ai/v1/api-keys/workspace/user
              -H "x-portkey-api-key: PORTKEY_API_KEY" \
              -H "Content-Type: application/json" \
              -d '{
                "name":"User API Key",
                "type":"workspace",
                "sub-type":"user",
                "workspace_id":"WORKSPACE_ID",
                "user_id":"USER_ID",
                "scopes":[
                  "completions.write",
                  "logs.view"
                ]
              }'
components:
  schemas:
    CreateApiKeyObject:
      type: object
      required:
        - name
        - scopes
      properties:
        name:
          type: string
          example: Development API Key
        description:
          type: string
          example: API key for development environment
        organisation_id:
          type: string
          format: uuid
          description: >-
            Organisation ID. Optional when calling via an org-level API key
            (picked from auth context).
          example: a1b2c3d4-e5f6-7890-abcd-ef1234567890
        workspace_id:
          type: string
          example: ws-myworkspace
        user_id:
          type: string
          format: uuid
          description: >-
            **Required** when sub-type path parameter is 'user'. Not required
            when sub-type is 'service'.
          example: c3d4e5f6-a7b8-6c7d-0e1f-2a3b4c5d6e7f
        rate_limits:
          type: array
          nullable: true
          items:
            type: object
            required:
              - type
              - unit
              - value
            properties:
              type:
                type: string
                enum:
                  - requests
                  - tokens
                example: requests
              unit:
                type: string
                enum:
                  - rpd
                  - rph
                  - rpm
                  - rps
                  - rpw
                description: >-
                  Rate limit unit: requests/tokens per day, hour, minute,
                  second, or week.
                example: rpm
              value:
                type: integer
                minimum: 0
                example: 100
        usage_limits:
          $ref: '#/components/schemas/UsageLimits'
        scopes:
          type: array
          items:
            type: string
          example:
            - completions.write
        defaults:
          type: object
          properties:
            metadata:
              type: object
              additionalProperties: true
              example:
                environment: development
                team: backend
            config_id:
              type: string
              example: config-abc
            allow_config_override:
              type: boolean
              default: true
              description: >-
                When false, the pinned config_id cannot be overridden at request
                time.
              example: false
        alert_emails:
          type: array
          items:
            type: string
            format: email
            example: foo@bar.com
        expires_at:
          type: string
          format: date-time
          nullable: true
          description: >-
            ISO 8601 datetime at which this key expires. Must not exceed the
            org/workspace maximum TTL if configured.
          example: '2026-12-31T23:59:59Z'
        rotation_policy:
          type: object
          nullable: true
          description: >-
            Automatic key rotation configuration. Requires either
            rotation_period or next_rotation_at.
          properties:
            rotation_period:
              type: string
              enum:
                - weekly
                - monthly
              nullable: true
              description: How often to automatically rotate the key.
              example: monthly
            next_rotation_at:
              type: string
              format: date-time
              nullable: true
              description: >-
                Explicit datetime for the next rotation. Mutually exclusive with
                rotation_period.
              example: '2026-06-01T00:00:00Z'
            key_transition_period_ms:
              type: integer
              minimum: 1800000
              description: >-
                Duration in milliseconds during which the previous key remains
                valid after rotation. Minimum 30 minutes (1800000). Must be less
                than the full rotation period.
              example: 3600000
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
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````