> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create or update prompt version

> Create or update a prompt version.



## OpenAPI

````yaml /openapi.json put /api/ml/v1/prompt-versions
openapi: 3.1.0
info:
  title: TrueFoundry API
  description: API for TrueFoundry
  version: 0.1.0
  contact: {}
servers:
  - url: https://{controlPlaneURL}
    variables:
      controlPlaneURL:
        default: app.truefoundry.com
        description: Control Plane URL
security: []
tags:
  - name: Agent
  - name: Applications
  - name: Apply
  - name: Artifacts
  - name: Audit Logs
  - name: Clusters
  - name: Jobs
  - name: Logs
  - name: MCP Servers
  - name: MLRepos
  - name: Model Deployments
  - name: Models
  - name: Personal Access Tokens
  - name: Prompts
  - name: Provider Integrations
  - name: SCIM v2
  - name: Secret Groups
  - name: Secrets
  - name: Teams
  - name: Traces
  - name: Users
  - name: Virtual Accounts
  - name: Workspaces
paths:
  /api/ml/v1/prompt-versions:
    put:
      tags:
        - Prompts
      summary: Create or update prompt version
      description: Create or update a prompt version.
      operationId: apply_prompt_version_v1_prompt_versions_put
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ApplyPromptRequest'
        required: true
      responses:
        '200':
          description: The created or updated prompt version
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetPromptVersionResponse'
        '422':
          description: Validation Error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HTTPValidationError'
      security:
        - jwt: []
