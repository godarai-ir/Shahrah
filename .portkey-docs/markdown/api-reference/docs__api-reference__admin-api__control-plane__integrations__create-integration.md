> ## Documentation Index
> Fetch the complete documentation index at: https://docs.portkey.ai/docs/llms.txt
> Use this file to discover all available pages before exploring further.

# Create Integration



## OpenAPI

````yaml post /integrations
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
  /integrations:
    servers:
      - url: https://api.portkey.ai/v1
        description: Portkey API Public Endpoint
      - url: SELF_HOSTED_CONTROL_PLANE_URL
        description: Self-Hosted Control Plane URL
    post:
      tags:
        - Integrations
      summary: Create a Integration
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateIntegrationRequest'
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

            # Add a new integration
            integration = portkey.integrations.create(
                name="openai-production",
                ai_provider_id="openai",
                key="sk-..."
            )

            print(integration)
        - lang: javascript
          label: Default
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
            })

            const integration = await portkey.integrations.create({
                name:"openai-production",
                ai_provider_id:"openai",
                key:"sk-...",
            })
            console.log(integration);
        - lang: curl
          label: Default
          source: |
            curl -X POST https://api.portkey.ai/v1/integrations \
            -H "x-portkey-api-key: PORTKEY_API_KEY" \
            -H "Content-Type: application/json" \
            -d '{
                "name": "openai-production",
                "ai_provider_id": "openai",
                "key": "sk-..."
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

            # Add a new integration
            integration = portkey.integrations.create(
                name="openai-production",
                ai_provider_id="openai",
                key="sk-..."
            )

            print(integration)
        - lang: javascript
          label: Self-Hosted
          source: |
            import { Portkey } from "portkey-ai";

            const portkey = new Portkey({
                apiKey: "PORTKEY_API_KEY",
                baseUrl: "SELF_HOSTED_CONTROL_PLANE_URL"
            })

            const integration = await portkey.integrations.create({
                name: "openai-production",
                ai_provider_id: "openai",
                key: "sk-...",
            })
            console.log(integration);
        - lang: curl
          label: Self-Hosted
          source: |
            curl -X POST SELF_HOSTED_CONTROL_PLANE_URL/integrations \
            -H "x-portkey-api-key: PORTKEY_API_KEY" \
            -H "Content-Type: application/json" \
            -d '{
                "name": "openai-production",
                "ai_provider_id": "openai",
                "key": "sk-..."
            }'
