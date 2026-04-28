> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List artifacts

> List artifacts with optional filtering by FQN, ML Repo, name, or run ID.



## OpenAPI

````yaml /openapi.json get /api/ml/v1/artifacts
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
  /api/ml/v1/artifacts:
    get:
      tags:
        - Artifacts
      summary: List artifacts
      description: List artifacts with optional filtering by FQN, ML Repo, name, or run ID.
      operationId: list_artifacts_v1_artifacts_get
      parameters:
        - description: >-
            Fully qualified name to filter artifacts by (format:
            '{artifact_type}:{tenant_name}/{ml_repo_name}/{artifact_name}')
          required: false
          schema:
            type: string
            title: Fqn
            description: >-
              Fully qualified name to filter artifacts by (format:
              '{artifact_type}:{tenant_name}/{ml_repo_name}/{artifact_name}')
          name: fqn
          in: query
        - description: ID of the ML Repo to filter artifacts by
          required: false
          schema:
            type: string
            title: Ml Repo Id
            description: ID of the ML Repo to filter artifacts by
          name: ml_repo_id
          in: query
        - description: Name of the artifact to filter by
          required: false
          schema:
            type: string
            title: Name
            description: Name of the artifact to filter by
          name: name
          in: query
        - description: Number of artifacts to skip for pagination
          required: false
          schema:
            type: integer
            title: Offset
            description: Number of artifacts to skip for pagination
            default: 0
          name: offset
          in: query
        - description: Maximum number of artifacts to return
          required: false
          schema:
            type: integer
            title: Limit
            description: Maximum number of artifacts to return
          name: limit
          in: query
        - description: ID of the run to filter artifacts by
          required: false
          schema:
            type: string
            title: Run Id
            description: ID of the run to filter artifacts by
          name: run_id
          in: query
        - description: Whether to include artifacts that have no versions
          required: false
          schema:
            type: boolean
            title: Include Empty Artifacts
            description: Whether to include artifacts that have no versions
            default: true
          name: include_empty_artifacts
          in: query
      responses:
        '200':
          description: List of artifacts matching the query with pagination information
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListArtifactsResponse'
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
    ListArtifactsResponse:
      properties:
        data:
          items:
            $ref: '#/components/schemas/Artifact'
          type: array
          title: Data
          description: List of artifacts matching the query
        pagination:
          allOf:
            - $ref: '#/components/schemas/Pagination'
          title: Pagination
          description: Pagination information including total count, offset, and limit
      type: object
      required:
        - data
        - pagination
      title: ListArtifactsResponse
      x-fern-type-name: ListArtifactsResponse
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
    Artifact:
      properties:
        id:
          type: string
          title: Id
          description: Unique identifier for the artifact
        ml_repo_id:
          type: string
          title: Ml Repo Id
          description: ID of the ML Repo that this artifact belongs to
        type:
          allOf:
            - $ref: '#/components/schemas/ArtifactType'
          description: >-
            Type of the artifact (e.g., 'artifact', 'model', 'chat_prompt',
            'agent_skill', 'plot', 'image')
        name:
          type: string
          title: Name
          description: >-
            Name of the artifact (alphanumeric characters, hyphens, and
            underscores only, max 256 characters)
        fqn:
          type: string
          title: Fqn
          description: >-
            Fully qualified name of the artifact in the format
            '{artifact_type}:{tenant_name}/{ml_repo_name}/{artifact_name}'
        created_by_subject:
          allOf:
            - $ref: '#/components/schemas/Subject'
          title: Created By Subject
          description: Subject (user, team, or service account) that created this artifact
        created_at:
          type: string
          format: date-time
          title: Created At
          description: Timestamp when the artifact was created
        updated_at:
          type: string
          format: date-time
          title: Updated At
          description: Timestamp when the artifact was last updated
        latest_version:
          allOf:
            - $ref: '#/components/schemas/BaseArtifactVersion'
          title: Latest Version
          description: The most recent version of this artifact
        run_steps:
          items:
            type: integer
          type: array
          title: Run Steps
          description: List of run step numbers where this artifact was created or updated
      type: object
      required:
        - id
        - ml_repo_id
        - type
        - name
        - fqn
        - created_by_subject
      title: Artifact
      description: Generic artifact API DTO including run-step linkage for ML runs.
      x-fern-type-name: Artifact
    Pagination:
      type: object
      properties:
        total:
          type: integer
          description: Total number of items
          example: 100
        offset:
          type: integer
          description: Number of items skipped
          example: 0
        limit:
          type: integer
          description: Number of items per page
          example: 10
      required:
        - total
      title: Pagination
      x-fern-type-name: Pagination
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
    ArtifactType:
      type: string
      enum:
        - artifact
        - model
        - plot
        - image
        - chat_prompt
        - agent_skill
      title: ArtifactType
      description: An enumeration.
      x-fern-type-name: ArtifactType
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
    BaseArtifactVersion:
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
          anyOf:
            - $ref: '#/components/schemas/ModelManifest'
            - $ref: '#/components/schemas/ChatPromptManifest'
            - $ref: '#/components/schemas/ArtifactManifest'
            - $ref: '#/components/schemas/AgentSkillManifest'
          title: Manifest
        ml_repo_id:
          type: string
          title: Ml Repo Id
          description: ID of the ML Repo that this artifact version belongs to
      type: object
      required:
        - id
        - fqn
        - created_by_subject
        - ml_repo_id
      title: BaseArtifactVersion
      description: Shared artifact-version fields (identity, manifest union, ML repo).
      x-fern-type-name: BaseArtifactVersion
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
    ModelManifest:
      properties:
        name:
          type: string
          maxLength: 256
          pattern: ^[A-Za-z0-9_\-]+$
          title: Name
          description: >-
            Name of the model (alphanumeric characters, hyphens, and underscores
            only, max 256 characters)
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
            Name of the ML Repo that this model belongs to (must start and end
            with alphanumeric, 2-100 characters)
        version:
          type: integer
          minimum: 1
          title: Version
          description: Version of the entity
        type:
          type: string
          enum:
            - model-version
          title: Type
          default: model-version
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
        source:
          oneOf:
            - $ref: '#/components/schemas/TrueFoundryManagedSource'
            - $ref: '#/components/schemas/ExternalBlobStorageSource'
            - $ref: '#/components/schemas/LocalModelSource'
          title: Source
          discriminator:
            propertyName: type
            mapping:
              truefoundry:
                $ref: '#/components/schemas/TrueFoundryManagedSource'
              external:
                $ref: '#/components/schemas/ExternalBlobStorageSource'
              local:
                $ref: '#/components/schemas/LocalModelSource'
        framework:
          oneOf:
            - $ref: '#/components/schemas/TransformersFramework'
            - $ref: '#/components/schemas/TensorFlowFramework'
            - $ref: '#/components/schemas/SklearnFramework'
            - $ref: '#/components/schemas/PyTorchFramework'
            - $ref: '#/components/schemas/KerasFramework'
            - $ref: '#/components/schemas/XGBoostFramework'
            - $ref: '#/components/schemas/LightGBMFramework'
            - $ref: '#/components/schemas/FastAIFramework'
            - $ref: '#/components/schemas/H2OFramework'
            - $ref: '#/components/schemas/ONNXFramework'
            - $ref: '#/components/schemas/SpaCyFramework'
            - $ref: '#/components/schemas/StatsModelsFramework'
            - $ref: '#/components/schemas/GluonFramework'
            - $ref: '#/components/schemas/PaddleFramework'
          title: Framework
          description: >-
            Framework for the model version like Transformers, PyTorch, Sklearn,
            Xgboost etc with framework specific metadata. This will be used to
            infer model deployment configuration
          discriminator:
            propertyName: type
            mapping:
              transformers:
                $ref: '#/components/schemas/TransformersFramework'
              tensorflow:
                $ref: '#/components/schemas/TensorFlowFramework'
              sklearn:
                $ref: '#/components/schemas/SklearnFramework'
              pytorch:
                $ref: '#/components/schemas/PyTorchFramework'
              keras:
                $ref: '#/components/schemas/KerasFramework'
              xgboost:
                $ref: '#/components/schemas/XGBoostFramework'
              lightgbm:
                $ref: '#/components/schemas/LightGBMFramework'
              fastai:
                $ref: '#/components/schemas/FastAIFramework'
              h2o:
                $ref: '#/components/schemas/H2OFramework'
              onnx:
                $ref: '#/components/schemas/ONNXFramework'
              spacy:
                $ref: '#/components/schemas/SpaCyFramework'
              statsmodels:
                $ref: '#/components/schemas/StatsModelsFramework'
              gluon:
                $ref: '#/components/schemas/GluonFramework'
              paddle:
                $ref: '#/components/schemas/PaddleFramework'
        environment:
          $ref: '#/components/schemas/ModelVersionEnvironment'
        step:
          type: integer
          minimum: 0
          title: Step
          description: >-
            Step/Epoch number in an iterative training loop the model version
            was created. Generally useful when logging a model version from a
            MLRepo Run
          default: 0
        run_id:
          type: string
          title: Run Id
          description: ID of the MLRepo Run that generated the model version
      type: object
      required:
        - name
        - metadata
        - ml_repo
        - source
      title: ModelManifest
      description: Model Version artifact.
      x-fern-type-name: ModelManifest
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
    ArtifactManifest:
      properties:
        name:
          type: string
          maxLength: 256
          pattern: ^[A-Za-z0-9_\-]+$
          title: Name
          description: >-
            Name of the artifact (alphanumeric characters, hyphens, and
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
            Name of the ML Repo that this artifact belongs to (must start and
            end with alphanumeric, 2-100 characters)
        version:
          type: integer
          minimum: 1
          title: Version
          description: Version of the entity
        type:
          type: string
          enum:
            - artifact-version
          title: Type
          default: artifact-version
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
        source:
          oneOf:
            - $ref: '#/components/schemas/TrueFoundryManagedSource'
            - $ref: '#/components/schemas/ExternalBlobStorageSource'
            - $ref: '#/components/schemas/LocalArtifactSource'
          title: Source
          discriminator:
            propertyName: type
            mapping:
              truefoundry:
                $ref: '#/components/schemas/TrueFoundryManagedSource'
              external:
                $ref: '#/components/schemas/ExternalBlobStorageSource'
              local:
                $ref: '#/components/schemas/LocalArtifactSource'
        step:
          type: integer
          minimum: 0
          title: Step
          description: >-
            Step/Epoch number in an iterative training loop the artifact version
            was created. Generally useful when logging a model version from a
            MLRepo Run
          default: 0
        run_id:
          type: string
          title: Run Id
          description: ID of the MLRepo Run that generated the artifact version
      type: object
      required:
        - name
        - metadata
        - ml_repo
        - source
      title: ArtifactManifest
      description: Artifact Version manifest.
      x-fern-type-name: ArtifactManifest
    AgentSkillManifest:
      properties:
        name:
          type: string
          maxLength: 64
          minLength: 1
          pattern: ^(?!.*anthropic)(?!.*claude)[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$
          title: Name
          description: Name of the skill.
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
          description: Name of the ML Repo that this agent skill belongs to
        version:
          type: integer
          minimum: 1
          title: Version
          description: Version of the entity
        type:
          type: string
          enum:
            - agent_skill
          title: Type
          default: agent_skill
        source:
          oneOf:
            - $ref: '#/components/schemas/AgentSkillSourceInline'
            - $ref: '#/components/schemas/AgentSkillSourceBlobStorage'
          title: Source
          discriminator:
            propertyName: type
            mapping:
              inline:
                $ref: '#/components/schemas/AgentSkillSourceInline'
              blob-storage:
                $ref: '#/components/schemas/AgentSkillSourceBlobStorage'
      type: object
      required:
        - name
        - metadata
        - ml_repo
        - source
      title: AgentSkillManifest
      description: >-
        Apply/API manifest: ML Repo and versioning; ``source`` is inline (apply)
        or blob-storage (stored).
      x-fern-type-name: AgentSkillManifest
    TrueFoundryManagedSource:
      properties:
        type:
          type: string
          enum:
            - truefoundry
          title: Type
          description: Type of the source
        uri:
          type: string
          pattern: ^(s3|gs|wasbs)://.*
          title: Uri
      type: object
      required:
        - type
      title: TrueFoundryManagedSource
      x-fern-type-name: TrueFoundryManagedSource
    ExternalBlobStorageSource:
      properties:
        type:
          type: string
          enum:
            - external
          title: Type
          description: Type of the source
        uri:
          type: string
          pattern: ^(s3|gs|wasbs)://.*
          title: Uri
          description: >-
            URI referencing a path in the blob storage bucket linked to the
            MLRepo
      type: object
      required:
        - type
        - uri
      title: ExternalBlobStorageSource
      x-fern-type-name: ExternalBlobStorageSource
    LocalModelSource:
      properties:
        type:
          type: string
          enum:
            - local
          title: Type
          description: Type of the source
        file_or_folder:
          type: string
          title: File Or Folder
          description: Path to the model file or folder
      type: object
      required:
        - type
        - file_or_folder
      title: LocalModelSource
      x-fern-type-name: LocalModelSource
    TransformersFramework:
      properties:
        type:
          type: string
          enum:
            - transformers
          title: Type
          description: Type of the framework
        library_name:
          allOf:
            - $ref: '#/components/schemas/LibraryName'
          description: Name of the library for the framework
          default: transformers
        pipeline_tag:
          type: string
          title: Pipeline Tag
          description: >-
            The `pipeline()` task this model can be used with e.g.
            `text-generation`. See [huggingface
            docs](https://huggingface.co/docs/transformers/main/en/main_classes/pipelines#transformers.pipeline.task)
            for all possible values
        base_model:
          type: string
          title: Base Model
          description: >-
            Base model Id from HuggingFace Hub. If this is a finetuned model,
            this points to the base model id used for finetuning.
      type: object
      required:
        - type
      title: TransformersFramework
      x-fern-type-name: TransformersFramework
    TensorFlowFramework:
      properties:
        type:
          type: string
          enum:
            - tensorflow
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: TensorFlowFramework
      x-fern-type-name: TensorFlowFramework
    SklearnFramework:
      properties:
        type:
          type: string
          enum:
            - sklearn
          title: Type
          description: Type of the framework
        model_filepath:
          type: string
          title: Model Filepath
          description: Relative path to the model file in the model version contents
        serialization_format:
          $ref: '#/components/schemas/SklearnSerializationFormat'
        model_schema:
          $ref: '#/components/schemas/SklearnModelSchema'
      type: object
      required:
        - type
      title: SklearnFramework
      x-fern-type-name: SklearnFramework
    PyTorchFramework:
      properties:
        type:
          type: string
          enum:
            - pytorch
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: PyTorchFramework
      x-fern-type-name: PyTorchFramework
    KerasFramework:
      properties:
        type:
          type: string
          enum:
            - keras
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: KerasFramework
      x-fern-type-name: KerasFramework
    XGBoostFramework:
      properties:
        type:
          type: string
          enum:
            - xgboost
          title: Type
          description: Type of the framework
        serialization_format:
          $ref: '#/components/schemas/XGBoostSerializationFormat'
        model_filepath:
          type: string
          title: Model Filepath
          description: Relative path to the model file in the model version contents
        model_schema:
          $ref: '#/components/schemas/XGBoostModelSchema'
      type: object
      required:
        - type
      title: XGBoostFramework
      x-fern-type-name: XGBoostFramework
    LightGBMFramework:
      properties:
        type:
          type: string
          enum:
            - lightgbm
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: LightGBMFramework
      x-fern-type-name: LightGBMFramework
    FastAIFramework:
      properties:
        type:
          type: string
          enum:
            - fastai
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: FastAIFramework
      x-fern-type-name: FastAIFramework
    H2OFramework:
      properties:
        type:
          type: string
          enum:
            - h2o
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: H2OFramework
      x-fern-type-name: H2OFramework
    ONNXFramework:
      properties:
        type:
          type: string
          enum:
            - onnx
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: ONNXFramework
      x-fern-type-name: ONNXFramework
    SpaCyFramework:
      properties:
        type:
          type: string
          enum:
            - spacy
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: SpaCyFramework
      x-fern-type-name: SpaCyFramework
    StatsModelsFramework:
      properties:
        type:
          type: string
          enum:
            - statsmodels
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: StatsModelsFramework
      x-fern-type-name: StatsModelsFramework
    GluonFramework:
      properties:
        type:
          type: string
          enum:
            - gluon
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: GluonFramework
      x-fern-type-name: GluonFramework
    PaddleFramework:
      properties:
        type:
          type: string
          enum:
            - paddle
          title: Type
          description: Type of the framework
      type: object
      required:
        - type
      title: PaddleFramework
      x-fern-type-name: PaddleFramework
    ModelVersionEnvironment:
      properties:
        python_version:
          type: string
          pattern: ^\d+(\.\d+){1,2}([\-\.a-z0-9]+)?$
          title: Python Version
          description: Python version to use when running the model
        pip_packages:
          items:
            type: string
          type: array
          title: Pip Packages
          description: pip dependencies needed for running the model
      type: object
      title: ModelVersionEnvironment
      description: >-
        Runtime environment describing python version and dependencies for
        running the model
      x-fern-type-name: ModelVersionEnvironment
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
    LocalArtifactSource:
      properties:
        type:
          type: string
          enum:
            - local
          title: Type
          description: Type of the source
        paths:
          items:
            $ref: '#/components/schemas/ArtifactPath'
          type: array
          title: Paths
          description: >-
            Array of ArtifactPath objects representing the source and
            destination paths
      type: object
      required:
        - type
        - paths
      title: LocalArtifactSource
      x-fern-type-name: LocalArtifactSource
    AgentSkillSourceInline:
      properties:
        type:
          type: string
          enum:
            - inline
          title: Type
        skill_md:
          type: string
          maxLength: 20000
          minLength: 1
          title: Skill Md
          description: Main body of SKILL.md file.
      type: object
      required:
        - type
        - skill_md
      title: AgentSkillSourceInline
      description: SKILL.md for this version when using inline source.
      x-fern-type-name: AgentSkillSourceInline
    AgentSkillSourceBlobStorage:
      properties:
        type:
          type: string
          enum:
            - blob-storage
          title: Type
        description:
          type: string
          maxLength: 1024
          minLength: 1
          title: Description
          description: 1–1024 chars.
      type: object
      required:
        - type
        - description
      title: AgentSkillSourceBlobStorage
      description: SKILL.md content stored on blob
      x-fern-type-name: AgentSkillSourceBlobStorage
    LibraryName:
      type: string
      enum:
        - transformers
        - sentence-transformers
        - diffusers
      title: LibraryName
      description: Name of the library for the framework
      x-fern-type-name: LibraryName
    SklearnSerializationFormat:
      type: string
      enum:
        - cloudpickle
        - joblib
        - pickle
      title: SklearnSerializationFormat
      description: Serialization format used to save the sklearn model
      x-fern-type-name: SklearnSerializationFormat
    SklearnModelSchema:
      properties:
        infer_method_name:
          allOf:
            - $ref: '#/components/schemas/InferMethodName'
          description: Name of the method used for inference
        inputs:
          items:
            type: object
          type: array
          title: Inputs
          description: Schema of the input
        outputs:
          items:
            type: object
          type: array
          title: Outputs
          description: Schema of the output
      type: object
      required:
        - infer_method_name
        - inputs
        - outputs
      title: SklearnModelSchema
      x-fern-type-name: SklearnModelSchema
    XGBoostSerializationFormat:
      type: string
      enum:
        - cloudpickle
        - joblib
        - pickle
        - json
      title: XGBoostSerializationFormat
      description: Serialization format used to save the xgboost model
      x-fern-type-name: XGBoostSerializationFormat
    XGBoostModelSchema:
      properties:
        infer_method_name:
          type: string
          enum:
            - predict
          title: Infer Method Name
          description: Name of the method used for inference
        inputs:
          items:
            type: object
          type: array
          title: Inputs
          description: Schema of the input
        outputs:
          items:
            type: object
          type: array
          title: Outputs
          description: Schema of the output
      type: object
      required:
        - infer_method_name
        - inputs
        - outputs
      title: XGBoostModelSchema
      description: Schema of the XGBoost model
      x-fern-type-name: XGBoostModelSchema
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
    ArtifactPath:
      properties:
        src:
          type: string
          title: Src
          description: Local file or folder path
        dest:
          type: string
          title: Dest
          description: >-
            Relative path where the file or folder will be uploaded to in the
            artifact
      type: object
      required:
        - src
      title: ArtifactPath
      x-fern-type-name: ArtifactPath
    InferMethodName:
      type: string
      enum:
        - predict
        - predict_proba
      title: InferMethodName
      description: Name of the method used for inference
      x-fern-type-name: InferMethodName
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