components:
  schemas:
    ApplyPromptRequest:
      properties:
        manifest:
          allOf:
            - $ref: '#/components/schemas/ChatPromptManifest'
          title: Manifest
          description: Manifest containing metadata for the prompt to apply
      type: object
      required:
        - manifest
      title: ApplyPromptRequest
      x-fern-type-name: ApplyPromptRequest
    GetPromptVersionResponse:
      properties:
        data:
          allOf:
            - $ref: '#/components/schemas/PromptVersion'
          title: Data
          description: The prompt version data
      type: object
      required:
        - data
      title: GetPromptVersionResponse
      x-fern-type-name: GetPromptVersionResponse
    HTTPValidationError:
      properties:
        detail:
          items:
            $ref: '#/components/schemas/ValidationError'
          type: array
          title: Detail
      type: object
      title: HTTPValidationError
      x-fern-type-name: HTTPValidationError
    ChatPromptManifest:
      properties:
        name:
          type: string
          maxLength: 256
          pattern: ^[A-Za-z0-9_\-]+$
          title: Name
          description: >-
            Name of the prompt (alphanumeric characters, hyphens, and
            underscores only, max 256 characters)
        metadata:
          type: object
          title: Metadata
          description: >-
            Key value metadata. Should be valid JSON. For e.g.
            `{"business-unit": "sales", "quality": "good", "rating": 4.5}`
        ml_repo:
          type: string
          pattern: ^[a-zA-Z][a-zA-Z0-9\-]{1,98}[a-zA-Z0-9]$
          title: Ml Repo
          description: >-
            Name of the ML Repo that this prompt belongs to (must start and end
            with alphanumeric, 2-100 characters)
        version:
          type: integer
          minimum: 1
          title: Version
          description: Version of the entity
        type:
          type: string
          enum:
            - chat_prompt
          title: Type
          default: chat_prompt
        description:
          type: string
          maxLength: 512
          title: Description
        version_alias:
          type: string
          maxLength: 128
          pattern: ^v[a-zA-Z0-9.-]*([a-zA-Z0-9]+)$
          title: Version Alias
          description: >-
            Version alias is alternate, ideally human readable, version string
            to reference an artifact version. It should start with `v` followed
            by alphanumeric and it can include `.` and `-` in between (e.g.
            `v1.0.0`, `v1-prod`, `v3-dev`, etc)
        messages:
          items:
            anyOf:
              - $ref: '#/components/schemas/SystemMessage'
              - $ref: '#/components/schemas/AssistantMessage'
              - $ref: '#/components/schemas/UserMessage'
              - $ref: '#/components/schemas/ToolMessage'
              - $ref: '#/components/schemas/DeveloperMessage'
          type: array
          title: Messages
          description: List of messages in the chat conversation, must be non-empty
        variables:
          additionalProperties:
            type: string
          type: object
          title: Variables
          description: >-
            Variables referenced in messages and that can be replaced when
            running generation
        model_configuration:
          $ref: '#/components/schemas/ModelConfiguration'
        tools:
          items:
            $ref: '#/components/schemas/ToolSchema'
          type: array
          title: Tools
          description: List of tools to be used in the chat prompt
        mcp_servers:
          items:
            anyOf:
              - $ref: '#/components/schemas/MCPServerWithFQN'
              - $ref: '#/components/schemas/MCPServerWithURL'
          type: array
          title: Mcp Servers
          description: A list of MCP servers FQNs or URLs and their tools
        guardrails:
          $ref: '#/components/schemas/Guardrails'
        response_format:
          oneOf:
            - $ref: '#/components/schemas/JsonObjectResponseFormat'
            - $ref: '#/components/schemas/JsonSchemaResponseFormat'
          title: Response Format
          description: Response format configuration for structured outputs
          discriminator:
            propertyName: type
            mapping:
              json_object:
                $ref: '#/components/schemas/JsonObjectResponseFormat'
              json_schema:
                $ref: '#/components/schemas/JsonSchemaResponseFormat'
        routing_config:
          oneOf:
            - $ref: '#/components/schemas/WeightBasedLoadBalancing'
            - $ref: '#/components/schemas/LatencyBasedLoadBalancing'
            - $ref: '#/components/schemas/PriorityBasedLoadBalancing'
          title: Routing Config
          description: Configuration for routing requests to different model targets
          discriminator:
            propertyName: type
            mapping:
              weight-based-routing:
                $ref: '#/components/schemas/WeightBasedLoadBalancing'
              latency-based-routing:
                $ref: '#/components/schemas/LatencyBasedLoadBalancing'
              priority-based-routing:
                $ref: '#/components/schemas/PriorityBasedLoadBalancing'
        cache_config:
          oneOf:
            - $ref: '#/components/schemas/SemanticCacheConfig'
            - $ref: '#/components/schemas/ExactMatchCacheConfig'
          title: Cache Config
          description: Cache configuration for the chat prompt
          discriminator:
            propertyName: type
            mapping:
              semantic:
                $ref: '#/components/schemas/SemanticCacheConfig'
              exact-match:
                $ref: '#/components/schemas/ExactMatchCacheConfig'
        tool_call_to_mcp_mapping:
          additionalProperties:
            $ref: '#/components/schemas/MCPServerToolDetails'
          type: object
          title: Tool Call To Mcp Mapping
          description: Mapping of tool calls to MCP server integration IDs and tool names
        logging_config:
          $ref: '#/components/schemas/LoggingConfig'
        sub_agents:
          items:
            $ref: '#/components/schemas/SubAgent'
          type: array
          title: Sub Agents
          description: Sub agents to be used in the chat prompt
      type: object
      required:
        - name
        - metadata
        - ml_repo
        - messages
      title: ChatPromptManifest
      description: Chat Prompt manifest.
      x-fern-type-name: ChatPromptManifest
    PromptVersion:
      properties:
        id:
          type: string
          title: Id
          description: Unique identifier for the artifact version
        fqn:
          type: string
          title: Fqn
          description: >-
            Fully qualified name of the artifact version in the format
            '{artifact_type}:{tenant_name}/{ml_repo_name}/{artifact_name}:{version}'
        created_by_subject:
          allOf:
            - $ref: '#/components/schemas/Subject'
          title: Created By Subject
          description: >-
            Subject (user, team, or service account) that created this artifact
            version
        created_at:
          type: string
          format: date-time
          title: Created At
          description: Timestamp when the artifact version was created
        updated_at:
          type: string
          format: date-time
          title: Updated At
          description: Timestamp when the artifact version was last updated
        manifest:
          allOf:
            - $ref: '#/components/schemas/ChatPromptManifest'
          title: Manifest
          description: Manifest containing metadata specific to the chat prompt version
        ml_repo_id:
          type: string
          title: Ml Repo Id
          description: ID of the ML Repo that this artifact version belongs to
        tags:
          items:
            type: string
          type: array
          title: Tags
          description: >-
            List of tags associated with this artifact version for filtering and
            organization
        version_alias:
          type: string
          title: Version Alias
          description: Optional human-readable version alias (e.g. v1.0.0)
        usage_code_snippet:
          type: string
          title: Usage Code Snippet
          description: Code snippet demonstrating how to use this artifact version
        usage_code_snippets:
          items:
            $ref: '#/components/schemas/UsageCodeSnippet'
          type: array
          title: Usage Code Snippets
          description: >-
            List of code snippets demonstrating how to use this prompt version
            in different languages
        prompt_id:
          type: string
          title: Prompt Id
          description: ID of the parent prompt that this version belongs to
      type: object
      required:
        - id
        - fqn
        - created_by_subject
        - manifest
        - ml_repo_id
        - prompt_id
      title: PromptVersion
      description: >-
        Tags, optional version alias, and SDK usage snippet (models, prompts,
        generic artifacts).
      x-fern-type-name: PromptVersion
    ValidationError:
      properties:
        loc:
          items:
            anyOf:
              - type: string
              - type: integer
          type: array
          title: Location
        msg:
          type: string
          title: Message
        type:
          type: string
          title: Error Type
      type: object
      required:
        - loc
        - msg
        - type
      title: ValidationError
      x-fern-type-name: ValidationError
    SystemMessage:
      description: |-
        +label=System Message
        +usage=System message for the chat
      type: object
      required:
        - role
        - content
      properties:
        role:
          description: |-
            +value=system
            +label=Role
            +usage=Role of the message
            +uiType=Ignore
          type: string
          enum:
            - system
        content:
          description: |-
            +label=Content
            +usage=Text content for the system message
            +uiType=CodeEditor
            +uiProps={"language": "text"}
          oneOf:
            - type: string
              pattern: ^.[\s\S]*$
            - $ref: '#/components/schemas/BlobStorageReference'
        name:
          description: |-
            +label=Name
            +usage=Name of the system
          type: string
      title: SystemMessage
      x-fern-type-name: SystemMessage
    AssistantMessage:
      description: |-
        +label=Assistant Message
        +usage=Assistant message turn
      type: object
      required:
        - role
      properties:
        role:
          description: |-
            +value=assistant
            +label=Role
            +usage=Role of the message
            +uiType=Ignore
          type: string
          enum:
            - assistant
        content:
          description: |-
            +label=Content
            +usage=Text content for the assistant message
            +uiType=CodeEditor
            +uiProps={"language": "text"}
          oneOf:
            - type: string
            - type: array
              items:
                type: object
                oneOf:
                  - $ref: '#/components/schemas/TextContentPart'
                  - $ref: '#/components/schemas/RefusalContentPart'
            - $ref: '#/components/schemas/BlobStorageReference'
        tool_calls:
          description: |-
            +label=Tool Calls
            +usage=List of tool calls made by the assistant
          type: array
          items:
            $ref: '#/components/schemas/ToolCall'
        name:
          description: |-
            +label=Name
            +usage=Name of the assistant this message is from
          type: string
      title: AssistantMessage
      x-fern-type-name: AssistantMessage
    UserMessage:
      description: |-
        +label=User Message
        +usage=User message turn
      type: object
      required:
        - role
        - content
      properties:
        role:
          description: |-
            +value=user
            +usage=Role of the message
            +uiType=Ignore
          type: string
          enum:
            - user
        content:
          description: |-
            +label=Content
            +usage=Content of the user message. can be a mix of text and images
            +uiType=CodeEditor
            +uiProps={"language": "text"}
          oneOf:
            - type: string
              pattern: ^.[\s\S]*$
            - type: array
              items:
                type: object
                oneOf:
                  - $ref: '#/components/schemas/TextContentPart'
                  - $ref: '#/components/schemas/ImageContentPart'
            - $ref: '#/components/schemas/BlobStorageReference'
        name:
          description: |-
            +label=Name
            +usage=Name of the user this message is from
          type: string
      title: UserMessage
      x-fern-type-name: UserMessage
    ToolMessage:
      description: |-
        +label=Tool Message
        +usage=Tool message for the chat
      type: object
      required:
        - role
        - content
        - tool_call_id
      properties:
        role:
          description: |-
            +label=Role
            +usage=Role of the message
          type: string
          enum:
            - tool
        content:
          description: |-
            +label=Content
            +usage=Content of the tool call result
          oneOf:
            - type: string
              pattern: ^.[\s\S]*$
            - $ref: '#/components/schemas/BlobStorageReference'
        tool_call_id:
          description: |-
            +label=Tool Call ID
            +usage=Unique identifier for the tool call
          type: string
      title: ToolMessage
      x-fern-type-name: ToolMessage
    DeveloperMessage:
      description: |-
        +label=Developer Message
        +usage=Developer message for the chat
      type: object
      required:
        - role
        - content
      properties:
        role:
          description: |-
            +value=Developer
            +label=Role
            +usage=The role of the messages author, in this case developer.
          type: string
          enum:
            - developer
        content:
          description: |-
            +label=Content
            +usage=The contents of the developer message.
          oneOf:
            - type: string
              pattern: ^.[\s\S]*$
            - $ref: '#/components/schemas/BlobStorageReference'
        name:
          description: |-
            +label=Name
            +usage=An optional name for the participant.
          type: string
      title: DeveloperMessage
      x-fern-type-name: DeveloperMessage
    ModelConfiguration:
      properties:
        provider:
          type: string
          title: Provider
          description: >-
            Name of the provider, must be one of the integration providers
            configured for the Gateway
        model:
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-.]{0,60}[a-zA-Z0-9])$
          title: Model
          description: >-
            Name of the model to be used for generations. This model should be
            available in the provider
        parameters:
          allOf:
            - $ref: '#/components/schemas/Parameters'
          title: Parameters
          description: Parameters to pass to the model when generating
        extra_parameters:
          type: object
          title: Extra Parameters
          description: Arbitrary extra parameters to pass to the model when generating
      type: object
      required:
        - provider
        - model
      title: ModelConfiguration
      description: >-
        Add a model from Gateway along with parameters to be used for chat
        completion
      x-fern-type-name: ModelConfiguration
    ToolSchema:
      description: |-
        +label=Tool Schema
        +usage=Schema defining a tool for the chat prompt
      type: object
      required:
        - type
        - function
      properties:
        type:
          description: |-
            +label=Type
            +usage=Type of the tool
          type: string
          enum:
            - function
        function:
          $ref: '#/components/schemas/FunctionSchema'
      title: ToolSchema
      x-fern-type-name: ToolSchema
    MCPServerWithFQN:
      description: |-
        +label=MCP Server FQN
        +icon=puzzle-piece
        +usage=MCP server with FQN
      type: object
      required:
        - type
        - integration_fqn
        - enable_all_tools
      properties:
        type:
          description: |-
            +label=MCP Server FQN Type
            +usage=The FQN of the MCP server
            +value=mcp-server-fqn
          type: string
          enum:
            - mcp-server-fqn
        integration_fqn:
          description: |-
            +label=MCP Server Integration FQN
            +usage=The FQN of the MCP server
          type: string
        enable_all_tools:
          description: |-
            +label=Enable All Tools
            +usage=Whether to enable all tools from the MCP server
          type: boolean
          default: false
        tools:
          description: |-
            +label=Tools
            +usage=The tools to enable from the MCP server
          type: array
          items:
            $ref: '#/components/schemas/MCPTool'
      title: MCPServerWithFQN
      x-fern-type-name: MCPServerWithFQN
    MCPServerWithURL:
      description: |-
        +label=MCP Server URL
        +icon=puzzle-piece
        +usage=MCP server with URL
      type: object
      required:
        - type
        - url
        - enable_all_tools
      properties:
        type:
          description: |-
            +label=URL
            +usage=The URL of the MCP server
            +value=mcp-server-url
            +sort=1
          type: string
          enum:
            - mcp-server-url
        url:
          description: |-
            +label=URL
            +usage=The URL of the MCP server
            +sort=2
          type: string
          pattern: ^https?://.+
        headers:
          description: |-
            +label=Headers
            +usage=The headers to send to the MCP server
            +sort=3
          type: object
          additionalProperties:
            type: string
        enable_all_tools:
          description: |-
            +label=Enable All Tools
            +usage=Whether to enable all tools from the MCP server
            +sort=4
          type: boolean
          default: false
        tools:
          description: |-
            +label=Tools
            +usage=The tools to enable from the MCP server
            +sort=5
          type: array
          items:
            $ref: '#/components/schemas/MCPTool'
      title: MCPServerWithURL
      x-fern-type-name: MCPServerWithURL
    Guardrails:
      description: +label=Guardrails
      type: object
      properties:
        input_guardrails:
          description: >-
            +label=Input Guardrails

            +usage=List of guardrail names to apply to user input messages
            before processing.
          type: array
          items:
            type: string
        output_guardrails:
          description: >-
            +label=Output Guardrails

            +usage=List of guardrail names to apply to AI response messages
            before returning to user.
          type: array
          items:
            type: string
      title: Guardrails
      x-fern-type-name: Guardrails
    JsonObjectResponseFormat:
      description: |-
        +label=JSON Object Response Format
        +usage=Simple JSON mode response format
      type: object
      required:
        - type
      properties:
        type:
          description: |-
            +label=Type
            +usage=Type of response format
          type: string
          enum:
            - json_object
      title: JsonObjectResponseFormat
      x-fern-type-name: JsonObjectResponseFormat
    JsonSchemaResponseFormat:
      description: |-
        +label=JSON Schema Response Format
        +usage=Structured JSON response format with schema validation
      type: object
      required:
        - type
        - json_schema
      properties:
        type:
          description: |-
            +label=Type
            +usage=Type of response format
          type: string
          enum:
            - json_schema
        json_schema:
          $ref: '#/components/schemas/JsonSchema'
      title: JsonSchemaResponseFormat
      x-fern-type-name: JsonSchemaResponseFormat
    WeightBasedLoadBalancing:
      description: +label=Weight-based Load Balancing
      type: object
      required:
        - type
        - load_balance_targets
      properties:
        type:
          description: |-
            +value=weight-based-routing
            +sort=1
            +label=Routing Type
          type: string
          enum:
            - weight-based-routing
        load_balance_targets:
          description: |-
            +usage=List of targets for load balancing with weights
            +uiProps={"descriptionInline":true}
            +sort=4
            +label=Load Balance Targets
          type: array
          items:
            $ref: '#/components/schemas/LoadBalanceTarget'
        sticky_routing:
          $ref: '#/components/schemas/StickyRouting'
      title: WeightBasedLoadBalancing
      x-fern-type-name: WeightBasedLoadBalancing
    LatencyBasedLoadBalancing:
      description: +label=Latency-based Load Balancing
      type: object
      required:
        - type
        - load_balance_targets
      properties:
        type:
          description: |-
            +value=latency-based-routing
            +sort=1
            +label=Routing Type
          type: string
          enum:
            - latency-based-routing
        load_balance_targets:
          description: |-
            +usage=List of targets for latency-based load balancing
            +uiProps={"descriptionInline":true}
            +sort=4
            +label=Load Balance Targets
          type: array
          items:
            $ref: '#/components/schemas/LatencyBasedLoadBalanceTarget'
      title: LatencyBasedLoadBalancing
      x-fern-type-name: LatencyBasedLoadBalancing
    PriorityBasedLoadBalancing:
      description: +label=Priority-based Load Balancing
      type: object
      required:
        - type
        - load_balance_targets
      properties:
        type:
          description: |-
            +value=priority-based-routing
            +sort=1
            +label=Routing Type
          type: string
          enum:
            - priority-based-routing
        load_balance_targets:
          description: |-
            +usage=List of targets for priority-based load balancing
            +uiProps={"descriptionInline":true}
            +sort=4
            +label=Load Balance Targets
          type: array
          items:
            $ref: '#/components/schemas/PriorityBasedLoadBalanceTarget'
      title: PriorityBasedLoadBalancing
      x-fern-type-name: PriorityBasedLoadBalancing
    SemanticCacheConfig:
      description: |-
        +label=Semantic Cache Configuration
        +uiType=Ignore
        +required=true
      type: object
      required:
        - type
        - similarity_threshold
        - ttl
      properties:
        type:
          description: |-
            +value=semantic
            +usage=Cache type
          type: string
          enum:
            - semantic
        similarity_threshold:
          description: |-
            +usage=Similarity threshold for semantic matching (between 0 and 1)
            +label=Similarity Threshold
            +sort=100
            +placeholder=Enter value between 0 and 1
          type: number
          minimum: 0
          exclusiveMinimum: true
          maximum: 1
        namespace:
          description: |-
            +usage=Cache namespace (defaults to 'default' if not provided)
            +label=Namespace
            +sort=200
            +placeholder=Enter namespace
          type: string
        ttl:
          description: |-
            +usage=Time-to-live for cached entries in seconds (max 3 days)
            +label=TTL (seconds)
            +sort=300
            +placeholder=Enter TTL in seconds (max: 259200 ~ 3 days)
          type: number
          minimum: 0
          exclusiveMinimum: true
          maximum: 259200
          default: 3600
      title: SemanticCacheConfig
      x-fern-type-name: SemanticCacheConfig
    ExactMatchCacheConfig:
      description: |-
        +label=Exact Match Cache Configuration
        +uiType=Ignore
        +required=true
      type: object
      required:
        - type
        - ttl
      properties:
        type:
          description: |-
            +value=exact-match
            +usage=Cache type
          type: string
          enum:
            - exact-match
        namespace:
          description: |-
            +usage=Cache namespace (defaults to 'default' if not provided)
            +label=Namespace
            +sort=100
            +placeholder=Enter namespace
          type: string
        ttl:
          description: |-
            +usage=Time-to-live for cached entries in seconds (max 3 days)
            +label=TTL (seconds)
            +sort=200
            +placeholder=Enter TTL in seconds (max: 259200 ~ 3 days)
          type: number
          minimum: 0
          exclusiveMinimum: true
          maximum: 259200
          default: 3600
      title: ExactMatchCacheConfig
      x-fern-type-name: ExactMatchCacheConfig
    MCPServerToolDetails:
      description: >-
        +label=Mcp Server Tool Details

        +usage=Mapping configuration for tool calls to MCP server integration
        and tool names
      type: object
      required:
        - mcp_server_integration_id
        - tool_name
      properties:
        mcp_server_integration_id:
          description: |-
            +label=MCP Server Integration ID
            +usage=The integration ID of the MCP server
          type: string
        tool_name:
          description: |-
            +label=Tool Name
            +usage=The name of the tool in the MCP server
          type: string
      title: MCPServerToolDetails
      x-fern-type-name: MCPServerToolDetails
    LoggingConfig:
      description: |-
        +label=Logging Configuration
        +usage=Logging configuration for the chat prompt
      type: object
      required:
        - enabled
      properties:
        enabled:
          description: |-
            +label=Enabled
            +usage=Whether logging is enabled for the chat prompt
          type: boolean
          default: true
        tracing_project_fqn:
          description: |-
            +label=Tracing Project FQN
            +usage=The FQN of the tracing project where the logs will be stored
          type: string
      title: LoggingConfig
      x-fern-type-name: LoggingConfig
    SubAgent:
      description: |-
        +label=Sub Agent
        +usage=Sub agent configuration for the chat prompt
      type: object
      required:
        - name
      properties:
        name:
          description: |-
            +label=Name
            +usage=Name of the sub agent
          type: string
      title: SubAgent
      x-fern-type-name: SubAgent
    Subject:
      type: object
      properties:
        subjectId:
          type: string
          description: Subject ID
        subjectType:
          description: Subject type
          allOf:
            - $ref: '#/components/schemas/SubjectType'
        subjectSlug:
          type: string
          description: Subject slug
        subjectDisplayName:
          type: string
          description: Subject display name
        subjectPatName:
          type: string
          description: Subject PAT name
        subjectControllerName:
          type: string
        subjectExternalIdentitySlug:
          type: string
          description: External identity slug (external_identity_id:sub:email)
      required:
        - subjectId
        - subjectType
      title: Subject
      x-fern-type-name: Subject
    UsageCodeSnippet:
      properties:
        display_name:
          type: string
          title: Display Name
          description: Display name for the code snippet (e.g., 'Python', 'JavaScript')
        language:
          type: string
          title: Language
          description: >-
            Programming language of the code snippet (e.g., 'python',
            'javascript')
        libraries:
          items:
            type: string
          type: array
          title: Libraries
          description: List of libraries/frameworks used in the code snippet
        code:
          type: string
          title: Code
          description: Code snippet demonstrating how to use the prompt version
      type: object
      required:
        - display_name
        - language
        - code
      title: UsageCodeSnippet
      x-fern-type-name: UsageCodeSnippet
    BlobStorageReference:
      properties:
        type:
          type: string
          enum:
            - blob-storage
          title: Type
        path:
          type: string
          pattern: ^.{1,}$
          title: Path
          description: Path to the content in blob storage
        mime_type:
          allOf:
            - $ref: '#/components/schemas/MimeType'
          description: MIME type of the content
      type: object
      required:
        - type
        - path
        - mime_type
      title: BlobStorageReference
      description: A reference to content offloaded to blob storage
      x-fern-type-name: BlobStorageReference
    TextContentPart:
      description: |-
        +label=Text Content
        +usage=Text content for the message
      type: object
      required:
        - type
        - text
      properties:
        type:
          description: |-
            +value=text
            +usage=Type of the content part
          type: string
          enum:
            - text
        text:
          description: |-
            +label=Text
            +usage=Text content for the message
          oneOf:
            - type: string
              pattern: ^.[\s\S]*$
            - $ref: '#/components/schemas/BlobStorageReference'
      title: TextContentPart
      x-fern-type-name: TextContentPart
    RefusalContentPart:
      type: object
      required:
        - type
        - refusal
      properties:
        type:
          description: |-
            +value=refusal
            +usage=Type of the content part
          type: string
        refusal:
          description: |-
            +label=Refusal
            +usage=Reason
          type: string
      title: RefusalContentPart
      x-fern-type-name: RefusalContentPart
    ToolCall:
      type: object
      required:
        - type
        - id
        - function
      properties:
        type:
          description: |-
            +label=Type
            +usage=Type of the tool call
          type: string
          enum:
            - function
        id:
          description: |-
            +label=ID
            +usage=Unique identifier for the tool call
          type: string
        function:
          $ref: '#/components/schemas/Function'
      title: ToolCall
      x-fern-type-name: ToolCall
    ImageContentPart:
      description: |-
        +label=Image Content
        +usage=Image content for the message
      type: object
      required:
        - type
        - image_url
      properties:
        type:
          description: |-
            +value=image_url
            +usage=Type of the content part
          type: string
          enum:
            - image_url
        image_url:
          description: |-
            +label=Image URL
            +usage=Image URL linking to the image
          type: object
          required:
            - url
          properties:
            url:
              description: |-
                +label=URL
                +usage=Image URL linking to the image
              oneOf:
                - type: string
                  pattern: "^\b((https?://)?(?:www\\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\\.[a-zA-Z0-9()]{1,6}(?:[-a-zA-Z0-9()@:%_\\+.~#?&/=]*)|data:image/[a-zA-Z]+;base64,[a-zA-Z0-9+/=\\s]+)$"
                - $ref: '#/components/schemas/BlobStorageReference'
            detail:
              description: +label=Details
              type: string
      title: ImageContentPart
      x-fern-type-name: ImageContentPart
    Parameters:
      properties:
        max_tokens:
          type: integer
          title: Max Tokens
        temperature:
          type: number
          title: Temperature
        top_k:
          type: number
          title: Top K
        top_p:
          type: number
          title: Top P
        stop:
          anyOf:
            - type: array
              items:
                type: string
            - type: string
          title: Stop
      type: object
      title: Parameters
      description: Parameters to pass to the model when generating
      x-fern-type-name: Parameters
    FunctionSchema:
      description: |-
        +label=Function Schema
        +usage=Schema defining a function for tool calls
      type: object
      required:
        - name
      properties:
        name:
          description: |-
            +label=Function Name
            +usage=Name of the function
          type: string
        description:
          description: |-
            +label=Description
            +usage=Description of the function
          type: string
        parameters:
          description: |-
            +label=Parameters
            +usage=Parameters schema for the function
            +uiType=JsonInput
            +uiProps={"descriptionInline":true}
          type: object
        strict:
          description: |-
            +label=Strict
            +usage=Indicates if the function should be called strictly
          type: boolean
          default: false
      title: FunctionSchema
      x-fern-type-name: FunctionSchema
    MCPTool:
      description: |-
        +label=MCP Tool
        +usage=Tool configuration for MCP server
      type: object
      required:
        - name
      properties:
        name:
          description: |-
            +label=Name
            +usage=The name of the tool as it appears in the MCP server
          type: string
      title: MCPTool
      x-fern-type-name: MCPTool
    JsonSchema:
      description: |-
        +label=JSON Schema
        +usage=JSON schema for structured output
      type: object
      required:
        - name
        - schema
      properties:
        name:
          description: |-
            +label=Name
            +usage=Name of the schema
          type: string
        schema:
          description: |-
            +label=Schema
            +usage=JSON schema object defining the expected structure
            +uiType=JsonInput
            +uiProps={"descriptionInline":true}
          type: object
      title: JsonSchema
      x-fern-type-name: JsonSchema
    LoadBalanceTarget:
      description: +label=Weight-based Target
      type: object
      required:
        - target
        - weight
      properties:
        target:
          description: >-
            +usage=Target model or provider FQN

            +uiProps={"integrationType":"model", "valuePattern":
            "${providerAccount.name}/${manifest.name}"}

            +sort=1

            +label=Target
          type: string
        weight:
          description: >-
            +usage=Relative weight for routing decisions (higher values mean
            more traffic)

            +uiProps={"descriptionInline":true}

            +sort=2

            +label=Weight
          type: integer
        retry_config:
          $ref: '#/components/schemas/RetryConfig'
        fallback_status_codes:
          description: >-
            +usage=Status Codes for which the request will fallback to other
            targets. If the status code is not present in fallback_status_codes,
            it fails immediately.

            +uiProps={"descriptionInline":true}

            +label=Fallback Status Codes
          type: array
          items:
            type: string
          default:
            - '401'
            - '403'
            - '404'
            - '429'
            - '500'
            - '502'
            - '503'
        fallback_candidate:
          description: >-
            +usage=Whether this target is a fallback candidate.  If set to
            false, this model will not be considered as a fallback option for
            targets of this load-balance-rule

            +uiProps={"descriptionInline":true}

            +sort=4

            +label=Fallback Candidate
          type: boolean
          default: true
        override_params:
          description: |-
            +usage=Optional parameters to override in the request
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Override Parameters
          type: object
        headers_override:
          $ref: '#/components/schemas/HeadersOverride'
        metadata_match:
          description: >-
            +usage=Optional metadata key-value pairs that must match incoming
            request metadata headers for this target to be considered for
            routing.

            +uiProps={"descriptionInline":true}

            +label=Metadata Match
          type: object
          additionalProperties:
            type: string
      title: LoadBalanceTarget
      x-fern-type-name: LoadBalanceTarget
    StickyRouting:
      description: +label=Sticky Routing
      type: object
      required:
        - ttl_seconds
        - session_identifiers
      properties:
        ttl_seconds:
          description: |-
            +usage=Sticky mapping TTL in seconds
            +uiProps={"descriptionInline":true}
            +label=TTL (seconds)
          type: integer
          minimum: 1800
          maximum: 86400
        session_identifiers:
          description: |-
            +usage=List of identifier selectors used to build sticky key
            +uiProps={"descriptionInline":true}
            +label=Session Identifiers
          type: array
          items:
            $ref: '#/components/schemas/StickySessionIdentifier'
      title: StickyRouting
      x-fern-type-name: StickyRouting
    LatencyBasedLoadBalanceTarget:
      description: +label=Latency-based Target
      type: object
      required:
        - target
      properties:
        target:
          description: |-
            +usage=Target model or provider FQN
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Target
          type: string
        retry_config:
          $ref: '#/components/schemas/RetryConfig'
        fallback_status_codes:
          description: >-
            +usage=Status Codes for which the request will fallback to other
            targets. If the status code is not present in fallback_status_codes,
            it fails immediately.

            +uiProps={"descriptionInline":true}

            +label=Fallback Status Codes
          type: array
          items:
            type: string
          default:
            - '401'
            - '403'
            - '404'
            - '429'
            - '500'
            - '502'
            - '503'
        fallback_candidate:
          description: >-
            +usage=Whether this target is a fallback candidate.  If set to
            false, this model will not be considered as a fallback option for
            targets of this load-balance-rule

            +uiProps={"descriptionInline":true}

            +sort=4

            +label=Fallback Candidate
          type: boolean
          default: true
        override_params:
          description: |-
            +usage=Optional parameters to override in the request
            +uiProps={"descriptionInline":true}
            +sort=2
            +label=Override Parameters
          type: object
        headers_override:
          $ref: '#/components/schemas/HeadersOverride'
        metadata_match:
          description: >-
            +usage=Optional metadata key-value pairs that must match incoming
            request metadata headers for this target to be considered for
            routing.

            +uiProps={"descriptionInline":true}

            +label=Metadata Match
          type: object
          additionalProperties:
            type: string
      title: LatencyBasedLoadBalanceTarget
      x-fern-type-name: LatencyBasedLoadBalanceTarget
    PriorityBasedLoadBalanceTarget:
      type: object
      required:
        - target
        - priority
      properties:
        target:
          description: >-
            +usage=Target model or provider FQN

            +uiProps={"integrationType":"model", "valuePattern":
            "${providerAccount.name}/${manifest.name}"}

            +sort=1

            +label=Target
          type: string
        priority:
          description: >-
            +usage=Priority for the target, Lower the number, higher the
            priority (0 is the highest priority)

            +uiProps={"descriptionInline":true}

            +sort=2

            +label=Priority
          type: integer
          minimum: 0
          default: 0
        sla_cutoff:
          $ref: '#/components/schemas/SLACutoff'
        retry_config:
          $ref: '#/components/schemas/RetryConfig'
        fallback_status_codes:
          description: >-
            +usage=Status Codes for which the request will fallback to other
            targets. If the status code is not present in fallback_status_codes,
            it fails immediately.

            +uiProps={"descriptionInline":true}

            +sort=5

            +label=Fallback Status Codes
          type: array
          items:
            type: string
          default:
            - '401'
            - '403'
            - '404'
            - '429'
            - '500'
            - '502'
            - '503'
        fallback_candidate:
          description: >-
            +usage=Whether this target is a fallback candidate.  If set to
            false, this model will not be considered as a fallback option for
            targets of this load-balance-rule

            +uiProps={"descriptionInline":true}

            +sort=6

            +label=Fallback Candidate
          type: boolean
          default: true
        override_params:
          description: |-
            +usage=Optional parameters to override in the request
            +uiProps={"descriptionInline":true}
            +sort=7
            +label=Override Parameters
          type: object
        headers_override:
          $ref: '#/components/schemas/HeadersOverride'
        metadata_match:
          description: >-
            +usage=Optional metadata key-value pairs that must match incoming
            request metadata headers for this target to be considered for
            routing.

            +uiProps={"descriptionInline":true}

            +label=Metadata Match
          type: object
          additionalProperties:
            type: string
      title: PriorityBasedLoadBalanceTarget
      x-fern-type-name: PriorityBasedLoadBalanceTarget
    SubjectType:
      type: string
      enum:
        - user
        - team
        - serviceaccount
        - virtualaccount
        - external-identity
        - role
      title: SubjectType
      x-fern-type-name: SubjectType
    MimeType:
      type: string
      enum:
        - text/plain
        - application/json
        - image/png
        - image/jpeg
        - application/x-directory
      title: MimeType
      description: MIME type of the content
      x-fern-type-name: MimeType
    Function:
      type: object
      required:
        - name
        - arguments
      properties:
        name:
          description: |-
            +label=Name
            +usage=Name of the tool/function being called
          type: string
        arguments:
          description: |-
            +label=Arguments
            +usage=Arguments passed to the tool/function
          type: string
      title: Function
      x-fern-type-name: Function
    RetryConfig:
      type: object
      required:
        - attempts
      properties:
        attempts:
          description: |-
            +usage=Number of attempts to retry the request
            +uiProps={"descriptionInline":true}
            +label=Attempts
          type: integer
          minimum: 0
        delay:
          description: |-
            +usage=Delay between retries in milliseconds
            +uiProps={"descriptionInline":true}
            +label=Delay
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 100
        on_status_codes:
          description: |-
            +usage=Status codes to retry on
            +uiProps={"descriptionInline":true}
            +label=Status Codes
          type: array
          items:
            type: string
          default:
            - '429'
            - '500'
            - '502'
            - '503'
      title: RetryConfig
      x-fern-type-name: RetryConfig
    HeadersOverride:
      description: +label=Headers Override
      type: object
      properties:
        remove:
          description: |-
            +usage=List of headers to remove from upstream request
            +uiProps={"descriptionInline":true}
            +label=Remove Headers
          type: array
          items:
            type: string
        set:
          description: |-
            +usage=Headers to add or overwrite in upstream request
            +uiProps={"descriptionInline":true}
            +label=Set Headers
          type: object
          additionalProperties:
            type: string
      title: HeadersOverride
      x-fern-type-name: HeadersOverride
    StickySessionIdentifier:
      description: +label=Sticky Session Identifier
      type: object
      required:
        - source
        - key
      properties:
        source:
          description: |-
            +usage=Source from which session identifier value will be extracted
            +uiProps={"descriptionInline":true}
            +label=Source
          type: string
          enum:
            - headers
            - metadata
        key:
          description: |-
            +usage=Key to read from selected source
            +uiProps={"descriptionInline":true}
            +label=Key
          type: string
          pattern: ^[!#$%&'*+.^_`|~0-9A-Za-z-]{1,128}$
      title: StickySessionIdentifier
      x-fern-type-name: StickySessionIdentifier
    SLACutoff:
      type: object
      properties:
        time_per_output_token_ms:
          description: >-
            +usage=Time per output token threshold in milliseconds. If time per
            output token exceeds this value, the target will be marked as
            unhealthy

            +uiProps={"descriptionInline":true}

            +label=Time Per Output Token (ms)
          type: integer
          minimum: 0
          exclusiveMinimum: true
      title: SLACutoff
      x-fern-type-name: SLACutoff
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````