components:
  schemas:
    CreateIntegrationRequest:
      type: object
      required:
        - name
        - ai_provider_id
      properties:
        name:
          type: string
          description: Human-readable name for the integration
          example: Production OpenAI
        slug:
          type: string
          pattern: ^[a-zA-Z0-9_-]+$
          description: URL-friendly identifier (auto-generated if not provided)
          example: production-openai
        ai_provider_id:
          type: string
          description: ID of the base AI provider
          example: openai
        key:
          type: string
          description: API key for the provider (if required)
          example: sk-...
        description:
          type: string
          description: Optional description of the integration
          example: Production OpenAI integration for customer-facing applications
        workspace_id:
          type: string
          description: Workspace ID (for workspace-scoped integrations)
          example: ws-my-team-1234
        configurations:
          type: object
          description: Provider-specific configuration object
          oneOf:
            - $ref: '#/components/schemas/OpenAIConfiguration'
              title: OpenAI
            - $ref: '#/components/schemas/AzureOpenAIConfiguration'
              title: Azure OpenAI
            - $ref: '#/components/schemas/BedrockConfiguration'
              title: AWS Bedrock
            - $ref: '#/components/schemas/VertexAIConfiguration'
              title: Vertex AI
            - $ref: '#/components/schemas/AzureAIConfiguration'
              title: Azure AI
            - $ref: '#/components/schemas/WorkersAIConfiguration'
              title: Workers AI
            - $ref: '#/components/schemas/SageMakerConfiguration'
              title: AWS Sagemaker
            - $ref: '#/components/schemas/HuggingFaceConfiguration'
              title: Hugginface
            - $ref: '#/components/schemas/CortexConfiguration'
              title: Cortex
            - $ref: '#/components/schemas/CustomHostConfiguration'
              title: Custom Base URL
        create_default_provider:
          type: boolean
          default: true
          description: >-
            Whether to automatically create a default provider when creating a
            workspace-scoped integration. Defaults to true.
        default_provider_slug:
          type: string
          pattern: ^[a-zA-Z0-9_-]+$
          maxLength: 255
          description: >-
            Custom slug for the auto-created default provider. Only applicable
            for workspace-scoped integrations. If the slug already exists in the
            workspace, the request will fail with a validation error.
        secret_mappings:
          type: array
          items:
            $ref: '#/components/schemas/SecretMapping'
          description: >-
            Dynamically resolve secrets from secret references at runtime. Valid
            target_field values are "key" or "configurations.<field>" (e.g.
            "configurations.aws_secret_access_key",
            "configurations.azure_entra_client_secret"). Each target_field must
            be unique. When "key" is mapped, the key body field can be omitted.
    OpenAIConfiguration:
      type: object
      properties:
        openai_organization:
          type: string
          description: OpenAI organization ID
        openai_project:
          type: string
          description: OpenAI project ID
    AzureOpenAIConfiguration:
      type: object
      required:
        - azure_resource_name
        - azure_deployment_config
        - azure_auth_mode
      properties:
        azure_auth_mode:
          type: string
          enum:
            - default
            - entra
            - managed
          description: Authentication mode for Azure
        azure_resource_name:
          type: string
          description: Azure OpenAI resource name
        azure_deployment_config:
          type: array
          minItems: 1
          items:
            $ref: '#/components/schemas/AzureDeploymentConfig'
        azure_entra_tenant_id:
          type: string
          description: Azure AD tenant ID (required for entra auth)
        azure_entra_client_id:
          type: string
          description: Azure AD client ID (required for entra auth)
        azure_entra_client_secret:
          type: string
          description: Azure AD client secret (required for entra auth)
        azure_managed_client_id:
          type: string
          description: Managed identity client ID (optional for managed auth)
    BedrockConfiguration:
      type: object
      required:
        - aws_auth_type
        - aws_region
      properties:
        aws_auth_type:
          type: string
          enum:
            - accessKey
            - assumedRole
          description: AWS authentication type
        aws_region:
          type: string
          description: AWS region
        aws_access_key_id:
          type: string
          description: AWS access key ID (required for accessKey auth)
        aws_secret_access_key:
          type: string
          description: AWS secret access key (required for accessKey auth)
        aws_role_arn:
          type: string
          description: AWS role ARN (required for assumedRole auth)
        aws_external_id:
          type: string
          nullable: true
          description: AWS external ID (optional for assumedRole auth)
    VertexAIConfiguration:
      type: object
      required:
        - vertex_auth_type
        - vertex_region
      properties:
        vertex_auth_type:
          type: string
          enum:
            - basic
            - serviceAccount
          description: Vertex AI authentication type
        vertex_region:
          type: string
          description: GCP region
        vertex_project_id:
          type: string
          description: GCP project ID (required for basic auth)
        vertex_service_account_json:
          type: object
          description: Service account JSON (required for serviceAccount auth)
    AzureAIConfiguration:
      type: object
      required:
        - azure_foundry_url
        - azure_auth_mode
      properties:
        azure_auth_mode:
          type: string
          enum:
            - default
            - entra
            - managed
          description: Authentication mode for Azure AI
        azure_foundry_url:
          type: string
          description: Azure AI Foundry URL
        azure_api_version:
          type: string
          maxLength: 30
          description: Azure API version
        azure_deployment_name:
          type: string
          description: Azure deployment name
        azure_entra_tenant_id:
          type: string
          description: Azure AD tenant ID (required for entra auth)
        azure_entra_client_id:
          type: string
          description: Azure AD client ID (required for entra auth)
        azure_entra_client_secret:
          type: string
          description: Azure AD client secret (required for entra auth)
        azure_managed_client_id:
          type: string
          description: Managed identity client ID (optional for managed auth)
    WorkersAIConfiguration:
      type: object
      required:
        - workers_ai_account_id
      properties:
        workers_ai_account_id:
          type: string
          description: Cloudflare Workers AI account ID
    SageMakerConfiguration:
      allOf:
        - $ref: '#/components/schemas/BedrockConfiguration'
        - type: object
          properties:
            amzn_sagemaker_custom_attributes:
              type: string
              description: Custom attributes for SageMaker
            amzn_sagemaker_target_model:
              type: string
              description: Target model for SageMaker
            amzn_sagemaker_target_variant:
              type: string
              description: Target variant for SageMaker
            amzn_sagemaker_target_container_hostname:
              type: string
              description: Target container hostname
            amzn_sagemaker_inference_id:
              type: string
              description: Inference ID
            amzn_sagemaker_enable_explanations:
              type: string
              description: Enable explanations
            amzn_sagemaker_inference_component:
              type: string
              description: Inference component
            amzn_sagemaker_session_id:
              type: string
              description: Session ID
            amzn_sagemaker_model_name:
              type: string
              description: Model name
    HuggingFaceConfiguration:
      type: object
      properties:
        huggingface_base_url:
          type: string
          description: Custom Hugging Face base URL
    CortexConfiguration:
      type: object
      required:
        - snowflake_account
      properties:
        snowflake_account:
          type: string
          description: Snowflake account identifier
    CustomHostConfiguration:
      type: object
      properties:
        custom_host:
          type: string
          description: >-
            Custom host URL (can be used along with other provider specific
            configuration fields)
        custom_headers:
          type: object
          additionalProperties:
            type: string
          description: >-
            Custom headers to send with requests (can be used along with other
            provider specific configuration fields)
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
    AzureDeploymentConfig:
      type: object
      required:
        - azure_api_version
        - azure_deployment_name
        - azure_model_slug
      properties:
        alias:
          type: string
          description: Alias for the deployment
        azure_api_version:
          type: string
          maxLength: 30
          description: Azure API version
        azure_deployment_name:
          type: string
          description: Azure deployment name
        is_default:
          type: boolean
          default: false
          description: Whether this is the default deployment
        azure_model_slug:
          type: string
          description: Azure model slug
  securitySchemes:
    Portkey-Key:
      type: apiKey
      in: header
      name: x-portkey-api-key

````