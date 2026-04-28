> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Secret Reference



## OpenAPI

````yaml post /secret-references
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
  /secret-references:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_CONTROL_PLANE_URL
        description: Self-Hosted Control Plane URL
    post:
      tags:
        - Secret-References
      summary: Create a Secret Reference
      operationId: createSecretReference
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateSecretReferenceRequest'
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
                    format: uuid
                  slug:
                    type: string
                  object:
                    type: string
                    enum:
                      - secret-reference
        '400':
          description: >-
            Validation failure (schema, duplicate slug, invalid workspaces,
            conflicting allow_all_workspaces + allowed_workspaces)
        '403':
          description: secretReferences feature not enabled on subscription
      security:
        - Portkey-Key: []
      x-code-samples:
        - lang: curl
          label: Default
          source: |
            curl -X POST https://api.portkey.ai/v1/secret-references \
            -H "x-portkey-api-key: PORTKEY_API_KEY" \
            -H "Content-Type: application/json" \
            -d '{
                "name": "my-aws-secret",
                "manager_type": "aws_sm",
                "auth_config": {
                    "aws_auth_type": "accessKey",
                    "aws_access_key_id": "AKIA...",
                    "aws_secret_access_key": "wJal...",
                    "aws_region": "us-east-1"
                },
                "secret_path": "prod/api-keys/openai"
            }'
        - lang: curl
          label: Self-Hosted
          source: |
            curl -X POST SELF_HOSTED_CONTROL_PLANE_URL/secret-references \
            -H "x-portkey-api-key: PORTKEY_API_KEY" \
            -H "Content-Type: application/json" \
            -d '{
                "name": "my-aws-secret",
                "manager_type": "aws_sm",
                "auth_config": {
                    "aws_auth_type": "accessKey",
                    "aws_access_key_id": "AKIA...",
                    "aws_secret_access_key": "wJal...",
                    "aws_region": "us-east-1"
                },
                "secret_path": "prod/api-keys/openai"
            }'
components:
  schemas:
    CreateSecretReferenceRequest:
      type: object
      required:
        - name
        - manager_type
        - auth_config
        - secret_path
      properties:
        organisation_id:
          type: string
          format: uuid
          description: Required if not using API key auth
        name:
          type: string
          minLength: 1
          maxLength: 255
        slug:
          type: string
          pattern: ^[a-zA-Z0-9_-]+$
          maxLength: 255
          description: Auto-generated from name if omitted
        description:
          type: string
          maxLength: 1024
          nullable: true
        manager_type:
          type: string
          enum:
            - aws_sm
            - azure_kv
            - hashicorp_vault
        auth_config:
          oneOf:
            - $ref: '#/components/schemas/AwsAccessKeyAuthConfig'
            - $ref: '#/components/schemas/AwsAssumedRoleAuthConfig'
            - $ref: '#/components/schemas/AwsServiceRoleAuthConfig'
            - $ref: '#/components/schemas/AzureEntraAuthConfig'
            - $ref: '#/components/schemas/AzureManagedAuthConfig'
            - $ref: '#/components/schemas/AzureDefaultAuthConfig'
            - $ref: '#/components/schemas/HashicorpTokenAuthConfig'
            - $ref: '#/components/schemas/HashicorpAppRoleAuthConfig'
            - $ref: '#/components/schemas/HashicorpKubernetesAuthConfig'
          discriminator:
            propertyName: aws_auth_type
        secret_path:
          type: string
          maxLength: 1024
        secret_key:
          type: string
          maxLength: 255
          nullable: true
        allow_all_workspaces:
          type: boolean
          default: true
          description: Cannot be true when allowed_workspaces is provided
        allowed_workspaces:
          type: array
          items:
            type: string
          minItems: 1
          description: >-
            Array of workspace UUIDs or slugs. Mutually exclusive with
            allow_all_workspaces=true.
        tags:
          type: object
          additionalProperties:
            type: string
          nullable: true
    AwsAccessKeyAuthConfig:
      type: object
      required:
        - aws_auth_type
        - aws_access_key_id
        - aws_secret_access_key
        - aws_region
      properties:
        aws_auth_type:
          type: string
          enum:
            - accessKey
        aws_access_key_id:
          type: string
        aws_secret_access_key:
          type: string
        aws_region:
          type: string
    AwsAssumedRoleAuthConfig:
      type: object
      required:
        - aws_auth_type
        - aws_role_arn
        - aws_region
      properties:
        aws_auth_type:
          type: string
          enum:
            - assumedRole
        aws_role_arn:
          type: string
        aws_external_id:
          type: string
          nullable: true
        aws_region:
          type: string
    AwsServiceRoleAuthConfig:
      type: object
      required:
        - aws_auth_type
      properties:
        aws_auth_type:
          type: string
          enum:
            - serviceRole
        aws_region:
          type: string
    AzureEntraAuthConfig:
      type: object
      required:
        - azure_auth_mode
        - azure_entra_tenant_id
        - azure_entra_client_id
        - azure_entra_client_secret
        - azure_vault_url
      properties:
        azure_auth_mode:
          type: string
          enum:
            - entra
        azure_entra_tenant_id:
          type: string
        azure_entra_client_id:
          type: string
        azure_entra_client_secret:
          type: string
        azure_vault_url:
          type: string
          format: uri
    AzureManagedAuthConfig:
      type: object
      required:
        - azure_auth_mode
        - azure_vault_url
      properties:
        azure_auth_mode:
          type: string
          enum:
            - managed
        azure_managed_client_id:
          type: string
        azure_vault_url:
          type: string
          format: uri
    AzureDefaultAuthConfig:
      type: object
      required:
        - azure_auth_mode
        - azure_vault_url
      properties:
        azure_auth_mode:
          type: string
          enum:
            - default
        azure_vault_url:
          type: string
          format: uri
    HashicorpTokenAuthConfig:
      type: object
      required:
        - vault_auth_type
        - vault_addr
        - vault_token
      properties:
        vault_auth_type:
          type: string
          enum:
            - token
        vault_addr:
          type: string
          format: uri
        vault_token:
          type: string
        vault_namespace:
          type: string
    HashicorpAppRoleAuthConfig:
      type: object
      required:
        - vault_auth_type
        - vault_addr
        - vault_role_id
        - vault_secret_id
      properties:
        vault_auth_type:
          type: string
          enum:
            - approle
        vault_addr:
          type: string
          format: uri
        vault_role_id:
          type: string
        vault_secret_id:
          type: string
        vault_namespace:
          type: string
    HashicorpKubernetesAuthConfig:
      type: object
      required:
        - vault_auth_type
        - vault_addr
        - vault_role
      properties:
        vault_auth_type:
          type: string
          enum:
            - kubernetes
        vault_addr:
          type: string
          format: uri
        vault_role:
          type: string
        vault_namespace:
          type: string
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````