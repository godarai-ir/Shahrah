> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Delete Resources

> Deletes resources of specific types, such as provider-account, cluster, workspace, or application.



## OpenAPI

````yaml /openapi.json post /api/svc/v1/delete
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
  /api/svc/v1/delete:
    post:
      tags:
        - Apply
      summary: Delete Resources
      description: >-
        Deletes resources of specific types, such as provider-account, cluster,
        workspace, or application.
      operationId: TfyApply.delete
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/TrueFoundryDeleteRequest'
      responses:
        '200':
          description: The resource has been successfully deleted.
      security:
        - jwt: []
components:
  schemas:
    TrueFoundryDeleteRequest:
      type: object
      properties:
        manifest:
          description: manifest of the resource to be deleted
          oneOf:
            - $ref: '#/components/schemas/MLRepoManifest'
            - $ref: '#/components/schemas/ArtifactManifest'
            - $ref: '#/components/schemas/ModelManifest'
            - $ref: '#/components/schemas/ChatPromptManifest'
            - bcbba96f-4365-4578-ae47-ce00df9232bc
            - $ref: '#/components/schemas/Service'
            - $ref: '#/components/schemas/ApplicationSet'
            - $ref: '#/components/schemas/ProviderAccounts'
            - $ref: '#/components/schemas/ClusterManifest'
            - $ref: '#/components/schemas/WorkspaceManifest'
            - $ref: '#/components/schemas/Job'
            - $ref: '#/components/schemas/Helm'
            - $ref: '#/components/schemas/Volume'
            - $ref: '#/components/schemas/Notebook'
            - $ref: '#/components/schemas/RStudio'
            - $ref: '#/components/schemas/Workflow'
            - $ref: '#/components/schemas/AsyncService'
            - $ref: '#/components/schemas/SSHServer'
            - $ref: '#/components/schemas/SparkJob'
            - $ref: '#/components/schemas/GatewayConfig'
            - $ref: '#/components/schemas/TeamManifest'
            - $ref: '#/components/schemas/PolicyManifest'
            - $ref: '#/components/schemas/RoleManifest'
            - $ref: '#/components/schemas/AlertConfig'
            - $ref: '#/components/schemas/VirtualAccountManifest'
            - $ref: '#/components/schemas/CommonToolsSettings'
            - $ref: '#/components/schemas/AIFeaturesSettings'
            - $ref: '#/components/schemas/SecretGroupManifest'
            - $ref: '#/components/schemas/AgentManifest'
            - $ref: '#/components/schemas/EnvironmentManifest'
            - $ref: '#/components/schemas/TracingProjectManifest'
            - $ref: '#/components/schemas/MCPServerManifest'
            - $ref: '#/components/schemas/RoleBindingManifest'
      required:
        - manifest
      title: TrueFoundryDeleteRequest
      x-fern-type-name: TrueFoundryDeleteRequest
    MLRepoManifest:
      description: >-
        +label=MLRepo

        +usage=MLRepo is a repository ML training runs that log params, metrics,
        plots, images and versioned entities like artifacts, models, prompts,
        tools, agents
      type: object
      required:
        - type
        - name
        - storage_integration_fqn
        - collaborators
      properties:
        type:
          description: +value=ml-repo
          type: string
          enum:
            - ml-repo
        name:
          description: >-
            +label=Name

            +icon=fa-desktop:#black

            +usage=Name of the ML Repo.

            +message=Alphanumeric word, may contain '-' with a maximum length of
            100 characters

            +sort=1
          type: string
          pattern: ^[a-zA-Z][a-zA-Z0-9\-]{1,98}[a-zA-Z0-9]$
        description:
          description: >-
            +label=Description

            +icon=fa-desktop:#black

            +usage=Description for the ML Repo.

            +message=Description can be maximum 500 characters

            +sort=2

            +placeholder=MLRepo to track text-classification experiments and
            models

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^.{0,500}$
        storage_integration_fqn:
          description: >-
            +label=Storage Integration

            +icon=hard-drive

            +usage=Storage Integration to store artifacts and models. A storage
            integration represents a connected blob storage like AWS S3 / Azure
            Blob Storage / Google Cloud Storage.

            +message=Select the storage-integration where you want to save your
            artifacts and models

            [View
            Docs](https://www.truefoundry.com/docs/integrations-overview#integrations)

            +sort=4
          type: string
        collaborators:
          description: |-
            +label=Collaborators
            +sort=5
            +usage=Users and Teams that have access to MLRepo
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: MLRepoManifest
      x-fern-type-name: MLRepoManifest
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
    Service:
      description: +docs=Describes the configuration for the service
      type: object
      properties:
        type:
          description: +value=service
          type: string
          enum:
            - service
        replicas:
          description: >-
            +label=Replicas

            +usage=Deploy multiple instances of your pods to distribute incoming
            traffic across them, ensuring effective load balancing.

            +icon=fa-clone

            +sort=4
          default: 1
          oneOf:
            - type: number
              minimum: 0
              maximum: 500
            - $ref: '#/components/schemas/ServiceAutoscaling'
        auto_shutdown:
          $ref: '#/components/schemas/Autoshutdown'
        allow_interception:
          description: >-
            +label=Allow intercepts

            +usage=Whether to allow intercepts to be applied for this service.

            This would inject an additional sidecar in each pod of the service.
            Not recommended on production
          type: boolean
          default: false
        rollout_strategy:
          description: >-
            +label=Rollout strategy

            +usage=Strategy to dictate how a rollout should happen when a new
            release for this service is made
            [Docs](https://docs.truefoundry.com/docs/rollout-strategy)

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/Rolling'
            - $ref: '#/components/schemas/Canary'
            - $ref: '#/components/schemas/BlueGreen'
          discriminator:
            propertyName: type
      allOf:
        - $ref: '#/components/schemas/BaseService'
      required:
        - type
        - replicas
      title: Service
      x-fern-type-name: Service
    ApplicationSet:
      description: +docs=Describes the configuration for the application set
      type: object
      required:
        - type
        - name
      properties:
        type:
          description: +value=application-set
          type: string
          enum:
            - application-set
        name:
          description: |-
            +label=Name
            +usage=Name of the application set.
            +icon=fa-font
          type: string
        components:
          description: |-
            +label=Components
            +usage=Array of components with their specifications.
            +icon=fa-puzzle-piece
            +uiType=AppComponents
            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/Service'
              - $ref: '#/components/schemas/AsyncService'
              - $ref: '#/components/schemas/Job'
              - $ref: '#/components/schemas/Helm'
        template:
          description: |-
            +label=Template
            +usage=Template to be used for the application set.
            +icon=fa-file
            +uiType=Hidden
          type: string
        values:
          description: >-
            +label=Values

            +usage=Values to be used to render components for the application
            set.

            +icon=fa-file

            +uiType=YamlInput

            +uiProps={"allowAllValues":true}
          type: object
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
        convert_template_manifest:
          description: >-
            +label=Convert Template Manifest

            +docs=Flag to indicate if the template manifest should be converted
            to TrueFoundry manifest

            +uiType=Hidden
          type: boolean
      title: ApplicationSet
      x-fern-type-name: ApplicationSet
    ProviderAccounts:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/InfraProviderAccount'
        - $ref: '#/components/schemas/ModelProviderAccount'
        - $ref: '#/components/schemas/MCPServerProviderAccount'
        - $ref: '#/components/schemas/GuardrailConfigGroup'
      discriminator:
        propertyName: type
      title: ProviderAccounts
      x-fern-type-name: ProviderAccounts
    ClusterManifest:
      type: object
      required:
        - type
        - name
        - cluster_type
        - environment_names
        - collaborators
      properties:
        type:
          description: +value=cluster
          type: string
          enum:
            - cluster
        name:
          description: >-
            +label=Name

            +icon=fa-desktop:#black

            +sort=10

            +message=3 to 35 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,33}[a-z0-9]$
        cluster_type:
          description: |-
            +label=Cluster Type
            +icon=fa-cloud:#7EC8E3
            +sort=20
            +usage=The cluster type of this cluster
          type: string
          enum:
            - aws-eks
            - gcp-gke-standard
            - azure-aks
            - generic
            - civo-talos
        environment_names:
          description: >-
            +label=Environments

            +icon=fa-tags:#black

            +usage=Tag the environment like dev, staging or production. You will
            need to [create
            environments](https://docs.truefoundry.com/docs/key-concepts#creating-environments)
            if you don't have already.

            +sort=30
          type: array
          items:
            type: string
        base_domains:
          description: >-
            +label=Base Domain URLs

            +placeholder=Enter Base Domain URL eg. devtest.mydomain.com or
            *.mydomain.com

            +usage=One or more base URLs, which can be either a wildcard domain
            (resulting in app URLs like
            `<app-name>-<workspace-name>.<base-domain-URL>`) or a non-wildcard
            domain (resulting in app URLs like
            `<base-domain-URL>/<app-name>-<workspace-name>`)

            > Read more about routing
            [here](https://docs.truefoundry.com/docs/routing) and about 'A'
            record and domain mapping
            [here](https://docs.truefoundry.com/docs/setting-up-domain)

            +icon=fa-globe:#68BBE3

            +sort=40

            +message=Domain URL should be valid domain or IP address
          type: array
          items:
            type: string
            pattern: >-
              ^((((\*|[a-zA-Z0-9\-]{1,63})\.)([a-zA-Z0-9\-]{1,63}\.)*([A-Za-z]{1,63}))|(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))$
        monitoring:
          description: |-
            +label=Monitoring
            +icon=fa-gear:#68BBE3
            +sort=50
          type: object
          properties:
            loki_url:
              description: +label=Cluster Loki URL
              type: string
            victoria_logs_url:
              description: +label=Cluster VictoriaLogs URL
              type: string
            prometheus_url:
              description: +label=Cluster Prometheus URL
              type: string
            kubecost_url:
              description: +label=Cluster Kubecost URL
              type: string
        default_registry_fqn:
          description: |-
            +label=Cluster Default Registry
            +sort=60
          type: string
        workbench_config:
          description: |-
            +label=Workbench Config
            +sort=70
          type: object
          properties:
            notebook_config:
              $ref: '#/components/schemas/NotebookConfig'
            ssh_server_config:
              $ref: '#/components/schemas/SSHServerConfig'
            default_storage_class:
              description: >-
                +label=Default Storage Class

                +usage=The default storage class for the home directory of
                workbench

                +message=Must not contain any spaces.
              type: string
              pattern: ^[^\s]*$
        spark_config:
          $ref: '#/components/schemas/SparkConfig'
        ingress_controller_config:
          $ref: '#/components/schemas/IngressControllerConfig'
        cluster_integration_fqn:
          description: >-
            +label=Cluster Integration FQN

            +sort=75

            +uiType=IntegrationSelect

            +uiProps={"descriptionKey": "cluster_name","integrationType":
            "cluster"}
          type: string
        workflow_storage_integration_fqn:
          description: >-
            +label=Workflow Storage Integration

            +sort=80

            +uiType=IntegrationSelect

            +uiProps={"descriptionKey": "cluster_name","integrationType":
            "blob-storage"}
          type: string
        supported_nodepools:
          description: >-
            +label=Nodepools

            +usage=Add nodepools that are already created in your cluster.

            When deploying, applications can choose to schedule from these
            nodepools.

            +sort=90
          type: array
          items:
            $ref: '#/components/schemas/Nodepool'
        node_label_keys:
          description: >-
            +label=Node Label Keys

            +usage=The node label keys that this cluster supports.

            Note: You will additionally need to add `truefoundry.com/gpu_type`
            label for GPU-supported node pools.

            For more information, check out [this
            documentation](https://docs.truefoundry.com/docs/generic-control-plane#configuring-node-pools-for-truefoundry).

            +sort=100
          type: object
          required:
            - nodepool_selector_label
          properties:
            nodepool_selector_label:
              description: |-
                +label=Nodepool Selector Label
                +usage=The nodepool selector label that this cluster supports
              type: string
        collaborators:
          description: |-
            +label=Collaborators
            +usage=Collaborators who can access this cluster
            +sort=110
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: ClusterManifest
      x-fern-type-name: ClusterManifest
    WorkspaceManifest:
      type: object
      required:
        - type
        - cluster_fqn
        - name
      properties:
        type:
          description: +value=workspace
          type: string
          enum:
            - workspace
        cluster_fqn:
          description: |-
            +label=Cluster FQN
            +sort=150
            +icon=fa-spider-web:#326ce5
            +uiProps={"disableCloudProvider": true}
          type: string
        name:
          description: >-
            +label=Name

            +sort=200

            +icon=fa-desktop:#black

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        environment_name:
          description: >-
            +label=Environment

            +sort=300

            +icon=fa-tags:#black

            +usage=Tag the environment like dev, staging or production. You will
            need to [create
            environments](https://docs.truefoundry.com/docs/key-concepts#creating-environments)
            if you don't have already.
          type: string
        labels:
          description: +sort=700
          type: object
          additionalProperties:
            type: string
        annotations:
          description: +sort=750
          type: object
          additionalProperties:
            type: string
        collaborators:
          description: |-
            +label=Collaborators
            +sort=500
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        permissions:
          description: >-
            +label=Permissions

            +sort=500

            +uiType=Permissions

            +uiProps={"allowedResources": {"mlf-project": {"title": "ML Repo",
            "description": "ML Repos are used to store models and artifacts like
            git repos. MLRepos are backed by S3 / GCS / Azure bucket and you can
            have different ml repos with different set of permissions. [View
            Docs](https://docs.truefoundry.com/docs/introduction-to-ml-repo)"}}}
          type: array
          items:
            $ref: '#/components/schemas/Permissions'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: WorkspaceManifest
      x-fern-type-name: WorkspaceManifest
    Job:
      description: +docs=Describes the configuration for the job
      type: object
      required:
        - type
        - name
        - image
        - trigger
      properties:
        type:
          description: +value=job
          type: string
          enum:
            - job
        name:
          description: >-
            +usage=Name of the job

            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        image:
          description: >-
            +docs=Specify whether you want to deploy a Docker image or build and
            deploy from source code

            +label=Deploy a Docker image or build and deploy from source code

            +icon=fa-solid fa-cloud-arrow-up:#21B6A8

            +sort=200

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/Build'
            - $ref: '#/components/schemas/Image'
          discriminator:
            propertyName: type
        trigger:
          description: |-
            +docs=Specify the trigger
            +sort=300
            +discriminator=type
          type: object
          default:
            type: manual
          oneOf:
            - $ref: '#/components/schemas/Manual'
            - $ref: '#/components/schemas/Schedule'
          discriminator:
            propertyName: type
        trigger_on_deploy:
          description: |-
            +docs=Trigger on deploy
            +sort=350
            +usage=Trigger the job after deploy immediately
          type: boolean
        params:
          description: |-
            +label=Params for input
            +uiType=StructsTable
            +usage=Configure params and pass it to create different job runs
            +sort=400
          type: array
          items:
            $ref: '#/components/schemas/Param'
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the service
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=500
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        resources:
          $ref: '#/components/schemas/Resources'
        alerts:
          description: >-
            +label=Alerts

            +usage=Configure alerts to be sent when the job
            starts/fails/completes

            +icon=fa-bell

            +sort=650
          type: array
          items:
            $ref: '#/components/schemas/JobAlert'
        retries:
          description: >-
            +label=Retries

            +usage=Specify the maximum number of attempts to retry a job before
            it is marked as failed.

            +icon=fa-repeat

            +sort=700
          type: integer
          minimum: 0
          maximum: 10
          default: 0
        timeout:
          description: |-
            +label=Timeout
            +usage=Job timeout in seconds.
            +icon=fa-clock
            +sort=800
          type: integer
          minimum: 0
          exclusiveMinimum: true
          maximum: 432000
        concurrency_limit:
          description: |-
            +label=Concurrency Limit
            +usage=Number of runs that can run concurrently
            +icon=fa-copy
            +sort=900
          type: integer
          minimum: 0
          exclusiveMinimum: true
        service_account:
          description: +sort=1000
          type: string
        mounts:
          description: >-
            +usage=Configure data to be mounted to job pod(s) as a string,
            secret or volume.
            [Docs](https://docs.truefoundry.com/docs/mounting-volumes-job)

            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/SecretMount'
              - $ref: '#/components/schemas/StringDataMount'
              - $ref: '#/components/schemas/VolumeMount'
        labels:
          description: +label=Labels
          type: object
          additionalProperties:
            type: string
        kustomize:
          $ref: '#/components/schemas/Kustomize'
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: Job
      x-fern-type-name: Job
    Helm:
      type: object
      required:
        - type
        - name
        - source
      properties:
        type:
          description: +value=helm
          type: string
          enum:
            - helm
        name:
          description: >-
            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +usage=Name of the Helm deployment. This will be set as the release
            name of the chart you are deploying.
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        labels:
          description: |-
            +label=Labels
            +usage=Add labels to base argo app
          type: object
          additionalProperties:
            type: string
        source:
          description: |-
            +label=Source helm repository
            +sort=2
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/HelmRepo'
            - $ref: '#/components/schemas/OCIRepo'
            - $ref: '#/components/schemas/GitHelmRepo'
          discriminator:
            propertyName: type
        values:
          description: |-
            +label=Values
            +usage=Values file as block file
          type: object
        kustomize:
          $ref: '#/components/schemas/Kustomize'
        ignoreDifferences:
          type: array
          items:
            type: object
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: Helm
      x-fern-type-name: Helm
    Volume:
      type: object
      required:
        - type
        - name
        - config
      properties:
        type:
          description: +value=volume
          type: string
          enum:
            - volume
        name:
          description: >-
            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +usage=Name of the Volume. This will be set as the volume name.
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        config:
          description: >-
            +sort=2

            +label=Volume Config

            +message=Volume Configuration, can be either Dynamically provisioned
            or statically provisioned.

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/DynamicVolumeConfig'
            - $ref: '#/components/schemas/StaticVolumeConfig'
          discriminator:
            propertyName: type
        volume_browser:
          $ref: '#/components/schemas/VolumeBrowser'
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: Volume
      x-fern-type-name: Volume
    Notebook:
      description: +docs=Describes the configuration for the service
      type: object
      properties:
        type:
          description: +value=notebook
          type: string
          enum:
            - notebook
        image:
          $ref: '#/components/schemas/WorkbenchImage'
        cull_timeout:
          description: >-
            +label=Stop after (minutes of inactivity)

            +usage=Stop the notebook instance after this much time in minutes of
            inactivity.

            The notebook instance will be stopped even if the notebook is open
            in your browser, but nothing is running on the notebook.

            +sort=5

            +uiProps={"descriptionInline":true}
          type: integer
          minimum: 5
          default: 30
      allOf:
        - $ref: '#/components/schemas/BaseWorkbenchInput'
      required:
        - type
        - image
      title: Notebook
      x-fern-type-name: Notebook
    RStudio:
      description: +docs=Describes the configuration for the Rstudio server
      type: object
      properties:
        type:
          description: +value=rstudio
          type: string
          enum:
            - rstudio
        image:
          $ref: '#/components/schemas/WorkbenchImage'
      allOf:
        - $ref: '#/components/schemas/BaseWorkbenchInput'
      required:
        - type
        - image
      title: RStudio
      x-fern-type-name: RStudio
    Workflow:
      description: +docs=Describes the configuration for the worflow
      type: object
      required:
        - type
        - name
        - source
        - workflow_file_path
      properties:
        type:
          description: +value=workflow
          type: string
          enum:
            - workflow
        name:
          description: >-
            +usage=Name of the workflow

            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        source:
          description: |-
            +docs=Source Code for the workflow, either local or remote
            +label=Source Code for your workflow
            +icon=fa-solid fa-cloud-arrow-up:#21B6A8
            +sort=200
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/LocalSource'
            - $ref: '#/components/schemas/RemoteSource'
          discriminator:
            propertyName: type
        workflow_file_path:
          description: |-
            +label=Workflow File Path
            +docs=Path to the workflow file relative to the project root path
            +sort=550
          type: string
        flyte_entities:
          description: |-
            +label=Flyte Entities
            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/FlyteTask'
              - $ref: '#/components/schemas/FlyteWorkflow'
              - $ref: '#/components/schemas/FlyteLaunchPlan'
        alerts:
          description: +label=Alerts
          type: array
          items:
            $ref: '#/components/schemas/WorkflowAlert'
      title: Workflow
      x-fern-type-name: Workflow
    AsyncService:
      description: +docs=Describes the configuration for the async-service
      type: object
      properties:
        type:
          description: +value=async-service
          type: string
          enum:
            - async-service
        replicas:
          description: >-
            +label=Replicas

            +usage=Deploy multiple instances of your pods to distribute incoming
            traffic across them, ensuring effective load balancing.

            +icon=fa-clone
          default: 1
          oneOf:
            - type: number
              minimum: 0
              maximum: 500
            - $ref: '#/components/schemas/AsyncServiceAutoscaling'
        rollout_strategy:
          $ref: '#/components/schemas/Rolling'
        worker_config:
          $ref: '#/components/schemas/WorkerConfig'
        sidecar:
          $ref: '#/components/schemas/AsyncProcessorSidecar'
      allOf:
        - $ref: '#/components/schemas/BaseService'
      required:
        - type
        - replicas
        - worker_config
      title: AsyncService
      x-fern-type-name: AsyncService
    SSHServer:
      description: +docs=Describes the configuration for the ssh server
      type: object
      properties:
        type:
          description: +value=ssh-server
          type: string
          enum:
            - ssh-server
        image:
          $ref: '#/components/schemas/WorkbenchImage'
        ssh_public_key:
          description: >-
            +label: SSH Public Key

            +usage=Add Your SSH Public Key, this will be used to authenticate
            you to the SSH Server.  \

            You can find it using `cat ~/.ssh/id_rsa.pub` in Mac/Linux or `type
            $home\.ssh\id_rsa.pub` in Windows Powershell.  \

            You can also generate a new SSH key pair using `ssh-keygen -t rsa`
            in your local terminal. (same for both Mac/Linux and Windows
            Powershell)

            +uiType=TextArea

            +sort=4
          type: string
        cull_timeout:
          description: >-
            +label=Stop after (minutes of inactivity)

            +usage=Stop the SSH Server instance after this much time in minutes
            of inactivity. The instance is considered active if there is at
            least one active SSH connection (a client connected to the SSH
            server), or if a background job is running using tmux or screen, or
            if the pod has restarted.

            +sort=5

            +uiProps={"descriptionInline":true, "warningMessage":"Please note
            that stop after inactivity is only available for images with
            tag(including custom images) >= v0.3.10"}
          type: integer
          minimum: 5
      allOf:
        - $ref: '#/components/schemas/BaseWorkbenchInput'
      required:
        - type
        - image
        - ssh_public_key
      title: SSHServer
      x-fern-type-name: SSHServer
    SparkJob:
      type: object
      required:
        - type
        - name
        - image
        - entrypoint
        - driver_config
        - executor_config
      properties:
        type:
          description: |-
            +value=spark-job
            +sort=1
          type: string
          enum:
            - spark-job
        name:
          description: >-
            +label=Name

            +usage=Name of the job

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +sort=2
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        image:
          description: >-
            +label=Deploy a Docker image & Specify Spark Version

            +usage=The image to use for driver and executors. Must have spark
            installed. Spark version must match the version in the image.

            +sort=500

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/SparkImage'
            - $ref: '#/components/schemas/SparkImageBuild'
          discriminator:
            propertyName: type
        entrypoint:
          description: |-
            +label=Entrypoint
            +sort=1000
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/SparkJobPythonEntrypoint'
            - $ref: '#/components/schemas/SparkJobScalaEntrypoint'
            - $ref: '#/components/schemas/SparkJobJavaEntrypoint'
            - $ref: '#/components/schemas/SparkJobPythonNotebookEntrypoint'
            - $ref: '#/components/schemas/SparkJobScalaNotebookEntrypoint'
          discriminator:
            propertyName: type
        driver_config:
          $ref: '#/components/schemas/SparkDriverConfig'
        executor_config:
          $ref: '#/components/schemas/SparkExecutorConfig'
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the service
            either as plain text.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=21000
          type: object
          default: null
          nullable: true
        spark_conf:
          description: >-
            +label=Spark Config Properties

            +usage=Extra configuration properties to be passed to the spark job.
            [Docs](https://spark.apache.org/docs/latest/configuration.html)

            +icon=fa-gear:#68BBE3

            +sort=21500
          type: object
          default: null
          additionalProperties:
            type: string
            minLength: 1
          nullable: true
        mounts:
          description: >-
            +label=Mounts

            +usage=Configure volumes to be mounted to driver and executors.
            [Docs](https://docs.truefoundry.com/docs/mounting-volumes-job)

            +sort=22000

            +uiType=Mounts
          type: array
          items:
            $ref: '#/components/schemas/VolumeMount'
        retries:
          description: >-
            +label=Retries

            +usage=Specify the maximum number of attempts to retry a job before
            it is marked as failed.

            +icon=fa-repeat

            +sort=23000
          type: integer
          minimum: 0
          maximum: 10
          default: 0
        service_account:
          description: |-
            +label=Service Account
            +sort=24000
          type: string
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: SparkJob
      x-fern-type-name: SparkJob
    GatewayConfig:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/RateLimitConfig'
        - $ref: '#/components/schemas/LoadBalancingConfig'
        - $ref: '#/components/schemas/GuardrailsConfig'
        - $ref: '#/components/schemas/BudgetConfig'
        - $ref: '#/components/schemas/GatewayOTELConfig'
        - $ref: '#/components/schemas/GatewayMetadataConfig'
        - $ref: '#/components/schemas/GatewayLoggingConfig'
        - $ref: '#/components/schemas/GlobalSettings'
        - $ref: '#/components/schemas/GatewayDataAccessConfig'
        - $ref: '#/components/schemas/GatewayDataRoutingConfig'
      discriminator:
        propertyName: type
      title: GatewayConfig
      x-fern-type-name: GatewayConfig
    TeamManifest:
      type: object
      required:
        - type
        - name
        - members
      properties:
        type:
          description: +value=team
          type: string
          enum:
            - team
        name:
          description: >-
            +sort=1

            +message=Only alphanumeric characters, "_" and "-" are allowed in
            the team name, and it can be up to 36 characters in length

            +usage=Name of the Team
          type: string
          pattern: ^[a-zA-Z][a-zA-Z0-9_-]{1,34}[a-zA-Z0-9]$
        managers:
          description: >-
            +sort=3

            +label=Team Managers

            +usage=Emails of each of the user who can manage the members of the
            team. Any tenant admin is by default a team manager.

            +uiType=UserSelect

            +uiProps={"optionTypes": ["users"]}
          type: array
          items:
            type: string
        members:
          description: |-
            +sort=4
            +label=Team Members
            +usage=Enter email of each of the user you want to add in the team.
            +uiType=UserSelect
            +uiProps={"optionTypes": ["users"]}
          type: array
          items:
            type: string
        ownedBy:
          $ref: '#/components/schemas/TeamOwnedBy'
        tags:
          description: >-
            +label=Tags

            +usage=Key-value pairs to categorize this Team (e.g., by owner or
            environment).

            +sort=6
          type: object
      title: TeamManifest
      x-fern-type-name: TeamManifest
    PolicyManifest:
      type: object
      required:
        - type
        - name
        - description
        - operation
        - mode
        - entities
        - actions
        - code
      properties:
        type:
          description: +value=policy
          type: string
          enum:
            - policy
        name:
          description: >-
            +label=Name

            +usage=Unique identifier of the policy across the organisation

            +sort=2

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        description:
          description: >-
            +label=Description

            +usage=Description of the policy that explains what the policy does,
            its purpose, and how it affects resources. This helps other users
            understand the policy's behavior and impact.

            +sort=3

            +uiProps={"descriptionInline":true}

            +message=Description must be between 1 byte and 1024 bytes
          type: string
          minLength: 1
          maxLength: 1024
        operation:
          description: >-
            +label=Policy Action

            +usage=Defines if the policy validates or mutates resources.
            Validation policies check TrueFoundry manifests to ensure they meet
            conditions and can block deployments. Mutation policies modify
            Kubernetes manifests before they're applied to the cluster. See this
            [documentation](https://docs.truefoundry.com/docs/applying-custom-policies#applying-custom-policies)
            for more details.

            +sort=4

            +discriminator=type
          type: object
          default:
            type: validate
          oneOf:
            - $ref: '#/components/schemas/PolicyMutationOperation'
            - $ref: '#/components/schemas/PolicyValidationOperation'
          discriminator:
            propertyName: type
        mode:
          description: >-
            +label=Mode

            +usage=Mode of the policy: `Audit` logs all policy evaluations
            without blocking deployments. `Enforce` blocks deployments if the
            policy fails. `Disabled` deactivates the policy.

            +uiType=MultiSelectPills

            +sort=6
          type: string
          enum:
            - audit
            - enforce
            - disabled
        entities:
          description: |-
            +label=Entities
            +usage=Types of applications this policy applies to
            +uiType=MultiSelect
            +sort=8
          type: array
          items:
            $ref: '#/components/schemas/PolicyEntityTypes'
        actions:
          description: |-
            +label=Actions
            +usage=Actions this policy applies to
            +uiType=Hidden
            +sort=9
          type: array
          items:
            $ref: '#/components/schemas/PolicyActions'
          default:
            - apply
        filters:
          $ref: '#/components/schemas/PolicyFilters'
        code:
          description: >-
            +label=Policy Code

            +usage=TypeScript code for Policy logic. To write and test your
            policy code, please refer to this
            [repository](https://github.com/truefoundry/tfy-typescript-policy)

            +sort=13

            +uiType=CodeEditor

            +uiProps={"language":"typescript"}

            +message=Code size must be between 1 byte and 10 KB
          type: string
          minLength: 1
          maxLength: 10240
      title: PolicyManifest
      x-fern-type-name: PolicyManifest
    RoleManifest:
      type: object
      required:
        - type
        - name
        - displayName
        - description
        - resourceType
        - permissions
      properties:
        type:
          description: +value=role
          type: string
          enum:
            - role
        name:
          description: >-
            +sort=1

            +label=Name

            +usage=Unique identifier of the role across the organization

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,33}[a-z0-9]$
        displayName:
          description: |-
            +sort=2
            +label=Display Name
            +usage=Human-readable name for the role
            +message=Display name must be between 1 and 255 characters
          type: string
          minLength: 1
          maxLength: 255
        description:
          description: >-
            +sort=3

            +label=Description

            +usage=Description of the role that explains its purpose and
            permissions

            +message=Description must be between 1 and 1024 characters
          type: string
          minLength: 1
          maxLength: 1024
        resourceType:
          description: |-
            +sort=4
            +label=Resource Type
            +usage=Type of resource this role applies to
          type: string
        permissions:
          description: >-
            +sort=5

            +label=Permissions

            +usage=Define the resource type and the corresponding actions this
            role can perform on the resource
          type: array
          items:
            type: string
      title: RoleManifest
      x-fern-type-name: RoleManifest
    AlertConfig:
      type: object
      required:
        - type
        - name
        - resource
        - notification_targets
        - rules
      properties:
        type:
          description: |-
            +label=Type
            +value=alert-config
          type: string
          enum:
            - alert-config
        name:
          description: >-
            +label=Name

            +message=3 to 35 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"descriptionInline":true}

            +uiType=Hidden
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,33}[a-z0-9]$
        resource:
          $ref: '#/components/schemas/AlertConfigResource'
        notification_targets:
          description: >-
            +label=Notification Targets

            +usage=Configure one or more notification targets where alerts will
            be sent. Each target specifies a notification channel (email or
            slack).

            +uiProps={"descriptionInline":true}

            +sort=1
          type: array
          items:
            $ref: '#/components/schemas/NotificationTargetForAlertRule'
        rules:
          description: >-
            +label=Rules

            +usage=Define one or more alert rules that specify the conditions to
            monitor, when to trigger alerts, and how they should be handled.

            +uiProps={"descriptionInline":true}

            +uiType=Structs
          type: array
          items:
            $ref: '#/components/schemas/PrometheusAlertRule'
      title: AlertConfig
      x-fern-type-name: AlertConfig
    VirtualAccountManifest:
      type: object
      required:
        - name
        - type
        - permissions
      properties:
        name:
          description: >-
            +label=Name

            +sort=1

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +usage=Virtual Account Name
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        type:
          description: +value=virtual-account
          type: string
          enum:
            - virtual-account
        permissions:
          description: |-
            +label=Permissions
            +sort=2
            +uiType=ServiceAccountPermissions
          type: array
          items:
            $ref: '#/components/schemas/Permissions'
        expiration_date:
          description: >-
            +label=Expiration Date (UTC)

            +message=Expiration date of the virtual account

            +usage=Expiration Date of the Virtual Account (should be in the
            format yyyy-mm-dd)

            +sort=3

            +uiType=DatePicker
          type: string
          pattern: ^\d{4}-\d{2}-\d{2}$
        auto_rotate:
          $ref: '#/components/schemas/AutoRotate'
        notification_target:
          $ref: '#/components/schemas/NotificationTarget'
        secret_store_config:
          $ref: '#/components/schemas/SecretStoreConfig'
        ownedBy:
          $ref: '#/components/schemas/VirtualAccountOwnedBy'
        tags:
          description: >-
            +label=Tags

            +usage=Key-value pairs to categorize this Virtual Account (e.g., by
            owner or environment).

            +sort=8
          type: object
      title: VirtualAccountManifest
      x-fern-type-name: VirtualAccountManifest
    CommonToolsSettings:
      description: >-
        +docs=Describes the configuration for Common Tools MCP Server in AI
        Gateway

        +label=Common Tools Settings
      type: object
      required:
        - type
        - web_search
        - code_executor
        - sandbox_exec
        - sequential_thinking
        - web_scrape
      properties:
        type:
          description: +value=settings/common-tools
          type: string
          enum:
            - settings/common-tools
        web_search:
          description: |-
            +label=Web Search
            +usage=Enable web search tool
            Allows model to search the web for information.
            +icon=fa-searchengin
            +sort=1
          type: boolean
          default: true
        code_executor:
          description: |-
            +label=Code Executor
            +usage=Enable code executor tool
            Allows model to execute code and return the results.
            +icon=fa-code
            +sort=2
          type: boolean
          default: true
        sandbox_exec:
          description: >-
            +label=Sandbox

            +usage=Enable sandbox executor tool

            Allows model to execute shell command in an isolated stateful
            sandbox environment.

            +icon=fa-cube

            +sort=3
          type: boolean
          default: true
        sequential_thinking:
          description: |-
            +label=Sequential Thinking
            +usage=Enable sequential thinking tool
            Allows model to reason step-by-step to solve complex problems.
            +icon=fa-brain
            +sort=4
          type: boolean
          default: true
        web_scrape:
          description: >-
            +label=Web Scrape

            +usage=Enable web scraping tool.

            Allows model to scrape content from web pages with intelligent
            format selection for structured or unstructured data.

            +icon=fa-brands fa-readme

            +sort=6
          type: boolean
          default: true
      title: CommonToolsSettings
      x-fern-type-name: CommonToolsSettings
    AIFeaturesSettings:
      description: |-
        +docs=Describes the configuration for TrueFoundry AI features
        +label=AI Features Settings
      type: object
      required:
        - type
        - enabled
        - ai_provider
      properties:
        type:
          description: +value=settings/ai-features
          type: string
          enum:
            - settings/ai-features
        enabled:
          description: |-
            +label=Enabled
            +usage=Enable/Disable AI features.
            Other fields only apply when this is true
            +sort=1
          type: boolean
        ai_provider:
          description: >-
            +label=AI Provider

            +usage=AI provider to use for AI features

            When TrueFoundry is selected, TrueFoundry will manage and provide
            the AI models for the features.

            +uiProps={"disableEdit":true}
          type: string
          enum:
            - truefoundry
      title: AIFeaturesSettings
      x-fern-type-name: AIFeaturesSettings
    SecretGroupManifest:
      type: object
      required:
        - type
        - name
        - integration_fqn
        - collaborators
      properties:
        type:
          description: +value=secret-group
          type: string
          enum:
            - secret-group
        name:
          description: >-
            +label=Name

            +usage=Name of the secret group

            +message=Enter a name with 5 to 50 characters. It can include
            letters, numbers, and hyphens. Must not start with a digit. Hyphens
            can only be in between and not at the start or end. Must not have
            consecutive hyphens.

            +sort=10000
          type: string
          pattern: ^(?!.*--)[a-z][a-z0-9-]{3,48}[a-z0-9]$
        integration_fqn:
          description: |-
            +label=Secret Store
            +usage=Secret Store to be associated with the secret group
            +sort=10001
          type: string
          pattern: ^\S+$
        collaborators:
          description: |-
            +label=Collaborators
            +sort=10002
            +usage=Users and Teams that have access to Secret Group
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: SecretGroupManifest
      x-fern-type-name: SecretGroupManifest
    AgentManifest:
      type: object
      oneOf:
        - $ref: '#/components/schemas/TrueFoundryAgentManifest'
        - $ref: '#/components/schemas/RemoteAgent'
        - $ref: '#/components/schemas/LegacyAgentManifest'
      title: AgentManifest
      x-fern-type-name: AgentManifest
    EnvironmentManifest:
      type: object
      required:
        - type
        - name
        - color
        - isProduction
        - optimizeFor
      properties:
        type:
          description: +value=environment
          type: string
          enum:
            - environment
        name:
          description: >-
            +sort=100

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +usage=Environment Name
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        color:
          $ref: '#/components/schemas/EnvironmentColor'
        isProduction:
          description: |-
            +sort=400
            +label=Environment Type
            +message=Indicates if the environment is for production use
            +uiType=MultiSelectPills
            +uiProps={"hideRadioIcon":true}
          type: boolean
        optimizeFor:
          $ref: '#/components/schemas/EnvironmentOptimizeFor'
      title: EnvironmentManifest
      x-fern-type-name: EnvironmentManifest
    TracingProjectManifest:
      description: >-
        +label=Tracing Project

        +usage=Tracing Project is an entity that allows you to manage and
        monitor telemetry data.
      type: object
      required:
        - type
        - name
        - storage
        - collaborators
      properties:
        type:
          description: +value=tracing-project
          type: string
          enum:
            - tracing-project
        name:
          description: >-
            +label=Name

            +icon=fa-desktop:#black

            +usage=Name of the Tracing Project

            +message=Alphanumeric word, may contain '-' with a maximum length of
            63 characters

            +sort=1
          type: string
          pattern: ^[a-zA-Z][a-zA-Z0-9\-]{1,61}[a-zA-Z0-9]$
        storage:
          $ref: '#/components/schemas/TracingProjectStorageConfig'
        description:
          description: |-
            +label=Description
            +icon=fa-desktop:#black
            +usage=Description for the Tracing Project.
            +message=Description can be maximum 500 characters
            +sort=3
            +placeholder=Tracing Project for Team A applications
            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^.{0,500}$
        collaborators:
          description: |-
            +label=Collaborators
            +sort=4
            +usage=Users and Teams that have access to Tracing Project
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        traces_retention_duration_days:
          description: >-
            +label=Traces Retention Duration (in days)

            +sort=5

            +usage=Number of days to retain traces for (minimum 7 days).

            Traces older than these number of days will be deleted
            automatically.

            When not set, traces are retained indefinitely.

            Note: Metrics will be retained regardless of this setting.

            +message=Trace retention days must be greater than or equal to 7

            +placeholder=30

            +uiProps={"descriptionInline":true}
          type: number
          minimum: 7
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: TracingProjectManifest
      x-fern-type-name: TracingProjectManifest
    MCPServerManifest:
      type: object
      oneOf:
        - $ref: '#/components/schemas/RemoteMCPServerManifest'
        - $ref: '#/components/schemas/VirtualMCPServerManifest'
        - $ref: '#/components/schemas/OpenAPIMCPServerManifest'
        - $ref: '#/components/schemas/StdioMCPServerManifest'
      title: MCPServerManifest
      x-fern-type-name: MCPServerManifest
    RoleBindingManifest:
      type: object
      required:
        - type
        - name
        - subjects
        - permissions
      properties:
        type:
          description: +value=role-binding
          type: string
          enum:
            - role-binding
        name:
          description: >-
            +label=Name

            +usage=Unique name for this role binding.

            +message=3 to 64 chars: lowercase letter start/end; middle may
            include hyphens; cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9-]{1,62}[a-z0-9]$
        subjects:
          description: >-
            +label=Subjects

            +usage=One row per principal; set type and name (email, team name,
            virtual account name, or external identity name).

            +minItems=1
          type: array
          items:
            $ref: '#/components/schemas/RoleBindingSubject'
        permissions:
          description: >-
            +label=Permissions

            +usage=Resource-scoped role grants (resource type, FQN, and role
            name).

            +minItems=1
          type: array
          items:
            $ref: '#/components/schemas/RoleBindingPermission'
      title: RoleBindingManifest
      x-fern-type-name: RoleBindingManifest
    Collaborator:
      type: object
      required:
        - subject
        - role_id
      properties:
        subject:
          description: >-
            +label=Subject FQN

            +usage=Fully Qualified Name of the subject. eg: user:email or
            team:teamname
          type: string
        role_id:
          description: |-
            +label=Role ID
            +usage=Role ID for the resource
          type: string
      title: Collaborator
      x-fern-type-name: Collaborator
    OwnedBy:
      description: |-
        +label=Owner
        +usage=Names of the owners that own this resource
      type: object
      required:
        - account
      properties:
        account:
          description: |-
            +label=Account Name
            +sort=1
            +usage=The name of the account that owns this resource
            +placeholder=Search Account
            +uiType=AccountSelect
            +uiProps={"maxAllowedAccounts": 1}
          type: string
      title: OwnedBy
      x-fern-type-name: OwnedBy
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
    ServiceAutoscaling:
      type: object
      properties:
        metrics:
          description: |-
            +label=Autoscaling metrics
            +usage=Metrics to use for the autoscaler
            +sort=4
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/CPUUtilizationMetric'
            - $ref: '#/components/schemas/RPSMetric'
            - $ref: '#/components/schemas/CronMetric'
          discriminator:
            propertyName: type
      allOf:
        - $ref: '#/components/schemas/BaseAutoscaling'
      required:
        - metrics
      title: ServiceAutoscaling
      x-fern-type-name: ServiceAutoscaling
    Autoshutdown:
      type: object
      required:
        - wait_time
      properties:
        wait_time:
          description: >-
            +label=Wait Time

            +usage=The period to wait after the last received request before
            scaling the replicas to 0. This value should be high enough to allow
            for the replicas of the service to come up to avoid premature
            scaling down.
          type: integer
          minimum: 0
          maximum: 604800
          default: 900
      title: Autoshutdown
      x-fern-type-name: Autoshutdown
    Rolling:
      description: >-
        +docs=This strategy updates the pods in a rolling fashion such that a
        subset of the

        total pods are replaced with new version at one time.

        A commonly used strategy can be to have maxUnavailablePercentage close
        to 0 so that there

        is no downtime and keep the maxSurgePercentage to around 25%. If you are
        anyways running

        a large number of pods, the service can often tolerate a few pods going
        down - so you

        max maxUnavailablePercentage = 10 and maxSurgePercentage=0. You can read
        about it more

        [here](https://spot.io/resources/kubernetes-autoscaling/5-kubernetes-deployment-strategies-roll-out-like-the-pros/)

        +label=Rolling update strategy
      type: object
      required:
        - type
        - max_unavailable_percentage
        - max_surge_percentage
      properties:
        type:
          description: +value=rolling_update
          type: string
          enum:
            - rolling_update
        max_unavailable_percentage:
          description: >-
            +label=Max unavailable(%)

            +usage=Percentage of total replicas that can be brought down at one
            time.

            For a value of 25 when replicas are set to 12 this would mean
            minimum (25% of 12) = 3 pods might be unavailable during the
            deployment.

            Setting this to a higher value can help in speeding up the
            deployment process.
          type: integer
          minimum: 0
          maximum: 100
          default: 25
        max_surge_percentage:
          description: >-
            +label=Max Surge(%)

            +usage=Percentage of total replicas of updated image that can be
            brought up over the total replicas count.

            For a value of 25 when replicas are set to 12 this would mean
            (12+(25% of 12) = 15) pods might be running at one time.

            Setting this to a higher value can help in speeding up the
            deployment process.
          type: integer
          minimum: 0
          maximum: 100
          default: 25
      title: Rolling
      x-fern-type-name: Rolling
    Canary:
      description: >-
        +docs=This strategy brings up the new release without bringing the older
        release down. Traffic is shifted from the older release to the newer
        release in a staged manner.

        This can help with verifying the health of the new release without
        shifting complete traffic.

        +label=Canary strategy
      type: object
      required:
        - type
        - steps
      properties:
        type:
          description: +value=canary
          type: string
          enum:
            - canary
        steps:
          description: >-
            +docs=These steps would be executed in order to enable shifting of
            traffic slowly from stable to canary version

            +label=Steps
          type: array
          items:
            $ref: '#/components/schemas/CanaryStep'
      title: Canary
      x-fern-type-name: Canary
    BlueGreen:
      description: >-
        +docs=This strategy brings up the new release completely before
        switching the complete load to the new release.

        This minimizes the time that two versions are serving traffic at the
        same time.

        +label=Blue Green strategy
      type: object
      required:
        - type
      properties:
        type:
          description: +value=blue_green
          type: string
          enum:
            - blue_green
        enable_auto_promotion:
          description: >-
            +docs=Promote the new release to handle the complete traffic. A
            manual promotion would be needed if this is disabled

            +label=Auto-promotion
          type: boolean
          default: false
        auto_promotion_seconds:
          description: >-
            +docs=Promote the new release to handle the complete traffic after
            waiting for these many seconds

            +label=Auto-promotion seconds
          type: integer
          minimum: 0
          default: 30
      title: BlueGreen
      x-fern-type-name: BlueGreen
    BaseService:
      type: object
      required:
        - name
        - image
        - ports
      properties:
        name:
          description: >-
            +usage=Name of the service. This uniquely identifies this service in
            the workspace.

            > Name can only contain alphanumeric characters and '-' and can be
            atmost 25 characters long

            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        image:
          description: >-
            +docs=Specify whether you want to deploy a Docker image or build and
            deploy from source code

            +label=Deploy a Docker image or build and deploy from source code

            +icon=fa-solid fa-cloud-arrow-up:#21B6A8

            +sort=2

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/Build'
            - $ref: '#/components/schemas/Image'
          discriminator:
            propertyName: type
        artifacts_download:
          $ref: '#/components/schemas/ArtifactsDownload'
        resources:
          $ref: '#/components/schemas/Resources'
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the service
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=6
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        ports:
          description: >-
            +docs=Specify the ports you want the service to be exposed to

            +label=Configure ports and endpoints to route customer traffic

            +usage=Expose the deployment to make it accessible over the internet
            or keep it private. Implement authentication to restrict access.
            [Docs](https://docs.truefoundry.com/docs/define-ports-and-domains)

            +icon=fa-plug

            +sort=4
          type: array
          items:
            $ref: '#/components/schemas/Port'
        service_account:
          type: string
        mounts:
          description: >-
            +usage=Configure data to be mounted to service pod(s) as a string,
            secret or volume.
            [Docs](https://docs.truefoundry.com/docs/mounting-volumes-service)

            +sort=10011

            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/SecretMount'
              - $ref: '#/components/schemas/StringDataMount'
              - $ref: '#/components/schemas/VolumeMount'
        labels:
          description: +label=Labels
          type: object
          additionalProperties:
            type: string
        kustomize:
          $ref: '#/components/schemas/Kustomize'
        liveness_probe:
          $ref: '#/components/schemas/HealthProbe'
        readiness_probe:
          $ref: '#/components/schemas/HealthProbe'
        startup_probe:
          $ref: '#/components/schemas/HealthProbe'
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: BaseService
      x-fern-type-name: BaseService
    InfraProviderAccount:
      type: object
      oneOf:
        - $ref: '#/components/schemas/AwsProviderAccount'
        - $ref: '#/components/schemas/AzureProviderAccount'
        - $ref: '#/components/schemas/GcpProviderAccount'
        - $ref: '#/components/schemas/DockerhubProviderAccount'
        - $ref: '#/components/schemas/BitbucketProviderAccount'
        - $ref: '#/components/schemas/CustomProviderAccount'
        - $ref: '#/components/schemas/GithubProviderAccount'
        - $ref: '#/components/schemas/GitlabProviderAccount'
        - $ref: '#/components/schemas/JfrogProviderAccount'
        - $ref: '#/components/schemas/TTLProviderAccount'
        - $ref: '#/components/schemas/TrueFoundryProviderAccount'
        - $ref: '#/components/schemas/QuayProviderAccount'
        - $ref: '#/components/schemas/SlackProviderAccount'
        - $ref: '#/components/schemas/WebhookProviderAccount'
        - $ref: '#/components/schemas/PagerDutyProviderAccount'
        - $ref: '#/components/schemas/HashicorpProviderAccount'
      title: InfraProviderAccount
      x-fern-type-name: InfraProviderAccount
    ModelProviderAccount:
      type: object
      oneOf:
        - $ref: '#/components/schemas/AwsBedrockProviderAccount'
        - $ref: '#/components/schemas/GoogleVertexProviderAccount'
        - $ref: '#/components/schemas/GoogleGeminiProviderAccount'
        - $ref: '#/components/schemas/AzureOpenAIProviderAccount'
        - $ref: '#/components/schemas/AzureFoundryProviderAccount'
        - $ref: '#/components/schemas/CohereProviderAccount'
        - $ref: '#/components/schemas/AI21ProviderAccount'
        - $ref: '#/components/schemas/AnthropicProviderAccount'
        - $ref: '#/components/schemas/DeepinfraProviderAccount'
        - $ref: '#/components/schemas/GroqProviderAccount'
        - $ref: '#/components/schemas/MistralAIProviderAccount'
        - $ref: '#/components/schemas/PerplexityAIProviderAccount'
        - $ref: '#/components/schemas/TogetherAIProviderAccount'
        - $ref: '#/components/schemas/NomicProviderAccount'
        - $ref: '#/components/schemas/PalmProviderAccount'
        - $ref: '#/components/schemas/OllamaProviderAccount'
        - $ref: '#/components/schemas/OpenaiProviderAccount'
        - $ref: '#/components/schemas/DatabricksProviderAccount'
        - $ref: '#/components/schemas/ClouderaProviderAccount'
        - $ref: '#/components/schemas/SelfHostedModelProviderAccount'
        - $ref: '#/components/schemas/OpenRouterProviderAccount'
        - $ref: '#/components/schemas/SambaNovaProviderAccount'
        - $ref: '#/components/schemas/XAIProviderAccount'
        - $ref: '#/components/schemas/BasetenProviderAccount'
        - $ref: '#/components/schemas/AwsSagemakerProviderAccount'
        - $ref: '#/components/schemas/CerebrasProviderAccount'
        - $ref: '#/components/schemas/SnowflakeCortexProviderAccount'
        - $ref: '#/components/schemas/ElevenLabsProviderAccount'
        - $ref: '#/components/schemas/DeepgramProviderAccount'
        - $ref: '#/components/schemas/CartesiaProviderAccount'
        - $ref: '#/components/schemas/VirtualModelProviderAccount'
        - $ref: '#/components/schemas/CustomEndpointProviderAccount'
      title: ModelProviderAccount
      x-fern-type-name: ModelProviderAccount
    MCPServerProviderAccount:
      description: |-
        +label=MCP Server Group
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - collaborators
        - integrations
      properties:
        type:
          description: +value=provider-account/mcp-server-group
          type: string
          enum:
            - provider-account/mcp-server-group
        name:
          description: >-
            +label=Group Name

            +sort=100

            +usage=The name of the MCP Server Group.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        collaborators:
          description: >-
            +label=Managers

            +sort=200

            +usage=List of teams/users who can manage this MCP Server Group.
            These subjects can add new MCP server integrations, can edit
            existing ones and can also use the MCP server.

            +uiType=Collaborators

            +uiProps={"optionTypes": ["Manager"]}
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        integrations:
          description: |-
            +label=MCP Servers
            +usage=List of MCP Servers, which are part of this MCP Server Group.
            +sort=400
            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/MCPServerIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: MCPServerProviderAccount
      x-fern-type-name: MCPServerProviderAccount
    GuardrailConfigGroup:
      description: |-
        +label=Guardrail Configs Group
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: |-
            +uiType=Hidden
            +value=provider-account/guardrail-config-group
          type: string
          enum:
            - provider-account/guardrail-config-group
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Guardrail Config Group.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        collaborators:
          description: |-
            +label=Access Control
            +sort=200
            +usage=List of users who have access to this Guardrail Config Group.
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        integrations:
          description: >-
            +label=Guardrail Configs

            +usage=List of Guardrail Configs, which are part of this Guardrail
            Config Group.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/GuardrailConfigIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: GuardrailConfigGroup
      x-fern-type-name: GuardrailConfigGroup
    NotebookConfig:
      type: object
      required:
        - base_domain
      properties:
        base_domain:
          description: >-
            +label=Notebook Base Domain

            +usage=The base domain for the cluster with which you can access
            your Notebooks

            +message=Must not contain any spaces.
          type: string
          pattern: ^[^\s]*$
      title: NotebookConfig
      x-fern-type-name: NotebookConfig
    SSHServerConfig:
      type: object
      required:
        - base_domain
        - port
      properties:
        base_domain:
          description: >-
            +label=SSH Base Domain

            +usage=The base domain for the cluster with which you can access
            your SSH containers

            +sort=1000

            +message=Must not contain any spaces.
          type: string
          pattern: ^[^\s]*$
        port:
          description: >-
            +label=SSH Port

            +usage=The port for the cluster with which you can access your SSH
            containers

            +sort=2000
          type: number
      title: SSHServerConfig
      x-fern-type-name: SSHServerConfig
    SparkConfig:
      type: object
      required:
        - ui_base_domain
      properties:
        ui_base_domain:
          description: >-
            +label=Spark UI Base Domain

            +usage=The base domain for the cluster with which you can access
            your Spark UI

            +message=Must not contain any spaces.
          type: string
          pattern: ^[^\s]*$
      title: SparkConfig
      x-fern-type-name: SparkConfig
    IngressControllerConfig:
      type: object
      required:
        - ingress_class_name
      properties:
        ingress_class_name:
          description: +label=Ingress Class Name
          type: string
          default: nginx
        tls_enabled:
          description: >-
            +label=TLS Enabled

            +usage=Whether TLS is managed by the ingress controller. If enabled,
            the ingress object will have TLS configuration.
          type: boolean
          default: false
      title: IngressControllerConfig
      x-fern-type-name: IngressControllerConfig
    Nodepool:
      description: >-
        +label=Configured Nodepools

        +usage=The nodepools that are already created in your cluster. This will
        be used to schedule your workloads on particular nodepools.
      type: object
      required:
        - name
      properties:
        name:
          type: string
        description:
          type: string
      title: Nodepool
      x-fern-type-name: Nodepool
    Permissions:
      type: object
      required:
        - resource_fqn
        - resource_type
        - role_id
      properties:
        resource_fqn:
          description: |-
            +label= Resource FQN
            +usage=The fully qualified name of the resource
          type: string
        resource_type:
          description: |-
            +label=Resource Type
            +usage=The type of the resource (cluster, workspace, etc.)
          type: string
        role_id:
          description: >-
            +label=Role ID

            +usage=The role id of the role to be assigned to the service account
            for that resource
          type: string
      title: Permissions
      x-fern-type-name: Permissions
    Build:
      description: |-
        +docs=Describes how we build our code into a Docker image.
        +label=Source Code (Build and deploy source code)
        +icon=fa-code
      type: object
      required:
        - type
        - build_source
        - build_spec
      properties:
        type:
          description: +value=build
          type: string
          enum:
            - build
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page
          type: string
          pattern: ^\S+$
        build_source:
          description: |-
            +docs=Source code location.
            +label=Fetch source code to build and deploy
            +icon=fa-code
            +discriminator=type
            +sort=1
          type: object
          oneOf:
            - $ref: '#/components/schemas/RemoteSource'
            - $ref: '#/components/schemas/GitSource'
            - $ref: '#/components/schemas/LocalSource'
          discriminator:
            propertyName: type
        build_spec:
          description: >-
            +docs=Instructions to build a container image out of the build
            source

            +label=Build using DockerFile or using Buildpack

            +icon=fa-wrench

            +discriminator=type

            +sort=2
          type: object
          oneOf:
            - $ref: '#/components/schemas/DockerFileBuild'
            - $ref: '#/components/schemas/PythonBuild'
          discriminator:
            propertyName: type
      title: Build
      x-fern-type-name: Build
    Image:
      description: >-
        +docs=Describes that we are using a pre-built image stored in a Docker
        Image registry

        +label=Docker Image (Deploy an existing image)

        +icon=fa-brands fa-docker:#0db7ed
      type: object
      required:
        - type
        - image_uri
      properties:
        type:
          description: +value=image
          type: string
          enum:
            - image
        image_uri:
          description: >-
            +label=Image URI

            +usage=The image URI. Specify the name of the image and the tag.

            If the image is in Dockerhub, you can skip registry-url (for e.g.
            `tensorflow/tensorflow`).

            You can use an image from a private registry using Advanced fields

            +placeholder=registry-url/account/image:version (e.g.
            docker.io/tensorflow/tensorflow)
          type: string
          pattern: ^\S+$
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page
          type: string
          pattern: ^\S+$
        command:
          description: >-
            +label=Command Override

            +usage=Override the command to run when container starts.

            When deploying a Job, the command can be templatized by defining
            `params` and referencing them in command

            E.g. `python main.py --learning_rate {{learning_rate}}`
          oneOf:
            - type: string
            - type: array
              items:
                type: string
      title: Image
      x-fern-type-name: Image
    Manual:
      description: >-
        +docs=Describes that we are going to manually trigger our job.

        +label=Manual

        +usage=Trigger the job manually.
        [Docs](https://docs.truefoundry.com/docs/deploy-a-cron-job)
      type: object
      required:
        - type
      properties:
        type:
          description: +value=manual
          type: string
          enum:
            - manual
      title: Manual
      x-fern-type-name: Manual
    Schedule:
      description: >-
        +docs=Describes that we are going to schedule our job to run at a
        schedule, making our job a cron job.

        +label=Schedule

        +usage=Run the job on a schedule.
        [Docs](https://docs.truefoundry.com/docs/deploy-a-cron-job)
      type: object
      required:
        - type
        - schedule
        - concurrency_policy
      properties:
        type:
          description: +value=scheduled
          type: string
          enum:
            - scheduled
        schedule:
          description: >-
            +docs=Specify the schedule for this job to be run periodically in
            cron format. [Learn
            more](https://docs.truefoundry.com/docs/deploy-a-cron-job)

            +usage=Specify the schedule for this job to be run periodically in
            cron format.

            ```

            * * * * *

            | | | | |

            | | | | |___ day of week (0-6) (Sunday is 0)

            | | | |_____ month (1-12)

            | | |_______ day of month (1-31)

            | |_________ hour (0-23)

            |___________ minute (0-59)

            ```
          type: string
        concurrency_policy:
          description: >-
            +usage=Choose whether to allow this job to run while another
            instance of the job is running, or to replace the currently running
            instance. Allow

            will enable multiple instances of this job to run. Forbid will keep
            the current instance of the job running and stop a new instance from
            being run.

            Replace will terminate any currently running instance of the job and
            start a new one.
          type: string
          enum:
            - Forbid
            - Allow
            - Replace
          default: Forbid
        timezone:
          description: >-
            +usage=Timezone against which the cron schedule will be calculated,
            e.g. "Asia/Tokyo". Default is machine's local time.

            https://docs.truefoundry.com/docs/list-of-supported-timezones
          type: string
      title: Schedule
      x-fern-type-name: Schedule
    Param:
      type: object
      required:
        - name
      properties:
        name:
          description: >-
            +usage=Name of the param

            +message=name can contain lower case alphabets, digits, underscore
            (_) and hypen (-). It can be 32 characters long, should start with
            an alphabet, and should end with either an alphabet or digit.
          type: string
          pattern: ^[a-z][a-z0-9\-_]{0,30}[a-z0-9]$
        description:
          description: |-
            +usage=Description of param
            +message=description cannot be longer than 127 characters
          type: string
          pattern: ^.{1,127}$
        default:
          description: |-
            +usage=Default value or placeholder
            +message=default value cannot be longer than 127 characters
            +label=Default value
          type: string
          pattern: ^.{0,127}$
        param_type:
          type: string
          enum:
            - string
            - ml_repo
          default: string
      title: Param
      x-fern-type-name: Param
    Resources:
      description: >-
        +docs=Describes the resource constraints for the application so that it
        can be deployed accordingly on the cluster

        To learn more you can go
        [here](https://www.truefoundry.com/docs/resources-cpu-memory-storage)

        +icon=fa-microchip

        +label=Resources

        +usage=Configure resource allocations, specify node constraints and
        capacity types to improve performance and reduce expenses.
        [Docs](https://www.truefoundry.com/docs/resources-cpu-memory-storage)
      type: object
      required:
        - cpu_request
        - cpu_limit
        - memory_request
        - memory_limit
        - ephemeral_storage_request
        - ephemeral_storage_limit
      properties:
        cpu_request:
          description: >-
            +label=CPU Request

            +sort=1

            +usage=Requested CPU which determines the minimum cost incurred. The
            CPU usage can exceed the requested

            amount, but not the value specified in the limit. 1 CPU means 1 CPU
            core. Fractional CPU can be requested

            like `0.5` or `0.05`
          type: number
          minimum: 0.001
          maximum: 256
          default: 0.2
        cpu_limit:
          description: >-
            +label=CPU Limit

            +usage=CPU limit beyond which the usage cannot be exceeded. 1 CPU
            means 1 CPU core. Fractional CPU can be requested

            like `0.5`. CPU limit should be >= cpu request.

            +sort=2
          type: number
          minimum: 0.001
          maximum: 256
          default: 0.5
        memory_request:
          description: >-
            +label=Memory Request

            +usage=Requested memory which determines the minimum cost incurred.
            The unit of memory is in megabytes(MB).

            So 1 means 1 MB and 2000 means 2GB.

            +sort=3
          type: integer
          minimum: 1
          maximum: 4000000
          default: 200
        memory_limit:
          description: >-
            +label=Memory Limit

            +usage=Memory limit after which the application will be killed with
            an OOM error. The unit of memory is

            in megabytes(MB). So 1 means 1 MB and 2000 means 2GB. MemoryLimit
            should be greater than memory request.

            +sort=4
          type: integer
          minimum: 1
          maximum: 4000000
          default: 500
        ephemeral_storage_request:
          description: >-
            +label=Storage Request

            +usage=Requested disk storage. The unit of memory is in
            megabytes(MB).

            This is ephemeral storage and will be wiped out on pod restarts or
            eviction

            +sort=5
          type: integer
          minimum: 1
          maximum: 2000000
          default: 1000
        ephemeral_storage_limit:
          description: >-
            +label=Storage Limit

            +usage=Disk storage limit. The unit of memory is in megabytes(MB).
            Exceeding this limit will result in eviction.

            It should be greater than the request. This is ephemeral storage and
            will be wiped out on pod restarts or eviction

            +sort=6
          type: integer
          minimum: 1
          maximum: 2000000
          default: 2000
        shared_memory_size:
          description: >-
            +label=Shared Memory Size (MB)

            +usage=Define the shared memory requirements for your workload.
            Machine learning libraries like Pytorch can use Shared Memory

            for inter-process communication. If you use this, we will mount a
            `tmpfs` backed volume at the `/dev/shm` directory.

            Any usage will also count against the workload's memory limit
            (`resources.memory_limit`) along with your workload's memory usage.

            If the overall usage goes above `resources.memory_limit` the user
            process may get killed.

            Shared Memory Size cannot be more than the defined Memory Limit for
            the workload.
          type: integer
          minimum: 64
          maximum: 4000000
        node:
          description: >-
            +label=Node

            +usage=This field determines how the underlying node resource is to
            be utilized

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/NodeSelector'
            - $ref: '#/components/schemas/NodepoolSelector'
          discriminator:
            propertyName: type
        devices:
          description: >-
            +label=Devices

            +usage=Define custom device or accelerator requirements for your
            workload. We currently support NVIDIA GPUs, AWS Inferentia
            Accelerators, Single Host TPU Slices.

            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/NvidiaGPU'
              - $ref: '#/components/schemas/AWSInferentia'
              - $ref: '#/components/schemas/NvidiaMIGGPU'
              - $ref: '#/components/schemas/NvidiaTimeslicingGPU'
              - $ref: '#/components/schemas/GcpTPU'
      title: Resources
      x-fern-type-name: Resources
    JobAlert:
      description: |-
        +docs=Describes the configuration for the job alerts
        +label=Alert
      type: object
      properties:
        notification_channel:
          description: |-
            +label=Notification Channel
            +usage=Specify the notification channel to send alerts to
            +uiType=Hidden
            +sort=660
          type: string
          minLength: 1
        to_emails:
          description: >-
            +label=To Emails

            +usage=List of recipients' email addresses if the notification
            channel is Email.

            +docs=Specify the emails to send alerts to

            +uiType=Hidden

            +sort=665
          type: array
          items:
            type: string
            minLength: 1
        notification_target:
          $ref: '#/components/schemas/NotificationTarget'
        on_start:
          description: |-
            +label=On Start
            +usage=Send an alert when the job starts
            +sort=670
          type: boolean
          default: false
        on_completion:
          type: boolean
          default: false
        on_failure:
          description: |-
            +label=On Failure
            +usage=Send an alert when the job fails
            +sort=690
          type: boolean
          default: true
      title: JobAlert
      x-fern-type-name: JobAlert
    SecretMount:
      type: object
      required:
        - type
        - mount_path
        - secret_fqn
      properties:
        type:
          description: +value=secret
          type: string
          enum:
            - secret
        mount_path:
          description: |-
            +label=File path
            +usage=Absolute file path where the file will be created.
            +message=Please enter a valid file path
          type: string
          pattern: ^\/(?:[^/\n]+\/*)*[^/\n]+(\.[^/\n]+)?$
        secret_fqn:
          description: |-
            +label=Secret
            +usage=The TrueFoundry secret whose value will be the file content.
          type: string
          pattern: ^tfy-secret:\/\/.+:.+:.+$
      title: SecretMount
      x-fern-type-name: SecretMount
    StringDataMount:
      type: object
      required:
        - type
        - mount_path
        - data
      properties:
        type:
          description: +value=string
          type: string
          enum:
            - string
        mount_path:
          description: |-
            +label=File Path
            +usage=Absolute file path where the file will be created.
            +message=Please enter a valid file path
          type: string
          pattern: ^\/(?:[^/\n]+\/*)*[^/\n]+(\.[^/\n]+)?$
        data:
          description: |-
            +label=Data
            +usage=The file content.
          type: string
      title: StringDataMount
      x-fern-type-name: StringDataMount
    VolumeMount:
      type: object
      required:
        - type
        - mount_path
        - volume_fqn
      properties:
        type:
          description: +value=volume
          type: string
          enum:
            - volume
        mount_path:
          description: |-
            +label=Volume mount path
            +usage=Absolute file path where the volume will be mounted.
            +message=Please enter a valid mount path
          type: string
          pattern: ^\/(?:[^/\n]+\/*)*[^/\n]+(\.[^/\n]+)?$
        sub_path:
          description: >-
            +label=Sub Path

            +usage=Sub path within the volume to mount. Defaults to root of the
            volume.
          type: string
          pattern: ^(?:[^/\n]+/*)*[^/\n]+(\.[^/\n]+)?$
        volume_fqn:
          description: |-
            +label=Volume
            +usage=The TrueFoundry volume that needs to be mounted.
          type: string
          pattern: ^tfy-volume:\/\/.+:.+:.+$
      title: VolumeMount
      x-fern-type-name: VolumeMount
    Kustomize:
      type: object
      properties:
        patch:
          description: >-
            +label=Patch

            +usage=Content of kustomization.yaml to perform kustomize operation.
            Please do not include the `resources` section. It is filled in
            automatically
          type: object
        additions:
          description: >-
            +label=Additional Manifests

            +usage=Additional kubernetes manifests to be included in the
            application
          type: array
          items:
            type: object
      title: Kustomize
      x-fern-type-name: Kustomize
    HelmRepo:
      type: object
      required:
        - type
        - repo_url
        - chart
        - version
      properties:
        type:
          description: +value=helm-repo
          type: string
          enum:
            - helm-repo
        repo_url:
          description: |-
            +label=Helm repository URL
            +sort=1
            +message=Needs to be a valid URL.
          type: string
          pattern: >-
            ^(((https?|wss):\/\/)?(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*))$
        integration_fqn:
          description: >-
            +docs=FQN of the helm repo integration. You can use the FQN of your
            desired helm integration (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations) page

            +label=Helm repo integration

            +sort=2

            +usage=FQN of the helm repo integration. If you can't find your
            integration here,

            add it through the [Integrations](/integrations) page

            +uiType=IntegrationSelect

            +uiProps={"integrationType":"helm-repo"}
          type: string
        chart:
          description: |-
            +label=Chart name
            +sort=3
            +usage=The helm chart name
            +uiType=InputSelect
            +uiProps={"creatable":true, "searchable":true}
          type: string
        version:
          description: |-
            +label=Version
            +sort=4
            +usage=Helm chart version
            +uiType=InputSelect
            +uiProps={"creatable":true, "searchable":true}
          type: string
      title: HelmRepo
      x-fern-type-name: HelmRepo
    OCIRepo:
      description: +label=OCIRepo
      type: object
      required:
        - type
        - oci_chart_url
        - version
      properties:
        type:
          description: +value=oci-repo
          type: string
          enum:
            - oci-repo
        oci_chart_url:
          description: |-
            +label=OCI chart URL
            +message=Need to be a valid URL.
          type: string
          pattern: >-
            ^(((oci):\/\/)?(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*))$
        integration_fqn:
          description: >-
            +docs=FQN of the container registry. You can use the FQN of your
            desired container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations) page

            +label=Container Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations) page

            +uiType=IntegrationSelect

            +uiProps={"integrationType":"docker-registry"}
          type: string
        version:
          description: |-
            +label=Version
            +usage=Helm chart version
          type: string
      title: OCIRepo
      x-fern-type-name: OCIRepo
    GitHelmRepo:
      type: object
      required:
        - type
        - repo_url
        - revision
        - path
      properties:
        type:
          description: +value=git-helm-repo
          type: string
          enum:
            - git-helm-repo
        repo_url:
          description: |-
            TODO: Check this regex and add guidelines
            +label=Git repository URL
            +sort=1
            +message=Needs to be a valid URL.
          type: string
          pattern: >-
            ^(((https?|wss):\/\/)?(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*))$
        revision:
          description: |-
            +label=Revision
            +sort=2
            +usage=Branch/Commit SHA/Tag of the git repo.
          type: string
        path:
          description: |-
            +label=Path
            +sort=3
            +usage=Path to the chart.
          type: string
        value_files:
          description: |-
            +label=Value files
            +sort=3
            +usage=Helm values files for overriding values in the helm chart.
            The path is relative to the Path directory defined above
          type: array
          items:
            type: string
      title: GitHelmRepo
      x-fern-type-name: GitHelmRepo
    DynamicVolumeConfig:
      description: +label=Create new volume
      type: object
      required:
        - type
        - storage_class
        - size
      properties:
        type:
          description: |-
            +label=Volume Type
            +value=dynamic
            +usage=Volume Type for the volume.
          type: string
          enum:
            - dynamic
        storage_class:
          description: |-
            +label=Storage Class Name
            +usage=Name of the storage class to be used for the volume.
          type: string
        size:
          description: |-
            +label=Size(Gi)
            +unit=Gi
            +usage=Size of volume in Gi
          type: integer
          minimum: 1
          maximum: 64000
      title: DynamicVolumeConfig
      x-fern-type-name: DynamicVolumeConfig
    StaticVolumeConfig:
      description: +label=Use existing volume
      type: object
      required:
        - type
        - persistent_volume_name
      properties:
        type:
          description: |-
            +label=Volume Type
            +value=static
            +usage=Volume Type for the volume.
          type: string
          enum:
            - static
        persistent_volume_name:
          description: |-
            +label=Persistent Volume
            +usage=Persistent Volume Name of the volume to be used.
          type: string
      title: StaticVolumeConfig
      x-fern-type-name: StaticVolumeConfig
    VolumeBrowser:
      description: +label=Volume Browser
      type: object
      required:
        - endpoint
      properties:
        username:
          description: >-
            +message=3 to 10 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number.

            +usage=Username for logging in the volume browser.

            +sort=1
          type: string
          pattern: ^[a-z][a-z0-9]{1,8}[a-z0-9]$
        password_secret_fqn:
          description: >-
            +label=Password Secret FQN

            +usage=TFY Secret containing the password for logging in the volume
            browser.

            +sort=2
          type: string
          pattern: ^tfy-secret:\/\/.+:.+:.+$
        endpoint:
          $ref: '#/components/schemas/Endpoint'
        service_account:
          description: |-
            +label=Service Account Name
            +usage=Kubernetes Service account name for the volume browser.
            +sort=4
          type: string
      title: VolumeBrowser
      x-fern-type-name: VolumeBrowser
    WorkbenchImage:
      description: +usage=Workbench Image with persistent environment (Python 3.11.6)
      type: object
      required:
        - image_uri
      properties:
        image_uri:
          description: >-
            +label=Image URI

            +usage=The image URI. Specify the name of the image and the tag.

            If the image is in Dockerhub, you can skip registry-url (for e.g.
            `tensorflow/tensorflow`).

            You can use an image from a private registry using Advanced fields

            +placeholder=registry-url/account/image:version
          type: string
        build_script:
          description: >-
            +label=Build Script

            +usage=The build script to run when building the image.

            This will be executed as the last step in the docker build process
            as the root user (RUN DEBIAN_FRONTEND=noninteractive bash -ex
            build_script.sh)

            +placeholder=Enter the build script

            +uiType=CodeEditor

            +uiProps={"language":"shell"}
          type: string
          minLength: 1
          maxLength: 5120
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +uiType=IntegrationSelect

            +uiProps={"integrationType":"docker-registry"}

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page
          type: string
      title: WorkbenchImage
      x-fern-type-name: WorkbenchImage
    BaseWorkbenchInput:
      description: +docs=Describes the configuration for the service
      type: object
      required:
        - name
        - home_directory_size
      properties:
        name:
          description: >-
            +usage=Name of the workbench. This uniquely identifies this
            workbench in the workspace.

            > Name can only contain alphanumeric characters and '-' and can be
            atmost 25 characters long

            +sort=1

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        home_directory_size:
          description: >-
            +label=Home Directory Size in GB (Persistent)

            +usage=Size of the home directory for the workbench (Persistent
            Storage)

            +sort=6
          type: integer
          minimum: 5
          maximum: 64000
          default: 20
        resources:
          $ref: '#/components/schemas/Resources'
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the service
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/environment-variables-and-secrets-jobs)

            +sort=10110
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        mounts:
          description: >-
            +usage=Configure data to be mounted to workbench pod(s) as a string,
            secret or volume.
            [Docs](https://docs.truefoundry.com/docs/mounting-volumes-job)

            +sort=10111

            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/SecretMount'
              - $ref: '#/components/schemas/StringDataMount'
              - $ref: '#/components/schemas/VolumeMount'
        service_account:
          description: +sort=10113
          type: string
        kustomize:
          $ref: '#/components/schemas/Kustomize'
        workspace_fqn:
          description: |-
            +label=Workspace FQN
            +docs=Fully qualified name of the workspace
            +uiType=Hidden
          type: string
      title: BaseWorkbenchInput
      x-fern-type-name: BaseWorkbenchInput
    LocalSource:
      description: >-
        +docs=Describes that we are using code stored in a local developement
        environment to build our image

        +label=Local

        +icon=fa-folder:black

        +sort=100
      type: object
      required:
        - type
        - project_root_path
        - local_build
      properties:
        type:
          description: +value=local
          type: string
          enum:
            - local
        project_root_path:
          description: +usage=Local project root path.
          type: string
          default: ./
        local_build:
          description: run docker build locally
          type: boolean
          default: true
      title: LocalSource
      x-fern-type-name: LocalSource
    RemoteSource:
      description: >-
        +docs=Describes that we are using code stored in a remote repository to
        build our image

        +label=S3

        +icon=fa-brands fa-aws:black

        +sort=200
      type: object
      required:
        - type
        - remote_uri
      properties:
        type:
          description: +value=remote
          type: string
          enum:
            - remote
        remote_uri:
          description: |-
            +docs=Remote repository URI
            +label=Remote URI
          type: string
      title: RemoteSource
      x-fern-type-name: RemoteSource
    FlyteTask:
      type: object
      required:
        - template
      properties:
        template:
          $ref: '#/components/schemas/FlyteTaskTemplate'
        description: {}
      title: FlyteTask
      x-fern-type-name: FlyteTask
    FlyteWorkflow:
      type: object
      required:
        - template
      properties:
        template:
          $ref: '#/components/schemas/FlyteWorkflowTemplate'
        description: {}
      title: FlyteWorkflow
      x-fern-type-name: FlyteWorkflow
    FlyteLaunchPlan:
      type: object
      required:
        - id
        - spec
        - closure
      properties:
        id:
          $ref: '#/components/schemas/FlyteLaunchPlanID'
        spec:
          $ref: '#/components/schemas/FlyteLaunchPlanSpec'
        closure: {}
      title: FlyteLaunchPlan
      x-fern-type-name: FlyteLaunchPlan
    WorkflowAlert:
      description: |-
        +docs=Describes the configuration for the workflow alerts
        +label=Alert
      type: object
      properties:
        notification_target:
          $ref: '#/components/schemas/NotificationTarget'
        on_completion:
          description: |-
            +label=On Completion
            +usage=Send an alert when the job completes
            +sort=680
          type: boolean
          default: false
        on_failure:
          description: |-
            +label=On Failure
            +usage=Send an alert when the job fails
            +sort=690
          type: boolean
          default: true
      title: WorkflowAlert
      x-fern-type-name: WorkflowAlert
    AsyncServiceAutoscaling:
      type: object
      properties:
        metrics:
          description: |-
            +label=Autoscaling metrics
            +usage=Metrics to use for the autoscaler
            +sort=4
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/SQSQueueMetricConfig'
            - $ref: '#/components/schemas/NATSMetricConfig'
            - $ref: '#/components/schemas/KafkaMetricConfig'
            - $ref: '#/components/schemas/CronMetric'
            - $ref: '#/components/schemas/AMQPMetricConfig'
          discriminator:
            propertyName: type
      allOf:
        - $ref: '#/components/schemas/BaseAutoscaling'
      required:
        - metrics
      title: AsyncServiceAutoscaling
      x-fern-type-name: AsyncServiceAutoscaling
    WorkerConfig:
      type: object
      required:
        - input_config
        - num_concurrent_workers
      properties:
        input_config:
          description: |-
            +label=Input Config
            +usage=Input Config
            +sort=1
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/SQSInputConfig'
            - $ref: '#/components/schemas/NATSInputConfig'
            - $ref: '#/components/schemas/KafkaInputConfig'
            - $ref: '#/components/schemas/AMQPInputConfig'
          discriminator:
            propertyName: type
        output_config:
          description: |-
            +label=Output Config
            +usage=Output Config
            +sort=2
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/SQSOutputConfig'
            - $ref: '#/components/schemas/NATSOutputConfig'
            - $ref: '#/components/schemas/CoreNATSOutputConfig'
            - $ref: '#/components/schemas/KafkaOutputConfig'
            - $ref: '#/components/schemas/AMQPOutputConfig'
          discriminator:
            propertyName: type
        num_concurrent_workers:
          description: |-
            +label=Number of Concurrent Workers
            +usage=Number of concurrent workers to spawn for the processor
            +sort=3
          type: integer
          minimum: 1
          maximum: 10
          default: 1
      title: WorkerConfig
      x-fern-type-name: WorkerConfig
    AsyncProcessorSidecar:
      type: object
      required:
        - destination_url
      properties:
        destination_url:
          description: |-
            +label=Destination URL
            +usage=URL for the processor to invoke
          type: string
        request_timeout:
          description: |-
            +label=Request Timeout Seconds
            +usage=Timeout for the invoke request in seconds
          type: integer
          minimum: 1
          default: 10
        sidecar_image:
          description: >-
            +label=Sidecar Image

            +usage=Image for the processor sidecar (This field will be
            deprecated in the future)
          type: string
      title: AsyncProcessorSidecar
      x-fern-type-name: AsyncProcessorSidecar
    SparkImage:
      description: >-
        +docs=Describes that we are using a pre-built image stored in a Docker
        Image registry

        +label=Deploy an existing image

        +icon=fa-brands fa-docker:#0db7ed
      type: object
      required:
        - type
        - spark_version
        - image_uri
      properties:
        type:
          description: +value=spark-image
          type: string
          enum:
            - spark-image
        spark_version:
          description: >-
            --- Spark Specific Field ---

            +label=Spark Version

            +usage=Spark version should match the spark version installed in the
            image.

            +sort=1000
          type: string
          default: 3.5.2
        image_uri:
          description: >-
            +label=Image URI

            +usage=The image URI. Specify the name of the image and the tag.

            If the image is in Dockerhub, you can skip registry-url (for e.g.
            `tensorflow/tensorflow`).

            You can use an image from a private registry using Advanced fields

            +placeholder=registry-url/account/image:version (e.g.
            docker.io/tensorflow/tensorflow)

            +sort=1001
          type: string
          pattern: ^\S+$
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page

            +sort=1002
          type: string
          pattern: ^\S+$
      title: SparkImage
      x-fern-type-name: SparkImage
    SparkImageBuild:
      description: |-
        TODO (gw): Fix sorting for this such that it looks similar to SparkImage
        +docs=Describes that we are building a new image based on the spec
        +label=Build a new image
        +icon=fa-brands fa-docker:#0db7ed
      type: object
      required:
        - type
        - build_source
        - build_spec
      properties:
        type:
          description: +value=spark-image-build
          type: string
          enum:
            - spark-image-build
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page

            +sort=1002
          type: string
          pattern: ^\S+$
        build_source:
          description: >-
            TODO (gw): The following is a hack till the uiType GitSelect is
            fixed fron frontend

            +label=Fetch source code

            +sort=1003

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/GitSource'
            - $ref: '#/components/schemas/RemoteSource'
          discriminator:
            propertyName: type
        build_spec:
          $ref: '#/components/schemas/SparkBuild'
      title: SparkImageBuild
      x-fern-type-name: SparkImageBuild
    SparkJobPythonEntrypoint:
      description: +label=python
      type: object
      required:
        - type
        - main_application_file
      properties:
        type:
          description: +value=python
          type: string
          enum:
            - python
        main_application_file:
          description: >-
            +label=Main Application File

            +usage=The main application file to be executed by the spark job.
            Relative path in case of git repository.

            +message=Filename should have .py extension

            +sort=5

            +placeholder=For example: local:///path/to/file.py,
            s3:///bucket/path/to/file.py, etc.
          type: string
        arguments:
          description: |-
            +label=Arguments
            +usage=Arguments to be passed to the main application file.
            +sort=6
          type: string
      title: SparkJobPythonEntrypoint
      x-fern-type-name: SparkJobPythonEntrypoint
    SparkJobScalaEntrypoint:
      description: +label=scala
      type: object
      required:
        - type
        - main_application_file
        - main_class
      properties:
        type:
          description: +value=scala
          type: string
          enum:
            - scala
        main_application_file:
          description: >-
            +label=Main Application File

            +usage=The main application file to be executed by the spark job.

            +message=Filename should have .jar extension

            +sort=5

            +placeholder=For example: local:///path/to/file.jar,
            s3:///bucket/path/to/file.jar, etc.
          type: string
        main_class:
          description: |-
            +label=Main Class
            +usage=The main class to be executed by the spark job.
            +sort=6
            +required=true
            +message=The main class must be a valid Java class name.
          type: string
        arguments:
          description: |-
            +label=Arguments
            +usage=Arguments to be passed to the main application file.
            +sort=7
          type: string
      title: SparkJobScalaEntrypoint
      x-fern-type-name: SparkJobScalaEntrypoint
    SparkJobJavaEntrypoint:
      description: +label=java
      type: object
      required:
        - type
        - main_application_file
        - main_class
      properties:
        type:
          description: +value=java
          type: string
          enum:
            - java
        main_application_file:
          description: >-
            +label=Main Application File

            +usage=The main application file to be executed by the spark job.

            +message=Filename should have .jar extension

            +sort=5

            +placeholder=For example: local:///path/to/file.jar,
            s3:///bucket/path/to/file.jar, etc.
          type: string
        main_class:
          description: |-
            +label=Main Class
            +usage=The main class to be executed by the spark job.
            +sort=6
            +required=true
            +message=The main class must be a valid Java class name.
          type: string
        arguments:
          description: |-
            +label=Arguments
            +usage=Arguments to be passed to the main application file.
            +sort=7
          type: string
      title: SparkJobJavaEntrypoint
      x-fern-type-name: SparkJobJavaEntrypoint
    SparkJobPythonNotebookEntrypoint:
      description: +label=python notebook
      type: object
      required:
        - type
        - main_application_file
      properties:
        type:
          description: +value=python-notebook
          type: string
          enum:
            - python-notebook
        main_application_file:
          description: >-
            +label=Main Application File

            +usage=The main application file to be executed by the spark job.
            Relative path in case of git repository.

            +message=Filename should have .ipynb extension

            +sort=5

            +placeholder=For example: local:///path/to/file.ipynb,
            s3:///bucket/path/to/file.ipynb, etc.
          type: string
      title: SparkJobPythonNotebookEntrypoint
      x-fern-type-name: SparkJobPythonNotebookEntrypoint
    SparkJobScalaNotebookEntrypoint:
      description: +label=scala notebook
      type: object
      required:
        - type
        - main_application_file
      properties:
        type:
          description: +value=scala-notebook
          type: string
          enum:
            - scala-notebook
        main_application_file:
          description: >-
            +label=Main Application File

            +usage=The main application file to be executed by the spark job.
            Relative path in case of git repository.

            +message=Filename should have .ipynb extension

            +sort=5

            +placeholder=For example: local:///path/to/file.ipynb,
            s3:///bucket/path/to/file.ipynb, etc.
          type: string
      title: SparkJobScalaNotebookEntrypoint
      x-fern-type-name: SparkJobScalaNotebookEntrypoint
    SparkDriverConfig:
      description: +label=Driver Config
      type: object
      properties:
        resources:
          $ref: '#/components/schemas/Resources'
      title: SparkDriverConfig
      x-fern-type-name: SparkDriverConfig
    SparkExecutorConfig:
      description: +label=Executor Config
      type: object
      required:
        - instances
      properties:
        instances:
          description: |-
            +label=Executor Instances
            +discriminator=type
          type: object
          default:
            type: fixed
            count: 1
          oneOf:
            - $ref: '#/components/schemas/SparkExecutorFixedInstances'
            - $ref: '#/components/schemas/SparkExecutorDynamicScaling'
          discriminator:
            propertyName: type
        resources:
          $ref: '#/components/schemas/Resources'
      title: SparkExecutorConfig
      x-fern-type-name: SparkExecutorConfig
    RateLimitConfig:
      description: +label=Rate Limit Configuration
      type: object
      required:
        - type
        - rules
      properties:
        type:
          description: |-
            +value=gateway-rate-limiting-config
            +sort=2
          type: string
          enum:
            - gateway-rate-limiting-config
        rules:
          description: |-
            +usage=List of rate limit rules
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Rate Limit Rules
          type: array
          items:
            $ref: '#/components/schemas/RateLimitRule'
      title: RateLimitConfig
      x-fern-type-name: RateLimitConfig
    LoadBalancingConfig:
      description: +label=Load Balancing Configuration
      type: object
      required:
        - type
        - rules
      properties:
        type:
          description: |-
            +value=gateway-load-balancing-config
            +sort=2
          type: string
          enum:
            - gateway-load-balancing-config
        rules:
          description: |-
            +usage=List of load balancing rules
            +uiProps={"descriptionInline":true}
            +sort=4
            +label=Load Balancing Rules
          type: array
          items:
            $ref: '#/components/schemas/LoadBalancingRule'
      title: LoadBalancingConfig
      x-fern-type-name: LoadBalancingConfig
    GuardrailsConfig:
      description: +label=Guardrails Configuration
      type: object
      required:
        - type
        - rules
      properties:
        type:
          description: |-
            +value=gateway-guardrails-config
            +sort=2
          type: string
          enum:
            - gateway-guardrails-config
        rules:
          description: |-
            +usage=List of guardrail rules
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Guardrail Rules
          type: array
          items:
            $ref: '#/components/schemas/GuardrailsRule'
      title: GuardrailsConfig
      x-fern-type-name: GuardrailsConfig
    BudgetConfig:
      description: +label=Budget Configuration
      type: object
      required:
        - type
        - rules
      properties:
        type:
          description: |-
            +value=gateway-budget-config
            +sort=2
          type: string
          enum:
            - gateway-budget-config
        rules:
          description: |-
            +usage=List of budget rules
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Budget Rules
          type: array
          items:
            $ref: '#/components/schemas/BudgetRule'
      title: BudgetConfig
      x-fern-type-name: BudgetConfig
    GatewayOTELConfig:
      type: object
      required:
        - type
      properties:
        type:
          description: |-
            +value=gateway-otel-config
            +sort=1
          type: string
          enum:
            - gateway-otel-config
        otel_traces_exporter_config:
          description: >-
            +usage=Set this configuration to export LLM gateway OTEL traces to
            an external platform. Note that we only export OTEL traces via this
            configuration.

            +label=Otel Traces Exporter Configuration

            +sort=300

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/OTELTracesExporterHttpConfig'
            - $ref: '#/components/schemas/OTELTracesExporterGrpcConfig'
          discriminator:
            propertyName: type
        otel_metrics_exporter_config:
          description: >-
            +usage=Set this configuration to export LLM gateway OTEL metrics to
            an external platform. Note that we only export OTEL metrics via this
            configuration.

            +label=Otel Metrics Exporter Configuration

            +sort=400

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/OTELMetricsExporterHttpConfig'
            - $ref: '#/components/schemas/OTELMetricsExporterGrpcConfig'
          discriminator:
            propertyName: type
      title: GatewayOTELConfig
      x-fern-type-name: GatewayOTELConfig
    GatewayMetadataConfig:
      description: +label=Metadata Configuration
      type: object
      required:
        - name
        - type
        - rules
      properties:
        name:
          description: |-
            +usage=Name of the metadata configuration
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Configuration Name
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-]{0,60}[a-zA-Z0-9])$
        type:
          description: |-
            +value=gateway-metadata-config
            +sort=2
          type: string
          enum:
            - gateway-metadata-config
        rules:
          description: |-
            +usage=List of metadata rules
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Metadata Rules
          type: array
          items:
            $ref: '#/components/schemas/GatewayMetadataRule'
      title: GatewayMetadataConfig
      x-fern-type-name: GatewayMetadataConfig
    GatewayLoggingConfig:
      description: +label=Logging Configuration
      type: object
      required:
        - name
        - type
        - rules
      properties:
        name:
          description: |-
            +usage=Name of the logging configuration
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Configuration Name
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-]{0,60}[a-zA-Z0-9])$
        type:
          description: |-
            +value=gateway-logging-config
            +sort=2
          type: string
          enum:
            - gateway-logging-config
        rules:
          description: |-
            +usage=List of logging rules
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Logging Rules
          type: array
          items:
            $ref: '#/components/schemas/GatewayLoggingRule'
      title: GatewayLoggingConfig
      x-fern-type-name: GatewayLoggingConfig
    GlobalSettings:
      description: >-
        +label=Global Settings

        +usage=Configure global gateway settings that apply to all requests
        processed by the gateway. These settings establish default behaviors for
        retry logic, guardrail enforcement, and request logging that can be
        overridden by route-specific configurations or request headers. Use
        these settings to define baseline policies for reliability, security,
        and observability across your entire gateway deployment.
      type: object
      required:
        - type
      properties:
        type:
          description: |-
            +value=gateway-global-settings
            +sort=2
          type: string
          enum:
            - gateway-global-settings
        retry_settings:
          $ref: '#/components/schemas/RetryConfig'
        logging_settings:
          $ref: '#/components/schemas/LoggingMode'
        semantic_cache_settings:
          $ref: '#/components/schemas/SemanticCacheSettings'
      title: GlobalSettings
      x-fern-type-name: GlobalSettings
    GatewayDataAccessConfig:
      description: +label=Gateway Data Access Configuration
      type: object
      required:
        - type
        - rules
      properties:
        type:
          description: |-
            +value=gateway-data-access-config
            +sort=1
          type: string
          enum:
            - gateway-data-access-config
        rules:
          description: |-
            +usage=List of data access rules
            +uiProps={"descriptionInline":true}
            +sort=2
            +label=Data Access Rules
          type: array
          items:
            $ref: '#/components/schemas/DataAccessRule'
      title: GatewayDataAccessConfig
      x-fern-type-name: GatewayDataAccessConfig
    GatewayDataRoutingConfig:
      description: >-
        +label=Gateway Data Routing Config

        +usage=Configuration for gateway data routing with destinations and
        default settings.
      type: object
      required:
        - type
        - default
      properties:
        type:
          description: |-
            +value=gateway-data-routing-config
            +label=Type
            +usage=Type of configuration
            +sort=1
          type: string
          enum:
            - gateway-data-routing-config
        destinations:
          description: |-
            +label=Destinations
            +usage=List of conditional routing destinations
            +sort=2
          type: array
          items:
            $ref: '#/components/schemas/GatewayDataRoutingConfigDestination'
        default:
          $ref: '#/components/schemas/GatewayDataRoutingConfigDestinationStorage'
      title: GatewayDataRoutingConfig
      x-fern-type-name: GatewayDataRoutingConfig
    TeamOwnedBy:
      description: |-
        +label=Owner
        +usage=Names of the owners that own the team
      type: object
      required:
        - account
      properties:
        account:
          description: |-
            +label=Account Name
            +sort=1
            +usage=The name of the account that owns the team
            +placeholder=Search Account
            +uiType=AccountSelect
            +uiProps={"maxAllowedAccounts": 1}
          type: string
      title: TeamOwnedBy
      x-fern-type-name: TeamOwnedBy
    PolicyMutationOperation:
      description: +label=Mutate
      type: object
      required:
        - type
        - order
      properties:
        type:
          description: |-
            +label=Action Type
            +usage=Mutates or Validates Resources
            +value=mutate
          type: string
          enum:
            - mutate
        order:
          description: >-
            +label=Execution Order

            +usage=Determines the sequence in which mutation policies are
            executed. Policies with lower order value run first, followed by
            higher value. For example, a policy with order 10 runs before one
            with order 20. This is crucial when multiple mutation policies need
            to run in a specific sequence. The order must be between 1 and 100.

            +sort=5

            +message=Order must be a positive integer less than or equal to 100

            +uiProps={"descriptionInline":true}
          type: integer
          minimum: 1
          maximum: 100
      title: PolicyMutationOperation
      x-fern-type-name: PolicyMutationOperation
    PolicyValidationOperation:
      description: +label=Validate
      type: object
      required:
        - type
      properties:
        type:
          description: |-
            +label=Action Type
            +usage=Mutates or Validates Resources
            +value=validate
          type: string
          enum:
            - validate
      title: PolicyValidationOperation
      x-fern-type-name: PolicyValidationOperation
    PolicyEntityTypes:
      type: string
      enum:
        - service
        - async-service
        - job
        - notebook
        - ssh-server
        - workflow
        - helm
        - volume
      title: PolicyEntityTypes
      x-fern-type-name: PolicyEntityTypes
    PolicyActions:
      type: string
      enum:
        - apply
      title: PolicyActions
      x-fern-type-name: PolicyActions
    PolicyFilters:
      description: +label=Filters
      type: object
      properties:
        cluster_names:
          description: >-
            +label=Clusters

            +usage=Cluster names where Policy applies

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: array
          items:
            type: string
            pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        workspace_names:
          description: >-
            +label=Workspaces

            +usage=Workspace names where Policy applies

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: array
          items:
            type: string
            pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        env_names:
          description: >-
            +label=Environments

            +usage=Environment names where Policy applies

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: array
          items:
            type: string
            pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
      title: PolicyFilters
      x-fern-type-name: PolicyFilters
    AlertConfigResource:
      type: object
      required:
        - type
        - fqn
      properties:
        type:
          description: >-
            +label=Resource Type

            +usage=The type of resource this alert config belongs to. Can be
            either an application or a cluster.

            +uiType=Hidden
          type: string
          enum:
            - application
            - cluster
        fqn:
          description: |-
            +label=Resource FQN
            +usage=FQN of the resource this alert config belongs to.
            +uiType=Hidden
          type: string
      title: AlertConfigResource
      x-fern-type-name: AlertConfigResource
    NotificationTargetForAlertRule:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/Email'
        - $ref: '#/components/schemas/SlackWebhook'
        - $ref: '#/components/schemas/SlackBot'
        - $ref: '#/components/schemas/PagerDuty'
      discriminator:
        propertyName: type
      title: NotificationTargetForAlertRule
      x-fern-type-name: NotificationTargetForAlertRule
    PrometheusAlertRule:
      description: +label=Prometheus Alert Rule
      type: object
      required:
        - type
        - name
        - expression
        - for
        - severity
        - notification_enabled
      properties:
        type:
          description: |-
            +label=Type
            +value=prometheus-alert-rule
            +uiType=Hidden
          type: string
          enum:
            - prometheus-alert-rule
        name:
          description: >-
            +label=Name

            +usage=Enter a unique, descriptive name for your Prometheus alert
            rule. Choose a name that clearly indicates the condition being
            monitored. Please note that reserved alert names for TrueFoundry
            managed alerts are not allowed. For checking the list of reserved
            alert names, please refer to the documentation.

            +uiProps={"descriptionInline":true}

            +message=3 to 35 alphanumeric characters long, can contain - in
            between, cannot start with a number and cannot end with a -

            +sort=1
          type: string
          pattern: ^[a-zA-Z][a-zA-Z0-9\-]{1,33}[a-zA-Z0-9]$
        expression:
          description: >-
            +label=Prometheus Expression

            +usage=Enter a valid PromQL expression that defines the condition
            for triggering this alert. The alert will fire when this expression
            evaluates to true for the duration specified in the duration to
            trigger alert field.

            +uiProps={"descriptionInline":true, "language": "promql"}

            +placeholder=sum(increase(container_cpu_cfs_throttled_periods_total{container="postgresql",
            namespace="prod-ws"}[5m])) by (container, namespace) /
            sum(increase(container_cpu_cfs_periods_total[5m])) by (container,
            namespace) > ( 25 / 100 )

            +sort=3

            +uiType=CodeEditor
          type: string
        for:
          description: >-
            +label=Trigger After (seconds)

            +placeholder=300

            +usage=The prometheus expression must remain true for this duration
            (in seconds) before the alert is triggered. If the condition becomes
            false before this time elapses, the alert will not fire.

            +sort=4
          type: number
          minimum: 60
        severity:
          $ref: '#/components/schemas/AlertSeverity'
        description:
          description: >-
            +label=Description

            +usage=Description of the alert rule which will be displayed in the
            alert rule list. This can be used to provide more context about the
            alert rule.

            +uiProps={"descriptionInline":true}

            +placeholder=CPU throttled for more than 25% of the time in the last
            5 minutes for

            +sort=2
          type: string
        notification_enabled:
          description: >-
            +label=Notification Enabled

            +usage=When enabled, notifications will be sent to all configured
            target channels when the alert conditions are met.

            +uiProps={"descriptionInline":true}

            +sort=7
          type: boolean
          default: true
      title: PrometheusAlertRule
      x-fern-type-name: PrometheusAlertRule
    AutoRotate:
      description: |-
        +label=Enable Auto Rotation
        +sort=4
        +usage=Enable Auto Rotation to automatically rotate the token
        +message=Enable Auto Rotation to automatically rotate the token
        +uiProps={"disableEdit":true}
      type: object
      required:
        - auto_rotate_interval
        - grace_period
      properties:
        auto_rotate_interval:
          description: >-
            +label=Rotation Interval in days

            +sort=1

            +usage=Rotation Interval in days after which the token will be
            rotated. Minimum value is 30.
          type: integer
          minimum: 30
          default: 360
        grace_period:
          description: >-
            +label=Grace Period in days

            +sort=2

            +usage=Grace Period in days for which the token will be valid after
            rotation interval. Minimum value is 1.
          type: integer
          minimum: 1
          default: 30
      title: AutoRotate
      x-fern-type-name: AutoRotate
    NotificationTarget:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/Email'
        - $ref: '#/components/schemas/SlackWebhook'
        - $ref: '#/components/schemas/SlackBot'
        - $ref: '#/components/schemas/PagerDuty'
      discriminator:
        propertyName: type
      title: NotificationTarget
      x-fern-type-name: NotificationTarget
    SecretStoreConfig:
      description: |-
        +label=Sync to Secret Manager
        +sort=6
        +usage=Enable to store the virtual account token in a secret store
        +message=Enable to store the virtual account token in a secret store
      type: object
      required:
        - integration_fqn
        - secret_path
      properties:
        integration_fqn:
          description: |-
            +label=Secret Store
            +usage=The secret store to store the virtual account token in
            +sort=1
            +uiType=IntegrationSelect
          type: string
        secret_path:
          description: >-
            +label=Secret Path

            +usage=The path where the virtual account token will be stored in
            the secret store (example: secret/path/to/virtual-account-token)

            +message=The path should only contain alphanumeric characters,
            hyphens, and slashes and it should start with a slash if giving the
            path containing slashes in between

            +sort=2
          type: string
          pattern: ^/?[A-Za-z0-9/-]*[A-Za-z0-9]$
      title: SecretStoreConfig
      x-fern-type-name: SecretStoreConfig
    VirtualAccountOwnedBy:
      description: |-
        +label=Owner
        +usage=Names of the owners that own the virtual account
      type: object
      required:
        - team
      properties:
        account:
          description: |-
            +label=Account Name
            +sort=1
            +usage=The name of the account that owns this resource
            +placeholder=Search Account
            +uiType=Hidden
            +uiProps={"maxAllowedAccounts": 1}
          type: string
        team:
          description: |-
            +label=Team Name
            +sort=2
            +usage=The name of the team that owns this resource
            +placeholder=Search Team
            +uiType=UserSelect
            +uiProps={"optionTypes": ["teams"], "maxAllowedUsers": 1}
          type: string
      title: VirtualAccountOwnedBy
      x-fern-type-name: VirtualAccountOwnedBy
    TrueFoundryAgentManifest:
      type: object
      required:
        - type
        - name
        - description
        - model
        - collaborators
      properties:
        type:
          description: |-
            +value=truefoundry-agent
            +label=Type
            +usage=Type of the manifest
            +uiType=Hidden
          type: string
          enum:
            - truefoundry-agent
        name:
          description: >-
            +label=Name

            +usage=The name of the Agent

            +sort=100

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"isManagementField":true}
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +usage=The description of the Agent
            +sort=200
            +uiProps={"isManagementField":true}
          type: string
          maxLength: 2000
        tags:
          description: |-
            +label=Tags
            +usage=Tags for the Agent
            +sort=250
            +uiProps={"isManagementField":true}
          type: object
        model:
          description: |-
            +label=Model
            +usage=Model for the Agent
            +sort=300
          type: string
        model_params:
          description: >-
            +label=Model Params

            +usage=Model parameters (default and extra). Keys are param names,
            values are float, int, bool, or string.

            +sort=310
          type: object
          properties:
            max_tokens:
              description: |-
                +label=Max Tokens
                +usage=Maximum number of tokens to generate.
                +sort=100
              type: integer
            reasoning_effort:
              description: >-
                +label=Reasoning Effort

                +usage=Reasoning depth when the model and provider support this
                parameter.

                +sort=200
              type: string
              enum:
                - none
                - minimal
                - low
                - medium
                - high
            temperature:
              description: >-
                +label=Temperature

                +usage=Sampling temperature (0–2). Higher values increase
                randomness.

                +sort=300
              type: number
            top_p:
              description: >-
                +label=Top P

                +usage=Nucleus sampling threshold (0–1). Alternative to
                temperature.

                +sort=400
              type: number
        skills:
          description: |-
            +label=Skills
            +usage=List of agent skills to attach to this agent
            +sort=315
          type: array
          items:
            $ref: '#/components/schemas/TrueFoundryAgentSkill'
        mcp_servers:
          description: >-
            +label=MCP Servers

            +usage=List of MCP servers with name, enable_all_tools, and optional
            tools filter

            +sort=320
          type: array
          items:
            $ref: '#/components/schemas/TrueFoundryAgentMcpServer'
        instruction:
          description: |-
            +label=Instruction
            +usage=System instruction or prompt for the agent
            +sort=330
          type: string
        messages:
          description: >-
            +label=Messages

            +usage=Pre-defined initial messages for the agent. Supports template
            variable replacement.

            +sort=335
          type: array
          items:
            $ref: '#/components/schemas/TrueFoundryAgentUserMessage'
        variables:
          description: >-
            +label=Variables

            +usage=Variables keyed by name. Each has optional default_value and
            optional description.

            +sort=340
          type: object
          additionalProperties:
            $ref: '#/components/schemas/TrueFoundryAgentVariable'
        response_format:
          description: |-
            +label=Response Format
            +usage=JSON schema or structure for the agent response format
            +sort=350
          type: object
          oneOf:
            - $ref: '#/components/schemas/ResponseFormatText'
            - $ref: '#/components/schemas/ResponseFormatJSONSchema'
            - $ref: '#/components/schemas/ResponseFormatJSONObject'
        sandbox:
          description: |-
            +label=Sandbox
            +usage=Sandbox execution settings
            +sort=360
          type: object
          required:
            - enabled
          properties:
            enabled:
              type: boolean
        iteration_limit:
          description: |-
            +label=Iteration Limit
            +usage=Maximum number of iterations for the agent
            +sort=370
          type: integer
          maximum: 100
          exclusiveMaximum: true
        collaborators:
          description: |-
            +label=Collaborators
            +usage=List of users who have access to this Agent
            +uiType=Collaborators
            +sort=1000
            +uiProps={"isManagementField":true}
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: TrueFoundryAgentManifest
      x-fern-type-name: TrueFoundryAgentManifest
    RemoteAgent:
      $ref: '#/components/schemas/SelfHostedAgent'
      title: RemoteAgent
      x-fern-type-name: RemoteAgent
    LegacyAgentManifest:
      type: object
      required:
        - type
        - name
        - description
        - source
        - collaborators
      properties:
        type:
          description: |-
            +value=agent
            +label=Type
            +usage=Type of the manifest
            +uiType=Hidden
          type: string
          enum:
            - agent
        name:
          description: >-
            +label=Name

            +usage=The name of the Agent

            +sort=100

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +usage=The description of the Agent
            +sort=200
          type: string
          maxLength: 2000
        source:
          $ref: '#/components/schemas/AgentSource'
        collaborators:
          description: |-
            +label=Collaborators
            +usage=List of users who have access to this Agent
            +uiType=Collaborators
            +sort=400
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        sample_inputs:
          description: >-
            +label=Sample Inputs

            +usage=Sample inputs for your agent. These inputs are shown as an
            example in the "Agent Chat" page. (Click on Try Now in the agent
            listing page)

            +sort=500
          type: array
          items:
            $ref: '#/components/schemas/SampleAgentInput'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: LegacyAgentManifest
      x-fern-type-name: LegacyAgentManifest
    EnvironmentColor:
      type: object
      properties:
        colorHex:
          type: string
        backgroundColorHex:
          type: string
        color:
          type: string
        backgroundColor:
          type: string
      title: EnvironmentColor
      x-fern-type-name: EnvironmentColor
    EnvironmentOptimizeFor:
      type: string
      enum:
        - COST
        - AVAILABILITY
      title: EnvironmentOptimizeFor
      x-fern-type-name: EnvironmentOptimizeFor
    TracingProjectStorageConfig:
      description: |-
        Storage configuration for tracing project
        +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/TracingProjectControlPlaneManagedStorage'
        - $ref: '#/components/schemas/TracingProjectCustomerManagedStorage'
      discriminator:
        propertyName: type
      title: TracingProjectStorageConfig
      x-fern-type-name: TracingProjectStorageConfig
    RemoteMCPServerManifest:
      description: |-
        +label=Remote MCP Server
        +icon=puzzle-piece
        +usage=Connect to an external MCP Server endpoint.
      type: object
      required:
        - type
        - name
        - description
        - url
        - collaborators
      properties:
        type:
          description: +value=mcp-server/remote
          type: string
          enum:
            - mcp-server/remote
        name:
          description: >-
            +label=Name

            +usage=The name of the MCP Server.

            +message=3 to 73 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +sort=100

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,71}[a-z0-9]$
        description:
          description: >-
            +label=Description

            +usage=Provide a brief description of the purpose of this MCP
            Server.

            +uiType=TextArea

            +message=1 to 1000 characters long, may contain any character except
            newlines

            +sort=200

            +uiProps={"descriptionInline":true}
          type: string
          minLength: 1
          maxLength: 1000
        url:
          description: |-
            +label=URL
            +usage=The endpoint URL for the MCP Server.
            +sort=300
          type: string
          pattern: ^https?:\/\/[^\s]+$
        tls_settings:
          $ref: '#/components/schemas/CustomTLSSettings'
        tool_settings:
          description: >-
            +label=Tool Settings

            +usage=Customize tool descriptions or enable/disable specific tools
            from the MCP Server.

            +uiType=Hidden

            +sort=450
          type: array
          items:
            $ref: '#/components/schemas/MCPToolSetting'
        auth_data:
          $ref: '#/components/schemas/MCPServerAuth'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=600
            +usage=Users and Teams that have access to this MCP Server
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
        tags:
          description: >-
            +label=Tags

            +usage=Key-value pairs to categorize this MCP Server (e.g., by owner
            or environment).

            +message=Maximum 10 tags, keys up to 32 characters, values up to 64
            characters.

            +isAdvanced=true

            +sort=1000
          type: object
      title: RemoteMCPServerManifest
      x-fern-type-name: RemoteMCPServerManifest
    VirtualMCPServerManifest:
      description: |-
        +label=Virtual MCP Server
        +icon=puzzle-piece
        +usage=TrueFoundry-hosted MCP Server composed of other MCP servers.
      type: object
      required:
        - type
        - name
        - description
        - servers
        - collaborators
      properties:
        type:
          description: +value=mcp-server/virtual
          type: string
          enum:
            - mcp-server/virtual
        name:
          description: >-
            +label=Name

            +usage=The name of the Virtual MCP Server.

            +message=3 to 73 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +sort=100

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,71}[a-z0-9]$
        description:
          description: >-
            +label=Description

            +usage=Provide a brief description of the purpose of this Virtual
            MCP Server.

            +uiType=TextArea

            +message=1 to 1000 characters long, may contain any character except
            newlines

            +sort=200

            +uiProps={"descriptionInline":true}
          type: string
          minLength: 1
          maxLength: 1000
        servers:
          description: |-
            +label=Source MCP Servers
            +usage=List of source MCP servers to include in this virtual server.
            +sort=300
          type: array
          items:
            $ref: '#/components/schemas/McpServerSource'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=Users and Teams that have access to this Virtual MCP Server
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
        tags:
          description: >-
            +label=Tags

            +usage=Key-value pairs to categorize this Virtual MCP Server (e.g.,
            by owner or environment).

            +message=Maximum 10 tags, keys up to 32 characters, values up to 64
            characters.

            +isAdvanced=true

            +sort=1000
          type: object
      title: VirtualMCPServerManifest
      x-fern-type-name: VirtualMCPServerManifest
    OpenAPIMCPServerManifest:
      description: |-
        +label=OpenAPI MCP Server
        +icon=puzzle-piece
        +usage=Connect to an external OpenAPI Server endpoint.
      type: object
      required:
        - type
        - name
        - description
        - url
        - openapi_spec_source
        - collaborators
      properties:
        type:
          description: +value=mcp-server/openapi
          type: string
          enum:
            - mcp-server/openapi
        name:
          description: >-
            +label=Name

            +usage=The name of the OpenAPI MCP Server.

            +message=3 to 73 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +sort=100

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,71}[a-z0-9]$
        description:
          description: >-
            +label=Description

            +usage=Provide a brief description of the purpose of this OpenAPI
            MCP Server.

            +uiType=TextArea

            +message=1 to 1000 characters long, may contain any character except
            newlines

            +sort=200

            +uiProps={"descriptionInline":true}
          type: string
          minLength: 1
          maxLength: 1000
        url:
          description: >-
            +label=OpenAPI Server URL

            +usage=The base URL where API requests will be sent. This should be
            the root endpoint of your API server, without any path suffix.

            +placeholder=https://myapp.example.org

            +message=Must be a valid HTTP/HTTPS URL

            +sort=300
          type: string
          pattern: ^https?:\/\/[^\s]+$
        openapi_spec_source:
          $ref: '#/components/schemas/OpenAPISpecSource'
        tls_settings:
          $ref: '#/components/schemas/CustomTLSSettings'
        enable_all_tools:
          description: >-
            +label=Enable All Tools

            +usage=When enabled, all tools from the OpenAPI spec will be
            available (up to 30 tools). Disable to select specific tools.

            +sort=600
          type: boolean
        tool_settings:
          description: |-
            +label=Tool Settings
            +usage=The tools to customize from the OpenAPI Server.
            +sort=700
          type: array
          items:
            $ref: '#/components/schemas/OpenAPIMCPToolSetting'
        auth_data:
          $ref: '#/components/schemas/MCPServerAuth'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=900
            +usage=Users and Teams that have access to this OpenAPI Server
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
        tags:
          description: >-
            +label=Tags

            +usage=Key-value pairs to categorize this OpenAPI MCP Server (e.g.,
            by owner or environment).

            +message=Maximum 10 tags, keys up to 32 characters, values up to 64
            characters.

            +isAdvanced=true

            +sort=1000
          type: object
      title: OpenAPIMCPServerManifest
      x-fern-type-name: OpenAPIMCPServerManifest
    StdioMCPServerManifest:
      description: |-
        +label=Hosted Stdio-based MCP Server
        +icon=puzzle-piece
        +usage=Stdio MCP server configuration.
      type: object
      required:
        - name
        - description
        - collaborators
        - type
        - command
      properties:
        name:
          description: >-
            +label=Name

            +usage=The name of the MCP Server.

            +message=3 to 73 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +sort=100

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,71}[a-z0-9]$
        description:
          description: >-
            +label=Description

            +usage=Provide a brief description of the purpose of this MCP
            Server.

            +uiType=TextArea

            +message=1 to 1000 characters long, may contain any character except
            newlines

            +sort=200

            +uiProps={"descriptionInline":true}
          type: string
          minLength: 1
          maxLength: 1000
        collaborators:
          description: |-
            +label=Collaborators
            +sort=300
            +usage=Users and Teams that have access to this MCP Server
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        type:
          description: |-
            +value=mcp-server/stdio
            +sort=400
          type: string
          enum:
            - mcp-server/stdio
        command:
          description: |-
            +label=Command
            +usage=Executable to run for the stdio MCP server process.
            +sort=500
            +placeholder=npx
          type: string
          minLength: 1
          default: npx
        args:
          description: |-
            +label=Arguments
            +usage=Arguments passed to the stdio MCP server command.
            +sort=600
          type: array
          items:
            type: string
        auth_data:
          $ref: '#/components/schemas/MCPServerEnvAuth'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
        tags:
          description: >-
            +label=Tags

            +usage=Key-value pairs to categorize this MCP Server (e.g., by owner
            or environment).

            +message=Maximum 10 tags, keys up to 32 characters, values up to 64
            characters.

            +isAdvanced=true

            +sort=1000
          type: object
      title: StdioMCPServerManifest
      x-fern-type-name: StdioMCPServerManifest
    RoleBindingSubject:
      description: >-
        Single row per principal: type selects how to interpret name (email vs
        names).

        Field names mirror the manifest root (type + name) for consistency in
        the UI and YAML.
      type: object
      required:
        - type
        - name
      properties:
        type:
          description: >-
            +label=Type

            +usage=Whether this row identifies a user (email), a team (name), a
            virtual account (name), or an external identity (name).
          type: string
          enum:
            - user
            - team
            - virtualaccount
            - external-identity
        name:
          description: >-
            +label=Name

            +usage=User email when type is user; team or virtual account name
            when type is team or virtualaccount; external identity name when
            type is external-identity.
          type: string
          minLength: 1
          maxLength: 1024
      title: RoleBindingSubject
      x-fern-type-name: RoleBindingSubject
    RoleBindingPermission:
      description: Grants a named account role on a specific resource identified by FQN.
      type: object
      required:
        - resourceType
        - resourceFqn
        - role
      properties:
        resourceType:
          description: |-
            +label=Resource type
            +usage=Target resource kind (e.g. workspace, cluster, ml-repo).
          type: string
          minLength: 1
          maxLength: 128
        resourceFqn:
          description: |-
            +label=Resource FQN
            +usage=Fully qualified name of the resource.
          type: string
          minLength: 1
          maxLength: 1024
        role:
          description: |-
            +label=Role
            +usage=Name of the role to bind (e.g. workspace-editor, model-user).
          type: string
          minLength: 1
          maxLength: 255
      title: RoleBindingPermission
      x-fern-type-name: RoleBindingPermission
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
    CPUUtilizationMetric:
      type: object
      required:
        - type
        - value
      properties:
        type:
          description: +value=cpu_utilization
          type: string
          enum:
            - cpu_utilization
        value:
          description: >-
            +label=CPU utilization %

            +usage=Percentage of cpu request averaged over all replicas which
            the autoscaler should try to maintain
          type: integer
          minimum: 1
          maximum: 100
      title: CPUUtilizationMetric
      x-fern-type-name: CPUUtilizationMetric
    RPSMetric:
      type: object
      required:
        - type
        - value
      properties:
        type:
          description: +value=rps
          type: string
          enum:
            - rps
        value:
          description: >-
            +label=Requests per second

            +usage=Average request per second averaged over all replicas that
            autoscaler should try to maintain
          type: number
          minimum: 0
          exclusiveMinimum: true
      title: RPSMetric
      x-fern-type-name: RPSMetric
    CronMetric:
      type: object
      required:
        - type
        - start
        - end
        - timezone
      properties:
        type:
          description: +value=cron
          type: string
          enum:
            - cron
        desired_replicas:
          description: >-
            +label=Desired Replicas

            +usage=Desired number of replicas during the given interval. Default
            value is max_replicas.
          type: integer
          minimum: 1
        start:
          description: |-
            +label=Start Schedule
            +docs=Cron expression indicating the start of the cron schedule.
            +usage=Cron expression indicating the start of the cron schedule.
            ```
            * * * * *
            | | | | |
            | | | | |___ day of week (0-6) (Sunday is 0)
            | | | |_____ month (1-12)
            | | |_______ day of month (1-31)
            | |_________ hour (0-23)
            |___________ minute (0-59)
            ```
          type: string
        end:
          description: |-
            +label=End Schedule
            +docs=Cron expression indicating the end of the cron schedule.
            +usage=Cron expression indicating the end of the cron schedule.
            ```
            * * * * *
            | | | | |
            | | | | |___ day of week (0-6) (Sunday is 0)
            | | | |_____ month (1-12)
            | | |_______ day of month (1-31)
            | |_________ hour (0-23)
            |___________ minute (0-59)
            ```
          type: string
        timezone:
          description: >-
            +usage=Timezone against which the cron schedule will be calculated,
            e.g. "Asia/Tokyo". Default is machine's local time.

            https://docs.truefoundry.com/docs/list-of-supported-timezones
          type: string
          default: UTC
      title: CronMetric
      x-fern-type-name: CronMetric
    BaseAutoscaling:
      type: object
      required:
        - min_replicas
        - max_replicas
      properties:
        min_replicas:
          description: |-
            +label=Minimum replicas
            +usage=Minimum number of replicas to keep available
            +sort=1
          type: integer
          minimum: 0
          default: 1
        max_replicas:
          description: |-
            +label=Maximum replicas
            +usage=Maximum number of replicas allowed for the component.
            +sort=2
          type: integer
          minimum: 1
          maximum: 500
        polling_interval:
          description: |-
            +label=Polling Interval
            +usage=This is the interval to check each trigger on.
          type: integer
          minimum: 0
          default: 30
      title: BaseAutoscaling
      x-fern-type-name: BaseAutoscaling
    CanaryStep:
      type: object
      required:
        - weight_percentage
      properties:
        weight_percentage:
          description: >-
            +docs=Percentage of total traffic to be shifted to the canary
            release.

            The rest will continue to go to the existing deployment

            +label=Canary weight percentage

            +unit=%

            +placeholder=Weight
          type: integer
          minimum: 0
          maximum: 100
        pause_duration:
          description: >-
            +docs=Duration for which to pause the release. The release process
            will wait for these seconds before proceeding to the next step.

            If this is not set, the step will pause indefinitely on this step

            +label=Pause duration

            +unit=seconds

            +placeholder=Duration
          type: integer
          minimum: 0
      title: CanaryStep
      x-fern-type-name: CanaryStep
    ArtifactsDownload:
      description: >-
        +docs=Describes the configuration for the artifacts cache

        +label=Artifacts Download

        +usage=Download and cache models in a volume to enhance loading speeds
        and reduce costs by avoiding repeated downloads.
        [Docs](https://docs.truefoundry.com/docs/download-and-cache-models)
      type: object
      required:
        - artifacts
      properties:
        cache_volume:
          $ref: '#/components/schemas/ArtifactsCacheVolume'
        artifacts:
          description: |-
            +label=Artifacts
            +usage=List of artifacts to be cached
            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/TrueFoundryArtifactSource'
              - $ref: '#/components/schemas/HuggingfaceArtifactSource'
      title: ArtifactsDownload
      x-fern-type-name: ArtifactsDownload
    Port:
      description: +docs=Describes the ports the service should be exposed to.
      type: object
      required:
        - port
        - protocol
        - expose
      properties:
        port:
          description: +usage=Port number to expose.
          type: integer
          minimum: 1
          maximum: 65535
          default: 80
        protocol:
          description: +usage=Protocol for the port.
          type: string
          enum:
            - TCP
            - UDP
          default: TCP
        expose:
          description: +usage=Expose the port
          type: boolean
          default: true
        app_protocol:
          description: >-
            +label=Application Protocol

            +usage=Application Protocol for the port.

            Select the application protocol used by your service. For most use
            cases, this should be `http`(HTTP/1.1).

            If you are running a gRPC server, select the `grpc` option.

            This is only applicable if `expose=true`.
          type: string
          enum:
            - http
            - grpc
            - tcp
          default: http
        host:
          description: >-
            +usage=Host e.g. ai.example.com, app.truefoundry.com

            +message=Upto 253 characters, each part of host should be at most 63
            characters long, can contain alphabets, digits and hypen, must begin
            and end with an alphanumeric characters. Parts must be separated by
            periods (.)
          type: string
          pattern: >-
            ^((([a-zA-Z0-9\-]{1,63}\.)([a-zA-Z0-9\-]{1,63}\.)*([A-Za-z]{1,63}))|(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))$
        path:
          description: >-
            +usage=Path e.g. /v1/api/ml/, /v2/docs/

            +message=Should begin and end with a forward slash (/). Each part
            can can contain alphabets, digits and hypen, must begin and end with
            an alphanumeric characters. Parts should be separated by forward
            slashes (/)
          type: string
          pattern: ^(/([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-_\.]*[a-zA-Z0-9]))*/$
        rewrite_path_to:
          description: >-
            +label=Rewrite Path to

            +usage=Rewrite the path prefix to a different path.

            If `path` is `/v1/api` and `rewrite_path_to` is `/api`. The URI in
            the HTTP request `http://0.0.0.0:8080/v1/api/houses` will be
            rewritten to `http://0.0.0.0:8080/api/houses` before the request is
            forwarded your service.

            Defaults to `/`.

            This is only applicable if `path` is given.

            +message=Should begin and end with a forward slash (/). Each part
            can can contain alphabets, digits and hypen, must begin and end with
            an alphanumeric characters. Parts should be separated by forward
            slashes (/)
          type: string
          pattern: ^(/([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-_\.]*[a-zA-Z0-9]))*/$
        auth:
          description: |-
            +usage=Authentication method for inbound traffic
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/BasicAuthCreds'
            - $ref: '#/components/schemas/JwtAuthConfig'
            - $ref: '#/components/schemas/TrueFoundryInteractiveLogin'
          discriminator:
            propertyName: type
      title: Port
      x-fern-type-name: Port
    HealthProbe:
      description: >-
        +docs=Describes the configuration for the Health Probe's

        To learn more you can go
        [here](https://docs.truefoundry.com/docs/liveness-readiness-probe)

        +icon=fa-heart

        +uiType=HealthProbe
      type: object
      required:
        - config
      properties:
        config:
          $ref: '#/components/schemas/HttpProbe'
        initial_delay_seconds:
          description: >-
            +usage=Time to wait after container has started before checking the
            endpoint
          type: integer
          minimum: 0
          maximum: 36000
          default: 0
        period_seconds:
          description: +usage=How often to check the endpoint
          type: integer
          minimum: 1
          maximum: 36000
          default: 10
        timeout_seconds:
          description: >-
            +usage=Time to wait for a response from the endpoint before
            considering it down
          type: integer
          minimum: 1
          maximum: 36000
          default: 1
        success_threshold:
          description: >-
            +usage=Number of successful responses from the endpoint before
            container is considered healthy
          type: integer
          minimum: 1
          maximum: 5000
          default: 1
        failure_threshold:
          description: >-
            +usage=Number of consecutive failures before the container is
            considered down
          type: integer
          minimum: 1
          maximum: 5000
          default: 3
      title: HealthProbe
      x-fern-type-name: HealthProbe
    AwsProviderAccount:
      description: |-
        +label=AWS
        +icon=aws
        +heroDataKey=aws_account_id
      type: object
      required:
        - type
        - name
        - aws_account_id
        - integrations
      properties:
        type:
          description: +value=provider-account/aws
          type: string
          enum:
            - provider-account/aws
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the AWS provider account.

            +sort=200

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        aws_account_id:
          description: |-
            +label=AWS Account ID
            +usage=The AWS account ID.
            +sort=300
          type: string
          pattern: ^\S+$
        auth_data:
          description: |-
            +label=AWS Account Auth Data
            +usage=Authentication data for the AWS account.
            +sort=400
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/AwsAssumedRoleBasedAuth'
          discriminator:
            propertyName: type
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the AWS
            provider account.

            +sort=500

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/AwsIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: AwsProviderAccount
      x-fern-type-name: AwsProviderAccount
    AzureProviderAccount:
      description: |-
        +label=Azure
        +icon=azure
        +heroDataKey=subscription_id
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/azure
          type: string
          enum:
            - provider-account/azure
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the Azure provider account.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        subscription_id:
          description: |-
            +label=Azure Subscription Id
            +usage=The ID of the target subscription. The value must be an UUID.
            +sort=200
          type: string
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the Azure
            provider account.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/AzureIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: AzureProviderAccount
      x-fern-type-name: AzureProviderAccount
    GcpProviderAccount:
      description: |-
        +label=GCP
        +icon=googleCloud
        +heroDataKey=project_id
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/gcp
          type: string
          enum:
            - provider-account/gcp
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the GCP provider account.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        project_id:
          description: |-
            +label=Project Id
            +usage=The project id of the GCP account.
            +sort=200
          type: string
        auth_data:
          $ref: '#/components/schemas/GcpKeyFileAuth'
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the GCP
            provider account.

            +sort=500

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/GcpIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: GcpProviderAccount
      x-fern-type-name: GcpProviderAccount
    DockerhubProviderAccount:
      description: |-
        +label=Dockerhub
        +icon=docker
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/dockerhub
          type: string
          enum:
            - provider-account/dockerhub
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        account_name:
          description: +label=Account Name
          type: string
          pattern: ^\S+$
        auth_data:
          $ref: '#/components/schemas/DockerhubBasicAuth'
        integrations:
          description: +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/DockerhubIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: DockerhubProviderAccount
      x-fern-type-name: DockerhubProviderAccount
    BitbucketProviderAccount:
      description: |-
        +label=Bitbucket
        +icon=bitbucket
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/bitbucket
          type: string
          enum:
            - provider-account/bitbucket
        name:
          description: +uiProps={"disableEdit":true}
          type: string
          pattern: ^\S+$
        integrations:
          description: +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/BitbucketIntegration'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: BitbucketProviderAccount
      x-fern-type-name: BitbucketProviderAccount
    CustomProviderAccount:
      description: |-
        +label=Custom
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/custom
          type: string
          enum:
            - provider-account/custom
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the provider account.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/CustomBasicAuth'
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the provider
            account.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/CustomIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: CustomProviderAccount
      x-fern-type-name: CustomProviderAccount
    GithubProviderAccount:
      description: |-
        +label=Github
        +icon=github
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/github
          type: string
          enum:
            - provider-account/github
        name:
          description: +uiProps={"disableEdit":true}
          type: string
          pattern: ^\S+$
        integrations:
          description: +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/GithubIntegration'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: GithubProviderAccount
      x-fern-type-name: GithubProviderAccount
    GitlabProviderAccount:
      description: |-
        +label=Gitlab
        +icon=gitlab
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/gitlab
          type: string
          enum:
            - provider-account/gitlab
        name:
          description: +uiProps={"disableEdit":true}
          type: string
          pattern: ^\S+$
        integrations:
          description: +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/GitlabIntegration'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: GitlabProviderAccount
      x-fern-type-name: GitlabProviderAccount
    JfrogProviderAccount:
      description: |-
        +label=Jfrog
        +icon=jfrog
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/jfrog
          type: string
          enum:
            - provider-account/jfrog
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the provider account.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        account_name:
          description: |-
            +label=Account Name
            +usage=The name of the JFrog account.
            +sort=200
          type: string
          pattern: ^\S+$
        auth_data:
          $ref: '#/components/schemas/JfrogBasicAuth'
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the provider
            account.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/JFrogIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: JfrogProviderAccount
      x-fern-type-name: JfrogProviderAccount
    TTLProviderAccount:
      description: |-
        +label=TTL
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/ttl
          type: string
          enum:
            - provider-account/ttl
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the provider account.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the provider
            account.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/TTLIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: TTLProviderAccount
      x-fern-type-name: TTLProviderAccount
    TrueFoundryProviderAccount:
      description: |-
        +label=TrueFoundry
        +icon=truefoundry
      type: object
      required:
        - type
        - name
      properties:
        type:
          description: +value=provider-account/truefoundry
          type: string
          enum:
            - provider-account/truefoundry
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the provider account.

            +sort=100

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the provider
            account.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/TrueFoundryIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: TrueFoundryProviderAccount
      x-fern-type-name: TrueFoundryProviderAccount
    QuayProviderAccount:
      description: |-
        +label=Quay
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/quay
          type: string
          enum:
            - provider-account/quay
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the provider account.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        account_name:
          description: |-
            +label=Account Name
            +usage=The name of the account on Quay.
            +sort=200
          type: string
          pattern: ^\S+$
        auth_data:
          $ref: '#/components/schemas/QuayBasicAuth'
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the provider
            account.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/QuayIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: QuayProviderAccount
      x-fern-type-name: QuayProviderAccount
    SlackProviderAccount:
      description: |-
        +label=Slack Provider Account
        +icon=slack
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/slack
          type: string
          enum:
            - provider-account/slack
        name:
          description: >-
            +uiProps={"disableEdit":false}

            +label=Name

            +usage=The name of the provider account.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the provider
            account.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/SlackIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: SlackProviderAccount
      x-fern-type-name: SlackProviderAccount
    WebhookProviderAccount:
      description: |-
        +label=Webhook Provider Account
        +icon=webhook
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/webhook
          type: string
          enum:
            - provider-account/webhook
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the provider account.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the provider
            account.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/WebhookIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: WebhookProviderAccount
      x-fern-type-name: WebhookProviderAccount
    PagerDutyProviderAccount:
      description: |-
        +label=PagerDuty Provider Account
        +icon=https://assets.production.truefoundry.com/pagerduty.png
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/pagerduty
          type: string
          enum:
            - provider-account/pagerduty
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the PagerDuty provider account

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,34}[a-z0-9]$
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the PagerDuty
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/PagerDutyIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: PagerDutyProviderAccount
      x-fern-type-name: PagerDutyProviderAccount
    HashicorpProviderAccount:
      description: |-
        +label=HashiCorp Vault
        +icon=https://assets.production.truefoundry.com/hashicorp.svg
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/hashicorp
          type: string
          enum:
            - provider-account/hashicorp
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the HashiCorp Vault provider account.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the HashiCorp
            Vault provider account.

            +sort=200

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/HashicorpIntegrations'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: HashicorpProviderAccount
      x-fern-type-name: HashicorpProviderAccount
    AwsBedrockProviderAccount:
      description: |-
        +label=AWS Bedrock
        +icon=aws
      type: object
      required:
        - type
        - name
        - region
        - integrations
      properties:
        type:
          description: +value=provider-account/aws-bedrock
          type: string
          enum:
            - provider-account/aws-bedrock
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the AWS Bedrock provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        region:
          $ref: '#/components/schemas/AWSRegion'
        auth_data:
          description: |-
            +label=AWS Account Auth Data
            +sort=400
            +usage=Authentication data for the AWS account
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/AwsBedrockAssumedRoleBasedAuth'
            - $ref: '#/components/schemas/AwsBedrockApiKeyAuth'
          discriminator:
            propertyName: type
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the AWS Bedrock
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/BedrockModel'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=600
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: AwsBedrockProviderAccount
      x-fern-type-name: AwsBedrockProviderAccount
    GoogleVertexProviderAccount:
      description: |-
        +label=Google Vertex Provider Account
        +icon=googleCloud
        +heroDataKey=project_id
      type: object
      required:
        - type
        - name
        - project_id
        - region
        - integrations
      properties:
        type:
          description: +value=provider-account/google-vertex
          type: string
          enum:
            - provider-account/google-vertex
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Google Vertex provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        project_id:
          description: |-
            +label=Project ID
            +sort=200
            +usage=The Google Cloud project ID where Vertex AI is enabled
            +message=Project ID must not be empty
          type: string
          pattern: ^\S+$
        region:
          $ref: '#/components/schemas/GCPRegion'
        auth_data:
          description: >-
            +label=Authentication Data

            +sort=400

            +usage=Service account key JSON, or Workload Identity Federation
            file (workload-identity-federation-file) with external_account JSON
            (e.g. create-cred-config).
            https://docs.truefoundry.com/gateway/google-vertex

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/VertexKeyFileAuth'
            - $ref: '#/components/schemas/VertexWifFileAuth'
          discriminator:
            propertyName: type
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the Google
            Vertex provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/VertexModel'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=600
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: GoogleVertexProviderAccount
      x-fern-type-name: GoogleVertexProviderAccount
    GoogleGeminiProviderAccount:
      description: |-
        +label=Google Gemini Provider Account
        +icon=googleCloud
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/google-gemini
          type: string
          enum:
            - provider-account/google-gemini
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Google Gemini provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/GcpApiKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the Google
            Gemini provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/GeminiModel'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: GoogleGeminiProviderAccount
      x-fern-type-name: GoogleGeminiProviderAccount
    AzureOpenAIProviderAccount:
      description: |-
        +label=Azure OpenAI Provider Account
        +icon=azure
      type: object
      required:
        - type
        - name
        - azure_endpoint
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/azure-openai
          type: string
          enum:
            - provider-account/azure-openai
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Azure OpenAI provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        azure_endpoint:
          description: >-
            +label=Azure Endpoint

            +sort=300

            +usage=The Azure OpenAI Service endpoint URL. Should look like
            https://{resource-name}.openai.azure.com

            +message=Endpoint URL must not be empty

            +uiProps={"warningValidation":{"message": "This URL looks like an
            Azure AI Foundry endpoint. Did you mean to use the Azure AI Foundry
            provider instead?", "pattern":
            "^(?!.*\\.models\\.ai\\.azure\\.com).*"}}
          type: string
          pattern: ^\S+$
        auth_data:
          description: |-
            +label=Authentication Data
            +sort=400
            +usage=Azure OpenAI authentication credentials
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AzureKeyAuth'
            - $ref: '#/components/schemas/AzureCertificateAuth'
          discriminator:
            propertyName: type
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the Azure
            OpenAI provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/AzureOpenAIModel'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: AzureOpenAIProviderAccount
      x-fern-type-name: AzureOpenAIProviderAccount
    AzureFoundryProviderAccount:
      description: |-
        +label=Azure AI Foundry Provider Account
        +icon=azure
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/azure-foundry
          type: string
          enum:
            - provider-account/azure-foundry
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Azure AI Foundry provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        integrations:
          description: >-
            +label=Integrations

            +sort=200

            +usage=List of integrations that are associated with the Azure
            Foundry provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/AzureFoundryModel'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=300
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: AzureFoundryProviderAccount
      x-fern-type-name: AzureFoundryProviderAccount
    CohereProviderAccount:
      description: |-
        +label=Cohere Provider Account
        +icon=cohere
      type: object
      required:
        - type
        - name
        - auth_data
      properties:
        type:
          description: +value=provider-account/cohere
          type: string
          enum:
            - provider-account/cohere
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Cohere provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/CohereKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the Cohere
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/CohereIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: CohereProviderAccount
      x-fern-type-name: CohereProviderAccount
    AI21ProviderAccount:
      description: |-
        +label=AI21
        +icon=ai21
      type: object
      required:
        - type
        - name
        - auth_data
      properties:
        type:
          description: +value=provider-account/ai21
          type: string
          enum:
            - provider-account/ai21
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the AI21 provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/AI21KeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=400

            +usage=List of integrations that are associated with the AI21
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/AI21Integrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: AI21ProviderAccount
      x-fern-type-name: AI21ProviderAccount
    AnthropicProviderAccount:
      description: |-
        +label=Anthropic
        +icon=anthropic
      type: object
      required:
        - type
        - name
        - auth_data
      properties:
        type:
          description: +value=provider-account/anthropic
          type: string
          enum:
            - provider-account/anthropic
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Anthropic provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/AnthropicKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the Anthropic
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/AnthropicIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: AnthropicProviderAccount
      x-fern-type-name: AnthropicProviderAccount
    DeepinfraProviderAccount:
      description: |-
        +label=DeepInfra Provider Account
        +icon=deepinfra
      type: object
      required:
        - type
        - name
        - auth_data
      properties:
        type:
          description: +value=provider-account/deepinfra
          type: string
          enum:
            - provider-account/deepinfra
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the DeepInfra provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/DeepinfraKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the DeepInfra
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/DeepinfraIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: DeepinfraProviderAccount
      x-fern-type-name: DeepinfraProviderAccount
    GroqProviderAccount:
      description: |-
        +label=Groq Provider Account
        +icon=groq
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/groq
          type: string
          enum:
            - provider-account/groq
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Groq provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/GroqKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the Groq
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/GroqIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: GroqProviderAccount
      x-fern-type-name: GroqProviderAccount
    MistralAIProviderAccount:
      description: |-
        +label=Mistral AI Provider Account
        +icon=mistral-ai
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/mistral-ai
          type: string
          enum:
            - provider-account/mistral-ai
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Mistral AI provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/MistralAIKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the Mistral AI
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/MistralAIIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: MistralAIProviderAccount
      x-fern-type-name: MistralAIProviderAccount
    PerplexityAIProviderAccount:
      description: |-
        +label=Perplexity AI Provider Account
        +icon=perplexity-ai
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/perplexity-ai
          type: string
          enum:
            - provider-account/perplexity-ai
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Perplexity AI provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/PerplexityAIKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the Perplexity
            AI provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/PerplexityIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: PerplexityAIProviderAccount
      x-fern-type-name: PerplexityAIProviderAccount
    TogetherAIProviderAccount:
      description: |-
        +label=Together AI Provider Account
        +icon=together-ai
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/together-ai
          type: string
          enum:
            - provider-account/together-ai
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Together AI provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/TogetherAIKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the Together AI
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/TogetherAIIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: TogetherAIProviderAccount
      x-fern-type-name: TogetherAIProviderAccount
    NomicProviderAccount:
      description: |-
        +label=Nomic Provider Account
        +icon=nomic
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/nomic
          type: string
          enum:
            - provider-account/nomic
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Nomic provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/NomicKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the Nomic
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/NomicIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: NomicProviderAccount
      x-fern-type-name: NomicProviderAccount
    PalmProviderAccount:
      description: |-
        +label=PaLM Provider Account
        +icon=palm
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/palm
          type: string
          enum:
            - provider-account/palm
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the PaLM provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/PalmKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the PaLM
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/PalmIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: PalmProviderAccount
      x-fern-type-name: PalmProviderAccount
    OllamaProviderAccount:
      description: |-
        +label=Ollama Provider Account
        +icon=ollama
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/ollama
          type: string
          enum:
            - provider-account/ollama
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Ollama provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/OllamaKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the Ollama
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/OllamaIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: OllamaProviderAccount
      x-fern-type-name: OllamaProviderAccount
    OpenaiProviderAccount:
      description: |-
        +label=OpenAI Provider Account
        +icon=openai
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/openai
          type: string
          enum:
            - provider-account/openai
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the OpenAI provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/OpenaiApiKeyAuth'
        base_url:
          description: |-
            +label=Base URL
            +sort=300
            +usage=Optional custom base URL for OpenAI API
            +message=Base URL must not be empty
          type: string
          pattern: ^\S+$
        integrations:
          description: >-
            +label=Integrations

            +sort=400

            +usage=List of integrations that are associated with the OpenAI
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/OpenAIIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=500
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: OpenaiProviderAccount
      x-fern-type-name: OpenaiProviderAccount
    DatabricksProviderAccount:
      description: |-
        +label=Databricks Provider Account
        +icon=databricks
      type: object
      required:
        - type
        - name
        - auth_data
        - base_url
        - integrations
      properties:
        type:
          description: +value=provider-account/databricks
          type: string
          enum:
            - provider-account/databricks
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Databricks provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          description: |-
            +label=Auth Data
            +sort=300
            +usage=Databricks authentication credentials
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/DatabricksServicePrincipalAuth'
            - $ref: '#/components/schemas/DatabricksApiKeyAuth'
          discriminator:
            propertyName: type
        base_url:
          description: |-
            +label=Base URL
            +sort=400
            +usage=The base URL of your Databricks workspace
            +message=Base URL must not be empty
          type: string
          pattern: ^\S+$
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the Databricks
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/DatabricksIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: DatabricksProviderAccount
      x-fern-type-name: DatabricksProviderAccount
    ClouderaProviderAccount:
      description: |-
        +label=Cloudera Provider Account
        +icon=cloudera
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/cloudera
          type: string
          enum:
            - provider-account/cloudera
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Cloudera provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          description: |-
            +label=Auth Data
            +sort=300
            +usage=Cloudera authentication credentials
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/ClouderaWorkloadAuthToken'
            - $ref: '#/components/schemas/ClouderaTokenAuth'
          discriminator:
            propertyName: type
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the Cloudera
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/ClouderaIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: ClouderaProviderAccount
      x-fern-type-name: ClouderaProviderAccount
    SelfHostedModelProviderAccount:
      description: |-
        +label=Self Hosted
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/self-hosted-model
          type: string
          enum:
            - provider-account/self-hosted-model
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the provider account.

            +sort=100

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        integrations:
          description: >-
            +label=Integrations

            +usage=List of integrations that are associated with the provider
            account.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/SelfHostedModelIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: SelfHostedModelProviderAccount
      x-fern-type-name: SelfHostedModelProviderAccount
    OpenRouterProviderAccount:
      description: |-
        +label=OpenRouter Provider Account
        +icon=openrouter
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/openrouter
          type: string
          enum:
            - provider-account/openrouter
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the OpenRouter provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/OpenRouterApiKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=400

            +usage=List of integrations that are associated with the OpenRouter
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/OpenRouterIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=500
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: OpenRouterProviderAccount
      x-fern-type-name: OpenRouterProviderAccount
    SambaNovaProviderAccount:
      description: |-
        +label=SambaNova Provider Account
        +icon=sambanova
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/sambanova
          type: string
          enum:
            - provider-account/sambanova
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the SambaNova provider account

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        auth_data:
          $ref: '#/components/schemas/SambaNovaKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the SambaNova
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/SambaNovaIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: SambaNovaProviderAccount
      x-fern-type-name: SambaNovaProviderAccount
    XAIProviderAccount:
      description: |-
        +label=xAI
        +icon=xai
      type: object
      required:
        - type
        - name
        - auth_data
      properties:
        type:
          description: +value=provider-account/xai
          type: string
          enum:
            - provider-account/xai
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the xAI provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/XAIKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=400

            +usage=List of integrations that are associated with the xAI
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/XAIIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: XAIProviderAccount
      x-fern-type-name: XAIProviderAccount
    BasetenProviderAccount:
      description: |-
        +label=Baseten
        +icon=baseten
      type: object
      required:
        - type
        - name
        - auth_data
      properties:
        type:
          description: +value=provider-account/baseten
          type: string
          enum:
            - provider-account/baseten
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Baseten provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/BasetenKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=400

            +usage=List of integrations that are associated with the Baseten
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/BasetenIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: BasetenProviderAccount
      x-fern-type-name: BasetenProviderAccount
    AwsSagemakerProviderAccount:
      description: |-
        +label=Sagemaker
        +icon=aws
      type: object
      required:
        - type
        - name
        - region
        - integrations
      properties:
        type:
          description: +value=provider-account/aws-sagemaker
          type: string
          enum:
            - provider-account/aws-sagemaker
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Sagemaker provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        region:
          $ref: '#/components/schemas/AWSRegion'
        auth_data:
          description: |-
            +label=AWS Account Auth Data
            +sort=400
            +usage=Authentication data for the AWS account
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/SagemakerAssumedRoleBasedAuth'
          discriminator:
            propertyName: type
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the Sagemaker
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/SagemakerModel'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=600
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: AwsSagemakerProviderAccount
      x-fern-type-name: AwsSagemakerProviderAccount
    CerebrasProviderAccount:
      description: |-
        +label=Cerebras Provider Account
        +icon=cerebras
      type: object
      required:
        - type
        - name
        - auth_data
      properties:
        type:
          description: +value=provider-account/cerebras
          type: string
          enum:
            - provider-account/cerebras
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Cerebras provider account

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        auth_data:
          $ref: '#/components/schemas/CerebrasKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=300

            +usage=List of integrations that are associated with the Cerebras
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/CerebrasIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=400
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: CerebrasProviderAccount
      x-fern-type-name: CerebrasProviderAccount
    SnowflakeCortexProviderAccount:
      description: |-
        +label=Snowflake Cortex Provider Account
        +icon=snowflake-cortex
      type: object
      required:
        - type
        - name
        - auth_data
        - account_identifier
        - integrations
      properties:
        type:
          description: +value=provider-account/snowflake-cortex
          type: string
          enum:
            - provider-account/snowflake-cortex
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Snowflake Cortex provider account

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        auth_data:
          $ref: '#/components/schemas/SnowflakeCortexPatTokenAuth'
        account_identifier:
          description: >-
            +label=Snowflake Account Identifier

            +sort=300

            +message=Snowflake account identifier must not be empty

            +usage=Snowflake account identifier. You can find it in the URL of
            your Snowflake account. Example:
            https://<account_identifier>.snowflakecomputing.com. Insert the
            account identifier only, without the URL.
          type: string
        integrations:
          description: >-
            +label=Integrations

            +sort=400

            +usage=List of integrations that are associated with the Snowflake
            Cortex provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/SnowflakeCortexIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=500
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: SnowflakeCortexProviderAccount
      x-fern-type-name: SnowflakeCortexProviderAccount
    ElevenLabsProviderAccount:
      description: |-
        +label=ElevenLabs Provider Account
        +icon=https://assets.production.truefoundry.com/elevenlabs.svg
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/elevenlabs
          type: string
          enum:
            - provider-account/elevenlabs
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the ElevenLabs provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/ElevenLabsApiKeyAuth'
        base_url:
          description: |-
            +label=Base URL
            +sort=350
            +usage=Optional custom base URL for ElevenLabs API
            +message=Base URL must not be empty
          type: string
          pattern: ^\S+$
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the ElevenLabs
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/ElevenLabsModel'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: ElevenLabsProviderAccount
      x-fern-type-name: ElevenLabsProviderAccount
    DeepgramProviderAccount:
      description: |-
        +label=Deepgram Provider Account
        +icon=https://assets.production.truefoundry.com/deepgram.svg
      type: object
      required:
        - type
        - name
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/deepgram
          type: string
          enum:
            - provider-account/deepgram
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Deepgram provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/DeepgramApiKeyAuth'
        base_url:
          description: |-
            +label=Base URL
            +sort=450
            +usage=Optional custom base URL for Deepgram API
            +message=Base URL must not be empty
          type: string
          pattern: ^\S+$
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the Deepgram
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/DeepgramModel'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: DeepgramProviderAccount
      x-fern-type-name: DeepgramProviderAccount
    CartesiaProviderAccount:
      description: |-
        +label=Cartesia Provider Account
        +icon=https://assets.production.truefoundry.com/cartesia.svg
      type: object
      required:
        - type
        - name
        - cartesia_version
        - auth_data
        - integrations
      properties:
        type:
          description: +value=provider-account/cartesia
          type: string
          enum:
            - provider-account/cartesia
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Cartesia provider account

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        base_url:
          description: |-
            +label=Base URL
            +sort=250
            +usage=Optional custom base URL for Cartesia API
            +message=Base URL must not be empty
          type: string
          pattern: ^\S+$
        cartesia_version:
          description: >-
            +label=Cartesia Version

            +sort=300

            +usage=The Cartesia API version header. Must be set to a supported
            API version.

            +uiType=Select
          type: string
          enum:
            - '2025-04-16'
            - '2024-11-13'
            - '2024-06-10'
          default: '2025-04-16'
        auth_data:
          $ref: '#/components/schemas/CartesiaApiKeyAuth'
        integrations:
          description: >-
            +label=Integrations

            +sort=500

            +usage=List of integrations that are associated with the Cartesia
            provider account

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/CartesiaModel'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +usage=List of users who have access to this provider account
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: CartesiaProviderAccount
      x-fern-type-name: CartesiaProviderAccount
    VirtualModelProviderAccount:
      description: +label=Virtual Model Provider Group
      type: object
      required:
        - name
        - type
        - integrations
      properties:
        name:
          description: >-
            +label=Name

            +sort=100

            +usage=The name of the Virtual Model Provider Group

            +message=3 to 64 characters long alphanumeric word, may contain - or
            _ in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-_]{1,62}[a-z0-9]$
        type:
          description: +value=provider-account/virtual-model
          type: string
          enum:
            - provider-account/virtual-model
        integrations:
          description: |-
            +label=Virtual Models
            +sort=200
            +usage=List of virtual models in this group
            +uiType=Structs
          type: array
          items:
            $ref: '#/components/schemas/VirtualModel'
        collaborators:
          description: >-
            +label=Collaborators

            +sort=150

            +usage=List of users who have access to this Virtual Model Provider
            Group

            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: VirtualModelProviderAccount
      x-fern-type-name: VirtualModelProviderAccount
    CustomEndpointProviderAccount:
      description: |-
        +label=Custom Endpoint
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - integrations
      properties:
        type:
          description: +value=provider-account/custom-endpoint
          type: string
          enum:
            - provider-account/custom-endpoint
        name:
          description: >-
            +uiProps={"disableEdit":true}

            +label=Name

            +usage=The name of the custom endpoint provider account.

            +sort=100

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        endpoint_type:
          description: >-
            +label=Endpoint Type

            +usage=The type of service behind this endpoint (used for tracking
            purposes)

            +sort=250
          type: string
          enum:
            - azure-speech-service
            - other
        auth_data:
          $ref: '#/components/schemas/CustomHeaderAuth'
        integrations:
          description: >-
            +label=Integrations

            +usage=List of endpoint integrations associated with this provider
            account.

            +sort=400

            +uiType=IntegrationsGroup
          type: array
          items:
            $ref: '#/components/schemas/CustomEndpointIntegrations'
        collaborators:
          description: |-
            +label=Collaborators
            +sort=200
            +uiType=Collaborators
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: CustomEndpointProviderAccount
      x-fern-type-name: CustomEndpointProviderAccount
    MCPServerIntegrations:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/MCPServerIntegration'
        - $ref: '#/components/schemas/VirtualMCPServerIntegration'
      discriminator:
        propertyName: type
      title: MCPServerIntegrations
      x-fern-type-name: MCPServerIntegrations
    GuardrailConfigIntegrations:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/OpenAIModerationsGuardrailConfig'
        - $ref: '#/components/schemas/AWSBedrockGuardrailConfig'
        - $ref: '#/components/schemas/CustomGuardrailConfig'
        - $ref: '#/components/schemas/AzurePIIGuardrailConfig'
        - $ref: '#/components/schemas/AzureContentSafetyGuardrailConfig'
        - $ref: '#/components/schemas/AzurePromptShieldGuardrailConfig'
        - $ref: '#/components/schemas/EnkryptAIGuardrailConfig'
        - $ref: '#/components/schemas/PaloAltoPrismaAIRSGuardrailConfig'
        - $ref: '#/components/schemas/FiddlerGuardrailConfig'
        - $ref: '#/components/schemas/PangeaGuardrailConfig'
        - $ref: '#/components/schemas/PatronusGuardrailConfig'
        - $ref: '#/components/schemas/SecretDetectionGuardrailConfig'
        - $ref: '#/components/schemas/CodeSafetyLinterGuardrailConfig'
        - $ref: '#/components/schemas/SQLSanitizerGuardrailConfig'
        - $ref: '#/components/schemas/RegexGuardrailConfig'
        - $ref: '#/components/schemas/TfyContentModerationGuardrailConfig'
        - $ref: '#/components/schemas/TfyPIIGuardrailConfig'
        - $ref: '#/components/schemas/TfyPromptInjectionGuardrailConfig'
        - $ref: '#/components/schemas/CedarGuardrailConfig'
        - $ref: '#/components/schemas/OPAGuardrailConfig'
        - $ref: '#/components/schemas/GoogleModelArmorGuardrailConfig'
        - $ref: '#/components/schemas/GraySwanCygnalGuardrailConfig'
        - $ref: '#/components/schemas/AktoGuardrailConfig'
        - $ref: '#/components/schemas/TrojAIGuardrailConfig'
      discriminator:
        propertyName: type
      title: GuardrailConfigIntegrations
      x-fern-type-name: GuardrailConfigIntegrations
    GitSource:
      description: >-
        +docs=Describes that we are using code stored in a git repository to
        build our image

        +label=Git Source

        +icon=fa-solid fa-code-branch:black

        +sort=300
      type: object
      required:
        - type
        - repo_url
        - ref
      properties:
        type:
          description: +value=git
          type: string
          enum:
            - git
        repo_url:
          description: >-
            +label=Repo URL

            +usage=The repository URL.

            +sort=1

            +message=Needs to be a valid Github, Bitbucket, Azure Repos or
            Gitlab link
          type: string
          pattern: ^(https?://)\S+$
        ref:
          description: |-
            +label=Commit SHA
            +usage=The commit SHA.
            +sort=2
          type: string
          pattern: ^\S+$
        branch_name:
          description: |-
            +label=Branch Name
            +usage=Selecting branch will select latest commit SHA of the branch.
            +sort=3
          type: string
          pattern: ^\S+$
      title: GitSource
      x-fern-type-name: GitSource
    DockerFileBuild:
      description: |-
        +docs=Describes that we are using a dockerfile to build our image
        +label=Docker File (I already have Docker File)
        +icon=fa-brands fa-docker:#0db7ed
      type: object
      required:
        - type
        - dockerfile_path
        - build_context_path
      properties:
        type:
          description: +value=dockerfile
          type: string
          enum:
            - dockerfile
        dockerfile_path:
          description: >-
            +label=Path to Dockerfile

            +usage=The file path of the Dockerfile relative to project root
            path.
          type: string
          default: ./Dockerfile
        build_context_path:
          description: >-
            +label=Path to build context

            +usage=Build context path for the Dockerfile relative to project
            root path.
          type: string
          default: ./
        command:
          description: >-
            +label=Command Override

            +usage=Override the command to run when the container starts

            When deploying a Job, the command can be templatized by defining
            `params` and referencing them in command

            E.g. `python main.py --learning_rate {{learning_rate}}`
          oneOf:
            - type: string
            - type: array
              items:
                type: string
        build_args:
          description: +label=Build arguments to pass to docker build
          type: object
          additionalProperties:
            type: string
        build_secrets:
          description: >-
            +label=Build secrets to pass to docker build

            +usage=Build secrets you can use in run statements inside dockerfile
            like RUN --mount=type=secret,id=pip_index_url ... you can learn more
            about build secrets here
            https://www.truefoundry.com/docs/docker-build-secrets

            +uiProps={"showTableView": true, "widthRatios": {"id":0.4286}}
          type: array
          items:
            $ref: '#/components/schemas/BuildSecret'
      title: DockerFileBuild
      x-fern-type-name: DockerFileBuild
    PythonBuild:
      description: >-
        +docs=Describes that we are using python to build a container image with
        a specific python version and pip packages installed.

        +label=Python Code (I don't have Dockerfile)

        +icon=fa-brands fa-python:#306998
      type: object
      required:
        - type
        - build_context_path
        - command
      properties:
        type:
          description: +value=tfy-python-buildpack
          type: string
          enum:
            - tfy-python-buildpack
        python_version:
          description: >-
            +label=Python version

            +usage=Python version to run your application. Should be one of the
            tags listed on [Official Python Docker
            Page](https://hub.docker.com/_/python)

            +message=Please enter a valid Python version tag

            +sort=10002
          type: string
          pattern: ^\d+(\.\d+){1,2}([\-\.a-z0-9]+)?$
        build_context_path:
          description: |-
            +label=Path to build context
            +usage=Build path relative to project root path.
            +sort=10000
          type: string
          default: ./
        requirements_path:
          description: |-
            `Path to build context`
            +label=Path to requirements
            +usage=Path to `requirements.txt` relative to
            `Path to build context`
            +uiType=Hidden
            +sort=10004
          type: string
        pip_packages:
          description: |-
            +label=Pip packages to install
            +usage=Define pip package requirements.
            In Python/YAML E.g. ["fastapi>=0.90,<1.0", "uvicorn"]
            +placeholder=Enter a pip package name E.g. fastapi>=0.90,<1.0
            +uiType=Hidden
          type: array
          items:
            type: string
        python_dependencies:
          description: |-
            +label=Python dependencies
            +usage=Python dependencies to install
            +sort=10004
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/Pip'
            - $ref: '#/components/schemas/UV'
            - $ref: '#/components/schemas/Poetry'
          discriminator:
            propertyName: type
        apt_packages:
          description: |-
            +label=List of Debian packages to install.
            +usage=Debian packages to install via `apt get`.
            In Python/YAML E.g. ["git", "ffmpeg", "htop"]
            +placeholder=Enter a debian package name E.g. ffmpeg
          type: array
          items:
            type: string
        command:
          description: >-
            Command will be set as the Entrypoint of the generated

            image.

            +label=Command

            +usage=Command to run when the container starts.

            Command will be set as the Entrypoint of the generated image.

            When deploying a Job, the command can be templatized by defining
            `params` and referencing them in command

            E.g. `python main.py --learning_rate {{learning_rate}}`

            +sort=10001
          oneOf:
            - type: string
            - type: array
              items:
                type: string
        cuda_version:
          description: >-
            +label=CUDA Version

            +usage=Version of CUDA Toolkit and CUDNN to install in the image

            These combinations are based off of publically available docker
            images on docker hub

            You can also specify a valid tag of the form
            {cuda_version_number}-cudnn{cudnn_version_number}-{runtime|devel}-ubuntu{ubuntu_version}

            Refer https://hub.docker.com/r/nvidia/cuda/tags for valid set of
            values

            Note: We use deadsnakes ubuntu ppa to add Python that currently
            supports only Ubuntu 18.04, 20.04 and 22.04

            +sort=10003
          type: string
          pattern: >-
            ^((\d+\.\d+(\.\d+)?-cudnn(\d+)?-(runtime|devel)-ubuntu\d+\.\d+)|11\.0-cudnn8|11\.1-cudnn8|11\.2-cudnn8|11\.3-cudnn8|11\.4-cudnn8|11\.5-cudnn8|11\.6-cudnn8|11\.7-cudnn8|11\.8-cudnn8|12\.0-cudnn8|12\.1-cudnn8|12\.2-cudnn8|12\.3-cudnn9|12\.4-cudnn9|12\.5-cudnn9|12\.6-cudnn9|12\.8-cudnn9|12\.9-cudnn9)$
      title: PythonBuild
      x-fern-type-name: PythonBuild
    NodeSelector:
      description: >-
        +label=Node selector

        +usage=Constraints to select a Node - Specific GPU / Instance Families,
        On-Demand/Spot.
      type: object
      required:
        - type
      properties:
        type:
          description: +value=node_selector
          type: string
          enum:
            - node_selector
        instance_families:
          description: >-
            +label=Instance family

            +usage=Instance family of the underlying machine to use. Multiple
            instance families can be supplied.

            The workload is guaranteed to be scheduled on one of them.
          type: array
          items:
            type: string
        capacity_type:
          description: >-
            +label=Capacity Type

            +usage=Configure what type of nodes to run the app. By default no
            placement logic is applied.

            "spot_fallback_on_demand" will try to place the application on spot
            nodes but will fallback to on-demand when spot nodes are not
            available.

            "spot" will strictly place the application on spot nodes.

            "on_demand" will strictly place the application on on-demand nodes.
          type: string
          enum:
            - spot_fallback_on_demand
            - spot
            - on_demand
      title: NodeSelector
      x-fern-type-name: NodeSelector
    NodepoolSelector:
      description: |-
        +label=Nodepool selector
        +usage=Specify one or more nodepools to run your application on.
      type: object
      required:
        - type
      properties:
        type:
          description: +value=nodepool_selector
          type: string
          enum:
            - nodepool_selector
        nodepools:
          description: >-
            +label=Nodepools

            +usage=Nodepools where you want to run your workload. Multiple
            nodepools can be selected.
             The workload is guaranteed to be scheduled on one of the nodepool
          type: array
          items:
            type: string
      title: NodepoolSelector
      x-fern-type-name: NodepoolSelector
    NvidiaGPU:
      type: object
      required:
        - type
        - count
      properties:
        type:
          description: +value=nvidia_gpu
          type: string
          enum:
            - nvidia_gpu
        name:
          description: >-
            +label=GPU Name

            +usage=Name of the Nvidia GPU. One of [P4, P100, V100, T4, A10G,
            A100_40GB, A100_80GB]

            This field is required for Node Selector and can be ignored in
            Nodepool Selector.

            One instance of the card contains the following amount of memory -

            P4: 8 GB, P100: 16 GB, V100: 16 GB, T4: 16 GB, A10G: 24 GB,
            A100_40GB: 40GB, A100_80GB: 80 GB
          type: string
        count:
          description: >-
            +label=GPU Count

            +usage=Count of GPUs to provide to the application

            Note the exact count and max count available for a given GPU type
            depends on cloud provider and cluster type.
          type: integer
          minimum: 1
          maximum: 16
      title: NvidiaGPU
      x-fern-type-name: NvidiaGPU
    AWSInferentia:
      type: object
      required:
        - type
        - count
      properties:
        type:
          description: +value=aws_inferentia
          type: string
          enum:
            - aws_inferentia
        name:
          description: >-
            +label=Inferentia accelerator name

            +usage=Name of the AWS Inferentia Accccelerator. One of [INF1,
            INF2].

            This field is required for Node Selector and can be ignored in
            Nodepool Selector.
          type: string
        count:
          description: >-
            +label=Count

            +usage=Count of Inferentia accelerator chips to provide to the
            application
          type: integer
          minimum: 1
          maximum: 16
      title: AWSInferentia
      x-fern-type-name: AWSInferentia
    NvidiaMIGGPU:
      type: object
      required:
        - type
        - profile
      properties:
        type:
          description: +value=nvidia_mig_gpu
          type: string
          enum:
            - nvidia_mig_gpu
        name:
          description: >-
            +label=GPU Name

            +usage=Name of the Nvidia GPU. One of [P4, P100, V100, T4, A10G,
            A100_40GB, A100_80GB]

            This field is required for Node Selector and can be ignored in
            Nodepool Selector.

            One instance of the card contains the following amount of memory -

            P4: 8 GB, P100: 16 GB, V100: 16 GB, T4: 16 GB, A10G: 24 GB,
            A100_40GB: 40GB, A100_80GB: 80 GB
          type: string
        profile:
          description: >-
            +label=MIG Profile

            +usage=Name of the MIG profile to use. One of the following based on
            gpu type

            Please refer to
            https://docs.nvidia.com/datacenter/tesla/mig-user-guide/#supported-mig-profiles
            for more details

            A30 - [1g.6gb, 2g.12gb, 4g.24gb]

            A100 40 GB - [1g.5gb, 1g.10gb, 2g.10gb, 3g.20gb, 4g.20gb, 7g.40gb]

            A100 80 GB / H100 80 GB - [1g.10gb, 1g.20gb, 2g.20gb, 3g.40gb,
            4g.40gb, 7g.80gb]

            H100 94 GB - [1g.12gb, 1g.24gb, 2g.24gb, 3g.47gb, 4g.47gb, 7g.94gb]

            H100 96 GB - [1g.12gb, 1g.24gb, 2g.24gb, 3g.48gb, 4g.48gb, 7g.96gb]

            H200 141 GB - [1g.18gb, 1g.35gb, 2g.35gb, 3g.71gb, 4g.71gb]

            B200 180 GB - [1g.23gb, 1g.45gb, 2g.45gb, 3g.90gb, 4g.90gb,
            7g.180gb]
          type: string
          enum:
            - 1g.6gb
            - 2g.12gb
            - 1g.5gb
            - 1g.10gb
            - 2g.10gb
            - 3g.20gb
            - 4g.20gb
            - 1g.20gb
            - 2g.20gb
            - 3g.40gb
            - 4g.40gb
            - 1g.12gb
            - 1g.24gb
            - 2g.24gb
            - 3g.47gb
            - 4g.47gb
            - 3g.48gb
            - 4g.48gb
            - 1g.18gb
            - 1g.35gb
            - 2g.35gb
            - 3g.71gb
            - 4g.71gb
            - 1g.23gb
            - 1g.45gb
            - 2g.45gb
            - 3g.90gb
            - 4g.90gb
            - 4g.24gb
            - 7g.40gb
            - 7g.80gb
            - 7g.94gb
            - 7g.96gb
            - 7g.180gb
      title: NvidiaMIGGPU
      x-fern-type-name: NvidiaMIGGPU
    NvidiaTimeslicingGPU:
      type: object
      required:
        - type
        - gpu_memory
      properties:
        type:
          description: +value=nvidia_timeslicing_gpu
          type: string
          enum:
            - nvidia_timeslicing_gpu
        name:
          description: >-
            +label=GPU Name

            +usage=Name of the Nvidia GPU. One of [P4, P100, V100, T4, A10G,
            A100_40GB, A100_80GB]

            This field is required for Node Selector and can be ignored in
            Nodepool Selector.

            One instance of the card contains the following amount of memory -

            P4: 8 GB, P100: 16 GB, V100: 16 GB, T4: 16 GB, A10G: 24 GB,
            A100_40GB: 40GB, A100_80GB: 80 GB
          type: string
        gpu_memory:
          description: >-
            +label=GPU Memory (MB)

            +usage=Amount of GPU memory (in MB) to allocate. Please note, this
            limit is not being enforced today but will be in future.
            Applications are expected to operate in co-opertative mode
          type: integer
          minimum: 1
          maximum: 400000
      title: NvidiaTimeslicingGPU
      x-fern-type-name: NvidiaTimeslicingGPU
    GcpTPU:
      type: object
      required:
        - type
        - name
        - topology
      properties:
        type:
          description: +value=gcp_tpu
          type: string
          enum:
            - gcp_tpu
        name:
          description: |-
            +label=TPU Type name
            +usage=Name of the TPU Type. One of
              - `tpu-v4-podslice` (TPU v4, ct4p)
              - `tpu-v5-lite-device` (TPU v5e, ct5l)
              - `tpu-v5-lite-podslice`  (TPU v5e, ct5lp)
              - `tpu-v5p-slice` (TPU v5p, ct5p)
          type: string
          pattern: ^tpu-[a-z\d\-]+$
        topology:
          description: >-
            +label=Slice Topology

            +usage=Topology of the TPU slices. Currently only single-host
            topology is supported.
             Please refer to [TPUs on GKE docs](https://cloud.google.com/kubernetes-engine/docs/concepts/tpus#plan-tpu-configuration)
             Allowed Values:
               - `2x2x1` for `tpu-v4-podslice`
               - One of `1x1`, `2x2`, `2x4` for `tpu-v5-lite-device` and `tpu-v5-lite-podslice`
               - `2x2x1` for `tpu-v5p-slice`
          type: string
          pattern: ^\d+x\d+(x\d+)?$
      title: GcpTPU
      x-fern-type-name: GcpTPU
    Endpoint:
      type: object
      required:
        - host
      properties:
        host:
          description: >-
            +usage=Host e.g. ai.example.com, app.truefoundry.com

            +message=Upto 253 characters, each part of host should be at most 63
            characters long, can contain alphabets, digits and hypen, must begin
            and end with an alphanumeric characters. Parts must be separated by
            periods (.)
          type: string
          pattern: >-
            ^((([a-zA-Z0-9\-]{1,63}\.)([a-zA-Z0-9\-]{1,63}\.)*([A-Za-z]{1,63}))|(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)))$
        path:
          description: >-
            +usage=Path e.g. /v1/api/ml/, /v2/docs/

            +message=Should begin and end with a forward slash (/). Each part
            can can contain alphabets, digits and hypen, must begin and end with
            an alphanumeric characters. Parts should be separated by forward
            slashes (/)
          type: string
          pattern: ^(/([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-_\.]*[a-zA-Z0-9]))*/$
      title: Endpoint
      x-fern-type-name: Endpoint
    FlyteTaskTemplate:
      type: object
      required:
        - id
        - custom
      properties:
        id:
          $ref: '#/components/schemas/FlyteTaskID'
        custom:
          $ref: '#/components/schemas/FlyteTaskCustom'
      title: FlyteTaskTemplate
      x-fern-type-name: FlyteTaskTemplate
    FlyteWorkflowTemplate:
      type: object
      required:
        - id
      properties:
        id:
          $ref: '#/components/schemas/FlyteWorkflowID'
      title: FlyteWorkflowTemplate
      x-fern-type-name: FlyteWorkflowTemplate
    FlyteLaunchPlanID:
      type: object
      required:
        - resourceType
        - name
      properties:
        resourceType:
          type: string
          enum:
            - LAUNCH_PLAN
        name:
          type: string
      title: FlyteLaunchPlanID
      x-fern-type-name: FlyteLaunchPlanID
    FlyteLaunchPlanSpec:
      type: object
      required:
        - workflowId
      properties:
        workflowId:
          $ref: '#/components/schemas/FlyteWorkflowID'
      title: FlyteLaunchPlanSpec
      x-fern-type-name: FlyteLaunchPlanSpec
    SQSQueueMetricConfig:
      type: object
      required:
        - type
        - queue_length
      properties:
        type:
          description: +value=sqs
          type: string
          enum:
            - sqs
        queue_length:
          description: >-
            +label=Queue length

            +usage=Upper limit of the number of backlog messages the auto-scaler
            will try to maintain per replica. If you set this number to 10 and
            have 30 messages in the queue and one replica, the auto-scaler will
            scale the number of replicas to 3.
          type: integer
          minimum: 1
        auth:
          description: |-
            +label=Authentication
            +usage=Receive permissions on the SQS Queue
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsSqsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/AwsSqsAssumedRoleBasedAuth'
          discriminator:
            propertyName: type
      title: SQSQueueMetricConfig
      x-fern-type-name: SQSQueueMetricConfig
    NATSMetricConfig:
      type: object
      required:
        - type
        - lag_threshold
      properties:
        type:
          description: +value=nats
          type: string
          enum:
            - nats
        lag_threshold:
          description: >-
            +label=Lag Threshold

            +usage=Upper limit of the number of backlog messages the auto-scaler
            will try to maintain per replica. If you set this number to 10 and
            have 30 messages in the stream and one replica, the auto-scaler will
            scale the number of replicas to 3.
          type: integer
          minimum: 1
      title: NATSMetricConfig
      x-fern-type-name: NATSMetricConfig
    KafkaMetricConfig:
      type: object
      required:
        - type
        - lag_threshold
      properties:
        type:
          description: +value=kafka
          type: string
          enum:
            - kafka
        lag_threshold:
          description: >-
            +label=Lag Threshold

            +usage=Upper limit of the number of backlog messages the auto-scaler
            will try to maintain per replica. If you set this number to 10 and
            have 30 messages in the stream and one replica, the auto-scaler will
            scale the number of replicas to 3.
          type: integer
          minimum: 1
      title: KafkaMetricConfig
      x-fern-type-name: KafkaMetricConfig
    AMQPMetricConfig:
      type: object
      required:
        - type
        - queue_length
      properties:
        type:
          description: +value=amqp
          type: string
          enum:
            - amqp
        queue_length:
          description: >-
            +label=Queue Length

            +usage=Upper limit of the number of backlog messages the auto-scaler
            will try to maintain per replica. If you set this number to 10 and
            have 30 messages in the stream and one replica, the auto-scaler will
            scale the number of replicas to 3.
          type: integer
          minimum: 1
      title: AMQPMetricConfig
      x-fern-type-name: AMQPMetricConfig
    SQSInputConfig:
      description: |-
        +docs=Describes the configuration for the input SQS worker
        +label=SQS
      type: object
      required:
        - type
        - queue_url
        - region_name
        - visibility_timeout
        - wait_time_seconds
      properties:
        type:
          description: +value=sqs
          type: string
          enum:
            - sqs
        queue_url:
          description: |-
            +label=Queue URL
            +usage=AWS SQS Queue URL of Subscriber
            +sort=1
          type: string
        region_name:
          description: |-
            +label=Region Name
            +usage=AWS Region Name
            +sort=2
          type: string
        visibility_timeout:
          description: >-
            +label=Visibility Timeout (seconds)

            +usage=A period during which Amazon SQS prevents all consumers from
            receiving and processing the message. If one message takes 5 seconds
            to process, you can set this number to 7 or any number higher than
            5. This will ensure that while the message is being processed, it
            will not be available to other replicas. For more information, see
            [here](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-visibility-timeout.html)

            +sort=3
          type: integer
          minimum: 1
          maximum: 43200
        wait_time_seconds:
          description: >-
            +label=Wait Time Seconds

            +usage=Wait timeout for long polling. For more information, see
            [here](https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-short-and-long-polling.html)
          type: integer
          minimum: 1
          maximum: 20
          default: 19
        auth:
          $ref: '#/components/schemas/AWSAccessKeyAuth'
      title: SQSInputConfig
      x-fern-type-name: SQSInputConfig
    NATSInputConfig:
      description: |-
        +docs=Describes the configuration for the input NATS worker
        +label=NATS
      type: object
      required:
        - type
        - nats_url
        - stream_name
        - root_subject
        - consumer_name
        - wait_time_seconds
      properties:
        type:
          description: +value=nats
          type: string
          enum:
            - nats
        nats_url:
          description: |-
            +label=NATS URL
            +usage=Input NATS URL
            +sort=1
          type: string
        stream_name:
          description: |-
            +label=Stream Name
            +usage=Name of the NATS stream
            +sort=2
          type: string
        root_subject:
          description: >-
            +label=Root Subject

            +usage=Root subject of input NATS

            +message=Input NATS root subject should only contain alphanumeric
            letters, dashes(-), and periods(.)

            +sort=3
          type: string
          pattern: ^[a-zA-Z0-9][a-zA-Z0-9\-.]+[a-zA-Z0-9]$
        consumer_name:
          description: >-
            +label=Consumer Name

            +usage=Consumer name of input NATS

            +message=Consumer name should only contain alphanumeric letters,
            dashes(-), and underscores(_)

            +sort=4
          type: string
          pattern: ^[a-zA-Z0-9][a-zA-Z0-9\-_]+[a-zA-Z0-9]$
        wait_time_seconds:
          description: |-
            +label=Wait Time Seconds
            +usage=Wait timeout for long polling.
            +sort=5
          type: integer
          minimum: 1
          maximum: 20
          default: 19
        nats_metrics_url:
          description: >-
            +label=NATS metrics URL

            +usage=URL for the NATS metrics endpoint. It is compulsory if you
            want to use NATS autoscaling.

            +message=NATS Metrics URL should be a valid HTTP/HTTPS URL
          type: string
          pattern: ^(http(s?)://).*$
        auth:
          $ref: '#/components/schemas/NATSUserPasswordAuth'
      title: NATSInputConfig
      x-fern-type-name: NATSInputConfig
    KafkaInputConfig:
      description: |-
        +docs=Describes the configuration for the input Kafka worker
        +label=Kafka
      type: object
      required:
        - type
        - bootstrap_servers
        - topic_name
        - consumer_group
        - tls
      properties:
        type:
          description: +value=kafka
          type: string
          enum:
            - kafka
        bootstrap_servers:
          description: >-
            +label=Bootstrap servers

            +usage='Kafka Bootstrap servers - Comma separated list of Kafka
            brokers "hostname:port" to connect to for bootstrap'

            +sort=1
          type: string
        topic_name:
          description: |-
            +label=Topic Name
            +usage=Kafka topic to subscribe to
            +sort=2
          type: string
        consumer_group:
          description: >-
            +label=Consumer Group Name

            +usage=The name of the consumer group to join for dynamic partition
            assignment

            +sort=3
          type: string
        tls:
          description: |-
            +label=TLS
            +usage=TLS configuration for SASL authentication
          type: boolean
          default: true
        wait_time_seconds:
          description: |-
            +label=Wait Time Seconds
            +usage=Wait timeout for long polling.
          type: integer
          minimum: 1
          maximum: 300
          default: 10
        auth:
          $ref: '#/components/schemas/KafkaSASLAuth'
      title: KafkaInputConfig
      x-fern-type-name: KafkaInputConfig
    AMQPInputConfig:
      description: |-
        +docs=Describes the configuration for the input AMQP worker
        +label=AMQP
      type: object
      required:
        - type
        - url
        - queue_name
        - wait_time_seconds
      properties:
        type:
          description: +value=amqp
          type: string
          enum:
            - amqp
        url:
          description: |-
            +label=Queue URL
            +usage=AMQP Queue URL of Subscriber
            +sort=1
          type: string
          pattern: >-
            ^(amqp|amqps?)://(?:[^:@]+(?::[^:@]+)?@)?([^/?]+)(?:/([^?]+))?/?([^?]+)?(?:\?(.*))?$
        queue_name:
          description: |-
            +label=Queue Name
            +usage=AMQP Queue Name
            +sort=2
          type: string
        wait_time_seconds:
          description: |-
            +label=Wait Time Seconds
            +usage=Wait timeout for long polling.
          type: integer
          minimum: 1
          default: 5
      title: AMQPInputConfig
      x-fern-type-name: AMQPInputConfig
    SQSOutputConfig:
      description: |-
        +docs=Describes the configuration for the output SQS worker
        +label=SQS
      type: object
      required:
        - type
        - queue_url
        - region_name
      properties:
        type:
          description: +value=sqs
          type: string
          enum:
            - sqs
        queue_url:
          description: |-
            +label=Queue URL
            +usage=AWS SQS Queue URL of Publisher
            +sort=1
          type: string
        region_name:
          description: |-
            +label=Region Name
            +usage=AWS Region Name
            +sort=2
          type: string
        auth:
          $ref: '#/components/schemas/AWSAccessKeyAuth'
      title: SQSOutputConfig
      x-fern-type-name: SQSOutputConfig
    NATSOutputConfig:
      description: |-
        +docs=Describes the configuration for the output NATS worker
        +label=NATS
      type: object
      required:
        - type
        - nats_url
        - root_subject
      properties:
        type:
          description: +value=nats
          type: string
          enum:
            - nats
        nats_url:
          description: |-
            +label=NATS URL
            +usage=Output NATS URL
            +sort=1
          type: string
        root_subject:
          description: >-
            +label=Root Subject

            +usage=Root subject of output NATS

            +message=Output NATS root subject should only contain alphanumeric
            letters, dashes(-), and periods(.)

            +sort=2
          type: string
          pattern: ^[a-zA-Z0-9][a-zA-Z0-9\-.]+[a-zA-Z0-9]$
        auth:
          $ref: '#/components/schemas/NATSUserPasswordAuth'
      title: NATSOutputConfig
      x-fern-type-name: NATSOutputConfig
    CoreNATSOutputConfig:
      description: |-
        +docs=Describes the configuration for the output Core NATS worker
        +label=Core NATS
      type: object
      required:
        - type
        - nats_url
        - root_subject
      properties:
        type:
          description: +value=core-nats
          type: string
          enum:
            - core-nats
        nats_url:
          description: |-
            +label=NATS URL
            +usage=Output NATS URL
            +sort=1
          type: string
        root_subject:
          description: >-
            +label=Root Subject

            +usage=Root subject of output NATS

            +message=Output NATS root subject should only contain alphanumeric
            letters, dashes(-), and periods(.)

            +sort=2
          type: string
          pattern: ^[a-zA-Z0-9][a-zA-Z0-9\-.]+[a-zA-Z0-9]$
        auth:
          $ref: '#/components/schemas/NATSUserPasswordAuth'
      title: CoreNATSOutputConfig
      x-fern-type-name: CoreNATSOutputConfig
    KafkaOutputConfig:
      description: |-
        +docs=Describes the configuration for the output Kafka worker
        +label=Kafka
      type: object
      required:
        - type
        - bootstrap_servers
        - topic_name
        - tls
      properties:
        type:
          description: +value=kafka
          type: string
          enum:
            - kafka
        bootstrap_servers:
          description: >-
            +label=Bootstrap servers

            +usage='Kafka Bootstrap servers - Comma separated list of Kafka
            brokers "hostname:port" to connect to for bootstrap'

            +sort=1
          type: string
        topic_name:
          description: |-
            +label=Topic Name
            +usage=Kafka topic to publish to
            +sort=2
          type: string
        tls:
          description: |-
            +label=TLS
            +usage=TLS configuration for SASL authentication
          type: boolean
          default: true
        auth:
          $ref: '#/components/schemas/KafkaSASLAuth'
      title: KafkaOutputConfig
      x-fern-type-name: KafkaOutputConfig
    AMQPOutputConfig:
      description: |-
        +docs=Describes the configuration for the output AMQP worker
        +label=AMQP
      type: object
      required:
        - type
        - url
        - routing_key
      properties:
        type:
          description: +value=amqp
          type: string
          enum:
            - amqp
        url:
          description: |-
            +label=Queue URL
            +usage=AMQP Queue URL of Publisher
            +sort=1
          type: string
          pattern: >-
            ^(amqp|amqps?)://(?:[^:@]+(?::[^:@]+)?@)?([^/?]+)(?:/([^?]+))?/?([^?]+)?(?:\?(.*))?$
        routing_key:
          description: |-
            +label=Routing Key
            +usage=AMQP Routing Key to publish to.
            +sort=2
          type: string
        exchange_name:
          description: |-
            +label=Exchange Name
            +usage=AMQP Exchange Name
            +sort=3
          type: string
      title: AMQPOutputConfig
      x-fern-type-name: AMQPOutputConfig
    SparkBuild:
      description: >-
        +docs=Describes that we are using python to build a container image with
        a specific python version and pip packages installed.

        +label=Build specification

        +icon=fa-brands fa-python:#306998
      type: object
      required:
        - type
        - spark_version
        - build_context_path
      properties:
        type:
          description: +value=tfy-spark-buildpack
          type: string
          enum:
            - tfy-spark-buildpack
        spark_version:
          description: >-
            --- Spark Specific Field ---

            +label=Spark Version

            +usage=Spark version should match the spark version installed in the
            image.

            +sort=1001
          type: string
          default: 3.5.2
        build_context_path:
          description: |-
            +label=Path to build context
            +usage=Build path relative to project root path.
          type: string
          default: ./
        requirements_path:
          description: |-
            `Path to build context`
            +label=Path to requirements
            +usage=Path to `requirements.txt` relative to
            `Path to build context`
          type: string
      title: SparkBuild
      x-fern-type-name: SparkBuild
    SparkExecutorFixedInstances:
      description: +label=Fixed Instances
      type: object
      required:
        - type
        - count
      properties:
        type:
          description: +value=fixed
          type: string
          enum:
            - fixed
        count:
          description: |-
            +label=Instances Count
            +usage=Number of instances to start
          type: integer
          minimum: 0
          maximum: 500
          default: 1
      title: SparkExecutorFixedInstances
      x-fern-type-name: SparkExecutorFixedInstances
    SparkExecutorDynamicScaling:
      description: +label=Dynamic Scaling
      type: object
      required:
        - type
        - min
        - max
      properties:
        type:
          description: +value=dynamic
          type: string
          enum:
            - dynamic
        min:
          description: |-
            +label=Min Instances
            +usage=Minimum number of instances to start / scale down to
            +sort=100
          type: integer
          minimum: 0
          maximum: 500
          default: 1
        max:
          description: |-
            +label=Max Instances
            +usage=Maximum number of instances to scale up to
            +sort=200
          type: integer
          minimum: 0
          maximum: 500
          default: 1
      title: SparkExecutorDynamicScaling
      x-fern-type-name: SparkExecutorDynamicScaling
    RateLimitRule:
      description: +label=Rate Limit Rule
      type: object
      required:
        - id
        - when
        - limit_to
        - unit
      properties:
        id:
          description: |-
            +usage=Unique identifier for the rule
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Rule ID
          type: string
          pattern: ^[A-Za-z{}][A-Za-z0-9{}._-]{0,60}[A-Za-z0-9{}._]$
        when:
          $ref: '#/components/schemas/RateLimitWhen'
        limit_to:
          description: |-
            +usage=Maximum number of requests or tokens allowed
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Limit Value
          type: integer
        unit:
          $ref: '#/components/schemas/RateLimitUnit'
        rate_limit_applies_per:
          description: >-
            +usage=Create separate rate limiting rules for each unique value of
            the selected entity. For example, if "user" is selected, a separate
            rate limit rule will be created for each unique user making
            requests. Options: user, virtualaccount, model, or a metadata key
            (e.g., metadata.appId).

            +uiProps={"descriptionInline":true}

            +sort=5

            +label=Rate Limit Applies Per
          type: array
          items:
            type: string
      title: RateLimitRule
      x-fern-type-name: RateLimitRule
    LoadBalancingRule:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/WeightBasedLoadBalancingRule'
        - $ref: '#/components/schemas/LatencyBasedLoadBalancingRule'
        - $ref: '#/components/schemas/PriorityBasedLoadBalancingRule'
      discriminator:
        propertyName: type
      title: LoadBalancingRule
      x-fern-type-name: LoadBalancingRule
    GuardrailsRule:
      description: +label=Guardrails Rule
      type: object
      required:
        - id
        - when
        - llm_input_guardrails
        - llm_output_guardrails
        - mcp_tool_pre_invoke_guardrails
        - mcp_tool_post_invoke_guardrails
      properties:
        id:
          description: |-
            +usage=Unique identifier for the rule
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Rule ID
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-]{0,60}[a-zA-Z0-9])$
        when:
          $ref: '#/components/schemas/GuardrailsWhen'
        llm_input_guardrails:
          description: >-
            +usage=Guardrail selectors of the guardrails to apply to the input
            (format: groupName/guardrailName)

            +uiProps={"descriptionInline":true}

            +sort=4

            +label=Input Guardrails
          type: array
          items:
            type: string
        llm_output_guardrails:
          description: >-
            +usage=Guardrail selectors of the guardrails to apply to the output
            (format: groupName/guardrailName)

            +uiProps={"descriptionInline":true}

            +sort=5

            +label=Output Guardrails
          type: array
          items:
            type: string
        mcp_tool_pre_invoke_guardrails:
          description: >-
            +usage=Guardrail selectors of the guardrails to apply to the MCP
            tool pre invoke (format: groupName/guardrailName)

            +uiProps={"descriptionInline":true}

            +sort=4

            +label=MCP Tool Pre Invoke Guardrails
          type: array
          items:
            type: string
        mcp_tool_post_invoke_guardrails:
          description: >-
            +usage=Guardrail selectors of the guardrails to apply to the MCP
            tool post invoke (format: groupName/guardrailName)

            +uiProps={"descriptionInline":true}

            +sort=5

            +label=MCP Tool Post Invoke Guardrails
          type: array
          items:
            type: string
      title: GuardrailsRule
      x-fern-type-name: GuardrailsRule
    BudgetRule:
      description: +label=Budget Rule
      type: object
      required:
        - id
        - when
        - limit_to
        - unit
      properties:
        id:
          description: |-
            +usage=Unique identifier for the rule
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Rule ID
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-]{0,60}[a-zA-Z0-9])$
        when:
          $ref: '#/components/schemas/BudgetWhen'
        limit_to:
          description: |-
            +usage=Numeric limit value for the budget
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Limit Value
          type: number
        unit:
          $ref: '#/components/schemas/BudgetLimitUnit'
        budget_applies_per:
          description: >-
            +usage=Create separate budget rules for each unique value of the
            selected entity. For example, if "user" is selected, a separate
            budget rule will be created for each unique user making requests.
            Options: user, virtualaccount, model, or a metadata key (e.g.,
            metadata.appId).

            +uiProps={"descriptionInline":true}

            +sort=5

            +label=Budget Applies Per
          type: array
          items:
            type: string
        alerts:
          $ref: '#/components/schemas/BudgetAlert'
        audit_mode:
          description: >-
            +usage=When enabled, requests exceeding the budget are tracked but
            not blocked

            +uiProps={"descriptionInline":true}

            +sort=7

            +label=Audit Mode
          type: boolean
          default: false
      title: BudgetRule
      x-fern-type-name: BudgetRule
    OTELTracesExporterHttpConfig:
      description: |-
        +label=HTTP Configuration
        +uiType=Ignore
        +required=true
      type: object
      properties:
        type:
          description: |-
            +value=http
            +usage=Configuration type
          type: string
          enum:
            - http
      allOf:
        - $ref: '#/components/schemas/OTELExporterHttpConfigBase'
        - $ref: '#/components/schemas/OTELTracesExporterCommonConfig'
      required:
        - type
      title: OTELTracesExporterHttpConfig
      x-fern-type-name: OTELTracesExporterHttpConfig
    OTELTracesExporterGrpcConfig:
      description: |-
        +label=gRPC Configuration
        +uiType=Ignore
        +required=true
      type: object
      properties:
        type:
          description: |-
            +value=grpc
            +usage=Configuration type
          type: string
          enum:
            - grpc
      allOf:
        - $ref: '#/components/schemas/OTELExporterGrpcConfigBase'
        - $ref: '#/components/schemas/OTELTracesExporterCommonConfig'
      required:
        - type
      title: OTELTracesExporterGrpcConfig
      x-fern-type-name: OTELTracesExporterGrpcConfig
    OTELMetricsExporterHttpConfig:
      description: |-
        +label=HTTP Configuration
        +uiType=Ignore
        +required=true
      type: object
      properties:
        type:
          description: |-
            +value=http
            +usage=Configuration type
          type: string
          enum:
            - http
      allOf:
        - $ref: '#/components/schemas/OTELExporterHttpConfigBase'
      required:
        - type
      title: OTELMetricsExporterHttpConfig
      x-fern-type-name: OTELMetricsExporterHttpConfig
    OTELMetricsExporterGrpcConfig:
      description: |-
        +label=gRPC Configuration
        +uiType=Ignore
        +required=true
      type: object
      properties:
        type:
          description: |-
            +value=grpc
            +usage=Configuration type
          type: string
          enum:
            - grpc
      allOf:
        - $ref: '#/components/schemas/OTELExporterGrpcConfigBase'
      required:
        - type
      title: OTELMetricsExporterGrpcConfig
      x-fern-type-name: OTELMetricsExporterGrpcConfig
    GatewayMetadataRule:
      description: +label=Metadata Rule
      type: object
      required:
        - id
        - when
        - metadata_override
      properties:
        id:
          description: |-
            +usage=Unique identifier for the rule
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Rule ID
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-]{0,60}[a-zA-Z0-9])$
        when:
          $ref: '#/components/schemas/GatewayMetadataWhen'
        metadata_override:
          description: |-
            +usage=Metadata key-value pairs to override in the request metadata
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Metadata Override
          type: object
          additionalProperties:
            type: string
      title: GatewayMetadataRule
      x-fern-type-name: GatewayMetadataRule
    GatewayLoggingRule:
      description: +label=Logging Rule
      type: object
      required:
        - id
        - when
        - tracing_project_fqn
      properties:
        id:
          description: |-
            +usage=Unique identifier for the rule
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Rule ID
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-]{0,60}[a-zA-Z0-9])$
        when:
          $ref: '#/components/schemas/GatewayLoggingWhen'
        tracing_project_fqn:
          description: |-
            +usage=FQN of the tracing project to send logs to
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Tracing Project FQN
          type: string
      title: GatewayLoggingRule
      x-fern-type-name: GatewayLoggingRule
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
    LoggingMode:
      description: +label=Logging Mode
      type: object
      oneOf:
        - $ref: '#/components/schemas/LoggingModeAlways'
        - $ref: '#/components/schemas/LoggingModeNever'
        - $ref: '#/components/schemas/LoggingModeHeaderControlled'
      title: LoggingMode
      x-fern-type-name: LoggingMode
    SemanticCacheSettings:
      description: +label=Semantic Cache Settings
      type: object
      required:
        - semantic_cache_embedding_model
      properties:
        semantic_cache_embedding_model:
          description: >-
            +usage=Select the embedding model to use for semantic caching across
            the gateway. This model will be used to generate embeddings for
            cache lookups when semantic caching is enabled. The embedding model
            helps identify semantically similar requests to serve cached
            responses, improving performance for similar queries.

            +uiProps={"descriptionInline":true,"integrationType":"model","valuePattern":"${providerAccount.name}/${manifest.name}","filterConfig":{"model_types":["embedding"]}}

            +label=Semantic Cache Embedding Model

            +uiType=IntegrationSelect
          type: string
      title: SemanticCacheSettings
      x-fern-type-name: SemanticCacheSettings
    DataAccessRule:
      description: |-
        +label=Scope
        +discriminator=scope
        +uiType=Select
        +usage=Select the scope for the rule
      type: object
      oneOf:
        - $ref: '#/components/schemas/OwnDataAccessRule'
        - $ref: '#/components/schemas/TeamDataAccessRule'
        - $ref: '#/components/schemas/AllDataAccessRule'
      discriminator:
        propertyName: scope
      title: DataAccessRule
      x-fern-type-name: DataAccessRule
    GatewayDataRoutingConfigDestination:
      description: |-
        +label=Data Routing Destination
        +usage=Destination for data routing based on conditions.
      type: object
      properties:
        name:
          description: |-
            +label=Name
            +usage=Name for the destination
            +sort=2
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-]{0,60}[a-zA-Z0-9])$
        enabled:
          description: |-
            +label=Enabled
            +usage=Whether this destination is enabled
            +sort=3
          type: boolean
          default: true
        description:
          description: |-
            +label=Description
            +usage=Description of the destination
            +sort=4
          type: string
          minLength: 1
        when:
          description: |-
            +label=When
            +usage=Conditions that must be met to use this destination
            +sort=6
          type: array
          items:
            $ref: '#/components/schemas/GatewayDataRoutingConfigDestinationCondition'
      allOf:
        - $ref: '#/components/schemas/GatewayDataRoutingConfigDestinationStorage'
      required:
        - name
        - enabled
        - when
      title: GatewayDataRoutingConfigDestination
      x-fern-type-name: GatewayDataRoutingConfigDestination
    GatewayDataRoutingConfigDestinationStorage:
      type: object
      required:
        - storage
      properties:
        storage:
          description: |-
            +label=Storage
            +usage=Storage configuration for traces and spans
            +sort=100
            +discriminator=type
          type: object
          default:
            type: controlplane-managed
          oneOf:
            - $ref: >-
                #/components/schemas/GatewayDataRoutingConfigDestinationControlPlaneManagedStorage
            - $ref: >-
                #/components/schemas/GatewayDataRoutingConfigDestinationCustomerManagedStorage
          discriminator:
            propertyName: type
        traces_retention_period_days:
          description: >-
            +label=Traces Retention Period (Days)

            +usage=Number of days to retain traces (this does not affect metrics
            storage)

            +sort=200
          type: integer
          minimum: 7
      title: GatewayDataRoutingConfigDestinationStorage
      x-fern-type-name: GatewayDataRoutingConfigDestinationStorage
    Email:
      description: +label=Email
      type: object
      required:
        - type
        - notification_channel
        - to_emails
      properties:
        type:
          description: +value=email
          type: string
          enum:
            - email
        notification_channel:
          description: |-
            +label=Notification Channel
            +usage=Specify the notification channel to send alerts to
            +uiType=AlertNotificationChannel
            +uiProps={"integrationType":"notification-channel"}
            +sort=660
          type: string
          minLength: 1
        to_emails:
          description: >-
            +label=To Emails

            +usage=List of recipients' email addresses if the notification
            channel is Email.

            +docs=Specify the emails to send alerts to

            +sort=665
          type: array
          items:
            type: string
            minLength: 1
      title: Email
      x-fern-type-name: Email
    SlackWebhook:
      description: +label=Slack Webhook
      type: object
      required:
        - type
        - notification_channel
      properties:
        type:
          description: +value=slack-webhook
          type: string
          enum:
            - slack-webhook
        notification_channel:
          description: |-
            +label=Notification Channel
            +usage=Specify the notification channel to send alerts to
            +uiType=AlertNotificationChannel
            +uiProps={"integrationType":"notification-channel"}
            +sort=660
          type: string
          minLength: 1
      title: SlackWebhook
      x-fern-type-name: SlackWebhook
    SlackBot:
      type: object
      required:
        - type
        - notification_channel
        - channels
      properties:
        type:
          description: +value=slack-bot
          type: string
          enum:
            - slack-bot
        notification_channel:
          description: |-
            +label=Notification Channel
            +usage=Specify the notification channel to send alerts to
            +uiType=AlertNotificationChannel
            +uiProps={"integrationType":"notification-channel"}
            +sort=660
          type: string
          minLength: 1
        channels:
          description: >-
            +label=Slack Channel Names

            +usage=List of channels to send messages to.

            +message=3 to 81 lower case characters long alphanumeric word, may
            contain - or _ in between, must start with #

            +sort=665
          type: array
          items:
            type: string
            pattern: ^#[a-z0-9\-_]{2,80}$
      title: SlackBot
      x-fern-type-name: SlackBot
    PagerDuty:
      type: object
      required:
        - type
        - notification_channel
      properties:
        type:
          description: +value=pagerduty
          type: string
          enum:
            - pagerduty
        notification_channel:
          description: |-
            +label=Notification Channel
            +usage=Specify the notification channel to send alerts to
            +uiType=AlertNotificationChannel
            +uiProps={"integrationType":"notification-channel"}
            +sort=660
          type: string
          minLength: 1
      title: PagerDuty
      x-fern-type-name: PagerDuty
    AlertSeverity:
      type: string
      enum:
        - warning
        - critical
      title: AlertSeverity
      x-fern-type-name: AlertSeverity
    TrueFoundryAgentSkill:
      type: object
      required:
        - fqn
        - preload_skill_md
      properties:
        fqn:
          description: |-
            +label=FQN
            +usage=Fully qualified name of the agent skill
            +sort=100
          type: string
          minLength: 1
          pattern: ^agent_skill:[^/]+/[^/]+/[^/:]+:[0-9]+$
        preload_skill_md:
          description: >-
            +label=Preload SKILL.md Content

            +usage=If true, the SKILL.md content is injected into the agent
            context.

            +sort=200

            +uiType=Switch
          type: boolean
          default: false
      title: TrueFoundryAgentSkill
      x-fern-type-name: TrueFoundryAgentSkill
    TrueFoundryAgentMcpServer:
      type: object
      required:
        - name
      properties:
        name:
          description: |-
            +label=Name
            +usage=MCP server name
            +sort=100
          type: string
        deferred:
          description: >-
            +label=Deferred

            +usage=When true, tools from this server are loaded lazily (deferred
            loading).

            +sort=300

            +uiType=Switch
          type: boolean
          default: false
        tools:
          description: |-
            +label=Tools
            +usage=List of tools to enable from this server
            +sort=400
          type: array
          items:
            $ref: '#/components/schemas/TrueFoundryAgentMcpTool'
      title: TrueFoundryAgentMcpServer
      x-fern-type-name: TrueFoundryAgentMcpServer
    TrueFoundryAgentUserMessage:
      type: object
      required:
        - role
        - content
      properties:
        role:
          description: |-
            +label=Role
            +usage=Role of the message sender.
            +uiType=Hidden
          type: string
          enum:
            - user
        content:
          description: >-
            +label=Content

            +usage=Content of the user message. Supports template variables like
            {{variable_name}}.

            +message=Content must not be empty.
          type: string
          minLength: 1
      title: TrueFoundryAgentUserMessage
      x-fern-type-name: TrueFoundryAgentUserMessage
    TrueFoundryAgentVariable:
      type: object
      properties:
        default_value:
          description: |-
            +label=Default Value
            +usage=Default value for the variable
          type: string
        description:
          description: |-
            +label=Description
            +usage=Description of the variable
          type: string
      title: TrueFoundryAgentVariable
      x-fern-type-name: TrueFoundryAgentVariable
    ResponseFormatText:
      type: object
      required:
        - type
      properties:
        type:
          type: string
          enum:
            - text
      title: ResponseFormatText
      x-fern-type-name: ResponseFormatText
    ResponseFormatJSONSchema:
      type: object
      required:
        - type
        - json_schema
      properties:
        type:
          type: string
          enum:
            - json_schema
        json_schema:
          type: object
          required:
            - name
          properties:
            name:
              type: string
            description:
              type: string
            schema:
              type: object
            strict:
              type: boolean
      title: ResponseFormatJSONSchema
      x-fern-type-name: ResponseFormatJSONSchema
    ResponseFormatJSONObject:
      type: object
      required:
        - type
      properties:
        type:
          type: string
          enum:
            - json_object
      title: ResponseFormatJSONObject
      x-fern-type-name: ResponseFormatJSONObject
    SelfHostedAgent:
      description: +label=Self Hosted
      type: object
      properties:
        type:
          description: |-
            +value=remote-agent/self-hosted
            +label=Type
            +usage=Type of the manifest
            +uiType=Hidden
          type: string
          enum:
            - remote-agent/self-hosted
        url:
          description: |-
            +label=URL
            +usage=The URL of the server
            +sort=200
          type: string
          pattern: ^https?:\/\/[^\s]+$
        auth_data:
          $ref: '#/components/schemas/CustomAgentServerAuth'
        framework:
          $ref: '#/components/schemas/AgentFramework'
      allOf:
        - $ref: '#/components/schemas/BaseRemoteAgent'
      required:
        - type
        - url
        - framework
      title: SelfHostedAgent
      x-fern-type-name: SelfHostedAgent
    AgentSource:
      description: |-
        +discriminator=type
        +sort=300
      type: object
      oneOf:
        - $ref: '#/components/schemas/PromptSource'
        - $ref: '#/components/schemas/HostedA2AAgent'
      discriminator:
        propertyName: type
      title: AgentSource
      x-fern-type-name: AgentSource
    SampleAgentInput:
      type: object
      properties:
        text:
          description: |-
            +label=Text
            +usage=Input for the agent
          type: string
        variables:
          description: |-
            +label=Variables
            +usage=Variable inputs for Prompt Variables defined in the Prompt
          type: object
          additionalProperties:
            type: string
      title: SampleAgentInput
      x-fern-type-name: SampleAgentInput
    TracingProjectControlPlaneManagedStorage:
      description: |-
        +label=Controlplane Managed Storage
        +usage=Managed Storage in Control-Plane.
      type: object
      required:
        - type
      properties:
        type:
          description: |-
            +value=controlplane-managed
            +label=Storage Type
            +usage=Type of storage configuration
            +message=Select controlplane-managed or customer-managed storage
            +sort=1
            +uiType=Select
          type: string
          enum:
            - controlplane-managed
      title: TracingProjectControlPlaneManagedStorage
      x-fern-type-name: TracingProjectControlPlaneManagedStorage
    TracingProjectCustomerManagedStorage:
      description: >-
        +label=Customer Managed Storage

        +usage=Bring your own storage. Please note that this cannot be changed
        once set.
      type: object
      required:
        - type
        - storage_integration_fqn
      properties:
        type:
          description: |-
            +value=customer-managed
            +label=Storage Type
            +usage=Type of storage configuration
            +message=Select controlplane-managed or customer-managed storage
            +sort=1
          type: string
          enum:
            - customer-managed
        storage_integration_fqn:
          description: >-
            +label=Storage Integration

            +icon=hard-drive

            +usage=Storage Integration FQN to store traces and spans. A storage
            integration represents a connected blob storage like AWS S3 / Azure
            Blob Storage / Google Cloud Storage. Please note that this cannot be
            changed once set.

            +message=Select the storage-integration where you want to save your
            traces and spans. If no integration is listed, please contact your
            admin.\n[View
            Docs](https://www.truefoundry.com/docs/integrations-overview#integrations)

            +sort=2

            +uiType=IntegrationSelect

            +uiProps={"descriptionKey":"storage_root","integrationType":"blob-storage"}
          type: string
          pattern: ^[a-zA-Z0-9][a-zA-Z0-9._:-]*[a-zA-Z0-9]$
      title: TracingProjectCustomerManagedStorage
      x-fern-type-name: TracingProjectCustomerManagedStorage
    CustomTLSSettings:
      description: >-
        +label=TLS Settings

        +usage=Configure TLS settings for secure connections with custom CA
        certificates.
      type: object
      required:
        - reject_unauthorized
      properties:
        reject_unauthorized:
          description: >-
            +label=Reject Unauthorized

            +usage=When set to true, it will reject any connection which is not
            authorized with the list of supplied CAs.

            +uiType=Switch
          type: boolean
          default: true
        ca_cert:
          description: |-
            +label=Custom CA Certificates
            +usage=Add custom CA certificates to use for TLS verification.
            +uiType=CodeEditor
          type: string
      title: CustomTLSSettings
      x-fern-type-name: CustomTLSSettings
    MCPToolSetting:
      description: +label=MCP Tool Setting
      type: object
      required:
        - name
      properties:
        name:
          description: |-
            +label=Tool Name
            +usage=The unique name of the tool.
            +message=1-128 characters: letters, digits, underscore, hyphen
            +sort=100
            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^[A-Za-z0-9_-]{1,128}$
        description:
          description: >-
            +label=Tool Description

            +usage=Describe what this tool does. This helps AI models understand
            when to use it.

            +uiType=TextArea

            +message=1 to 20000 characters long

            +sort=200

            +uiProps={"descriptionInline":true}
          type: string
          minLength: 1
          maxLength: 20000
        disabled:
          description: |-
            +label=Disable Tool
            +usage=Allows to disable the tool.
            +sort=300
          type: boolean
      title: MCPToolSetting
      x-fern-type-name: MCPToolSetting
    MCPServerAuth:
      type: object
      oneOf:
        - $ref: '#/components/schemas/MCPServerHeaderAuth'
        - $ref: '#/components/schemas/MCPServerOAuth2'
        - $ref: '#/components/schemas/MCPServerPassthrough'
      title: MCPServerAuth
      x-fern-type-name: MCPServerAuth
    McpServerSource:
      description: +label=MCP Server Source
      type: object
      required:
        - name
      properties:
        name:
          description: |-
            +label=MCP Server Name
            +usage=Name of the MCP server
            +uiType=VirtualMCPServerSelection
            +sort=100
          type: string
        enabled_tools:
          description: |-
            +label=Enabled Tools
            +usage=List of tools to enable from this source server.
            +uiType=Hidden
            +sort=200
          type: array
          items:
            type: string
        auth_data:
          $ref: '#/components/schemas/MCPServerHeaderOverrideAuth'
      title: McpServerSource
      x-fern-type-name: McpServerSource
    OpenAPISpecSource:
      type: object
      oneOf:
        - $ref: '#/components/schemas/RemoteSpecSource'
        - $ref: '#/components/schemas/UploadSpecSource'
        - $ref: '#/components/schemas/InlineSpecSource'
      title: OpenAPISpecSource
      x-fern-type-name: OpenAPISpecSource
    OpenAPIMCPToolSetting:
      description: +label=OpenAPI MCP Tool Setting
      type: object
      required:
        - method
        - path
        - name
        - description
      properties:
        method:
          description: |-
            +label=HTTP Method
            +usage=The HTTP method for this endpoint.
            +sort=100
          type: string
          enum:
            - get
            - post
            - put
            - patch
            - delete
            - head
        path:
          description: |-
            +label=API Path
            +usage=The API endpoint path (e.g., /users/{id}).
            +sort=200
          type: string
        name:
          description: >-
            +label=Tool Name

            +usage=The unique name of this tool, derived from OpenAPI
            operationId.

            +message=1-128 characters: letters, digits, underscore, hyphen, dot

            +sort=300

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^[A-Za-z0-9_.-]{1,128}$
        description:
          description: >-
            +label=Tool Description

            +usage=Describe what this tool does. This helps AI models understand
            when to use it.

            +uiType=TextArea

            +message=1 to 20000 characters long

            +sort=400

            +uiProps={"descriptionInline":true}
          type: string
          minLength: 1
          maxLength: 20000
      title: OpenAPIMCPToolSetting
      x-fern-type-name: OpenAPIMCPToolSetting
    MCPServerEnvAuth:
      description: >-
        +label=Env Auth

        +usage=Environment variable authentication for stdio MCP. For individual
        credentials, use a single {{placeholder}} name in values (same rules as
        header auth).
      type: object
      required:
        - type
        - env
      properties:
        type:
          description: +value=env
          type: string
          enum:
            - env
        env:
          description: >-
            +usage=Map each env variable to its value. For shared, enter the
            actual value. For individual, use a placeholder that will be filled
            in by each user (e.g. API_KEY {{API_KEY}}).
          type: object
          additionalProperties:
            type: string
        auth_level:
          description: >-
            +label=Auth level

            +usage=Level at which this authentication will be applied. Values:
            global (default), per_user

            +uiType=RadioCard

            +sort=200
          type: string
          enum:
            - global
            - per_user
          default: global
      title: MCPServerEnvAuth
      x-fern-type-name: MCPServerEnvAuth
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
    ArtifactsCacheVolume:
      description: |-
        +docs=Describes the volume that will be used to cache the models
        +label=Artifacts Cache Volume
      type: object
      required:
        - storage_class
        - cache_size
      properties:
        storage_class:
          description: |-
            +label=Storage Class
            +usage=Storage class of the Volume where artifacts will be cached
          type: string
        cache_size:
          description: >-
            +label=Cache Size (GB)

            +usage=Size of the Volume (in GB) where artifacts will be cached.
            Should be greater than twice the size of artifacts getting cached
          type: integer
          minimum: 1
          maximum: 1000
          default: 200
      title: ArtifactsCacheVolume
      x-fern-type-name: ArtifactsCacheVolume
    TrueFoundryArtifactSource:
      description: |-
        +docs=Input for Artifact from TrueFoundry Artifact Registry
        +label=TrueFoundry Artifact Source
      type: object
      required:
        - type
        - artifact_version_fqn
        - download_path_env_variable
      properties:
        type:
          description: +value=truefoundry-artifact
          type: string
          enum:
            - truefoundry-artifact
        artifact_version_fqn:
          description: >-
            +label=Artifact or Model Version FQN

            +usage=Artifact or Model Version FQN of the artifact to be
            downloaded
          type: string
        download_path_env_variable:
          description: >-
            +label=Download Path Environment Variable

            +usage=Environment variable which will contain the download path of
            the artifact
          type: string
      title: TrueFoundryArtifactSource
      x-fern-type-name: TrueFoundryArtifactSource
    HuggingfaceArtifactSource:
      description: |-
        +docs=Input for Artifact from Huggingface Model Hub
        +label=Huggingface Model Source
      type: object
      required:
        - type
        - model_id
        - revision
        - ignore_patterns
        - download_path_env_variable
      properties:
        type:
          description: +value=huggingface-hub
          type: string
          enum:
            - huggingface-hub
        model_id:
          description: |-
            +label=Model ID
            +usage=Model ID of the artifact to be downloaded
          type: string
        revision:
          description: |-
            +label=Revision
            +usage=Revision of the artifact to be downloaded
          type: string
        ignore_patterns:
          description: |-
            +label=Ignore Patterns
            +usage=List of patterns to ignore while downloading the artifact
          type: array
          items:
            type: string
          default:
            - '*.h5'
            - '*.ot'
            - '*.tflite'
            - '*.msgpack'
        download_path_env_variable:
          description: >-
            +label=Download Path Environment Variable

            +usage=Environment variable which will contain the download path of
            the artifact
          type: string
      title: HuggingfaceArtifactSource
      x-fern-type-name: HuggingfaceArtifactSource
    BasicAuthCreds:
      description: +label=Username and password
      type: object
      required:
        - type
        - username
        - password
      properties:
        type:
          description: +value=basic_auth
          type: string
          enum:
            - basic_auth
        username:
          description: |-
            +label=Username for service auth
            +message=Username for the user to authenticate with
            +sort=1
          type: string
        password:
          description: |-
            +label=Password for service auth
            +message=Password for the user to authenticate with
            +sort=2
          type: string
      title: BasicAuthCreds
      x-fern-type-name: BasicAuthCreds
    JwtAuthConfig:
      description: |-
        +label=JWT Authentication
        +usage=Configure JWT-based authentication using JWKS
      type: object
      required:
        - type
        - integration_fqn
      properties:
        type:
          description: +value=jwt_auth
          type: string
          enum:
            - jwt_auth
        integration_fqn:
          description: >-
            +docs=FQN of the JWT Auth integration. You can use the FQN of your
            desired jwt auth integration (or add one)

            in the  Integrations page[Integrations](/integrations) page

            +label=JWT Auth Integration

            +sort=200

            +usage=FQN of the JWT Auth integration. If you can't find your
            integration here,

            add it through the [Integrations](/integrations) page

            +uiType=IntegrationSelect

            +uiProps={"integrationType":"jwt-auth"}
          type: string
        enable_login:
          description: |-
            +label=Enable Login
            +usage=Enable login for the service
            +sort=300
          type: boolean
        claims:
          description: |-
            +label=Claims
            +usage=List of key-value pairs of claims to verify in the JWT token
            +sort=400
          type: array
          items:
            type: object
            required:
              - key
              - values
            properties:
              key:
                type: string
              values:
                type: array
                items:
                  type: string
        bypass_auth_paths:
          description: >-
            +label=Paths that will bypass auth

            +usage=List of paths that will bypass auth.

            needs to start with a forward slash(/) and should not contain
            wildcards(*)

            +sort=500
          type: array
          items:
            type: string
            pattern: ^/[^*]*
      title: JwtAuthConfig
      x-fern-type-name: JwtAuthConfig
    TrueFoundryInteractiveLogin:
      description: +label=Login with truefoundry
      type: object
      required:
        - type
      properties:
        type:
          description: +value=truefoundry_oauth
          type: string
          enum:
            - truefoundry_oauth
        bypass_auth_paths:
          description: >-
            +label=Paths that will bypass auth

            +usage=List of paths that will bypass auth.

            needs to start with a forward slash(/) and should not contain
            wildcards(*)
          type: array
          items:
            type: string
            pattern: ^/[^*]*
      title: TrueFoundryInteractiveLogin
      x-fern-type-name: TrueFoundryInteractiveLogin
    HttpProbe:
      description: >-
        +docs=Describes the Instructions for assessing container health by
        executing an HTTP GET request.

        To learn more you can go
        [here](https://docs.truefoundry.com/docs/liveness-readiness-probe)

        +label=Instructions for assessing container health by executing an HTTP
        GET request.
      type: object
      required:
        - type
        - path
        - port
      properties:
        type:
          description: |-
            +sort=1
            +label=Request Type
            +value=http
          type: string
          enum:
            - http
        path:
          description: |-
            +usage=Path to the health check endpoint
            +sort=2
          type: string
        port:
          description: |-
            +usage=Listening port for the health check endpoint
            +sort=3
          type: integer
          minimum: 0
          maximum: 65535
        host:
          description: |-
            +sort=4
            +usage=Host name to connect to, defaults to the pod IP
          type: string
        scheme:
          description: |-
            +sort=5
            +usage=Scheme to use for connecting to the host
          type: string
          default: HTTP
      title: HttpProbe
      x-fern-type-name: HttpProbe
    AwsAccessKeyBasedAuth:
      description: +label=AWS Access Key Based Auth
      type: object
      required:
        - type
        - access_key_id
        - secret_access_key
      properties:
        type:
          description: +value=access-key-based
          type: string
          enum:
            - access-key-based
        access_key_id:
          description: |-
            +label=AWS access key ID
            +usage=AWS access key ID that you want to use for authentication.
            +sort=100
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
        secret_access_key:
          description: >-
            +label=AWS access key secret

            +usage=AWS access key secret that you want to use for
            authentication.

            +sort=200

            +uiType=SecretSelect

            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: AwsAccessKeyBasedAuth
      x-fern-type-name: AwsAccessKeyBasedAuth
    AwsAssumedRoleBasedAuth:
      description: +label=AWS Assumed Role Based Auth
      type: object
      required:
        - type
        - assumed_role_arn
      properties:
        type:
          description: +value=assumed-role-based
          type: string
          enum:
            - assumed-role-based
        assumed_role_arn:
          description: |-
            +label=AWS Assumed Role ARN
            +usage=ARN of the role you want to assume.
            +sort=100
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^arn:(aws|aws-us-gov|aws-cn):iam::[0-9]{12}:role(/[\w+=,.@-]+)+$
      title: AwsAssumedRoleBasedAuth
      x-fern-type-name: AwsAssumedRoleBasedAuth
    AwsIntegrations:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/AwsEcr'
        - $ref: '#/components/schemas/AwsS3'
        - $ref: '#/components/schemas/AwsParameterStore'
        - $ref: '#/components/schemas/AwsSecretsManager'
        - $ref: '#/components/schemas/AwsEksIntegration'
      discriminator:
        propertyName: type
      title: AwsIntegrations
      x-fern-type-name: AwsIntegrations
    AzureIntegrations:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/AzureAksIntegration'
        - $ref: '#/components/schemas/AzureContainerRegistry'
        - $ref: '#/components/schemas/AzureBlobStorage'
        - $ref: '#/components/schemas/AzureVault'
        - $ref: '#/components/schemas/AzureReposIntegration'
      discriminator:
        propertyName: type
      title: AzureIntegrations
      x-fern-type-name: AzureIntegrations
    GcpKeyFileAuth:
      description: +label=GCP Key File Auth
      type: object
      required:
        - type
        - key_file_content
      properties:
        type:
          description: +value=key-file
          type: string
          enum:
            - key-file
        key_file_content:
          description: |-
            +uiType=SecretSelect
            +label=Key File Content
            +sort=100
            +message=Key file content must be valid JSON
            +uiProps={"typeOptionsList":["json","secret"]}
          oneOf:
            - type: string
            - type: object
      title: GcpKeyFileAuth
      x-fern-type-name: GcpKeyFileAuth
    GcpIntegrations:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/GcpGcr'
        - $ref: '#/components/schemas/GcpGcs'
        - $ref: '#/components/schemas/GcpGsm'
        - $ref: '#/components/schemas/GcpGkeIntegration'
      discriminator:
        propertyName: type
      title: GcpIntegrations
      x-fern-type-name: GcpIntegrations
    DockerhubBasicAuth:
      description: +label=Dockerhub Basic Auth
      type: object
      required:
        - type
        - username
        - password
      properties:
        type:
          description: +value=basic-auth
          type: string
          enum:
            - basic-auth
        username:
          description: +sort=100
          type: string
        password:
          description: |-
            +sort=200
            +uiType=SecretSelect
          type: string
      title: DockerhubBasicAuth
      x-fern-type-name: DockerhubBasicAuth
    DockerhubIntegrations:
      $ref: '#/components/schemas/DockerhubRegistry'
      title: DockerhubIntegrations
      x-fern-type-name: DockerhubIntegrations
    BitbucketIntegration:
      description: |-
        +label=Bitbucket
        +icon=bitbucket
      type: object
      required:
        - type
        - name
        - account_name
      properties:
        type:
          description: +value=integration/vcs/bitbucket
          type: string
          enum:
            - integration/vcs/bitbucket
        name:
          description: +sort=100
          type: string
          pattern: ^\S+$
        account_name:
          description: |-
            +label=Account Name
            +sort=200
          type: string
          pattern: ^\S+$
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: BitbucketIntegration
      x-fern-type-name: BitbucketIntegration
    CustomBasicAuth:
      description: +label=Custom Basic Auth
      type: object
      required:
        - type
        - username
        - password
      properties:
        type:
          description: +value=basic-auth
          type: string
          enum:
            - basic-auth
        username:
          description: |-
            +label=Username
            +usage=The username for the basic authentication.
            +sort=100
          type: string
          pattern: ^\S+$
        password:
          description: |-
            +label=Password
            +usage=The password for the basic authentication.
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
            +sort=200
          type: string
          minLength: 1
      title: CustomBasicAuth
      x-fern-type-name: CustomBasicAuth
    CustomIntegrations:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/CustomUsernamePasswordArtifactsRegistry'
        - $ref: '#/components/schemas/EmailNotificationChannel'
        - $ref: '#/components/schemas/CustomHelmRepo'
        - $ref: '#/components/schemas/CustomBlobStorage'
        - $ref: '#/components/schemas/CustomJWTAuthIntegration'
        - $ref: '#/components/schemas/GenericSecretStoreIntegration'
      discriminator:
        propertyName: type
      title: CustomIntegrations
      x-fern-type-name: CustomIntegrations
    GithubIntegration:
      description: |-
        +label=Github
        +icon=github
      type: object
      required:
        - type
        - name
        - account_name
      properties:
        type:
          description: +value=integration/vcs/github
          type: string
          enum:
            - integration/vcs/github
        name:
          description: +sort=100
          type: string
          pattern: ^\S+$
        account_name:
          description: |-
            +label=Account Name
            +sort=200
          type: string
          pattern: ^\S+$
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: GithubIntegration
      x-fern-type-name: GithubIntegration
    GitlabIntegration:
      description: |-
        +label=Gitlab
        +icon=gitlab
      type: object
      required:
        - type
        - name
        - account_name
      properties:
        type:
          description: +value=integration/vcs/gitlab
          type: string
          enum:
            - integration/vcs/gitlab
        name:
          description: +sort=100
          type: string
          pattern: ^\S+$
        account_name:
          description: |-
            +label=Account Name
            +sort=200
          type: string
          pattern: ^\S+$
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: GitlabIntegration
      x-fern-type-name: GitlabIntegration
    JfrogBasicAuth:
      description: +label=Jfrog Basic Auth
      type: object
      required:
        - type
        - username
        - password
      properties:
        type:
          description: +value=basic-auth
          type: string
          enum:
            - basic-auth
        username:
          description: |-
            +label=Username
            +usage=The username for the basic authentication.
            +sort=100
          type: string
          pattern: ^\S+$
        password:
          description: |-
            +sort=200
            +label=Password
            +usage=The password for the basic authentication.
            +uiType=SecretSelect
          type: string
      title: JfrogBasicAuth
      x-fern-type-name: JfrogBasicAuth
    JFrogIntegrations:
      $ref: '#/components/schemas/JfrogArtifactsRegistry'
      title: JFrogIntegrations
      x-fern-type-name: JFrogIntegrations
    TTLIntegrations:
      $ref: '#/components/schemas/TTLRegistry'
      title: TTLIntegrations
      x-fern-type-name: TTLIntegrations
    TrueFoundryIntegrations:
      $ref: '#/components/schemas/TrueFoundryDBSSM'
      title: TrueFoundryIntegrations
      x-fern-type-name: TrueFoundryIntegrations
    QuayBasicAuth:
      description: +label=Quay Basic Auth
      type: object
      required:
        - type
        - username
        - password
      properties:
        type:
          description: +value=basic-auth
          type: string
          enum:
            - basic-auth
        username:
          description: |-
            +label=Username
            +usage=The username for the basic authentication.
            +sort=100
          type: string
        password:
          description: |-
            +label=Password
            +usage=The password for the basic authentication.
            +sort=200
            +uiType=SecretSelect
          type: string
      title: QuayBasicAuth
      x-fern-type-name: QuayBasicAuth
    QuayIntegrations:
      $ref: '#/components/schemas/QuayArtifactsRegistry'
      title: QuayIntegrations
      x-fern-type-name: QuayIntegrations
    SlackIntegrations:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/SlackWebhookIntegration'
        - $ref: '#/components/schemas/SlackBotIntegration'
      discriminator:
        propertyName: type
      title: SlackIntegrations
      x-fern-type-name: SlackIntegrations
    WebhookIntegrations:
      $ref: '#/components/schemas/WebhookIntegration'
      title: WebhookIntegrations
      x-fern-type-name: WebhookIntegrations
    PagerDutyIntegrations:
      $ref: '#/components/schemas/PagerDutyIntegration'
      title: PagerDutyIntegrations
      x-fern-type-name: PagerDutyIntegrations
    HashicorpIntegrations:
      $ref: '#/components/schemas/HashicorpVaultIntegration'
      title: HashicorpIntegrations
      x-fern-type-name: HashicorpIntegrations
    AWSRegion:
      type: string
      enum:
        - af-south-1
        - ap-east-1
        - ap-northeast-1
        - ap-northeast-2
        - ap-northeast-3
        - ap-south-1
        - ap-south-2
        - ap-southeast-1
        - ap-southeast-2
        - ap-southeast-3
        - ap-southeast-4
        - ap-southeast-5
        - ap-southeast-7
        - ca-central-1
        - ca-west-1
        - cn-north-1
        - cn-northwest-1
        - eu-central-1
        - eu-central-2
        - eu-north-1
        - eu-south-1
        - eu-south-2
        - eu-west-1
        - eu-west-2
        - eu-west-3
        - il-central-1
        - me-central-1
        - me-south-1
        - mx-central-1
        - sa-east-1
        - us-east-1
        - us-east-2
        - us-gov-east-1
        - us-gov-west-1
        - us-west-1
        - us-west-2
        - ap-southeast-6
        - ap-east-2
      title: AWSRegion
      x-fern-type-name: AWSRegion
    AwsBedrockAssumedRoleBasedAuth:
      description: +label=AWS Assumed Role Based Auth for Bedrock
      type: object
      properties:
        external_id:
          description: >-
            +label=External ID

            +usage=Optional external ID to use when assuming the role. This
            provides an added layer of security by ensuring only parties with
            the correct external ID can assume the role.

            +sort=150
          type: string
      allOf:
        - $ref: '#/components/schemas/AwsAssumedRoleBasedAuth'
      title: AwsBedrockAssumedRoleBasedAuth
      x-fern-type-name: AwsBedrockAssumedRoleBasedAuth
    AwsBedrockApiKeyAuth:
      description: |-
        +label=AWS Bedrock API Key Auth
        +icon=aws-bedrock
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=Your AWS Bedrock API key for authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: AwsBedrockApiKeyAuth
      x-fern-type-name: AwsBedrockApiKeyAuth
    BedrockModel:
      description: |-
        +label=Bedrock Model
        +icon=aws-bedrock
      type: object
      required:
        - name
        - model_id
        - type
        - model_types
      properties:
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The AWS Bedrock model identifier or inference profile. Can be
            a foundation model ID (e.g.,
            anthropic.claude-3-5-sonnet-20240620-v1:0,
            amazon.titan-text-express-v1), an inference profile ID (e.g.,
            us.anthropic.claude-3-5-sonnet-20240620-v1:0), or an inference
            profile ARN (e.g.,
            arn:aws:bedrock:us-east-1:123456789012:inference-profile/my-profile).

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        type:
          description: +value=integration/model/bedrock
          type: string
          enum:
            - integration/model/bedrock
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Bedrock model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        region:
          $ref: '#/components/schemas/AWSRegion'
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=6

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: BedrockModel
      x-fern-type-name: BedrockModel
    GCPRegion:
      type: string
      enum:
        - global
        - northamerica-northeast1
        - northamerica-northeast2
        - southamerica-east1
        - us-central1
        - us-east1
        - us-east4
        - us-east5
        - us-south1
        - us-west1
        - us-west2
        - us-west3
        - us-west4
        - asia-east1
        - asia-east2
        - asia-northeast1
        - asia-northeast2
        - asia-northeast3
        - asia-south1
        - asia-south2
        - asia-southeast1
        - asia-southeast2
        - australia-southeast1
        - australia-southeast2
        - europe-central2
        - europe-north1
        - europe-southwest1
        - europe-west1
        - europe-west2
        - europe-west3
        - europe-west4
        - europe-west6
        - europe-west8
        - europe-west9
        - me-central1
        - me-central2
        - me-west1
      title: GCPRegion
      x-fern-type-name: GCPRegion
    VertexKeyFileAuth:
      description: |-
        +label=Service account key file
        +icon=googleCloud
      type: object
      required:
        - type
        - key_file_content
      properties:
        type:
          description: +value=key-file
          type: string
          enum:
            - key-file
        key_file_content:
          description: >-
            +label=Service account key JSON

            +sort=100

            +usage=JSON from a downloadable Google Cloud service account key
            (type service_account). Not for Workload Identity Federation; use
            authentication type Workload Identity Federation file
            (workload-identity-federation-file) for gcloud iam
            workload-identity-pools create-cred-config output.

            +message=Key file content must be valid JSON

            +uiType=SecretSelect

            +uiProps={"typeOptionsList":["json","secret"]}
          oneOf:
            - type: string
            - type: object
      title: VertexKeyFileAuth
      x-fern-type-name: VertexKeyFileAuth
    VertexWifFileAuth:
      description: |-
        +label=Workload Identity Federation file
        +icon=googleCloud
      type: object
      required:
        - type
        - key_file_content
      properties:
        type:
          description: +value=workload-identity-federation-file
          type: string
          enum:
            - workload-identity-federation-file
        key_file_content:
          description: >-
            +label=Key file content

            +sort=100

            +usage=JSON credential configuration for Workload Identity
            Federation (ADC type external_account), for example from gcloud iam
            workload-identity-pools create-cred-config. This is not a service
            account private key. Same field name as key-file auth
            (key_file_content); use auth_data.type
            workload-identity-federation-file vs key-file to distinguish.

            +message=Key file content must be valid JSON

            +uiType=SecretSelect

            +uiProps={"typeOptionsList":["json","secret"]}
          oneOf:
            - type: string
            - type: object
      title: VertexWifFileAuth
      x-fern-type-name: VertexWifFileAuth
    VertexModel:
      description: |-
        +label=Vertex Model
        +icon=googleCloud
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/vertex
          type: string
          enum:
            - integration/model/vertex
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Google Vertex AI model identifier (e.g., gemini-2.5-pro,
            anthropic/claude-sonnet-4-5@20250929). This is the model version
            name from Vertex AI's model catalog.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        region:
          $ref: '#/components/schemas/GCPRegion'
        model_types:
          description: >-
            +label=Model Types

            +sort=4

            +usage=Specify the type of the Vertex AI model (e.g., chat, text,
            etc.)

            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=6

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: VertexModel
      x-fern-type-name: VertexModel
    GcpApiKeyAuth:
      description: |-
        +label=Google Cloud API Key Auth
        +icon=googleCloud
      type: object
      required:
        - type
        - apiKey
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        apiKey:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for Google Cloud authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: GcpApiKeyAuth
      x-fern-type-name: GcpApiKeyAuth
    GeminiModel:
      description: |-
        +label=Gemini Model
        +icon=googleCloud
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/gemini
          type: string
          enum:
            - integration/model/gemini
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Google Gemini model identifier (e.g., gemini-2.0-flash,
            gemini-2.0-flash-lite-preview-02-05,
            gemini-2.0-flash-lite-preview-02-05). This is the standard model
            name from Google's Gemini API documentation.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Gemini model (e.g., chat, text, etc.)
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: GeminiModel
      x-fern-type-name: GeminiModel
    AzureKeyAuth:
      description: |-
        +label=Azure API Key Auth
        +icon=azure
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for Azure AI Foundry authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: AzureKeyAuth
      x-fern-type-name: AzureKeyAuth
    AzureCertificateAuth:
      description: +label=Azure Entra certificate based auth
      type: object
      required:
        - type
        - tenant_id
        - client_id
        - certificate
        - certificate_thumbprint
      properties:
        type:
          description: +value=certificate
          type: string
          enum:
            - certificate
        tenant_id:
          description: |-
            +label=Tenant ID
            +sort=100
            +usage=The Azure AD tenant ID where the application is registered
            +message=Tenant ID must not be empty
            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        client_id:
          description: >-
            +label=Client ID

            +sort=200

            +usage=The Azure AD application (client) ID for service principal
            authentication

            +message=Client ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        certificate:
          description: >-
            +label=Certificate

            +sort=300

            +usage=Certificate in PEM format (e.g., -----BEGIN CERTIFICATE-----
            ... -----END CERTIFICATE-----)

            +message=Certificate must not be empty

            +placeholder=Add your certificate or select a TrueFoundry secret

            +uiType=SecretSelect

            +uiProps={"isPassword": true, "enableNew": true, "multiline": true}
          type: string
          pattern: ^\S+$
        certificate_thumbprint:
          description: >-
            +label=Certificate Thumbprint

            +sort=400

            +usage=The base64url-encoded SHA-1 thumbprint of the certificate
            (x5t claim)

            +message=Certificate thumbprint must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
      title: AzureCertificateAuth
      x-fern-type-name: AzureCertificateAuth
    AzureOpenAIModel:
      description: |-
        +label=Azure OpenAI Model
        +icon=azure
      type: object
      required:
        - type
        - name
        - model_id
        - api_version
        - deployment_type
        - model_types
      properties:
        type:
          description: +value=integration/model/azure-openai
          type: string
          enum:
            - integration/model/azure-openai
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID / Deployment Name

            +sort=2

            +usage=The model deployment name created in your Azure OpenAI
            resource (e.g., gpt-4-deployment, gpt-35-turbo-deployment). This is
            NOT the base model name but the deployment name you configured in
            Azure.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        api_version:
          description: |-
            +label=API Version
            +sort=3
            +usage=The Azure OpenAI API version to use
            +message=API version must not be empty
          type: string
          pattern: ^\S+$
        foundation_model:
          description: |-
            +label=Foundation Model
            +sort=4
            +usage=The foundation model identifier (e.g., gpt-4o-2024-11-20)
            +placeholder=gpt-4o-2024-11-20
          type: string
        deployment_type:
          description: >-
            +label=Deployment Type

            +sort=5

            +usage=Global: worldwide processing; Data zone (US): US data zone
            processing; Data zone (EU): EU data zone processing; Standard:
            single-region processing

            +message=Deployment type must be global, datazone_us, datazone_eu,
            or standard
          type: string
          enum:
            - standard
            - datazone_us
            - datazone_eu
            - global
          default: standard
        model_types:
          description: |-
            +label=Model Types
            +sort=6
            +usage=Specify the type of the Azure OpenAI model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=8

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: AzureOpenAIModel
      x-fern-type-name: AzureOpenAIModel
    AzureFoundryModel:
      description: |-
        +label=Azure AI Foundry Model
        +icon=azure
      type: object
      required:
        - type
        - name
        - auth_data
        - model_types
        - model_id
        - azure_endpoint
      properties:
        type:
          description: +value=integration/model/azure-foundry
          type: string
          enum:
            - integration/model/azure-foundry
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        auth_data:
          description: |-
            +label=Authentication Data
            +sort=2
            +usage=Azure AI Foundry authentication credentials
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AzureKeyAuth'
            - $ref: '#/components/schemas/AzureCertificateAuth'
          discriminator:
            propertyName: type
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Types of models supported by this Azure AI Foundry deployment
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        model_id:
          description: >-
            +label=Azure Deployment Name

            +sort=3

            +usage=The model deployment name created in your Azure AI Foundry
            project. This is the deployment name you configured in Azure AI
            Foundry, not the base model name.

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        azure_endpoint:
          description: >-
            +label=Azure Endpoint

            +sort=4

            +usage=The Azure AI Foundry endpoint URL. Should look like
            https://{deployment}.{region}.models.ai.azure.com

            +message=Endpoint URL must not be empty

            +uiProps={"warningValidation":{"message": "This URL looks like an
            Azure OpenAI endpoint. Did you mean to use the Azure OpenAI provider
            instead?", "pattern": "^(?!.*\\.openai).*"}}
          type: string
          pattern: ^\S+$
        api_version:
          description: |-
            +label=API Version
            +sort=5
            +usage=The API version for the Azure Foundry model
          type: string
          pattern: ^\S+$
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=7

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: AzureFoundryModel
      x-fern-type-name: AzureFoundryModel
    CohereKeyAuth:
      description: |-
        +label=Cohere API Key Auth
        +icon=cohere
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for Cohere authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: CohereKeyAuth
      x-fern-type-name: CohereKeyAuth
    CohereIntegrations:
      $ref: '#/components/schemas/CohereModel'
      title: CohereIntegrations
      x-fern-type-name: CohereIntegrations
    AI21KeyAuth:
      description: |-
        +label=AI21 API Key Auth
        +usage=Authentication method using AI21 API key
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +sort=100
            +usage=The API key for AI21 authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: AI21KeyAuth
      x-fern-type-name: AI21KeyAuth
    AI21Integrations:
      $ref: '#/components/schemas/AI21Model'
      title: AI21Integrations
      x-fern-type-name: AI21Integrations
    AnthropicKeyAuth:
      description: |-
        +label=Anthropic API Key Auth
        +usage=Authentication method using Anthropic API key
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +sort=100
            +usage=The API key for Anthropic authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: AnthropicKeyAuth
      x-fern-type-name: AnthropicKeyAuth
    AnthropicIntegrations:
      $ref: '#/components/schemas/AnthropicModel'
      title: AnthropicIntegrations
      x-fern-type-name: AnthropicIntegrations
    DeepinfraKeyAuth:
      description: |-
        +label=DeepInfra API Key Auth
        +icon=deepinfra
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for DeepInfra authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: DeepinfraKeyAuth
      x-fern-type-name: DeepinfraKeyAuth
    DeepinfraIntegrations:
      $ref: '#/components/schemas/DeepinfraModel'
      title: DeepinfraIntegrations
      x-fern-type-name: DeepinfraIntegrations
    GroqKeyAuth:
      description: |-
        +label=Groq API Key Auth
        +icon=groq
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for Groq authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: GroqKeyAuth
      x-fern-type-name: GroqKeyAuth
    GroqIntegrations:
      $ref: '#/components/schemas/GroqModel'
      title: GroqIntegrations
      x-fern-type-name: GroqIntegrations
    MistralAIKeyAuth:
      description: |-
        +label=Mistral AI API Key Auth
        +icon=mistral-ai
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for Mistral AI authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: MistralAIKeyAuth
      x-fern-type-name: MistralAIKeyAuth
    MistralAIIntegrations:
      $ref: '#/components/schemas/MistralAIModel'
      title: MistralAIIntegrations
      x-fern-type-name: MistralAIIntegrations
    PerplexityAIKeyAuth:
      description: |-
        +label=Perplexity AI API Key Auth
        +icon=perplexity-ai
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for Perplexity AI authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: PerplexityAIKeyAuth
      x-fern-type-name: PerplexityAIKeyAuth
    PerplexityIntegrations:
      $ref: '#/components/schemas/PerplexityAIModel'
      title: PerplexityIntegrations
      x-fern-type-name: PerplexityIntegrations
    TogetherAIKeyAuth:
      description: |-
        +label=Together AI API Key Auth
        +icon=together-ai
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for Together AI authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: TogetherAIKeyAuth
      x-fern-type-name: TogetherAIKeyAuth
    TogetherAIIntegrations:
      $ref: '#/components/schemas/TogetherAIModel'
      title: TogetherAIIntegrations
      x-fern-type-name: TogetherAIIntegrations
    NomicKeyAuth:
      description: |-
        +label=Nomic API Key Auth
        +icon=nomic
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for Nomic authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: NomicKeyAuth
      x-fern-type-name: NomicKeyAuth
    NomicIntegrations:
      $ref: '#/components/schemas/NomicModel'
      title: NomicIntegrations
      x-fern-type-name: NomicIntegrations
    PalmKeyAuth:
      description: |-
        +label=PaLM API Key Auth
        +icon=palm
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=Your PaLM API key for authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: PalmKeyAuth
      x-fern-type-name: PalmKeyAuth
    PalmIntegrations:
      $ref: '#/components/schemas/PalmModel'
      title: PalmIntegrations
      x-fern-type-name: PalmIntegrations
    OllamaKeyAuth:
      description: |-
        +label=Ollama API Key Auth
        +icon=ollama
      type: object
      required:
        - custom_host
      properties:
        custom_host:
          description: |-
            +label=Custom Host
            +sort=100
            +usage=The custom host URL for Ollama API
            +message=Host URL must not be empty
          type: string
          pattern: ^\S+$
      title: OllamaKeyAuth
      x-fern-type-name: OllamaKeyAuth
    OllamaIntegrations:
      $ref: '#/components/schemas/OllamaModel'
      title: OllamaIntegrations
      x-fern-type-name: OllamaIntegrations
    OpenaiApiKeyAuth:
      description: |-
        +label=OpenAI API Key Auth
        +icon=openai
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=Your OpenAI API key for authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: OpenaiApiKeyAuth
      x-fern-type-name: OpenaiApiKeyAuth
    OpenAIIntegrations:
      $ref: '#/components/schemas/OpenAIModel'
      title: OpenAIIntegrations
      x-fern-type-name: OpenAIIntegrations
    DatabricksServicePrincipalAuth:
      description: >-
        +label=Databricks Service Principal Auth

        +usage=You can authenticate using Databricks service principal. To
        generate OAuth secret and Client ID, go to Databricks workspace ->
        Workspace Settings -> Identity and Access -> Service Principals ->
        Secrets.
      type: object
      required:
        - type
        - oauth_secret
        - client_id
      properties:
        type:
          description: +value=service-principal
          type: string
          enum:
            - service-principal
        oauth_secret:
          description: |-
            +label=OAuth Secret
            +usage=OAuth secret of the Databricks service principal.
            +sort=200
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
        client_id:
          description: >-
            +label=Client ID

            +usage=Client ID or Application ID of the Databricks service
            principal.

            +sort=100

            +uiType=SecretSelect

            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: DatabricksServicePrincipalAuth
      x-fern-type-name: DatabricksServicePrincipalAuth
    DatabricksApiKeyAuth:
      description: +label=Databricks API Key Based Auth
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +usage=API key for Databricks authentication.
            +sort=100
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: DatabricksApiKeyAuth
      x-fern-type-name: DatabricksApiKeyAuth
    DatabricksIntegrations:
      $ref: '#/components/schemas/DatabricksModel'
      title: DatabricksIntegrations
      x-fern-type-name: DatabricksIntegrations
    ClouderaWorkloadAuthToken:
      description: >-
        +label=Cloudera CDP Workload Auth Token

        +usage=Authenticate using Cloudera CDP workload authentication
        (recommended). This method generates a secure workload token using your
        API access key and private key, allowing you to securely access CDP
        workloads and services.
      type: object
      required:
        - type
        - cdp_base_url
        - access_key_id
        - private_key
        - workload_name
      properties:
        type:
          description: +value=workload-auth-token
          type: string
          enum:
            - workload-auth-token
        cdp_base_url:
          description: |-
            +label=CDP Base URL
            +usage=The base URL for CDP API services
            +sort=50
            +message=CDP Base URL must not be empty
          type: string
          pattern: ^\S+$
        access_key_id:
          description: >-
            +label=Access Key ID

            +usage=The access key ID for your Cloudera CDP workload
            authentication.

            +sort=100

            +uiType=SecretSelect

            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
        private_key:
          description: >-
            +label=Private Key

            +usage=The private key for your Cloudera CDP workload
            authentication.

            +sort=200

            +uiType=SecretSelect

            +uiProps={"isPassword": true, "enableNew": true}

            +message=Private key must not be empty
          type: string
          pattern: \S
        workload_name:
          description: |-
            +label=Workload Name
            +usage=The name of the CDP workload you want to authenticate with.
            +sort=300
            +message=Workload name must not be empty
          type: string
          pattern: ^\S+$
      title: ClouderaWorkloadAuthToken
      x-fern-type-name: ClouderaWorkloadAuthToken
    ClouderaTokenAuth:
      description: >-
        +label=Cloudera Token Auth

        +usage=Authenticate using a pre-generated API token. Alternative to
        workload authentication, this method uses a directly provided API token
        for accessing Cloudera CDP services.
      type: object
      required:
        - type
        - api_token
      properties:
        type:
          description: +value=token
          type: string
          enum:
            - token
        api_token:
          description: >-
            +label=API Token

            +usage=Pre-generated API token for Cloudera CDP authentication. You
            can generate this token using the CDP CLI or from your user profile
            settings. This token provides direct access to CDP services without
            requiring workload-specific authentication.

            +sort=100

            +uiType=SecretSelect

            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: ClouderaTokenAuth
      x-fern-type-name: ClouderaTokenAuth
    ClouderaIntegrations:
      $ref: '#/components/schemas/ClouderaModel'
      title: ClouderaIntegrations
      x-fern-type-name: ClouderaIntegrations
    SelfHostedModelIntegrations:
      $ref: '#/components/schemas/SelfHostedModel'
      title: SelfHostedModelIntegrations
      x-fern-type-name: SelfHostedModelIntegrations
    OpenRouterApiKeyAuth:
      description: |-
        +label=OpenRouter API Key Auth
        +icon=openrouter
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for OpenRouter authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: OpenRouterApiKeyAuth
      x-fern-type-name: OpenRouterApiKeyAuth
    OpenRouterIntegrations:
      $ref: '#/components/schemas/OpenRouterModel'
      title: OpenRouterIntegrations
      x-fern-type-name: OpenRouterIntegrations
    SambaNovaKeyAuth:
      description: |-
        +label=SambaNova Key Auth
        +icon=sambanova
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for SambaNova authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: SambaNovaKeyAuth
      x-fern-type-name: SambaNovaKeyAuth
    SambaNovaIntegrations:
      $ref: '#/components/schemas/SambaNovaModel'
      title: SambaNovaIntegrations
      x-fern-type-name: SambaNovaIntegrations
    XAIKeyAuth:
      description: |-
        +label=xAI API Key Auth
        +usage=Authentication method using xAI API key
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +sort=100
            +usage=The API key for xAI authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: XAIKeyAuth
      x-fern-type-name: XAIKeyAuth
    XAIIntegrations:
      $ref: '#/components/schemas/XAIModel'
      title: XAIIntegrations
      x-fern-type-name: XAIIntegrations
    BasetenKeyAuth:
      description: |-
        +label=Baseten API Key Auth
        +usage=Authentication method using Baseten API key
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +sort=100
            +usage=The API key for Baseten authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: BasetenKeyAuth
      x-fern-type-name: BasetenKeyAuth
    BasetenIntegrations:
      $ref: '#/components/schemas/BasetenModel'
      title: BasetenIntegrations
      x-fern-type-name: BasetenIntegrations
    SagemakerAssumedRoleBasedAuth:
      $ref: '#/components/schemas/AwsBedrockAssumedRoleBasedAuth'
      title: SagemakerAssumedRoleBasedAuth
      x-fern-type-name: SagemakerAssumedRoleBasedAuth
    SagemakerModel:
      description: |-
        +label=Sagemaker Model
        +icon=aws-sagemaker
      type: object
      required:
        - name
        - model_id
        - type
        - model_types
      properties:
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The SageMaker endpoint name you created for your model
            deployment. This is the name of the inference endpoint in AWS
            SageMaker, not the model name.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        type:
          description: +value=integration/model/sagemaker
          type: string
          enum:
            - integration/model/sagemaker
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Sagemaker model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        region:
          $ref: '#/components/schemas/AWSRegion'
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=6

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: SagemakerModel
      x-fern-type-name: SagemakerModel
    CerebrasKeyAuth:
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=API Key
            +sort=100
            +usage=The API key for Cerebras authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: CerebrasKeyAuth
      x-fern-type-name: CerebrasKeyAuth
    CerebrasIntegrations:
      $ref: '#/components/schemas/CerebrasModel'
      title: CerebrasIntegrations
      x-fern-type-name: CerebrasIntegrations
    SnowflakeCortexPatTokenAuth:
      type: object
      required:
        - type
        - pat_token
      properties:
        type:
          description: +value=pat-token
          type: string
          enum:
            - pat-token
        pat_token:
          description: |-
            +label=PAT Token
            +sort=100
            +usage=PAT token for Snowflake Cortex authentication
            +message=PAT token must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: SnowflakeCortexPatTokenAuth
      x-fern-type-name: SnowflakeCortexPatTokenAuth
    SnowflakeCortexIntegrations:
      $ref: '#/components/schemas/SnowflakeCortexModel'
      title: SnowflakeCortexIntegrations
      x-fern-type-name: SnowflakeCortexIntegrations
    ElevenLabsApiKeyAuth:
      description: |-
        +label=ElevenLabs API Key Auth
        +icon=https://assets.production.truefoundry.com/elevenlabs.svg
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=XI API Key
            +sort=100
            +usage=The XI API key for ElevenLabs authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: ElevenLabsApiKeyAuth
      x-fern-type-name: ElevenLabsApiKeyAuth
    ElevenLabsModel:
      description: |-
        +label=ElevenLabs Model
        +icon=https://assets.production.truefoundry.com/elevenlabs.svg
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/elevenlabs
          type: string
          enum:
            - integration/model/elevenlabs
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The ElevenLabs model identifier (for example,
            eleven_multilingual_v2, eleven_turbo_v2_5)

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: >-
            +label=Model Types

            +sort=3

            +usage=Specify the type of the ElevenLabs model (text_to_speech or
            audio_transcription)

            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - text_to_speech
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: ElevenLabsModel
      x-fern-type-name: ElevenLabsModel
    DeepgramApiKeyAuth:
      description: |-
        +label=Deepgram API Key Auth
        +icon=https://assets.production.truefoundry.com/deepgram.svg
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=Deepgram API Key
            +sort=100
            +usage=The API key for Deepgram authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: DeepgramApiKeyAuth
      x-fern-type-name: DeepgramApiKeyAuth
    DeepgramModel:
      description: |-
        +label=Deepgram Model
        +icon=https://assets.production.truefoundry.com/deepgram.svg
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/deepgram
          type: string
          enum:
            - integration/model/deepgram
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Deepgram model identifier (for example, nova-2,
            nova-2-general, aura-asteria-en)

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: >-
            +label=Model Types

            +sort=3

            +usage=Specify the type of the Deepgram model (text_to_speech or
            audio_transcription)

            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - audio_transcription
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: DeepgramModel
      x-fern-type-name: DeepgramModel
    CartesiaApiKeyAuth:
      description: |-
        +label=Cartesia API Key Auth
        +icon=https://assets.production.truefoundry.com/cartesia.svg
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +label=Cartesia API Key
            +sort=100
            +usage=The API key for Cartesia authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: CartesiaApiKeyAuth
      x-fern-type-name: CartesiaApiKeyAuth
    CartesiaModel:
      description: |-
        +label=Cartesia Model
        +icon=https://assets.production.truefoundry.com/cartesia.svg
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/cartesia
          type: string
          enum:
            - integration/model/cartesia
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Cartesia model identifier (for example, sonic-2,
            sonic-multilingual)

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: >-
            +label=Model Types

            +sort=3

            +usage=Specify the type of the Cartesia model (text_to_speech or
            audio_transcription)

            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - text_to_speech
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: CartesiaModel
      x-fern-type-name: CartesiaModel
    VirtualModel:
      description: +label=Virtual Model
      type: object
      required:
        - name
        - type
        - model_types
        - routing_config
      properties:
        name:
          description: >-
            +label=Name

            +sort=1

            +usage=The name of the Virtual Model

            +message=3 to 64 characters long alphanumeric word, may contain - , 
            _ or . in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-_.]{1,62}[a-z0-9]$
        type:
          description: +value=integration/model/virtual
          type: string
          enum:
            - integration/model/virtual
        model_types:
          description: |-
            +label=Model Types
            +sort=2
            +usage=Specify the type of the Virtual Model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/VirtualModelModelType'
          default:
            - chat
        slug:
          description: >-
            +label=Slug

            +sort=1

            +usage=It should be a unique identifier for the virtual model. It is
            used to uniquely identify the virtual model. It can be used in place
            of virtualGroupName/virtualModelName.

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        routing_config:
          $ref: '#/components/schemas/HeaderRoutingConfig'
      title: VirtualModel
      x-fern-type-name: VirtualModel
    CustomHeaderAuth:
      description: >-
        +label=Header Auth

        +usage=Static API key or token authentication via request headers. All
        users share the same credentials.
      type: object
      required:
        - type
        - headers
      properties:
        type:
          description: +value=header
          type: string
          enum:
            - header
        headers:
          description: >-
            +label=Headers

            +usage=Headers passed to the generic secret server with every
            request.
          type: object
          additionalProperties:
            type: string
            pattern: ^\S+$
      title: CustomHeaderAuth
      x-fern-type-name: CustomHeaderAuth
    CustomEndpointIntegrations:
      $ref: '#/components/schemas/CustomEndpoint'
      title: CustomEndpointIntegrations
      x-fern-type-name: CustomEndpointIntegrations
    MCPServerIntegration:
      description: |-
        +label=MCP Server
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - description
        - url
        - transport
      properties:
        type:
          description: +value=integration/mcp-server/remote
          type: string
          enum:
            - integration/mcp-server/remote
        name:
          description: |-
            +label=Name
            +usage=The name of the MCP Server.
            +sort=100
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: >-
            +label=Description

            +usage=Provide a brief description of the purpose of this MCP
            Server.

            +uiType=TextArea

            +message=1 to 1000 characters long, may contain any character except
            newlines

            +sort=200

            +uiProps={"descriptionInline":true}
          type: string
          minLength: 1
          maxLength: 1000
        url:
          description: |-
            +label=URL
            +usage=The endpoint URL for the MCP Server.
            +sort=300
          type: string
          pattern: ^https?:\/\/[^\s]+$
        transport:
          description: >-
            +label=Transport

            +usage=The transport protocol to use for communication with the MCP
            Server.

            +sort=350

            +uiType=RadioCard
          type: string
          enum:
            - streamable-http
            - sse
        auth_data:
          $ref: '#/components/schemas/MCPServerAuth'
        authorized_subjects:
          description: |-
            +label=Access Control
            +usage=List of teams/users who can use this MCP server.
            +sort=500
            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: MCPServerIntegration
      x-fern-type-name: MCPServerIntegration
    VirtualMCPServerIntegration:
      description: |-
        +label=Virtual MCP Server
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - description
        - servers
      properties:
        type:
          description: +value=integration/mcp-server/virtual
          type: string
          enum:
            - integration/mcp-server/virtual
        name:
          description: |-
            +label=Name
            +usage=The name of the Virtual MCP Server.
            +sort=100
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: >-
            +label=Description

            +usage=Provide a brief description of the purpose of this Virtual
            MCP Server.

            +uiType=TextArea

            +message=1 to 1000 characters long, may contain any character except
            newlines

            +sort=200

            +uiProps={"descriptionInline":true}
          type: string
          minLength: 1
          maxLength: 1000
        servers:
          description: |-
            +label=Source MCP Servers
            +usage=List of source MCP servers to include in this virtual server.
            +sort=300
          type: array
          items:
            $ref: '#/components/schemas/VirtualMCPServerSource'
        authorized_subjects:
          description: |-
            +label=Access Control
            +usage=List of teams/users who can use this MCP server.
            +sort=400
            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: VirtualMCPServerIntegration
      x-fern-type-name: VirtualMCPServerIntegration
    OpenAIModerationsGuardrailConfig:
      description: |-
        +label=OpenAI Moderation
        +icon=openai
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: >-
            OpenAI content moderation for hate, harassment, self-harm, sexual,
            violence, and illicit content
        type:
          description: |-
            +value=integration/guardrail-config/openai-moderations
            +sort=50
            +uiType=Hidden
          type: string
          enum:
            - integration/guardrail-config/openai-moderations
        auth_data:
          $ref: '#/components/schemas/OpenaiApiKeyAuth'
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. OpenAI Moderation
            guardrails can only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=150
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - model
          properties:
            model:
              description: |-
                +label=Model
                +sort=300
                +usage=The model to use for the OpenAI Moderation API.
              type: string
              default: omni-moderation-latest
            base_url:
              description: >-
                +label=Base URL

                +sort=100

                +usage=Optional custom base URL for OpenAI API. If not provided,
                the default base URL will be used.
              type: string
              pattern: ^\S+$
            category_thresholds:
              type: object
              additionalProperties:
                type: number
                minimum: 0
                maximum: 1
      title: OpenAIModerationsGuardrailConfig
      x-fern-type-name: OpenAIModerationsGuardrailConfig
    AWSBedrockGuardrailConfig:
      description: |-
        +label=AWS Bedrock
        +icon=aws-bedrock
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: AWS Bedrock Guardrails for content filtering and safety policies
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/aws-bedrock
          type: string
          enum:
            - integration/guardrail-config/aws-bedrock
        auth_data:
          description: |-
            +label=AWS Account Auth Data
            +usage=Authentication data for the AWS account
            +sort=85
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/AwsBedrockAssumedRoleBasedAuth'
            - $ref: '#/components/schemas/AwsBedrockApiKeyAuth'
          discriminator:
            propertyName: type
        operation:
          description: >-
            +label=Operation

            +usage=The operation type to use for the Guardrail. Validate
            guardrails are used to validate requests and mutate can validate as
            well as mutate requests.

            Validate guardrails are run in parallel while mutate guardrails are
            run sequentially.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
            - mutate
          default: validate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first. Only applicable when operation is mutate.

            +sort=150
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - guardrail_id
            - guardrail_version
            - region
          properties:
            guardrail_id:
              description: >-
                +label=Guardrail ID

                +usage=The unique identifier of the Bedrock Guardrail created in
                AWS console
              type: string
              pattern: ^\S+$
            guardrail_version:
              description: >-
                +label=Guardrail Version

                +usage=Version number of the guardrail to use (e.g., 2 for
                version 2)
              type: string
              pattern: ^\S+$
            region:
              $ref: '#/components/schemas/AWSRegion'
      title: AWSBedrockGuardrailConfig
      x-fern-type-name: AWSBedrockGuardrailConfig
    CustomGuardrailConfig:
      description: |-
        +label=Custom
        +icon=puzzle-piece
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
        - target
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: Custom guardrail server for validate or mutate via HTTP endpoint
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/custom
          type: string
          enum:
            - integration/guardrail-config/custom
        auth_data:
          description: |-
            +label=Auth Data
            +sort=85
            +usage=Authentication data for the Guardrail Server.
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/CustomBasicAuth'
            - $ref: '#/components/schemas/CustomBearerAuth'
          discriminator:
            propertyName: type
        operation:
          description: >-
            +label=Operation

            +usage=The operation type to use for the Guardrail. Validate
            guardrails are used to validate requests and mutate can validate as
            well as mutate requests.

            Validate guardrails are run in parallel while mutate guardrails are
            run sequentially.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
            - mutate
          default: validate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first. Only applicable when operation is mutate.

            +sort=160
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        target:
          description: >-
            +label=Target

            +usage=Specify whether the guardrail should be applied to the
            request or response. Guardrails with target "Request" can be only
            used in input guardrails and guardrails with target "Response" can
            only be used in output guardrails.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=200
          type: string
          enum:
            - request
            - response
          default: request
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - url
          properties:
            url:
              description: |-
                +label=URL
                +sort=300
                +usage=The URL of the Guardrail to send a post request to.
              type: string
              pattern: ^https?:\/\/[^\s\/]+(\/[^\s\/]*)*[^\/\s]$
            headers:
              description: >-
                +label=Headers

                +sort=500

                +usage=Headers for the Guardrail Server. Forwarded to the
                Guardrail Server as is. For example: `{"Authorization": "APIKey
                <token>"}`

                +uiProps={"defaultForCreate": {"": ""}}
              type: object
              additionalProperties:
                type: string
            config:
              description: >-
                +label=Config

                +sort=600

                +uiType=JsonInput

                +usage=The config for the Guardrail Server. This is a JSON
                object that will be sent as a config to Guardrail Server along
                with the request.
              type: object
      title: CustomGuardrailConfig
      x-fern-type-name: CustomGuardrailConfig
    AzurePIIGuardrailConfig:
      description: |-
        +label=Azure PII
        +icon=azure
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: Azure AI Language PII detection and redaction for sensitive data
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/azure-pii
          type: string
          enum:
            - integration/guardrail-config/azure-pii
        auth_data:
          description: |-
            +label=Azure Account Auth Data
            +usage=Authentication data for the Azure account
            +sort=85
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AzureKeyAuth'
            - $ref: '#/components/schemas/AzureCertificateAuth'
          discriminator:
            propertyName: type
        operation:
          description: >-
            +label=Operation

            +usage=The operation type to use for the Guardrail. Validate
            guardrails are used to validate requests and mutate can validate as
            well as mutate requests. Validate guardrails are run in parallel
            while mutate guardrails are run sequentially.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - mutate
            - validate
          default: mutate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first. Only applicable when operation is mutate.

            +sort=150
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - resource_name
            - api_version
            - domain
            - pii_categories
            - model_version
            - language
          properties:
            resource_name:
              description: >-
                +label=Resource Name

                +usage=Name of your Azure AI Language resource where the PII
                detection service is deployed (e.g., my-language-resource)

                +sort=50
              type: string
            api_version:
              description: |-
                +label=API Version
                +usage=API version for the PII detection API
                +sort=100
              type: string
              default: '2024-11-01'
            custom_host:
              description: >-
                +label=Custom Host

                +usage=Custom endpoint URL for the PII detection API (optional,
                uses default Azure endpoint if not specified)

                +sort=200

                +uiType=Hidden
              type: string
              pattern: ^(https?://)\S+$
            domain:
              description: >-
                +label=Domain

                +usage=Specialized domain for PII detection. Use phi for PHI
                (Protected Health Information) or none for general text

                +sort=400

                +uiType=MultiSelectPills

                +uiProps={"hideRadioIcon": true}
              type: string
              enum:
                - none
                - phi
              default: none
            pii_categories:
              description: |-
                +label=PII Categories
                +usage=Categories of PII to detect.
                +sort=500
                +uiType=MultiSelect
                +placeholder=Select PII categories to detect
              type: array
              items:
                $ref: '#/components/schemas/AzurePIICategory'
              default:
                - '*'
            model_version:
              description: >-
                +label=Model Version

                +usage=Version of the PII detection model to use, use latest for
                the newest model or specify a specific version for consistency

                +sort=600
              type: string
              default: latest
            language:
              description: |-
                +label=Language Code
                +usage=Language code for PII detection (e.g., en for English)
                +sort=700
              type: string
              default: en
      title: AzurePIIGuardrailConfig
      x-fern-type-name: AzurePIIGuardrailConfig
    AzureContentSafetyGuardrailConfig:
      description: |-
        +label=Azure Content Safety
        +icon=azure
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: >-
            Azure Content Safety for hate, self-harm, sexual, and violence
            detection
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/azure-content-safety
          type: string
          enum:
            - integration/guardrail-config/azure-content-safety
        auth_data:
          description: |-
            +label=Azure Account Auth Data
            +usage=Authentication data for the Azure account
            +sort=85
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AzureKeyAuth'
            - $ref: '#/components/schemas/AzureCertificateAuth'
          discriminator:
            propertyName: type
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. Azure Content Safety
            guardrails can only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - resource_name
            - api_version
            - severity_threshold
            - categories
          properties:
            resource_name:
              description: >-
                +label=Resource Name

                +usage=Name of your Azure Content Safety resource where the
                service is deployed (e.g., my-content-safety)

                +sort=200
              type: string
            api_version:
              description: |-
                +label=API Version
                +usage=API version for the Content Safety API
                +sort=200
              type: string
              default: '2024-09-01'
            custom_host:
              description: >-
                +label=Custom Host

                +usage=Custom endpoint URL for the Content Safety API (optional,
                uses default Azure endpoint if not specified)

                +sort=300

                +uiType=Hidden
              type: string
              pattern: ^(https?://)\S+$
            blocklist_names:
              description: >-
                +label=Blocklist Names

                +usage=Names of custom blocklists created in Azure Content
                Safety to check text against. Leave empty if not using custom
                blocklists

                +sort=400

                +uiProps={"defaultForCreate": [""]}
              type: array
              items:
                type: string
            severity_threshold:
              description: >-
                +label=Severity

                +usage=Minimum severity level (0-6) to flag content. Higher
                values are more restrictive. 0=Safe, 2=Low risk, 4=Medium risk,
                6=High risk

                +sort=500
              type: number
              default: 2
            categories:
              description: >-
                +label=Categories

                +usage=Types of harmful content to detect: Hate (hate speech),
                SelfHarm (self-injury), Sexual (sexual content), Violence
                (violent content)

                +uiType=Select

                +sort=600
              type: array
              items:
                $ref: '#/components/schemas/AzureContentSafetyCategory'
              default:
                - Hate
                - SelfHarm
                - Sexual
                - Violence
      title: AzureContentSafetyGuardrailConfig
      x-fern-type-name: AzureContentSafetyGuardrailConfig
    AzurePromptShieldGuardrailConfig:
      description: |-
        +label=Azure Prompt Shield
        +icon=azure
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: Azure Prompt Shield for jailbreak and prompt injection detection
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/azure-prompt-shield
          type: string
          enum:
            - integration/guardrail-config/azure-prompt-shield
        auth_data:
          description: |-
            +label=Azure Account Auth Data
            +usage=Authentication data for the Azure account
            +sort=85
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AzureKeyAuth'
            - $ref: '#/components/schemas/AzureCertificateAuth'
          discriminator:
            propertyName: type
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. Azure Prompt Shield
            guardrails can only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - resource_name
            - api_version
          properties:
            resource_name:
              description: >-
                +label=Resource Name

                +usage=Name of your Azure Content Safety resource where the
                Prompt Shield service is deployed (e.g., my-content-safety)

                +sort=200
              type: string
            api_version:
              description: |-
                +label=API Version
                +usage=API version for the Prompt Shield API
                +sort=200
              type: string
              default: '2024-09-01'
            custom_host:
              description: >-
                +label=Custom Host

                +usage=Custom endpoint URL for the Prompt Shield API (optional,
                uses default Azure endpoint if not specified)

                +sort=300

                +uiType=Hidden
              type: string
              pattern: ^(https?://)\S+$
      title: AzurePromptShieldGuardrailConfig
      x-fern-type-name: AzurePromptShieldGuardrailConfig
    EnkryptAIGuardrailConfig:
      description: |-
        +label=Enkrypt AI
        +icon=https://app.enkryptai.com/favicon.ico
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: Enkrypt AI guardrails for content safety and policy enforcement
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/enkrypt-ai
          type: string
          enum:
            - integration/guardrail-config/enkrypt-ai
        auth_data:
          $ref: '#/components/schemas/EnkryptAIKeyAuth'
        operation:
          description: >-
            +label=Operation

            +usage=The operation type to use for the Guardrail. Validate
            guardrails are used to validate requests and mutate can validate as
            well as mutate requests.

            Validate guardrails are run in parallel while mutate guardrails are
            run sequentially.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
            - mutate
          default: validate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first. Only applicable when operation is mutate.

            +sort=160
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - guardrail_name
          properties:
            guardrail_name:
              description: >-
                +label=Guardrail Name

                +usage=The name of the guardrail to use for the Enkrypt AI
                Guardrail.

                +sort=100
              type: string
      title: EnkryptAIGuardrailConfig
      x-fern-type-name: EnkryptAIGuardrailConfig
    PaloAltoPrismaAIRSGuardrailConfig:
      description: >-
        +label=Palo Alto Prisma AIRS

        +icon=https://assets.production.truefoundry.com/palo-alto-prisma-airs.svg
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: Palo Alto Prisma AIRS for AI security and content validation
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/palo-alto-prisma-airs
          type: string
          enum:
            - integration/guardrail-config/palo-alto-prisma-airs
        auth_data:
          $ref: '#/components/schemas/PaloAltoPrismaAIRSKeyAuth'
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. Palo Alto Prisma AIRS
            guardrails can only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - profile_name
          properties:
            base_url:
              description: >-
                +label=Base URL

                +usage=Optional override for the base URL for Palo Alto Prisma
                AIRS. Defaults to
                https://service.api.aisecurity.paloaltonetworks.com/v1/ if not
                provided.

                +sort=100
              type: string
              pattern: ^\S+$
            profile_name:
              description: |-
                +label=Profile Name
                +usage=The profile name associated with this Guardrail Config.
                +sort=60
              type: string
            mode:
              description: >-
                +label=Mode

                +usage=Execution mode for the guardrail. Sync waits for the
                guardrail check to complete before proceeding. Async triggers
                the check without waiting. Defaults to sync.

                +sort=70

                +uiType=Select
              type: string
              enum:
                - sync
                - async
              default: sync
            metadata_key_mapping:
              description: >-
                +label=Metadata Key Mapping

                +usage=Map Palo Alto AIRS metadata fields to TrueFoundry request
                metadata keys. Key is the Palo Alto AIRS metadata field name,
                value is the corresponding key from X-TFY-METADATA header on the
                TrueFoundry side. Example: {"app_user": "user_email",
                "ai_model": "model_name"}. Reserved AIRS keys: ai_model,
                app_user, app_name.
                [Docs](https://www.truefoundry.com/docs/ai-gateway/palo-alto-airs)

                +sort=250

                +uiType=KV

                +uiProps={}
              type: object
              additionalProperties:
                type: string
      title: PaloAltoPrismaAIRSGuardrailConfig
      x-fern-type-name: PaloAltoPrismaAIRSGuardrailConfig
    FiddlerGuardrailConfig:
      description: |-
        +label=Fiddler
        +icon=https://assets.production.truefoundry.com/fiddler.svg
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: >-
            Fiddler for harmful content detection or response faithfulness
            checks
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/fiddler
          type: string
          enum:
            - integration/guardrail-config/fiddler
        auth_data:
          $ref: '#/components/schemas/FiddlerKeyAuth'
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. Fiddler guardrails can
            only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=120
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - guard_type
          properties:
            guard_type:
              $ref: '#/components/schemas/FiddlerGuardType'
            threshold:
              description: >-
                +label=Threshold

                +usage=Confidence threshold (0.0-1.0) for flagging content.
                Content scoring above the threshold will be flagged

                +sort=200
              type: number
              minimum: 0
              maximum: 1
      title: FiddlerGuardrailConfig
      x-fern-type-name: FiddlerGuardrailConfig
    PangeaGuardrailConfig:
      description: |-
        +label=Pangea
        +icon=https://assets.production.truefoundry.com/pangea-logo.svg
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: Pangea textGuard for text security or PII detection and redaction
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/pangea
          type: string
          enum:
            - integration/guardrail-config/pangea
        auth_data:
          $ref: '#/components/schemas/PangeaKeyAuth'
        operation:
          description: >-
            +label=Operation

            +usage=The operation type to use for the Guardrail. Validate
            guardrails are used to validate requests and mutate can validate as
            well as mutate requests.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=130
          type: string
          enum:
            - validate
            - mutate
          default: validate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first. Only applicable when operation is mutate.

            +sort=160
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - guard_type
            - domain
          properties:
            guard_type:
              $ref: '#/components/schemas/PangeaGuardType'
            domain:
              description: >-
                +label=Domain

                +usage=Domain of the cloud provider and region where your Pangea
                project is configured. Example: if endpoint is:
                https://<service_name>.aws.us-west-2.pangea.cloud/v1/text/guard,
                the input should be: aws.us-west-2.pangea.cloud

                +message=Must be a valid domain name.

                +sort=200
              type: string
              pattern: ^[a-z0-9-]+\.[a-z0-9-]+\.pangea\.cloud$
            recipe:
              description: >-
                +label=Recipe

                +usage=Recipe key from Pangea console defining security rules
                and data types to apply

                +sort=200
              type: string
            overrides:
              description: |-
                +label=Overrides
                +usage=Overrides for the Pangea account
                +sort=300
                +uiType=KV
                +uiProps={"defaultForCreate": {"": ""}}
              type: object
              additionalProperties:
                type: string
      title: PangeaGuardrailConfig
      x-fern-type-name: PangeaGuardrailConfig
    PatronusGuardrailConfig:
      description: |-
        +label=Patronus
        +icon=https://assets.production.truefoundry.com/patronus.svg
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: >-
            Patronus evaluators for relevance, safety, PII, PHI, toxicity, and
            more
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/patronus
          type: string
          enum:
            - integration/guardrail-config/patronus
        auth_data:
          $ref: '#/components/schemas/PatronusKeyAuth'
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. Patronus guardrails
            can only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - target
            - evaluators
          properties:
            target:
              description: >-
                +label=Target

                +usage=Where to apply evaluation: request (evaluate user input)
                or response (evaluate model response)

                +sort=100

                +uiType=Select
              type: string
              enum:
                - response
                - request
              default: response
            evaluators:
              description: >-
                +label=Evaluators

                +usage=Patronus evaluators guard against GenAI failures by
                checking for evaluation criteria, refer to
                [docs](https://docs.patronus.ai/docs/evaluators/reference_guide)

                +sort=200
              type: array
              items:
                $ref: '#/components/schemas/PatronusEvaluator'
      title: PatronusGuardrailConfig
      x-fern-type-name: PatronusGuardrailConfig
    SecretDetectionGuardrailConfig:
      description: |-
        +label=Secrets Detection
        +icon=truefoundry
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: >-
            Detects likely credentials/secrets (AWS keys, API keys, JWT tokens,
            private keys)
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/secret-detection
          type: string
          enum:
            - integration/guardrail-config/secret-detection
        operation:
          description: >-
            +label=Operation

            +usage=The operation type to use for the Guardrail. Validate
            guardrails are used to validate requests and mutate can validate as
            well as mutate requests. Validate guardrails are run in parallel
            while mutate guardrails are run sequentially.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
            - mutate
          default: validate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first. Only applicable when operation is mutate.

            +sort=110
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          properties:
            redaction_text:
              description: >-
                +label=Redaction Text

                +usage=Text to use when redacting detected secrets in logs and
                error messages. Defaults to '[REDACTED]'.Only applicable in
                mutate mode.

                +sort=200
              type: string
              default: '[REDACTED]'
      title: SecretDetectionGuardrailConfig
      x-fern-type-name: SecretDetectionGuardrailConfig
    CodeSafetyLinterGuardrailConfig:
      description: |-
        +label=Code Safety Linter
        +icon=truefoundry
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: >-
            Detects unsafe code patterns in tool outputs (eval, exec, os.system,
            subprocess, rm -rf)
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/code-safety-linter
          type: string
          enum:
            - integration/guardrail-config/code-safety-linter
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. Code Safety Linter
            guardrails can only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"disableEdit": true , "hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
      title: CodeSafetyLinterGuardrailConfig
      x-fern-type-name: CodeSafetyLinterGuardrailConfig
    SQLSanitizerGuardrailConfig:
      description: |-
        +label=SQL Sanitizer
        +icon=truefoundry
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: >-
            Detects and sanitizes risky SQL patterns (DROP, TRUNCATE,
            DELETE/UPDATE without WHERE, string interpolation)
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/sql-sanitizer
          type: string
          enum:
            - integration/guardrail-config/sql-sanitizer
        operation:
          description: >-
            +label=Operation

            +usage=Validate (detect and block) or Mutate (detect, sanitize
            comments, and continue).

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
            - mutate
          default: validate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first. Only applicable when operation is mutate.

            +sort=110
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          type: object
          properties:
            block_drop:
              description: |-
                +label=Block DROP Statements
                +usage=Block SQL DROP statements as dangerous operations.
                +sort=101
              type: boolean
              default: true
            block_truncate:
              description: |-
                +label=Block TRUNCATE Statements
                +usage=Block SQL TRUNCATE statements as dangerous operations.
                +sort=102
              type: boolean
              default: true
            block_alter:
              description: |-
                +label=Block ALTER Statements
                +usage=Block SQL ALTER statements as dangerous operations.
                +sort=103
              type: boolean
              default: true
            block_grant:
              description: |-
                +label=Block GRANT Statements
                +usage=Block SQL GRANT statements as dangerous operations.
                +sort=104
              type: boolean
              default: true
            block_revoke:
              description: |-
                +label=Block REVOKE Statements
                +usage=Block SQL REVOKE statements as dangerous operations.
                +sort=105
              type: boolean
              default: true
            strip_comments:
              description: >-
                +label=Strip SQL Comments

                +usage=Remove SQL comments (-- and block comments) from queries.
                In mutate mode, comments are stripped from the output.

                +sort=200
              type: boolean
              default: true
            block_delete_without_where:
              description: >-
                +label=Block DELETE without WHERE

                +usage=Flag DELETE FROM statements that lack a WHERE clause as
                potentially dangerous.

                +sort=300
              type: boolean
              default: true
            block_update_without_where:
              description: >-
                +label=Block UPDATE without WHERE

                +usage=Flag UPDATE statements that lack a WHERE clause as
                potentially dangerous.

                +sort=400
              type: boolean
              default: true
            require_parameterization:
              description: >-
                +label=Require Parameterization

                +usage=Detect naive string interpolation patterns (+, %., {})
                that may indicate SQL injection vulnerabilities.

                +sort=500
              type: boolean
              default: false
      title: SQLSanitizerGuardrailConfig
      x-fern-type-name: SQLSanitizerGuardrailConfig
    RegexGuardrailConfig:
      description: |-
        +label=Regex Pattern Match
        +icon=truefoundry
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: Regex pattern matching for content filtering or redaction
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/regex
          type: string
          enum:
            - integration/guardrail-config/regex
        operation:
          description: >-
            +label=Operation

            +usage=Validate blocks when pattern matches. Mutate replaces matched
            text and continues.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
            - mutate
          default: validate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first. Only applicable when operation is mutate.

            +sort=110
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          properties:
            preset_patterns:
              $ref: '#/components/schemas/PresetRegexPattern'
            custom_patterns:
              description: >-
                +label=Custom Patterns

                +usage=Enter your own regex patterns to match against content,
                each with its own redaction text (used in mutate mode).

                +sort=200
              type: array
              items:
                $ref: '#/components/schemas/CustomRegexPattern'
      title: RegexGuardrailConfig
      x-fern-type-name: RegexGuardrailConfig
    TfyContentModerationGuardrailConfig:
      description: |-
        +label=Content Moderation
        +icon=truefoundry
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: Content moderation guardrail, managed by TrueFoundry
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/tfy-content-moderation
          type: string
          enum:
            - integration/guardrail-config/tfy-content-moderation
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. TrueFoundry Content
            Moderation guardrails can only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - severity_threshold
            - categories
          properties:
            severity_threshold:
              description: >-
                +label=Severity Threshold

                +usage=Minimum severity level (0-6) to flag content. 0=Safe,
                2=Low risk, 4=Medium risk, 6=High risk

                +sort=500
              type: number
              default: 2
            categories:
              description: >-
                +label=Categories

                +usage=Types of harmful content to detect: Hate (hate speech),
                SelfHarm (self-injury), Sexual (sexual content), Violence
                (violent content)

                +uiType=Select

                +sort=600
              type: array
              items:
                $ref: '#/components/schemas/AzureContentSafetyCategory'
              default:
                - Hate
                - SelfHarm
                - Sexual
                - Violence
      title: TfyContentModerationGuardrailConfig
      x-fern-type-name: TfyContentModerationGuardrailConfig
    TfyPIIGuardrailConfig:
      description: |-
        +label=PII/PHI
        +icon=truefoundry
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: PII detection and redaction guardrail, managed by TrueFoundry
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/tfy-pii
          type: string
          enum:
            - integration/guardrail-config/tfy-pii
        operation:
          description: >-
            +label=Operation

            +usage=The operation type to use for the Guardrail. Validate
            guardrails are used to validate requests and mutate can validate as
            well as mutate requests. Validate guardrails are run in parallel
            while mutate guardrails are run sequentially.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - mutate
            - validate
          default: mutate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first.

            +sort=150
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - pii_categories
          properties:
            pii_categories:
              description: |-
                +label=PII Categories
                +usage=Categories of PII to detect.
                +sort=500
                +uiType=MultiSelectPills
                +placeholder=Select PII categories to detect
              type: array
              items:
                $ref: '#/components/schemas/AzurePIICategory'
              default:
                - '*'
      title: TfyPIIGuardrailConfig
      x-fern-type-name: TfyPIIGuardrailConfig
    TfyPromptInjectionGuardrailConfig:
      description: |-
        +label=Prompt Injection
        +icon=truefoundry
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: Prompt injection guardrail, managed by TrueFoundry
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/tfy-prompt-injection
          type: string
          enum:
            - integration/guardrail-config/tfy-prompt-injection
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. TrueFoundry Prompt
            Shield guardrails can only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
      title: TfyPromptInjectionGuardrailConfig
      x-fern-type-name: TfyPromptInjectionGuardrailConfig
    CedarGuardrailConfig:
      description: |-
        +label=Cedar
        +icon=truefoundry
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: Enforces access control policies using Cedar policy language
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/cedar
          type: string
          enum:
            - integration/guardrail-config/cedar
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. Cedar guardrails can
            only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - policy
          properties:
            policy:
              description: >-
                +label=Policy

                +usage=Enter a Cedar policy string to define access control
                rules. Supports principal types (user, team, virtualaccount),
                MCPTool actions, MCPServer resources, and when/unless clauses,
                refer to
                [Docs](https://www.truefoundry.com/docs/ai-gateway/cedar-guardrails)

                +uiType=TextArea

                +uiProps={"descriptionInline":true}

                +sort=200
              type: string
      title: CedarGuardrailConfig
      x-fern-type-name: CedarGuardrailConfig
    OPAGuardrailConfig:
      description: |-
        +label=OPA
        +icon=truefoundry
      type: object
      required:
        - name
        - type
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: >-
            Open Policy Agent (OPA) for policy-based access control and
            validation
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/opa
          type: string
          enum:
            - integration/guardrail-config/opa
        auth_data:
          $ref: '#/components/schemas/OPAAuth'
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. OPA guardrails can
            only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - opa_url
            - policy_path
          properties:
            opa_url:
              description: >-
                +label=OPA URL

                +usage=The base URL of the OPA server (e.g.,
                https://opa.example.com:8181). The policy will be queried at
                {opa_url}/v1/data/{policy_path}.

                +sort=100
              type: string
              pattern: ^https?:\/\/[^\s\/]+(\/[^\s\/]*)*[^\/\s]?$
            policy_path:
              description: >-
                +label=Policy Path

                +usage=The policy path to evaluate in OPA (e.g., "authz/allow",
                "guardrails/validate"). This corresponds to the package and rule
                name in your Rego policy. The full query URL will be
                {opa_url}/v1/data/{policy_path}.

                +sort=200
              type: string
              pattern: ^[a-zA-Z_][a-zA-Z0-9_]*(\/[a-zA-Z_][a-zA-Z0-9_]*)*$
      title: OPAGuardrailConfig
      x-fern-type-name: OPAGuardrailConfig
    GoogleModelArmorGuardrailConfig:
      description: |-
        +label=Google Model Armor
        +icon=googleCloud
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: >-
            Google Cloud Model Armor for prompt injection, harmful content, PII,
            and malicious URI detection
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/google-model-armor
          type: string
          enum:
            - integration/guardrail-config/google-model-armor
        auth_data:
          description: |-
            +label=Google Cloud Auth Data
            +usage=Authentication credentials for Google Cloud Model Armor
            +sort=85
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/GoogleModelArmorApiKeyAuth'
            - $ref: '#/components/schemas/GoogleModelArmorKeyFileAuth'
          discriminator:
            propertyName: type
        operation:
          description: >-
            +label=Operation

            +usage=The operation type to use for the Guardrail. Validate
            guardrails are used to validate requests and mutate can validate as
            well as mutate requests.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
            - mutate
          default: validate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first. Only applicable when operation is mutate.

            +sort=150
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - project_id
            - location
            - template_id
          properties:
            project_id:
              description: |-
                +label=Project ID
                +usage=The Google Cloud project ID where Model Armor is enabled
                +sort=100
              type: string
              pattern: ^\S+$
            location:
              $ref: '#/components/schemas/GCPRegion'
            template_id:
              description: >-
                +label=Template ID

                +usage=The Model Armor template name that defines which filters
                and confidence thresholds to apply (e.g. my-template)

                +sort=300
              type: string
              pattern: ^\S+$
      title: GoogleModelArmorGuardrailConfig
      x-fern-type-name: GoogleModelArmorGuardrailConfig
    GraySwanCygnalGuardrailConfig:
      description: |-
        +label=GraySwan Cygnal
        +icon=https://assets.production.truefoundry.com/grayswan.ico
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: GraySwan Cygnal for policy violation and content safety monitoring
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/grayswan-cygnal
          type: string
          enum:
            - integration/guardrail-config/grayswan-cygnal
        auth_data:
          $ref: '#/components/schemas/GraySwanCygnalApiKeyAuth'
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. GraySwan Cygnal
            guardrails can only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - reasoning_mode
          properties:
            base_url:
              description: >-
                +label=Base URL

                +usage=Optional override for the Cygnal monitor endpoint.
                Defaults to https://api.grayswan.ai/cygnal/monitor if not
                provided.

                +sort=50
              type: string
              pattern: ^\S+$
            policy_ids:
              description: >-
                +label=Policy IDs

                +usage=Custom policy IDs to use for monitoring. Rules from all
                policies are merged in order, with earlier policies taking
                precedence. If not provided, the default Basic Content Safety
                policy is applied, refer to
                [Docs](https://docs.grayswan.ai/cygnal/monitor-requests)

                +sort=100
              type: array
              items:
                type: string
            rules:
              description: >-
                +label=Rules

                +usage=Custom rule definitions for monitoring. Each key is a
                rule name and its value is the rule description, refer to
                [Docs](https://docs.grayswan.ai/cygnal/monitor-requests)

                +sort=200

                +uiType=KV

                +uiProps={"defaultForCreate": {"": ""}}
              type: object
              additionalProperties:
                type: string
            reasoning_mode:
              description: >-
                +label=Reasoning Mode

                +usage=Controls whether Cygnal uses internal reasoning steps
                before determining if content violates policy. "off" is fastest,
                "hybrid" adds moderate reasoning, "thinking" performs guided
                reasoning for highest quality.

                +sort=300

                +uiType=Select
              type: string
              enum:
                - 'off'
                - hybrid
                - thinking
              default: 'off'
      title: GraySwanCygnalGuardrailConfig
      x-fern-type-name: GraySwanCygnalGuardrailConfig
    AktoGuardrailConfig:
      description: |-
        +label=Akto
        +icon=https://assets.production.truefoundry.com/akto.svg
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: >-
            Akto guardrail for LLM security, prompt injection detection, and
            policy violation monitoring
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/akto
          type: string
          enum:
            - integration/guardrail-config/akto
        auth_data:
          $ref: '#/components/schemas/AktoTokenAuth'
        operation:
          description: >-
            +label=Operation

            +usage=The operation type for this guardrail. Akto guardrails can
            only be used for validation.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - base_url
          properties:
            base_url:
              description: >-
                +label=Base URL

                +usage=The base URL of the Akto guardrails service (e.g.
                https://your-akto-host:port)

                +sort=100
              type: string
              pattern: ^https?:\/\/\S+$
      title: AktoGuardrailConfig
      x-fern-type-name: AktoGuardrailConfig
    TrojAIGuardrailConfig:
      description: |-
        +label=TrojAI
        +icon=https://assets.production.truefoundry.com/trojai.svg
      type: object
      required:
        - name
        - type
        - auth_data
        - operation
        - enforcing_strategy
        - config
      properties:
        name:
          description: >-
            +label=Name

            +sort=50

            +usage=The name of the Guardrail Config.

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number

            +uiProps={"disableEdit":true}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +sort=75
            +usage=Optional description for this Guardrail Config.
          type: string
          default: TrojAI DEFEND firewall for real-time AI security
        type:
          description: |-
            +uiType=Hidden
            +value=integration/guardrail-config/trojai
          type: string
          enum:
            - integration/guardrail-config/trojai
        auth_data:
          $ref: '#/components/schemas/TrojAIClientIdAuth'
        operation:
          description: >-
            +label=Operation

            +usage=The operation type to use for the Guardrail. Validate
            guardrails are used to validate requests and mutate can validate as
            well as mutate requests.

            Validate guardrails are run in parallel while mutate guardrails are
            run sequentially.

            +uiType=MultiSelectPills

            +uiProps={"hideRadioIcon": true}

            +sort=100
          type: string
          enum:
            - validate
            - mutate
          default: validate
        priority:
          description: >-
            +label=Priority

            +usage=Execution order for mutate guardrails. Lower values run
            first. Only applicable when operation is mutate.

            +sort=150
          type: integer
          minimum: 0
          exclusiveMinimum: true
          default: 1
        enforcing_strategy:
          $ref: '#/components/schemas/EnforcingStrategy'
        config:
          description: |-
            +uiType=Ignore
            +uiProps={"forwardJsonKey": true}
          type: object
          required:
            - base_url
          properties:
            base_url:
              description: >-
                +label=Base URL

                +usage=The base URL of the TrojAI firewall (e.g.
                https://your-trojai-firewall-host)

                +sort=100
              type: string
              pattern: ^https?:\/\/\S+$
      title: TrojAIGuardrailConfig
      x-fern-type-name: TrojAIGuardrailConfig
    BuildSecret:
      type: object
      required:
        - id
        - value
      properties:
        id:
          description: |-
            +label=ID
            +usage=The ID of the build secret.
            +sort=1
          type: string
          pattern: ^\S+$
        value:
          description: |-
            +label=Value
            +usage=The value of the build secret.
            +uiType=SecretSelect
            +uiProps={"typeOptionsList":["password","secret"]}
            +sort=2
          type: string
        env:
          description: >-
            +label=Environment Variable Name

            +usage=The name of the environment variable to inject the secret
            into.

            +uiType=Hidden

            +sort=3
          type: string
          pattern: ^[_a-zA-Z0-9][_a-zA-Z0-9]{1,31}$
      title: BuildSecret
      x-fern-type-name: BuildSecret
    Pip:
      description: |-
        +usage=Use pip to install requirements files and packages
        +label=Pip
      type: object
      required:
        - type
      properties:
        type:
          description: +value=pip
          type: string
          enum:
            - pip
        requirements_path:
          description: >-
            +label=Path to requirements

            +usage=Path to `requirements.txt` relative to `Path to build
            context`
          type: string
        pip_packages:
          description: |-
            +label=Pip packages to install
            +usage=Define pip package requirements.
            In Python/YAML E.g. ["fastapi>=0.90,<1.0", "uvicorn"]
            +placeholder=Enter a pip package name E.g. fastapi>=0.90,<1.0
          type: array
          items:
            type: string
      title: Pip
      x-fern-type-name: Pip
    UV:
      description: |-
        +usage=Use `uv` to setup env.
        Your build context root must contain `pyproject.toml` and `uv.lock`
        +label=uv
      type: object
      required:
        - type
      properties:
        type:
          description: +value=uv
          type: string
          enum:
            - uv
        uv_version:
          description: |-
            +label=UV version
            +usage=UV version to use
          type: string
          default: latest
        sync_options:
          description: |-
            +label=Sync options
            +usage=Sync options to pass to uv command
          type: string
      title: UV
      x-fern-type-name: UV
    Poetry:
      description: |-
        +usage=Use `poetry` to setup env.
        Your build context root must contain `pyproject.toml` and `poetry.lock`
      type: object
      required:
        - type
      properties:
        type:
          description: +value=poetry
          type: string
          enum:
            - poetry
        poetry_version:
          description: |-
            +label=Poetry version
            +usage=Poetry version to use
          type: string
          default: latest
          oneOf:
            - enum:
                - latest
            - pattern: ^\d+(\.\d+){1,2}([\-\.a-z0-9]+)?$
        install_options:
          description: |-
            +label=Install options
            +usage=install options to pass to poetry command
          type: string
      title: Poetry
      x-fern-type-name: Poetry
    FlyteTaskID:
      type: object
      required:
        - resourceType
        - name
      properties:
        resourceType:
          type: string
          enum:
            - TASK
        name:
          type: string
      title: FlyteTaskID
      x-fern-type-name: FlyteTaskID
    FlyteTaskCustom:
      type: object
      required:
        - truefoundry
      properties:
        truefoundry:
          description: +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/PythonTaskConfig'
            - $ref: '#/components/schemas/ContainerTaskConfig'
            - $ref: '#/components/schemas/PySparkTaskConfig'
            - $ref: '#/components/schemas/DatabricksJobTaskConfig'
          discriminator:
            propertyName: type
      title: FlyteTaskCustom
      x-fern-type-name: FlyteTaskCustom
    FlyteWorkflowID:
      type: object
      required:
        - resourceType
        - name
      properties:
        resourceType:
          type: string
          enum:
            - WORKFLOW
        name:
          type: string
      title: FlyteWorkflowID
      x-fern-type-name: FlyteWorkflowID
    AwsSqsAccessKeyBasedAuth:
      description: +label=AWS Access Key Based Auth
      type: object
      required:
        - type
        - access_key_id
        - secret_access_key
      properties:
        type:
          description: +value=access-key-based
          type: string
          enum:
            - access-key-based
        access_key_id:
          description: |-
            +label=AWS access key ID
            +usage=AWS access key ID that you want to use for authentication.
            +sort=100
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
        secret_access_key:
          description: >-
            +label=AWS access key secret

            +usage=AWS access key secret that you want to use for
            authentication.

            +sort=200

            +uiType=SecretSelect

            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: AwsSqsAccessKeyBasedAuth
      x-fern-type-name: AwsSqsAccessKeyBasedAuth
    AwsSqsAssumedRoleBasedAuth:
      description: +label=AWS Assumed Role Based Auth
      type: object
      required:
        - type
        - assumed_role_arn
      properties:
        type:
          description: +value=assumed-role-based
          type: string
          enum:
            - assumed-role-based
        assumed_role_arn:
          description: |-
            +label=AWS Assumed Role ARN
            +usage=ARN of the role you want to assume.
            +sort=100
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^arn:(aws|aws-us-gov|aws-cn):iam::[0-9]{12}:role(/[\w+=,.@-]+)+$
      title: AwsSqsAssumedRoleBasedAuth
      x-fern-type-name: AwsSqsAssumedRoleBasedAuth
    AWSAccessKeyAuth:
      type: object
      required:
        - aws_access_key_id
        - aws_secret_access_key
      properties:
        aws_access_key_id:
          description: |-
            +label=AWS Access Key ID
            +usage=AWS Access Key ID
            +sort=1
          type: string
        aws_secret_access_key:
          description: |-
            +label=AWS Secret Access Key
            +usage=AWS Secret Access Key for the user to authenticate with
            +sort=2
          type: string
        aws_session_token:
          description: >-
            +label=AWS Session Token

            +usage=AWS Session Token, only required when using temporary
            credentials

            +sort=3
          type: string
      title: AWSAccessKeyAuth
      x-fern-type-name: AWSAccessKeyAuth
    NATSUserPasswordAuth:
      description: |-
        +docs=NATS User Password Authentication
        +label=NATS User Password Authentication
      type: object
      required:
        - account_name
        - user
        - password
      properties:
        account_name:
          description: |-
            +label=Account Name
            +usage=Name of the NATS account
            +sort=1
          type: string
          default: $G
        user:
          description: |-
            +label=User
            +usage=User for NATS authentication
            +sort=2
          type: string
        password:
          description: |-
            +label=Password
            +usage=Password for NATS authentication
            +sort=3
          type: string
      title: NATSUserPasswordAuth
      x-fern-type-name: NATSUserPasswordAuth
    KafkaSASLAuth:
      type: object
      required:
        - username
        - password
      properties:
        username:
          description: |-
            +label=Username
            +usage=Username for SASL authentication
            +sort=1
          type: string
        password:
          description: |-
            +label=Password
            +usage=Password for SASL authentication
            +sort=2
          type: string
      title: KafkaSASLAuth
      x-fern-type-name: KafkaSASLAuth
    RateLimitWhen:
      description: +label=Rule Conditions
      type: object
      properties:
        subjects:
          description: >-
            +usage=List of subjects that this rule applies to

            +uiProps={"descriptionInline":true, "optionTypes":
            ["users","teams","virtualAccounts"]}

            +label=Subjects
          type: array
          items:
            type: string
        models:
          description: >-
            +usage=List of models that this rule applies to

            +uiType=IntegrationSelect

            +uiProps={"integrationType":"model", "isMultiSelect": true,
            "valuePattern": "${providerAccount.name}/${manifest.name}"}

            +label=Models
          type: array
          items:
            type: string
        metadata:
          description: |-
            +usage=Metadata key-value pairs that this rule applies to
            +uiProps={"descriptionInline":true}
            +label=Metadata
          type: object
          additionalProperties:
            type: string
      title: RateLimitWhen
      x-fern-type-name: RateLimitWhen
    RateLimitUnit:
      description: +usage=Unit for rate limit calculation
      type: string
      enum:
        - requests_per_day
        - tokens_per_minute
        - requests_per_minute
        - tokens_per_day
        - requests_per_hour
        - tokens_per_hour
      title: RateLimitUnit
      x-fern-type-name: RateLimitUnit
    WeightBasedLoadBalancingRule:
      description: +label=Weight-based Load Balancing Rule
      type: object
      properties:
        id:
          description: |-
            +usage=Unique identifier for the rule
            +uiProps={"descriptionInline":true}
            +sort=2
            +label=Rule ID
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-]{0,60}[a-zA-Z0-9])$
        when:
          $ref: '#/components/schemas/LoadBalancingWhen'
      allOf:
        - $ref: '#/components/schemas/WeightBasedLoadBalancing'
      required:
        - id
        - when
      title: WeightBasedLoadBalancingRule
      x-fern-type-name: WeightBasedLoadBalancingRule
    LatencyBasedLoadBalancingRule:
      description: +label=Latency-based Load Balancing Rule
      type: object
      properties:
        id:
          description: |-
            +usage=Unique identifier for the rule
            +uiProps={"descriptionInline":true}
            +sort=2
            +label=Rule ID
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-]{0,60}[a-zA-Z0-9])$
        when:
          $ref: '#/components/schemas/LoadBalancingWhen'
      allOf:
        - $ref: '#/components/schemas/LatencyBasedLoadBalancing'
      required:
        - id
        - when
      title: LatencyBasedLoadBalancingRule
      x-fern-type-name: LatencyBasedLoadBalancingRule
    PriorityBasedLoadBalancingRule:
      description: +label=Priority-based Load Balancing Rule
      type: object
      properties:
        id:
          description: |-
            +usage=Unique identifier for the rule
            +uiProps={"descriptionInline":true}
            +sort=2
            +label=Rule ID
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-]{0,60}[a-zA-Z0-9])$
        when:
          $ref: '#/components/schemas/LoadBalancingWhen'
      allOf:
        - $ref: '#/components/schemas/PriorityBasedLoadBalancing'
      required:
        - id
        - when
      title: PriorityBasedLoadBalancingRule
      x-fern-type-name: PriorityBasedLoadBalancingRule
    GuardrailsWhen:
      type: object
      properties:
        target:
          $ref: '#/components/schemas/TargetConditionGroup'
        subjects:
          $ref: '#/components/schemas/SubjectConditionGroup'
        metadata:
          type: object
          additionalProperties:
            type: string
      title: GuardrailsWhen
      x-fern-type-name: GuardrailsWhen
    BudgetWhen:
      description: +label=Rule Conditions
      type: object
      properties:
        subjects:
          description: >-
            +usage=List of subjects that this rule applies to

            +uiProps={"descriptionInline":true, "optionTypes":
            ["users","teams","virtualAccounts"]}

            +label=Subjects
          type: array
          items:
            type: string
        models:
          description: >-
            +usage=List of models that this rule applies to

            +uiType=IntegrationSelect

            +uiProps={"integrationType":"model", "isMultiSelect": true,
            "valuePattern": "${providerAccount.name}/${manifest.name}"}

            +label=Models
          type: array
          items:
            type: string
        metadata:
          description: |-
            +usage=Metadata key-value pairs that this rule applies to
            +uiProps={"descriptionInline":true}
            +label=Metadata
          type: object
          additionalProperties:
            type: string
      title: BudgetWhen
      x-fern-type-name: BudgetWhen
    BudgetLimitUnit:
      description: +usage=Unit for budget limit calculation
      type: string
      enum:
        - cost_per_day
        - cost_per_month
        - cost_per_week
        - tokens_per_day
        - tokens_per_month
      title: BudgetLimitUnit
      x-fern-type-name: BudgetLimitUnit
    BudgetAlert:
      description: +label=Budget Alert
      type: object
      required:
        - thresholds
        - notification_target
      properties:
        thresholds:
          description: >-
            +usage=List of usage percentages (0-100) at which alerts should be
            triggered. Default thresholds are [75, 90, 95, 100].

            +uiProps={"descriptionInline":true}

            +sort=2

            +label=Alert Thresholds
          type: array
          items:
            type: number
            minimum: 0
            maximum: 100
          default:
            - 75
            - 90
            - 95
            - 100
        notification_target:
          description: |-
            +label=Notification Target
            +usage=Select where to send budget alert notifications
            +sort=5
          type: array
          items:
            $ref: '#/components/schemas/NotificationTarget'
      title: BudgetAlert
      x-fern-type-name: BudgetAlert
    OTELExporterHttpConfigBase:
      description: Base HTTP exporter fields (without type discriminator)
      type: object
      required:
        - endpoint
        - encoding
      properties:
        endpoint:
          description: >-
            +usage=Endpoint URL where OpenTelemetry data will be sent

            +label=Endpoint

            +sort=100

            +placeholder=Enter complete endpoint URL (e.g.,
            http://localhost:4318/v1/traces)
          type: string
        encoding:
          description: |-
            +usage=Encoding for OpenTelemetry data
            +label=Encoding
            +sort=200
          type: string
          enum:
            - proto
            - json
          default: proto
        headers:
          description: |-
            +usage=HTTP headers to include in OpenTelemetry export requests
            +label=Headers
            +sort=300
            +placeholder=Enter headers key-value pairs
          type: object
          additionalProperties:
            type: string
      title: OTELExporterHttpConfigBase
      x-fern-type-name: OTELExporterHttpConfigBase
    OTELTracesExporterCommonConfig:
      description: Common fields shared by HTTP and gRPC exporter configs
      type: object
      properties:
        additional_resource_attributes:
          description: >-
            +usage=Set additional attributes to append them to existing resource
            attributes of every exported trace

            +label=Additional Resource Attributes

            +sort=400

            +placeholder=Enter resource attributes key-value pairs
          type: object
          additionalProperties:
            type: string
        exclude_request_data:
          description: >-
            +usage=Enable to prevent forwarding the LLM or MCP request and
            response body in exported traces. The following span attributes are
            dropped: tfy.input, tfy.output, tfy.input_short_hand

            +label=Exclude Request Data

            +sort=500
          type: boolean
        when:
          description: >-
            +usage=Export spans only when they match all of the following
            filters. Only matching on spanAttributes with string values in
            supported.

            +label=Filters

            +sort=600
          type: array
          items:
            $ref: '#/components/schemas/OTELTracesExporterSpanAttributeFilter'
      title: OTELTracesExporterCommonConfig
      x-fern-type-name: OTELTracesExporterCommonConfig
    OTELExporterGrpcConfigBase:
      description: Base gRPC exporter fields (without type discriminator)
      type: object
      required:
        - endpoint
      properties:
        endpoint:
          description: >-
            +usage=Endpoint URL where OpenTelemetry data will be sent

            +label=Endpoint

            +sort=100

            +placeholder=Enter complete endpoint URL (e.g.,
            http://localhost:4318)
          type: string
        headers:
          description: >-
            In otel config the gRPC metadata to be passed to the endpoint is
            called headers

            +usage=gRPC metadata to include in OpenTelemetry export requests

            +label=Headers

            +sort=200

            +placeholder=Enter headers key-value pairs
          type: object
          additionalProperties:
            type: string
      title: OTELExporterGrpcConfigBase
      x-fern-type-name: OTELExporterGrpcConfigBase
    GatewayMetadataWhen:
      description: +label=Rule Conditions
      type: object
      properties:
        subjects:
          description: >-
            +usage=List of subjects that this rule applies to

            +uiProps={"descriptionInline":true, "optionTypes":
            ["users","teams","virtualAccounts"]}

            +label=Subjects
          type: array
          items:
            type: string
      title: GatewayMetadataWhen
      x-fern-type-name: GatewayMetadataWhen
    GatewayLoggingWhen:
      description: +label=Rule Conditions
      type: object
      properties:
        subjects:
          description: >-
            +usage=List of subjects that this rule applies to

            +uiProps={"descriptionInline":true, "optionTypes":
            ["users","teams","virtualAccounts"]}

            +label=Subjects
          type: array
          items:
            type: string
      title: GatewayLoggingWhen
      x-fern-type-name: GatewayLoggingWhen
    LoggingModeAlways:
      description: >-
        +label=Always Log Body

        +usage=Enable this mode to log request and response bodies for all
        incoming gateway requests, ignoring request headers. Use it for in-depth
        logging required for debugging, compliance, or auditing.
      type: object
      required:
        - type
      properties:
        type:
          description: +value=always
          type: string
          enum:
            - always
      title: LoggingModeAlways
      x-fern-type-name: LoggingModeAlways
    LoggingModeNever:
      description: >-
        +label=Never Log Body

        +usage=Enable this mode to disable request and response body logging for
        all gateway requests, regardless of headers. Request metadata and
        headers are still logged, but body content is excluded.
      type: object
      required:
        - type
      properties:
        type:
          description: +value=never
          type: string
          enum:
            - never
      title: LoggingModeNever
      x-fern-type-name: LoggingModeNever
    LoggingModeHeaderControlled:
      description: >-
        +label=Header Controlled Body Logging

        +usage=Enable this mode to allow per-request control of request body
        logging via HTTP headers. Individual requests can opt in or opt out of
        body logging using specific headers.
      type: object
      required:
        - type
        - log_body_by_default
      properties:
        type:
          description: +value=header_controlled
          type: string
          enum:
            - header_controlled
        log_body_by_default:
          description: >-
            +usage=Sets the default body logging behaviour when no header is
            present. Enabling it will log requests by default and clients can
            opt out. Disabling it will NOT log by default and clients can opt in
            to log requests.

            +uiProps={"descriptionInline":true}

            +label=Log Body By Default
          type: boolean
      title: LoggingModeHeaderControlled
      x-fern-type-name: LoggingModeHeaderControlled
    OwnDataAccessRule:
      description: +label=Own Data
      type: object
      properties:
        scope:
          description: |-
            +value=own_data
            +usage=Data access scope - access to own data only
            +sort=0
            +uiType=Ignore
          type: string
          enum:
            - own_data
      allOf:
        - $ref: '#/components/schemas/DataAccessRuleBase'
      required:
        - scope
      title: OwnDataAccessRule
      x-fern-type-name: OwnDataAccessRule
    TeamDataAccessRule:
      description: +label=Team Data
      type: object
      properties:
        scope:
          description: >-
            +value=team_data

            +usage=Data access scope - access to team data including owned
            virtual accounts

            +sort=0

            +uiType=Ignore
          type: string
          enum:
            - team_data
      allOf:
        - $ref: '#/components/schemas/DataAccessRuleBase'
      required:
        - scope
      title: TeamDataAccessRule
      x-fern-type-name: TeamDataAccessRule
    AllDataAccessRule:
      description: +label=All Data
      type: object
      properties:
        scope:
          description: |-
            +value=all_data
            +usage=Data access scope - access to all data with optional filters
            +sort=0
            +uiType=Ignore
          type: string
          enum:
            - all_data
        filters:
          description: >-
            +usage=Optional filters to restrict data access (only applicable for
            all_data scope)

            +uiProps={"descriptionInline":true}

            +sort=5

            +label=Filters
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/GatewayDataMetadataFilter'
              - $ref: '#/components/schemas/GatewayDataCreatedByFilter'
      allOf:
        - $ref: '#/components/schemas/DataAccessRuleBase'
      required:
        - scope
      title: AllDataAccessRule
      x-fern-type-name: AllDataAccessRule
    GatewayDataRoutingConfigDestinationCondition:
      description: +discriminator=field
      type: object
      oneOf:
        - $ref: >-
            #/components/schemas/GatewayDataRoutingConfigDestinationMetadataCondition
        - $ref: >-
            #/components/schemas/GatewayDataRoutingConfigDestinationCreatedByCondition
      discriminator:
        propertyName: field
      title: GatewayDataRoutingConfigDestinationCondition
      x-fern-type-name: GatewayDataRoutingConfigDestinationCondition
    GatewayDataRoutingConfigDestinationControlPlaneManagedStorage:
      description: |-
        +label=Controlplane Managed Storage
        +usage=Managed Storage in Control-Plane.
      type: object
      required:
        - type
      properties:
        type:
          description: |-
            +value=controlplane-managed
            +label=Storage Type
            +usage=Type of storage configuration
            +message=Select controlplane-managed or customer-managed storage
            +sort=1
            +uiType=Select
          type: string
          enum:
            - controlplane-managed
        region:
          description: >-
            +label=Storage Region for Control Plane

            +usage=Specifies the region where control plane–managed storage is
            located.

            +sort=100

            +message=Region must not be empty.
          type: string
          enum:
            - us
            - australia
            - india
      title: GatewayDataRoutingConfigDestinationControlPlaneManagedStorage
      x-fern-type-name: GatewayDataRoutingConfigDestinationControlPlaneManagedStorage
    GatewayDataRoutingConfigDestinationCustomerManagedStorage:
      description: |-
        +label=Customer Managed Storage
        +usage=Bring your own storage.
      type: object
      required:
        - type
        - storage_integration_fqn
      properties:
        type:
          description: |-
            +value=customer-managed
            +label=Storage Type
            +usage=Type of storage configuration
            +message=Select controlplane-managed or customer-managed storage
            +sort=1
          type: string
          enum:
            - customer-managed
        storage_integration_fqn:
          description: >-
            +label=Storage Integration

            +icon=hard-drive

            +usage=Storage Integration FQN to store traces and spans. A storage
            integration represents a connected blob storage like AWS S3 / Azure
            Blob Storage / Google Cloud Storage.

            +message=Select the storage-integration where you want to save your
            traces and spans. If no integration is listed, please contact your
            admin.\n[View
            Docs](https://www.truefoundry.com/docs/integrations-overview#integrations)

            +sort=2

            +uiType=IntegrationSelect

            +uiProps={"descriptionKey":"storage_root","integrationType":"blob-storage"}
          type: string
          pattern: ^[a-zA-Z0-9][a-zA-Z0-9._:-]*[a-zA-Z0-9]$
      title: GatewayDataRoutingConfigDestinationCustomerManagedStorage
      x-fern-type-name: GatewayDataRoutingConfigDestinationCustomerManagedStorage
    TrueFoundryAgentMcpTool:
      type: object
      required:
        - name
      properties:
        name:
          type: string
      title: TrueFoundryAgentMcpTool
      x-fern-type-name: TrueFoundryAgentMcpTool
    CustomAgentServerAuth:
      type: object
      oneOf:
        - $ref: '#/components/schemas/CustomServerHeaderAuth'
        - $ref: '#/components/schemas/CustomServerPassthrough'
      title: CustomAgentServerAuth
      x-fern-type-name: CustomAgentServerAuth
    AgentFramework:
      description: +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/A2AFramework'
        - $ref: '#/components/schemas/CustomFramework'
      discriminator:
        propertyName: type
      title: AgentFramework
      x-fern-type-name: AgentFramework
    BaseRemoteAgent:
      type: object
      required:
        - name
        - description
        - collaborators
      properties:
        name:
          description: >-
            +label=Name

            +usage=The name of the Agent

            +sort=100

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,30}[a-z0-9]$
        description:
          description: |-
            +label=Description
            +usage=The description of the Agent
            +sort=110
          type: string
          maxLength: 2000
        tags:
          description: |-
            +label=Tags
            +usage=Tags for the Agent
            +sort=120
          type: object
        collaborators:
          description: |-
            +label=Collaborators
            +usage=List of users who have access to this Agent
            +uiType=Collaborators
            +sort=130
          type: array
          items:
            $ref: '#/components/schemas/Collaborator'
        ownedBy:
          $ref: '#/components/schemas/OwnedBy'
      title: BaseRemoteAgent
      x-fern-type-name: BaseRemoteAgent
    PromptSource:
      description: +label=Prompt
      type: object
      required:
        - type
        - prompt_version_fqn
      properties:
        type:
          description: |-
            +value=prompt
            +label=Type
            +usage=Type of agent source
            +uiType=Hidden
          type: string
          enum:
            - prompt
        prompt_version_fqn:
          description: >-
            +label=Prompt Version FQN

            +usage=Fully qualified name of the prompt version to use as the
            agent source

            +sort=100

            +uiType=AgentFqn
          type: string
        skills:
          description: |-
            +label=Skills
            +usage=List of skills associated with the agent
            +sort=200
            +uiType=Hidden
          type: array
          items:
            $ref: '#/components/schemas/AgentSkill'
      title: PromptSource
      x-fern-type-name: PromptSource
    HostedA2AAgent:
      description: +label=Hosted A2A Agent
      type: object
      required:
        - type
        - agent_card_url
      properties:
        type:
          description: |-
            +value=hosted-a2a-agent
            +label=Type
            +usage=Type of agent source
            +uiType=Hidden
          type: string
          enum:
            - hosted-a2a-agent
        agent_card_url:
          description: |-
            +label=Agent Card URL
            +usage=Agent Card URL for the A2A Agent
            +sort=100
          type: string
        headers:
          description: >-
            +label=Headers

            +usage=Headers to be sent with the request to the hosted A2A agent.
            Primarily used for authentication.
          type: object
          additionalProperties:
            type: string
      title: HostedA2AAgent
      x-fern-type-name: HostedA2AAgent
    MCPServerHeaderAuth:
      description: >-
        +label=Header Auth

        +usage=Authenticate using a static key or token sent as a request
        header. Choose shared vs individual credentials below.
      type: object
      required:
        - type
        - headers
        - auth_level
      properties:
        type:
          description: +value=header
          type: string
          enum:
            - header
        headers:
          description: >-
            +usage=Map each header name to its value. For shared, enter the
            actual value. For individual, use a placeholder that will be filled
            in by each user (e.g. Bearer {{API_KEY}}).

            +uiProps={"keyPlaceholder": "Authorization", "valuePlaceholder":
            "Bearer {{TOKEN}}", "addEmptyFieldOnCreate": true}
          type: object
          additionalProperties:
            type: string
        auth_level:
          description: >-
            +label=API Key Configuration

            +usage=Level at which this authentication will be applied. Values:
            global (default), per_user

            +uiType=RadioCard

            +sort=200
          type: string
          enum:
            - global
            - per_user
          default: global
      title: MCPServerHeaderAuth
      x-fern-type-name: MCPServerHeaderAuth
    MCPServerOAuth2:
      description: +label=OAuth2
      type: object
      required:
        - type
        - grant_type
        - token_url
        - jwt_source
      properties:
        type:
          description: |-
            +value=oauth2
            +label=OAuth2
            +sort=100
            +usage=OAuth2 authentication
          type: string
          enum:
            - oauth2
        grant_type:
          description: |-
            +label=Grant Type
            +usage=The OAuth2 grant type to use for authentication.
            +sort=150
            +uiType=RadioCard
          type: string
          enum:
            - authorization_code
            - client_credentials
          default: authorization_code
        authorization_url:
          description: |-
            +label=Authorization URL
            +usage=URL for the authorization request
            +sort=200
          type: string
          pattern: ^https:\/\/(([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})(\/[^\s]*)?$
        token_url:
          description: |-
            +label=Token URL
            +usage=The endpoint to exchange auth code for tokens.
            +sort=300
          type: string
          pattern: ^https:\/\/(([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})(\/[^\s]*)?$
        client_id:
          description: >-
            +label=Client ID

            +usage=client ID for OAuth2 or the TrueFoundry secret FQN containing
            the client ID.

            +sort=400

            +uiType=SecretSelect
          type: string
          pattern: ^\S+$
        client_secret:
          description: >-
            +label=Client Secret

            +usage=Client secret or the TrueFoundry secret FQN containing the
            client secret for OAuth2.

            +sort=500

            +uiType=SecretSelect
          type: string
          minLength: 1
        registration_url:
          description: >-
            +label=Registration URL (Used for Dynamic Client Registration)

            +usage=URL for dynamic client registration (RFC 7591). If provided,
            client credentials will be obtained automatically using the Dynamic
            Client Registration (DCR) process.

            +sort=600
          type: string
          pattern: ^https:\/\/(([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})(\/[^\s]*)?$
        code_challenge_methods_supported:
          description: |-
            +label=Code Challenge Methods Supported
            +usage=List of supported PKCE code challenge methods (S256 only)
            +sort=700
          type: array
          items:
            type: string
            enum:
              - S256
        jwt_source:
          description: |-
            +label=JWT Source
            +usage=Source of the JWT token to be used for verification.
            +sort=800
          type: string
          enum:
            - access_token
            - id_token
          default: access_token
        scopes:
          description: |-
            +label=Scopes
            +usage=List of scopes to request from the OAuth2 provider.
            +sort=900
          type: array
          items:
            type: string
            minLength: 1
      title: MCPServerOAuth2
      x-fern-type-name: MCPServerOAuth2
    MCPServerPassthrough:
      description: >-
        +label=Token Passthrough

        +usage=Authenticate using your existing TrueFoundry account (PAT or
        Virtual Accounts). Your credentials are automatically passed to the MCP
        server.
      type: object
      required:
        - type
      properties:
        type:
          description: +value=passthrough
          type: string
          enum:
            - passthrough
      title: MCPServerPassthrough
      x-fern-type-name: MCPServerPassthrough
    MCPServerHeaderOverrideAuth:
      description: |-
        +label=Override authentication header
        +usage=Override authentication for this source server
      type: object
      required:
        - type
        - headers
      properties:
        type:
          description: +value=header
          type: string
          enum:
            - header
        headers:
          type: object
          additionalProperties:
            type: string
      title: MCPServerHeaderOverrideAuth
      x-fern-type-name: MCPServerHeaderOverrideAuth
    RemoteSpecSource:
      description: >-
        +label=Fetch from URL

        +usage=Fetch the OpenAPI specification from a remote HTTP/HTTPS
        endpoint.
      type: object
      required:
        - type
        - url
      properties:
        type:
          description: +value=remote
          type: string
          enum:
            - remote
        url:
          description: |-
            +label=OpenAPI Spec URL
            +usage=URL to fetch the OpenAPI specification (JSON or YAML format).
            +placeholder=https://myapp.example.org/openapi.json
            +message=Must be a valid HTTP/HTTPS URL
            +sort=10
          type: string
          pattern: ^https?:\/\/[^\s]+$
      title: RemoteSpecSource
      x-fern-type-name: RemoteSpecSource
    UploadSpecSource:
      description: |-
        +label=Upload Spec
        +usage=Fetch the stored/uploaded OpenAPI specification.
      type: object
      required:
        - type
        - url
      properties:
        type:
          description: +value=upload
          type: string
          enum:
            - upload
        url:
          description: >-
            +label=OpenAPI Spec Path

            +usage=Cloud storage path where the OpenAPI specification is stored
            (S3, GCS, or Azure Blob).

            +message=Must be a valid storage path: s3://bucket/...,
            gs://bucket/..., or
            https://<account>.blob.core.windows.net/<container>/...

            +sort=10
          type: string
          pattern: >-
            ^(s3://[^/]+|gs://[^/]+|https://[^.]+\.blob\.core\.windows\.net/[^/]+)/tfy-openapi-specs/(?:\d{4}-\d{2}/|[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+/)[a-f0-9-]+\.json$
      title: UploadSpecSource
      x-fern-type-name: UploadSpecSource
    InlineSpecSource:
      description: >-
        +label=Inline Spec

        +usage=Inline your OpenAPI specification directly. The spec will be
        stored securely.
      type: object
      required:
        - type
        - openapi_spec
      properties:
        type:
          description: +value=inline
          type: string
          enum:
            - inline
        openapi_spec:
          description: >-
            +label=OpenAPI Spec

            +usage=The OpenAPI specification as a JSON object or a JSON/YAML
            string.

            +message=Must be a valid OpenAPI spec (object or string)

            +sort=10

            +uiType=CodeEditor
          oneOf:
            - type: object
            - type: string
      title: InlineSpecSource
      x-fern-type-name: InlineSpecSource
    AwsEcr:
      description: |-
        +label=AWS ECR
        +icon=aws-ecr
      type: object
      required:
        - type
        - name
        - registry_url
      properties:
        type:
          description: +value=integration/docker-registry/aws/ecr
          type: string
          enum:
            - integration/docker-registry/aws/ecr
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        registry_url:
          description: |-
            +label=Registry URL
            +usage=The URL of the registry.
            +sort=200
          type: string
          pattern: >-
            ^([\w]+)\.dkr\.ecr\.([\w-]+)\.amazonaws\.com|(https?:\/\/)?(public\.ecr\.aws)(\/([\w.-]+))?(\/?)$
        auth_data:
          description: |-
            +label=Auth Data
            +usage=Custom authentication data for the integration.
            +sort=300
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/AwsAssumedRoleBasedAuth'
          discriminator:
            propertyName: type
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: AwsEcr
      x-fern-type-name: AwsEcr
    AwsS3:
      description: |-
        +label=AWS S3
        +icon=aws-s3
      type: object
      required:
        - type
        - name
        - storage_root
        - region
      properties:
        type:
          description: +value=integration/blob-storage/aws/s3
          type: string
          enum:
            - integration/blob-storage/aws/s3
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        storage_root:
          description: >-
            +label=Storage Root

            +usage=The S3 storage path in the format
            's3://<bucket-name>/<optional-path-prefix>'. Works with
            S3-compatible storage providers like MinIO, Ceph, and DigitalOcean
            Spaces. Example: 's3://mybucket' or
            's3://mybucket/artifacts/production'.

            +sort=200
          type: string
        region:
          $ref: '#/components/schemas/AWSRegion'
        auth_data:
          description: |-
            +label=Auth Data
            +usage=Custom authentication data for the integration.
            +sort=400
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/AwsAssumedRoleBasedAuth'
          discriminator:
            propertyName: type
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: AwsS3
      x-fern-type-name: AwsS3
    AwsParameterStore:
      description: |-
        +label=AWS Parameter Store
        +icon=aws-ssm
      type: object
      required:
        - type
        - name
        - region
      properties:
        type:
          description: +value=integration/secret-store/aws/parameter-store
          type: string
          enum:
            - integration/secret-store/aws/parameter-store
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=150

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        region:
          $ref: '#/components/schemas/AWSRegion'
        auth_data:
          description: |-
            +label=Auth Data
            +usage=Custom authentication data for the integration.
            +sort=300
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/AwsAssumedRoleBasedAuth'
          discriminator:
            propertyName: type
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
        kms_arn:
          description: |-
            +label=KMS ARN
            +usage=The ARN of the KMS key to use for encryption.
            +sort=700
            +isAdvanced=true
          type: string
          pattern: ^arn:aws[a-zA-Z-]*:kms:[a-z0-9-]+:[0-9]{12}:key/[a-fA-F0-9-]+$
      title: AwsParameterStore
      x-fern-type-name: AwsParameterStore
    AwsSecretsManager:
      description: |-
        +label=AWS Secrets Manager
        +icon=aws-secrets-manager
      type: object
      required:
        - type
        - name
        - region
      properties:
        type:
          description: +value=integration/secret-store/aws/secrets-manager
          type: string
          enum:
            - integration/secret-store/aws/secrets-manager
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=150

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        region:
          $ref: '#/components/schemas/AWSRegion'
        auth_data:
          description: |-
            +label=Auth Data
            +usage=Custom authentication data for the integration.
            +sort=300
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/AwsAssumedRoleBasedAuth'
          discriminator:
            propertyName: type
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
        kms_arn:
          description: |-
            +label=KMS ARN
            +usage=The ARN of the KMS key to use for encryption.
            +sort=700
            +isAdvanced=true
          type: string
          pattern: ^arn:aws[a-zA-Z-]*:kms:[a-z0-9-]+:[0-9]{12}:key/[a-fA-F0-9-]+$
      title: AwsSecretsManager
      x-fern-type-name: AwsSecretsManager
    AwsEksIntegration:
      description: |-
        +label=AWS EKS
        +icon=aws-eks
      type: object
      required:
        - type
        - name
        - region
        - cluster_name
      properties:
        type:
          description: +value=integration/cluster/aws/eks
          type: string
          enum:
            - integration/cluster/aws/eks
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        region:
          $ref: '#/components/schemas/AWSRegion'
        cluster_name:
          description: |-
            +label=Cluster Name
            +usage=Name of the cluster given in the cloud account.
            +sort=300
          type: string
          pattern: ^\S+$
        auth_data:
          description: |-
            +label=Auth Data
            +usage=Custom authentication data for the integration.
            +sort=400
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/AwsAccessKeyBasedAuth'
            - $ref: '#/components/schemas/AwsAssumedRoleBasedAuth'
          discriminator:
            propertyName: type
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: AwsEksIntegration
      x-fern-type-name: AwsEksIntegration
    AzureAksIntegration:
      description: |-
        +icon=azure-aks
        +label=Azure AKS
      type: object
      required:
        - type
        - name
        - resource_group
        - cluster_name
      properties:
        type:
          description: +value=integration/cluster/azure/aks
          type: string
          enum:
            - integration/cluster/azure/aks
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        resource_group:
          description: |-
            +label=Resource Group
            +usage=The resource group of the cluster in the cloud account.
            +sort=200
          type: string
          pattern: ^\S+$
        cluster_name:
          description: |-
            +label=Cluster Name
            +usage=Name of the cluster given in the cloud account.
            +sort=300
          type: string
          pattern: ^\S+$
        auth_data:
          $ref: '#/components/schemas/AzureOAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: AzureAksIntegration
      x-fern-type-name: AzureAksIntegration
    AzureContainerRegistry:
      description: |-
        +icon=azure-acr
        +label=Azure ACR
      type: object
      required:
        - type
        - name
        - registry_url
      properties:
        type:
          description: +value=integration/docker-registry/azure/acr
          type: string
          enum:
            - integration/docker-registry/azure/acr
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        registry_url:
          description: |-
            +label=Registry URL
            +usage=The URL of the registry.
            +sort=200
          type: string
          pattern: ^([a-z0-9-]{5,50}\.azurecr\.io)$
        auth_data:
          $ref: '#/components/schemas/AzureBasicAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: AzureContainerRegistry
      x-fern-type-name: AzureContainerRegistry
    AzureBlobStorage:
      description: |-
        +icon=azure-blob
        +label=Azure ABS
      type: object
      required:
        - type
        - name
        - storage_root
      properties:
        type:
          description: +value=integration/blob-storage/azure/blob
          type: string
          enum:
            - integration/blob-storage/azure/blob
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        storage_root:
          description: >-
            +label=Storage Root

            +usage=The full Azure Blob Storage URL in the format
            'https://<storage-account-name>.blob.core.windows.net/<container-name>/<optional-path-prefix>'.
            Example:
            'https://mystorageaccount.blob.core.windows.net/mycontainer' or
            'https://mystorageaccount.blob.core.windows.net/mycontainer/artifacts/production'.

            +sort=200
          type: string
          pattern: ^\S+$
        auth_data:
          $ref: '#/components/schemas/AzureConnectionStringAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: AzureBlobStorage
      x-fern-type-name: AzureBlobStorage
    AzureVault:
      description: |-
        +label=Azure Vault
        +icon=azure-vault
      type: object
      required:
        - type
        - name
        - key_vault_name
      properties:
        type:
          description: +value=integration/secret-store/azure/vault
          type: string
          enum:
            - integration/secret-store/azure/vault
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=150

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/AzureOAuth'
        key_vault_name:
          description: |-
            +label=Azure Key vault Name
            +usage=Name of key vault name.
            +sort=300
          type: string
          pattern: ^\S+$
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=400

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: AzureVault
      x-fern-type-name: AzureVault
    AzureReposIntegration:
      description: |-
        +label=Azure Repo
        +icon=azure
      type: object
      required:
        - type
        - name
        - account_name
      properties:
        type:
          description: +value=integration/vcs/azure
          type: string
          enum:
            - integration/vcs/azure
        name:
          description: +sort=100
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        account_name:
          description: |-
            +label=Account Name
            +sort=200
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,33}[a-z0-9]$
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: AzureReposIntegration
      x-fern-type-name: AzureReposIntegration
    GcpGcr:
      description: |-
        +icon=gcp-gcr
        +label=GCP GCR
      type: object
      required:
        - type
        - name
        - registry_url
      properties:
        type:
          description: +value=integration/docker-registry/gcp/gcr
          type: string
          enum:
            - integration/docker-registry/gcp/gcr
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        registry_url:
          description: |-
            +label=Registry URL
            +usage=The URL of the registry.
            +sort=200
          type: string
          pattern: ^([\w-]+-docker\.pkg\.dev\/[\w-]+|(?:[\w-]+\.)?gcr\.io\/[\w-]+)$
        auth_data:
          $ref: '#/components/schemas/GcpKeyFileAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: GcpGcr
      x-fern-type-name: GcpGcr
    GcpGcs:
      description: |-
        +icon=gcp-gcs
        +label=GCP GCS
      type: object
      required:
        - type
        - name
        - storage_root
      properties:
        type:
          description: +value=integration/blob-storage/gcp/gcs
          type: string
          enum:
            - integration/blob-storage/gcp/gcs
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        storage_root:
          description: >-
            +label=Storage Root

            +usage=The Google Cloud Storage path in the format
            'gs://<bucket-name>/<optional-path-prefix>'. Example:
            'gs://mybucket' or 'gs://mybucket/artifacts/production'.

            +sort=200
          type: string
        auth_data:
          $ref: '#/components/schemas/GcpKeyFileAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: GcpGcs
      x-fern-type-name: GcpGcs
    GcpGsm:
      description: |-
        +icon=gcp-gsm
        +label=GCP GSM
      type: object
      required:
        - type
        - name
      properties:
        type:
          description: +value=integration/secret-store/gcp/gsm
          type: string
          enum:
            - integration/secret-store/gcp/gsm
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/GcpKeyFileAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: GcpGsm
      x-fern-type-name: GcpGsm
    GcpGkeIntegration:
      description: |-
        +icon=gcp-gke
        +label=GCP GKE
      type: object
      required:
        - type
        - name
        - location
        - cluster_name
      properties:
        type:
          description: +value=integration/cluster/gcp/gke-standard
          type: string
          enum:
            - integration/cluster/gcp/gke-standard
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        location:
          $ref: '#/components/schemas/GCPRegion'
        cluster_name:
          description: |-
            +label=Cluster Name
            +usage=Name of the cluster given in the cloud account.
            +sort=300
          type: string
          pattern: ^\S+$
        auth_data:
          $ref: '#/components/schemas/GcpKeyFileAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: GcpGkeIntegration
      x-fern-type-name: GcpGkeIntegration
    DockerhubRegistry:
      description: |-
        +label=Dockerhub Registry
        +icon=docker
      type: object
      required:
        - type
        - name
        - registry_url
      properties:
        type:
          description: +value=integration/docker-registry/dockerhub
          type: string
          enum:
            - integration/docker-registry/dockerhub
        name:
          description: >-
            +label=Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        registry_url:
          description: |-
            +label=Registry URL
            +sort=200
          type: string
          pattern: >-
            ^(https?:\/\/)?(index\.docker\.io|docker\.io|registry\.hub\.docker\.com)(\/([\w.-]+))?(\/?)$
        auth_data:
          $ref: '#/components/schemas/DockerhubBasicAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: DockerhubRegistry
      x-fern-type-name: DockerhubRegistry
    CustomUsernamePasswordArtifactsRegistry:
      description: |-
        +label=Custom Docker Registry
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - registry_url
      properties:
        type:
          description: +value=integration/docker-registry/custom
          type: string
          enum:
            - integration/docker-registry/custom
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        registry_url:
          description: |-
            +label=Registry URL
            +usage=The URL of the registry.
            +sort=200
          type: string
          pattern: >-
            ^(https?:\/\/)?([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63}(\/[a-zA-Z0-9._~\-]+)*$
        auth_data:
          $ref: '#/components/schemas/CustomBasicAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: CustomUsernamePasswordArtifactsRegistry
      x-fern-type-name: CustomUsernamePasswordArtifactsRegistry
    EmailNotificationChannel:
      description: |-
        +label=Email Notification Channel
        +icon=envelope
      type: object
      required:
        - type
        - name
        - auth_data
        - from_email
      properties:
        type:
          description: +value=integration/notification-channel/email
          type: string
          enum:
            - integration/notification-channel/email
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/SMTPCredentials'
        from_email:
          description: |-
            +label=From Email
            +usage=The email address from which the email will be sent.
            +sort=1000
          type: string
          minLength: 1
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=1200

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: EmailNotificationChannel
      x-fern-type-name: EmailNotificationChannel
    CustomHelmRepo:
      description: |-
        +label=Custom Helm Repo
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - repo_url
        - auth_data
      properties:
        type:
          description: +value=integration/helm-repo/custom
          type: string
          enum:
            - integration/helm-repo/custom
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        repo_url:
          description: |-
            +label=Helm Repo URL
            +usage=The URL of the Helm Repo.
            +sort=200
          type: string
          pattern: >-
            ^(((https?|wss):\/\/)?(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*))$
        auth_data:
          $ref: '#/components/schemas/CustomBasicAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: CustomHelmRepo
      x-fern-type-name: CustomHelmRepo
    CustomBlobStorage:
      description: |-
        +label=Custom Blob Storage
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - endpoint
        - storage_root
        - auth_data
      properties:
        type:
          description: +value=integration/blob-storage/custom
          type: string
          enum:
            - integration/blob-storage/custom
        name:
          description: >-
            +label=Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        endpoint:
          description: |-
            +label=Endpoint
            +usage=The endpoint of the storage provider.
            +sort=200
          type: string
          pattern: ^(https?:\/\/)?(([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})$
        storage_root:
          description: >-
            +label=Storage Root

            +usage=The S3-compatible storage path in the format
            's3://<bucket-name>/<optional-path-prefix>'. Works with
            S3-compatible storage providers like MinIO, Ceph, and DigitalOcean
            Spaces. Example: 's3://mybucket' or
            's3://mybucket/artifacts/production'.

            +sort=200
          type: string
          pattern: ^\S+$
        region:
          description: |-
            +label=Region
            +usage=The region where the storage is located.
            +sort=600
          type: string
        auth_data:
          $ref: '#/components/schemas/CustomBasicAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=400

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: CustomBlobStorage
      x-fern-type-name: CustomBlobStorage
    CustomJWTAuthIntegration:
      description: |-
        +label=Custom JWT Auth Integration
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - issuer
        - jwks_uri
      properties:
        type:
          description: +value=integration/jwt-auth/custom
          type: string
          enum:
            - integration/jwt-auth/custom
        name:
          description: >-
            +label=Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        issuer:
          description: |-
            +label=Issuer
            +usage=The base URL of the authentication provider.
            +sort=200
          type: string
          pattern: ^(https?:\/\/)?(([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})(\/[^\s]*)?$
        jwks_uri:
          description: |-
            +label=JWKS URI
            +usage=The JSON Web Key Set URI for JWT verification.
            +sort=300
          type: string
          pattern: ^https:\/\/(([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})(\/[^\s]*)?$
        login_provider:
          $ref: '#/components/schemas/OAuth2LoginProvider'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=500

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: CustomJWTAuthIntegration
      x-fern-type-name: CustomJWTAuthIntegration
    GenericSecretStoreIntegration:
      description: |-
        +label=Generic Secret Store
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - base_url
        - auth_data
      properties:
        type:
          description: +value=integration/secret-store/custom/generic
          type: string
          enum:
            - integration/secret-store/custom/generic
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        base_url:
          description: |-
            +label=Base URL
            +usage=Base URL of the generic secret server.
            +sort=200
          type: string
          pattern: ^https?://[^/\s?#]+/?$
        auth_data:
          $ref: '#/components/schemas/CustomHeaderAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=10010

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: GenericSecretStoreIntegration
      x-fern-type-name: GenericSecretStoreIntegration
    JfrogArtifactsRegistry:
      description: |-
        +label=Jfrog Artifacts Registry
        +icon=jfrog
      type: object
      required:
        - type
        - name
        - registry_url
      properties:
        type:
          description: +value=integration/docker-registry/jfrog
          type: string
          enum:
            - integration/docker-registry/jfrog
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        registry_url:
          description: |-
            +label=Registry URL
            +usage=The URL of the registry.
            +sort=200
          type: string
          pattern: >-
            ^(https?:\/\/)?([-a-zA-Z0-9._\-+/]{1,256}\.[a-zA-Z0-9()]{1,6})(\/([\w.-]+))?(\/?)$
        auth_data:
          $ref: '#/components/schemas/JfrogBasicAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: JfrogArtifactsRegistry
      x-fern-type-name: JfrogArtifactsRegistry
    TTLRegistry:
      description: |-
        +label=TTL
        +icon=docker
      type: object
      required:
        - type
        - name
        - registry_url
      properties:
        type:
          description: +value=integration/docker-registry/ttl
          type: string
          enum:
            - integration/docker-registry/ttl
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        registry_url:
          description: |-
            +label=Registry URL
            +usage=The URL of the registry.
            +sort=200
          type: string
          default: https://ttl.sh
          pattern: ^(https?:\/\/)?(ttl\.sh)$
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: TTLRegistry
      x-fern-type-name: TTLRegistry
    TrueFoundryDBSSM:
      description: |-
        +label=TrueFoundry DBSSM
        +icon=truefoundry
      type: object
      required:
        - type
        - name
      properties:
        type:
          description: +value=integration/secret-store/truefoundry/db
          type: string
          enum:
            - integration/secret-store/truefoundry/db
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 32 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,30}[a-z0-9]$
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: TrueFoundryDBSSM
      x-fern-type-name: TrueFoundryDBSSM
    QuayArtifactsRegistry:
      description: |-
        +label=Quay Artifacts Registry
        +icon=quay
      type: object
      required:
        - type
        - name
        - registry_url
      properties:
        type:
          description: +value=integration/docker-registry/quay
          type: string
          enum:
            - integration/docker-registry/quay
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        registry_url:
          description: |-
            +label=Registry URL
            +usage=The URL of the registry.
            +sort=200
          type: string
          pattern: ^(https?://)?(quay\.io)(/([\w.-]+))?(/?)$
        auth_data:
          $ref: '#/components/schemas/QuayBasicAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: QuayArtifactsRegistry
      x-fern-type-name: QuayArtifactsRegistry
    SlackWebhookIntegration:
      description: |-
        +label=Slack Webhook Integration
        +icon=https://assets.production.truefoundry.com/slack.svg
      type: object
      required:
        - type
        - name
        - auth_data
      properties:
        type:
          description: +value=integration/notification-channel/slack-webhook
          type: string
          enum:
            - integration/notification-channel/slack-webhook
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +uiProps={"disableEdit":false}

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/SlackWebhookAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=10005

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: SlackWebhookIntegration
      x-fern-type-name: SlackWebhookIntegration
    SlackBotIntegration:
      description: |-
        +label=Slack Bot Integration
        +icon=https://assets.production.truefoundry.com/slack.svg
      type: object
      required:
        - type
        - name
        - auth_data
      properties:
        type:
          description: +value=integration/notification-channel/slack-bot
          type: string
          enum:
            - integration/notification-channel/slack-bot
        name:
          description: >-
            +label=Display Name

            +uiProps={"disableEdit":false}

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/SlackBotAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=10005

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: SlackBotIntegration
      x-fern-type-name: SlackBotIntegration
    WebhookIntegration:
      description: |-
        +label=Webhook Integration
        +icon=https://assets.production.truefoundry.com/webhook.svg
      type: object
      required:
        - type
        - name
        - webhook_url
      properties:
        type:
          description: +value=integration/gateway-request-logger/webhook
          type: string
          enum:
            - integration/gateway-request-logger/webhook
        name:
          description: >-
            +label=Display Name

            +uiProps={"disableEdit":true}

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        webhook_url:
          description: |-
            +label=Webhook URL
            +usage=The URL of the webhook endpoint
            +sort=200
          type: string
        auth_data:
          description: |-
            +label=Authentication
            +usage=Authentication configuration for the webhook
            +sort=300
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/WebhookBasicAuth'
            - $ref: '#/components/schemas/WebhookBearerAuth'
          discriminator:
            propertyName: type
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=10005

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: WebhookIntegration
      x-fern-type-name: WebhookIntegration
    PagerDutyIntegration:
      description: |-
        +label=PagerDuty Integration
        +icon=https://assets.production.truefoundry.com/pagerduty.png
      type: object
      required:
        - type
        - name
        - auth_data
      properties:
        type:
          description: +value=integration/notification-channel/pagerduty
          type: string
          enum:
            - integration/notification-channel/pagerduty
        name:
          description: >-
            +label=Display Name

            +sort=100

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +uiProps={"disableEdit":false}
          type: string
          pattern: ^[a-z](?:[a-z0-9]|-(?!-)){1,34}[a-z0-9]$
        auth_data:
          $ref: '#/components/schemas/PagerDutyIntegrationKeyAuth'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=10005

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
      title: PagerDutyIntegration
      x-fern-type-name: PagerDutyIntegration
    HashicorpVaultIntegration:
      description: |-
        +label=HashiCorp Vault Integration
        +icon=https://assets.production.truefoundry.com/hashicorp-vault.svg
      type: object
      required:
        - type
        - name
        - vault_url
        - kv_mount_path
        - auth_data
      properties:
        type:
          description: +value=integration/secret-store/hashicorp/vault
          type: string
          enum:
            - integration/secret-store/hashicorp/vault
        name:
          description: >-
            +label=Display Name

            +usage=The name of the integration that will be displayed in the
            TrueFoundry UI.

            +sort=100

            +message=3 to 36 lower case characters long alphanumeric word, may
            contain - in between, cannot start with a number
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        vault_url:
          description: >-
            +label=Vault URL

            +usage=The URL of the HashiCorp Vault server (e.g.,
            https://vault.example.com:8200).

            +sort=200
          type: string
          pattern: ^https?://[^\s]+$
        kv_mount_path:
          description: |-
            +label=KV Mount Path
            +usage=Mount path of the KV v2 engine vault to use for secrets.
            +sort=250
          type: string
        root_path:
          type: string
          pattern: ^[^/\s]+(?:/[^/\s]+)*$
        auth_data:
          description: |-
            +label=Auth Data
            +usage=Authentication data for the Vault integration.
            +sort=300
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/HashicorpTokenAuth'
            - $ref: '#/components/schemas/HashicorpAppRoleAuth'
          discriminator:
            propertyName: type
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=10010

            +uiType=AuthorizedSubjects
          type: array
          items:
            type: string
        namespace:
          description: >-
            +label=Namespace

            +usage=Namespace for vault operations, used for multi-tenancy in
            HashiCorp Vault Enterprise. Optional for open source Vault.

            +label=Namespace

            +sort=3000

            +isAdvanced=true
          type: string
      title: HashicorpVaultIntegration
      x-fern-type-name: HashicorpVaultIntegration
    ModelType:
      description: |-
        +label=Model Type
        +uiType=Select
      type: string
      enum:
        - chat
        - completion
        - embedding
        - realtime
        - rerank
        - audio_transcription
        - audio_translation
        - text_to_speech
        - moderation
        - image
        - responses
      title: ModelType
      x-fern-type-name: ModelType
    ModelCostMetric:
      type: object
      oneOf:
        - $ref: '#/components/schemas/PerThousandTokensCostMetric'
        - $ref: '#/components/schemas/PerThousandEmbeddingTokensCostMetric'
        - $ref: '#/components/schemas/PublicCostMetric'
      title: ModelCostMetric
      x-fern-type-name: ModelCostMetric
    CohereModel:
      description: |-
        +label=Cohere Model
        +icon=cohere
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/cohere
          type: string
          enum:
            - integration/model/cohere
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Cohere model identifier (e.g., command-a-03-2025,
            command-r-03-2024, command-r-03-2024). This is the standard model id
            from Cohere's model catalog.

            +message=Model ID must not be empty
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Cohere model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: CohereModel
      x-fern-type-name: CohereModel
    AI21Model:
      description: |-
        +label=AI21 Model
        +icon=ai21
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/ai21
          type: string
          enum:
            - integration/model/ai21
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The AI21 model identifier (e.g., jamba-large, jamba-mini).
            This is the model name from AI21's model catalog.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the AI21 model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: AI21Model
      x-fern-type-name: AI21Model
    AnthropicModel:
      description: |-
        +label=Anthropic Model
        +icon=anthropic
      type: object
      required:
        - name
        - model_id
        - type
        - model_types
      properties:
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Anthropic model identifier (e.g.,
            claude-3-5-sonnet-20241022, claude-3-opus-20240229,
            claude-3-haiku-20240307). This is the standard model name from
            Anthropic's API documentation.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        type:
          description: +value=integration/model/anthropic
          type: string
          enum:
            - integration/model/anthropic
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Anthropic model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: AnthropicModel
      x-fern-type-name: AnthropicModel
    DeepinfraModel:
      description: |-
        +label=DeepInfra Model
        +icon=deepinfra
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/deepinfra
          type: string
          enum:
            - integration/model/deepinfra
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The DeepInfra model identifier (e.g.,
            meta-llama/Llama-3.3-70B-Instruct-Turbo,
            mistralai/Mistral-Small-24B-Instruct-2501). This is the full model
            path from DeepInfra's model catalog.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the DeepInfra model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: DeepinfraModel
      x-fern-type-name: DeepinfraModel
    GroqModel:
      description: |-
        +label=Groq Model
        +icon=groq
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/groq
          type: string
          enum:
            - integration/model/groq
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Groq model identifier (e.g., llama-3.1-8b-instant,
            openai/gpt-oss-20b). This is the standard model id from Groq's model
            catalog.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Groq model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: GroqModel
      x-fern-type-name: GroqModel
    MistralAIModel:
      description: |-
        +label=Mistral AI Model
        +icon=mistral-ai
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/mistral-ai
          type: string
          enum:
            - integration/model/mistral-ai
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Mistral AI model identifier (e.g., mistral-medium-2312,
            ministral-8b-2410). This is the API Endpoint name from Mistral AI's
            model catalog.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Mistral AI model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: MistralAIModel
      x-fern-type-name: MistralAIModel
    PerplexityAIModel:
      description: |-
        +label=Perplexity AI Model
        +icon=perplexity-ai
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/perplexity-ai
          type: string
          enum:
            - integration/model/perplexity-ai
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Perplexity AI model identifier (e.g., sonar-reasoning,
            sonar). This is the model name from Perplexity AI's API
            documentation.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Perplexity AI model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: PerplexityAIModel
      x-fern-type-name: PerplexityAIModel
    TogetherAIModel:
      description: |-
        +label=Together AI Model
        +icon=together-ai
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/together-ai
          type: string
          enum:
            - integration/model/together-ai
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Together AI model identifier (e.g.,
            deepseek-ai/DeepSeek-R1, moonshotai/Kimi-K2-Instruct-0905). This is
            the model name from Together AI's model catalog.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Together AI model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: TogetherAIModel
      x-fern-type-name: TogetherAIModel
    NomicModel:
      description: |-
        +label=Nomic Model
        +icon=nomic
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/nomic
          type: string
          enum:
            - integration/model/nomic
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: |-
            +label=Model ID
            +sort=2
            +usage=The name of the Nomic model to use
            +message=Model ID must not be empty
            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Nomic model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: NomicModel
      x-fern-type-name: NomicModel
    PalmModel:
      description: |-
        +label=PaLM Model
        +icon=palm
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/palm
          type: string
          enum:
            - integration/model/palm
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=Name to identify this PaLM model in the UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: |-
            +label=Model ID
            +sort=2
            +usage=The name of the PaLM model to use
            +message=Model ID must not be empty
            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the PaLM model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: PalmModel
      x-fern-type-name: PalmModel
    OllamaModel:
      description: |-
        +label=Ollama Model
        +icon=ollama
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/ollama
          type: string
          enum:
            - integration/model/ollama
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=Name to identify this Ollama model in the UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: |-
            +label=Model ID
            +sort=2
            +usage=The name of the Ollama model to use
            +message=Model ID must not be empty
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Ollama model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: OllamaModel
      x-fern-type-name: OllamaModel
    OpenAIModel:
      description: |-
        +label=OpenAI Model
        +icon=openai
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/openai
          type: string
          enum:
            - integration/model/openai
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The OpenAI model identifier (e.g., gpt-4o, gpt-4-turbo,
            gpt-3.5-turbo, o1-mini). This is the standard model name from
            OpenAI's API documentation.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the OpenAI model
            +uiType=Select
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: OpenAIModel
      x-fern-type-name: OpenAIModel
    DatabricksModel:
      description: |-
        +label=Databricks Model
        +icon=databricks
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/databricks
          type: string
          enum:
            - integration/model/databricks
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The serving endpoint name configured in your Databricks
            workspace. This is the endpoint you created in Databricks Model
            Serving, not the underlying model name.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Databricks model
            +uiType=Select
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: DatabricksModel
      x-fern-type-name: DatabricksModel
    ClouderaModel:
      description: |-
        +label=Cloudera Model
        +icon=cloudera
      type: object
      required:
        - type
        - name
        - model_id
        - base_url
        - model_types
      properties:
        type:
          description: +value=integration/model/cloudera
          type: string
          enum:
            - integration/model/cloudera
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The unique identifier of the Model in Cloudera Model
            Registry. This is the model ID assigned when you registered or
            deployed your model in the CDP Model Registry

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        base_url:
          description: >-
            +label=Base URL

            +sort=3

            +usage=The model endpoint URL up to the API version v1. To get the
            base_url, copy the model endpoint URL and delete the last two path
            components

            +message=Base URL must not be empty
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=4
            +usage=Specify the type of the Cloudera model
            +uiType=Select
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=6

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: ClouderaModel
      x-fern-type-name: ClouderaModel
    SelfHostedModel:
      description: |-
        +label=Self Hosted Model Integration
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - hosted_model_name
        - url
        - model_server
        - model_types
      properties:
        type:
          description: +value=integration/model/self-hosted-model
          type: string
          enum:
            - integration/model/self-hosted-model
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this self-hosted model in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        hosted_model_name:
          description: |-
            +label=Hosted Model Name
            +sort=2
            +usage=The name of the hosted model
          type: string
          pattern: ^\S+$
        url:
          description: >-
            +label=URL

            +sort=3

            +usage=The URL where the model is hosted

            +message=enter valid https/http URL that should not end with
            trailing slash
          type: string
          pattern: ^https?:\/\/[^\s\/]+(\/[^\s\/]*)*[^\/\s]$
        model_server:
          description: |-
            +sort=4
            +label=Model Server Type
            +usage=The type of model server being used
          type: string
          enum:
            - openai-compatible
            - cohere-rerank-v2
            - nemo-retriever
            - infinity
            - tei
            - vllm-openai
        tfy_application_id:
          description: +uiType=Hidden
          type: string
        model_types:
          description: |-
            +usage=Specify the type of the model
            +sort=4
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        tls_settings:
          $ref: '#/components/schemas/CustomTLSSettings'
        auth_data:
          description: |-
            +label=Auth Data
            +usage=Self Hosted Model authentication data for the integration.
            +sort=300
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/CustomBasicAuth'
            - $ref: '#/components/schemas/CustomBearerAuth'
          discriminator:
            propertyName: type
        headers:
          description: >-
            +label=Self Hosted Model Headers

            +usage=Self Hosted Model headers for the integration. Forwarded to
            the provider as is. For example: `{"Authorization": "APIKey
            <token>"}`

            +sort=500
          type: object
          additionalProperties:
            type: string
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=Hidden
          type: array
          items:
            type: string
      title: SelfHostedModel
      x-fern-type-name: SelfHostedModel
    OpenRouterModel:
      description: |-
        +label=OpenRouter Model
        +icon=openrouter
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/openrouter
          type: string
          enum:
            - integration/model/openrouter
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The OpenRouter model identifier (e.g.,
            anthropic/claude-3-haiku, openai/gpt-4o-mini). This is the model
            name from OpenRouter's model catalog.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the OpenRouter model
            +uiType=Select
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: OpenRouterModel
      x-fern-type-name: OpenRouterModel
    SambaNovaModel:
      description: |-
        +label=SambaNova Model
        +icon=sambanova
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/sambanova
          type: string
          enum:
            - integration/model/sambanova
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The SambaNova model identifier (e.g., DeepSeek-R1-0528,
            Meta-Llama-3.3-70B-Instruct). This is the standard Model ID from
            SambaNova's model catalog.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the SambaNova model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: SambaNovaModel
      x-fern-type-name: SambaNovaModel
    XAIModel:
      description: |-
        +label=xAI Model
        +icon=xai
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/xai
          type: string
          enum:
            - integration/model/xai
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The xAI model identifier (e.g., grok-4-1-fast-reasoning,
            grok-4-1-fast-non-reasoning, grok-4). This is the model name from
            xAI's model catalog.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the xAI model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: XAIModel
      x-fern-type-name: XAIModel
    BasetenModel:
      description: |-
        +label=Baseten Model
        +icon=baseten
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/baseten
          type: string
          enum:
            - integration/model/baseten
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Baseten model identifier. This is the model name from
            Baseten's model catalog.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Baseten model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: BasetenModel
      x-fern-type-name: BasetenModel
    CerebrasModel:
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/cerebras
          type: string
          enum:
            - integration/model/cerebras
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this model integration in the
            UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: >-
            +label=Model ID

            +sort=2

            +usage=The Cerebras model identifier (e.g.,
            llama-4-scout-17b-16e-instruct, llama3.1-70b). This is the standard
            model name from Cerebras's API documentation.

            +message=Model ID must not be empty

            +uiProps={"descriptionInline":true}
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Cerebras model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: CerebrasModel
      x-fern-type-name: CerebrasModel
    SnowflakeCortexModel:
      type: object
      required:
        - type
        - name
        - model_id
        - model_types
      properties:
        type:
          description: +value=integration/model/snowflake-cortex
          type: string
          enum:
            - integration/model/snowflake-cortex
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=Name to identify this Snowflake Cortex model in the UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        model_id:
          description: |-
            +label=Model ID
            +sort=2
            +usage=ID of the Snowflake Cortex model to use
            +message=Model ID must not be empty
          type: string
          pattern: ^\S+$
        model_types:
          description: |-
            +label=Model Types
            +sort=3
            +usage=Specify the type of the Snowflake Cortex model
            +uiProps={"disableAllOption": true}
          type: array
          items:
            $ref: '#/components/schemas/ModelType'
          default:
            - chat
        cost:
          $ref: '#/components/schemas/ModelCostMetric'
        authorized_subjects:
          description: >-
            +label=Access Control

            +sort=5

            +usage=List of subjects that are authorized to access this
            integration. List of user fqn in format <user_type>:<username>.

            +uiType=Hidden
          type: array
          items:
            type: string
      title: SnowflakeCortexModel
      x-fern-type-name: SnowflakeCortexModel
    VirtualModelModelType:
      type: string
      enum:
        - chat
        - completion
        - embedding
        - rerank
        - moderation
      title: VirtualModelModelType
      x-fern-type-name: VirtualModelModelType
    HeaderRoutingConfig:
      description: >-
        this union is mainly for gateway to parse the load balancing config from
        request header

        +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/WeightBasedLoadBalancing'
        - $ref: '#/components/schemas/LatencyBasedLoadBalancing'
        - $ref: '#/components/schemas/PriorityBasedLoadBalancing'
      discriminator:
        propertyName: type
      title: HeaderRoutingConfig
      x-fern-type-name: HeaderRoutingConfig
    CustomEndpoint:
      description: |-
        +label=Custom Endpoint
        +icon=puzzle-piece
      type: object
      required:
        - type
        - name
        - base_url
      properties:
        type:
          description: +value=integration/model/custom-endpoint
          type: string
          enum:
            - integration/model/custom-endpoint
        name:
          description: >-
            +label=Display Name

            +sort=1

            +usage=A descriptive name to identify this endpoint in the UI

            +message=2 to 62 characters long alphanumeric word, may contain - ,
            . or _ in between, cannot start with a number
          type: string
          pattern: ^[a-zA-Z](?:[a-zA-Z0-9\-._]{0,60}[a-zA-Z0-9])$
        base_url:
          description: >-
            +label=Base URL

            +sort=2

            +usage=The target base URL to proxy requests to (e.g.,
            https://my-service.example.com/v1)

            +message=Enter a valid https/http URL that should not end with a
            trailing slash
          type: string
          pattern: ^https?:\/\/[^\s\/]+(\/[^\s\/]*)*[^\/\s]$
        auth_data:
          $ref: '#/components/schemas/CustomHeaderAuth'
        headers:
          description: >-
            +label=Custom Headers

            +usage=Custom headers forwarded to the upstream endpoint with every
            request. For example: `{"X-Custom-Header": "value"}`

            +sort=500

            +isAdvanced=true
          type: object
          additionalProperties:
            type: string
        tls_settings:
          $ref: '#/components/schemas/CustomTLSSettings'
        authorized_subjects:
          description: >-
            +label=Access Control

            +usage=List of subjects that are authorized to access this endpoint.
            List of user fqn in format <user_type>:<username>.

            +sort=600

            +uiType=Hidden
          type: array
          items:
            type: string
      title: CustomEndpoint
      x-fern-type-name: CustomEndpoint
    VirtualMCPServerSource:
      description: +label=Virtual MCP Server Source
      type: object
      required:
        - name
      properties:
        name:
          description: >-
            +label=Server Name

            +usage=Name of the source MCP server within this group (use
            {provider_account_name}/{integration_name} format).

            +uiType=VirtualMCPServerSelection

            +sort=100

            +message=Must be in format:
            {provider_account_name}/{integration_name}
          type: string
          pattern: ^[a-z][a-z0-9\-]{1,34}[a-z0-9]/[a-z][a-z0-9\-]{1,34}[a-z0-9]$
        enabled_tools:
          description: |-
            +label=Enabled Tools
            +usage=List of tools to enable from this source server.
            +uiType=Hidden
            +sort=200
          type: array
          items:
            type: string
        auth_data:
          $ref: '#/components/schemas/MCPServerHeaderOverrideAuth'
      title: VirtualMCPServerSource
      x-fern-type-name: VirtualMCPServerSource
    EnforcingStrategy:
      description: >-
        +label=Enforcing Strategy

        +usage=Strategy for enforcing this guardrail: enforce (block on
        failure), enforce_but_ignore_on_error (continue on error but block on
        failure), or audit (only audit violations)

        +uiType=Select

        +sort=175
      type: string
      enum:
        - enforce_but_ignore_on_error
        - enforce
        - audit
      default: enforce_but_ignore_on_error
      title: EnforcingStrategy
      x-fern-type-name: EnforcingStrategy
    CustomBearerAuth:
      description: +label=Custom Bearer Auth
      type: object
      required:
        - type
        - bearer_token
      properties:
        type:
          description: +value=bearer-auth
          type: string
          enum:
            - bearer-auth
        bearer_token:
          description: |-
            +label=Bearer Token
            +usage=The bearer token for the custom bearer authentication.
            +sort=100
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
      title: CustomBearerAuth
      x-fern-type-name: CustomBearerAuth
    AzurePIICategory:
      type: string
      enum:
        - '*'
        - ABARoutingNumber
        - Address
        - Age
        - ARNationalIdentityNumber
        - AUBankAccountNumber
        - AUDriversLicenseNumber
        - AUMedicalAccountNumber
        - AUPassportNumber
        - AUTaxFileNumber
        - AUBusinessNumber
        - AUCompanyNumber
        - ATIdentityCard
        - ATTaxIdentificationNumber
        - ATValueAddedTaxNumber
        - AzureDocumentDBAuthKey
        - AzureIAASDatabaseConnectionAndSQLString
        - AzureIoTConnectionString
        - AzurePublishSettingPassword
        - AzureRedisCacheString
        - AzureSAS
        - AzureServiceBusString
        - AzureStorageAccountKey
        - AzureStorageAccountGeneric
        - BankAccountNumber
        - BENationalNumber
        - BEValueAddedTaxNumber
        - BRCPFNumber
        - BRLegalEntityNumber
        - BRNationalIDRG
        - BGUniformCivilNumber
        - CABankAccountNumber
        - CADriversLicenseNumber
        - CAHealthServiceNumber
        - CAPassportNumber
        - CAPersonalHealthID
        - CASocialInsuranceNumber
        - CLIdentityCardNumber
        - CNResidentIdentityCardNumber
        - CreditCardNumber
        - HRIdentityCardNumber
        - HRNationalIDCardNumber
        - HRPersonalIdentificationNumber
        - CYIdentityCard
        - CYTaxIdentificationNumber
        - CZPersonalIdentityNumber
        - Date
        - DateTime
        - DKPersonalIdentificationNumber
        - DEANumber
        - Email
        - EEPersonalIdentificationCode
        - EUDebitCardNumber
        - EUDriversLicenseNumber
        - EUGPSCoordinates
        - EUNationalIdentificationNumber
        - EUPassportNumber
        - EUSocialSecurityNumber
        - EUTaxIdentificationNumber
        - FIEuropeanHealthNumber
        - FINationalID
        - FIPassportNumber
        - FRDriversLicenseNumber
        - FRHealthInsuranceNumber
        - FRNationalID
        - FRPassportNumber
        - FRSocialSecurityNumber
        - FRTaxIdentificationNumber
        - FRValueAddedTaxNumber
        - DEDriversLicenseNumber
        - DEPassportNumber
        - DEIdentityCardNumber
        - DETaxIdentificationNumber
        - DEValueAddedTaxNumber
        - GRNationalIDCard
        - GRTaxIdentificationNumber
        - HKIdentityCardNumber
        - HUValueAddedTaxNumber
        - HUPersonalIdentificationNumber
        - HUTaxIdentificationNumber
        - INPermanentAccountNumber
        - INUniqueIdentificationNumber
        - IDIdentityCardNumber
        - InternationalBankingAccountNumber
        - IPAddress
        - IEPersonalPublicServiceNumber
        - ILBankAccountNumber
        - ILNationalID
        - ITDriversLicenseNumber
        - ITFiscalCode
        - ITValueAddedTaxNumber
        - JPBankAccountNumber
        - JPDriversLicenseNumber
        - JPPassportNumber
        - JPResidentRegistrationNumber
        - JPSocialInsuranceNumber
        - JPMyNumberCorporate
        - JPMyNumberPersonal
        - JPResidenceCardNumber
        - LVPersonalCode
        - LTPersonalCode
        - LUNationalIdentificationNumberNatural
        - LUNationalIdentificationNumberNonNatural
        - MYIdentityCardNumber
        - MTIdentityCardNumber
        - MTTaxIDNumber
        - NLCitizensServiceNumber
        - NLTaxIdentificationNumber
        - NLValueAddedTaxNumber
        - NZBankAccountNumber
        - NZDriversLicenseNumber
        - NZInlandRevenueNumber
        - NZMinistryOfHealthNumber
        - NZSocialWelfareNumber
        - NOIdentityNumber
        - Organization
        - Person
        - PersonType
        - PHUnifiedMultiPurposeIDNumber
        - PhoneNumber
        - PLIdentityCard
        - PLNationalID
        - PLPassportNumber
        - PLTaxIdentificationNumber
        - PLREGONNumber
        - PTCitizenCardNumber
        - PTTaxIdentificationNumber
        - ROPersonalNumericalCode
        - RUPassportNumberDomestic
        - RUPassportNumberInternational
        - SANationalID
        - SGNationalRegistrationIdentityCardNumber
        - SKPersonalNumber
        - SITaxIdentificationNumber
        - SIUniqueMasterCitizenNumber
        - ZAIdentificationNumber
        - KRResidentRegistrationNumber
        - ESDNI
        - ESSocialSecurityNumber
        - ESTaxIdentificationNumber
        - AzureSQLConnectionString
        - SENationalID
        - SEPassportNumber
        - SETaxIdentificationNumber
        - SWIFTCode
        - CHSocialSecurityNumber
        - TWNationalID
        - TWPassportNumber
        - TWResidentCertificate
        - THPopulationIdentificationCode
        - TRNationalIdentificationNumber
        - UKDriversLicenseNumber
        - UKElectoralRollNumber
        - UKNationalHealthNumber
        - UKNationalInsuranceNumber
        - UKUniqueTaxpayerNumber
        - USUKPassportNumber
        - USBankAccountNumber
        - USDriversLicenseNumber
        - USIndividualTaxpayerIdentification
        - USSocialSecurityNumber
        - URL
      title: AzurePIICategory
      x-fern-type-name: AzurePIICategory
    AzureContentSafetyCategory:
      type: string
      enum:
        - Hate
        - SelfHarm
        - Sexual
        - Violence
      title: AzureContentSafetyCategory
      x-fern-type-name: AzureContentSafetyCategory
    EnkryptAIKeyAuth:
      description: |-
        +label=Enkrypt AI API Key Auth
        +usage=Authentication method using Enkrypt AI API key
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +sort=100
            +usage=The API key for Enkrypt AI authentication
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: EnkryptAIKeyAuth
      x-fern-type-name: EnkryptAIKeyAuth
    PaloAltoPrismaAIRSKeyAuth:
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
            +sort=100
            +usage=API key for the Palo Alto  Prisma AIRS account
          type: string
          pattern: ^\S+$
      title: PaloAltoPrismaAIRSKeyAuth
      x-fern-type-name: PaloAltoPrismaAIRSKeyAuth
    FiddlerKeyAuth:
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
            +sort=100
            +usage=API key for the Fiddler account
          type: string
          pattern: ^\S+$
      title: FiddlerKeyAuth
      x-fern-type-name: FiddlerKeyAuth
    FiddlerGuardType:
      type: string
      enum:
        - fiddler-safety
        - fiddler-response-faithfulness
      title: FiddlerGuardType
      x-fern-type-name: FiddlerGuardType
    PangeaKeyAuth:
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
            +sort=100
            +usage=API key for the Pangea account
          type: string
          pattern: ^\S+$
      title: PangeaKeyAuth
      x-fern-type-name: PangeaKeyAuth
    PangeaGuardType:
      type: string
      enum:
        - textGuard
        - pii
      default: textGuard
      title: PangeaGuardType
      x-fern-type-name: PangeaGuardType
    PatronusKeyAuth:
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
            +sort=100
            +usage=API key for the Patronus account
          type: string
          pattern: ^\S+$
      title: PatronusKeyAuth
      x-fern-type-name: PatronusKeyAuth
    PatronusEvaluator:
      description: |-
        +label=Patronus Evaluators
        +discriminator=type
      type: object
      oneOf:
        - $ref: '#/components/schemas/PatronusAnswerRelevanceEvaluator'
        - $ref: '#/components/schemas/PatronusGliderEvaluator'
        - $ref: '#/components/schemas/PatronusJudgeEvaluator'
        - $ref: '#/components/schemas/PatronusPIIEvaluator'
        - $ref: '#/components/schemas/PatronusPhiEvaluator'
        - $ref: '#/components/schemas/PatronusToxicityEvaluator'
      discriminator:
        propertyName: type
      title: PatronusEvaluator
      x-fern-type-name: PatronusEvaluator
    PresetRegexPattern:
      description: >-
        +label=Predefined Regex Patterns

        +usage=Select one or more common regex patterns from the predefined
        list.
      type: object
      required:
        - presets
      properties:
        presets:
          description: >-
            +label=Predefined Patterns

            +usage=Select one or more common regex patterns from the predefined
            list.

            +placeholder=Select predefined patterns

            +sort=100

            +uiProps={"disableAllOption": true, "isSearchable": true, "noLabel":
            true, "descriptionInline":true}
          type: array
          items:
            $ref: '#/components/schemas/PresetRegexPatternEnum'
      title: PresetRegexPattern
      x-fern-type-name: PresetRegexPattern
    CustomRegexPattern:
      description: +label=Custom Regex Pattern
      type: object
      required:
        - pattern
      properties:
        pattern:
          description: >-
            +label=Regex Pattern

            +usage=A custom regex pattern to match against content (e.g.,
            "\\b\\d{3}-\\d{2}-\\d{4}\\b" for SSN)

            +sort=100
          type: string
        redaction_text:
          description: >-
            +label=Redaction Text

            +usage=Text to use when redacting matched content (only applicable
            in mutate mode). Defaults to '[REDACTED]'.

            +sort=200
          type: string
          default: '[REDACTED]'
      title: CustomRegexPattern
      x-fern-type-name: CustomRegexPattern
    OPAAuth:
      type: object
      oneOf:
        - $ref: '#/components/schemas/OPAHeaderAuth'
        - $ref: '#/components/schemas/OPABearerAuth'
      title: OPAAuth
      x-fern-type-name: OPAAuth
    GoogleModelArmorApiKeyAuth:
      description: |-
        +label=Google Model Armor API Key Auth
        +usage=Authentication method using a Google Cloud API key
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +sort=100
            +usage=Google Cloud API key with Model Armor access
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: GoogleModelArmorApiKeyAuth
      x-fern-type-name: GoogleModelArmorApiKeyAuth
    GoogleModelArmorKeyFileAuth:
      description: >-
        +label=Google Model Armor Key File Auth

        +usage=Authentication method using a Google Cloud service account key
        file
      type: object
      required:
        - type
        - key_file_content
      properties:
        type:
          description: +value=key-file
          type: string
          enum:
            - key-file
        key_file_content:
          description: >-
            +label=Key File Content

            +sort=100

            +usage=The JSON content of your Google Cloud service account key
            file with Model Armor permissions

            +message=Key file content must be valid JSON

            +uiType=SecretSelect

            +uiProps={"typeOptionsList":["json","secret"]}
          oneOf:
            - type: string
            - type: object
      title: GoogleModelArmorKeyFileAuth
      x-fern-type-name: GoogleModelArmorKeyFileAuth
    GraySwanCygnalApiKeyAuth:
      description: +label=GraySwan Cygnal API Key Auth
      type: object
      required:
        - type
        - api_key
      properties:
        type:
          description: +value=api-key
          type: string
          enum:
            - api-key
        api_key:
          description: |-
            +sort=100
            +usage=GraySwan API key for Cygnal access
            +message=API key must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: GraySwanCygnalApiKeyAuth
      x-fern-type-name: GraySwanCygnalApiKeyAuth
    AktoTokenAuth:
      description: |-
        +label=Akto Token Auth
        +usage=Authentication method using an Akto JWT token
      type: object
      required:
        - type
        - token
      properties:
        type:
          description: +value=token
          type: string
          enum:
            - token
        token:
          description: |-
            +sort=100
            +usage=JWT token for authenticating with the Akto guardrails service
            +message=Token must not be empty
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: AktoTokenAuth
      x-fern-type-name: AktoTokenAuth
    TrojAIClientIdAuth:
      description: |-
        +label=TrojAI Client ID Auth
        +usage=Authentication using a TrojAI firewall client ID (x-eag-clientid)
      type: object
      required:
        - type
        - client_id
      properties:
        type:
          description: +value=client-id
          type: string
          enum:
            - client-id
        client_id:
          description: >-
            +sort=100

            +usage=The client ID that identifies your TrojAI firewall policy
            (sent as x-eag-clientid header)

            +message=Client ID must not be empty

            +uiType=SecretSelect

            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: TrojAIClientIdAuth
      x-fern-type-name: TrojAIClientIdAuth
    PythonTaskConfig:
      description: +docs=Describes the configuration for the python function task
      type: object
      required:
        - type
        - image
      properties:
        type:
          description: +value=python-task-config
          type: string
          enum:
            - python-task-config
        image:
          description: |-
            +label=Image Spec
            +docs=Specification for the image to be used for the task
            +sort=100
            +usage=Specify the image spec for the task
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/TaskPythonBuild'
            - $ref: '#/components/schemas/TaskDockerFileBuild'
          discriminator:
            propertyName: type
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the task
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=200
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        resources:
          $ref: '#/components/schemas/Resources'
        mounts:
          description: |-
            +usage=Configure data to be mounted to Workflow pod(s) as a volume.
            +sort=400
            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/SecretMount'
              - $ref: '#/components/schemas/StringDataMount'
              - $ref: '#/components/schemas/VolumeMount'
        service_account:
          description: |-
            +label=Service Account
            +sort=500
          type: string
      title: PythonTaskConfig
      x-fern-type-name: PythonTaskConfig
    ContainerTaskConfig:
      type: object
      required:
        - type
        - image
      properties:
        type:
          description: +value=container-task-config
          type: string
          enum:
            - container-task-config
        image:
          description: >-
            +docs=Specify whether you want to deploy a Docker image or build and
            deploy from source code

            +label=Deploy a Docker image or build and deploy from source code

            +icon=fa-solid fa-cloud-arrow-up:#21B6A8

            +sort=200

            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/Build'
            - $ref: '#/components/schemas/Image'
          discriminator:
            propertyName: type
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the task
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=200
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        resources:
          $ref: '#/components/schemas/Resources'
        mounts:
          description: |-
            +usage=Configure data to be mounted to Workflow pod(s) as a volume.
            +sort=400
            +todo=Add discriminator for union types in array
          type: array
          items:
            type: object
            oneOf:
              - $ref: '#/components/schemas/SecretMount'
              - $ref: '#/components/schemas/StringDataMount'
              - $ref: '#/components/schemas/VolumeMount'
        service_account:
          description: |-
            +label=Service Account
            +sort=500
          type: string
      title: ContainerTaskConfig
      x-fern-type-name: ContainerTaskConfig
    PySparkTaskConfig:
      type: object
      required:
        - type
        - image
        - driver_config
        - executor_config
      properties:
        type:
          description: +value=pyspark-task-config
          type: string
          enum:
            - pyspark-task-config
        image:
          $ref: '#/components/schemas/TaskPySparkBuild'
        driver_config:
          $ref: '#/components/schemas/SparkDriverConfig'
        executor_config:
          $ref: '#/components/schemas/SparkExecutorConfig'
        spark_conf:
          description: >-
            +label=Spark Config Properties

            +usage=Extra configuration properties to be passed to the spark job.
            [Docs](https://spark.apache.org/docs/latest/configuration.html)

            +icon=fa-gear:#68BBE3

            +sort=3000
          type: object
          default: null
          additionalProperties:
            type: string
            minLength: 1
          nullable: true
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the task
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=4000
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        mounts:
          description: |-
            +usage=Configure data to be mounted to Workflow pod(s) as a volume.
            +sort=5000
          type: array
          items:
            $ref: '#/components/schemas/VolumeMount'
        service_account:
          description: |-
            +label=Service Account
            +sort=6000
          type: string
      title: PySparkTaskConfig
      x-fern-type-name: PySparkTaskConfig
    DatabricksJobTaskConfig:
      description: >-
        +docs=Task that triggers a Databricks job via API and polls until
        completion. Uses image and optional resources; no env or mounts.
        Execution is handled by the CLI (trigger + poll).

        +label=Databricks Job Task

        +icon=fa-solid fa-database:#FF3621
      type: object
      required:
        - type
        - image
        - workspace_host
        - job_id
      properties:
        type:
          description: +value=databricks-job-task-config
          type: string
          enum:
            - databricks-job-task-config
        image:
          description: |-
            +label=Image Spec
            +docs=Specification for the image to be used for the task
            +sort=100
            +usage=Specify the image spec for the task
            +discriminator=type
          type: object
          oneOf:
            - $ref: '#/components/schemas/TaskPythonBuild'
            - $ref: '#/components/schemas/TaskDockerFileBuild'
          discriminator:
            propertyName: type
        workspace_host:
          description: >-
            +label=Workspace Host

            +usage=Databricks workspace URL, e.g.
            https://<workspace>.cloud.databricks.com

            +sort=150
          type: string
        job_id:
          description: >-
            +label=Job ID

            +usage=The Databricks job ID to run (from Databricks workspace
            Jobs).

            +sort=200
          type: string
        service_account:
          description: |-
            +label=Service Account
            +sort=500
          type: string
        job_parameters:
          description: >-
            +label=Job Parameters

            +usage=Optional parameters to pass to the job run (e.g. notebook
            params, jar params).

            +sort=300
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        timeout_seconds:
          description: >-
            +label=Timeout (seconds)

            +usage=Maximum seconds to wait for the job run to complete. Used by
            CLI when polling.

            +sort=500
          type: number
          default: null
          nullable: true
        skip_wait_for_completion:
          description: >-
            +label=Skip Wait for Completion

            +usage=If false, the task waits for the Databricks job run to
            complete (trigger and poll). If true, only triggers the job and
            returns. Default false.

            +sort=505
          type: boolean
          default: false
        env:
          description: >-
            +label=Environment Variables

            +usage=Configure environment variables to be injected in the task
            either as plain text or secrets.
            [Docs](https://docs.truefoundry.com/docs/env-variables)

            +icon=fa-globe

            +sort=200
          type: object
          default: null
          additionalProperties:
            type: string
          nullable: true
        resources:
          $ref: '#/components/schemas/Resources'
      title: DatabricksJobTaskConfig
      x-fern-type-name: DatabricksJobTaskConfig
    LoadBalancingWhen:
      description: +label=Rule Conditions
      type: object
      properties:
        subjects:
          description: |-
            +usage=List of subjects that this rule applies to
            +uiProps={"descriptionInline":true}
            +label=Subjects
          type: array
          items:
            type: string
        models:
          description: |-
            +usage=List of models that this rule applies to
            +uiType=IntegrationSelect
            +uiProps={"integrationType":"model"}
            +label=Models
          type: array
          items:
            type: string
        metadata:
          description: |-
            +usage=Metadata key-value pairs that this rule applies to
            +uiProps={"descriptionInline":true}
            +label=Metadata
          type: object
          additionalProperties:
            type: string
      title: LoadBalancingWhen
      x-fern-type-name: LoadBalancingWhen
    TargetConditionGroup:
      type: object
      required:
        - operator
        - conditions
      properties:
        operator:
          type: string
          enum:
            - or
        conditions:
          $ref: '#/components/schemas/TargetClause'
      title: TargetConditionGroup
      x-fern-type-name: TargetConditionGroup
    SubjectConditionGroup:
      type: object
      required:
        - operator
        - conditions
      properties:
        operator:
          type: string
          enum:
            - and
            - or
          default: and
        conditions:
          $ref: '#/components/schemas/SubjectClause'
      title: SubjectConditionGroup
      x-fern-type-name: SubjectConditionGroup
    OTELTracesExporterSpanAttributeFilter:
      description: >-
        +label=Span Filter

        +usage=Filter configuration for filtering spans by attributes before
        exporting
      type: object
      required:
        - field
        - key
        - operator
        - value
      properties:
        field:
          description: |-
            Only keys that have string values are supported for now
            +usage=Type of attribute field to filter on
            +label=Field
          type: string
          enum:
            - spanAttributes
        key:
          description: |-
            +usage=Key name of the span attribute to filter on
            +label=Key
          type: string
        operator:
          description: |-
            +usage=Comparison operator to use for filtering
            +label=Operator
          type: string
          enum:
            - IN
        value:
          description: +label=Values
          type: array
          items:
            type: string
      title: OTELTracesExporterSpanAttributeFilter
      x-fern-type-name: OTELTracesExporterSpanAttributeFilter
    DataAccessRuleBase:
      description: |-
        Base rule structure shared by all data access rules
        +label=Data Access Rule
      type: object
      required:
        - id
        - subjects
        - data_types
      properties:
        id:
          description: |-
            +usage=Unique identifier for the rule
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Rule ID
          type: string
          pattern: ^[A-Za-z{}][A-Za-z0-9{}._-]{0,60}[A-Za-z0-9{}._]$
        description:
          description: |-
            +usage=Description of the rule
            +uiProps={"descriptionInline":true}
            +sort=2
            +label=Description
          type: string
        enabled:
          description: |-
            +usage=Whether this rule is enabled
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Enabled
          type: boolean
          default: true
        subjects:
          description: >-
            +usage=List of subjects that this rule applies to (users, teams,
            roles)

            +uiType=AuthorizedSubjects

            +uiProps={"descriptionInline":true, "optionTypes": ["users","teams",
            "roles"]}

            +sort=4

            +label=Subjects
          type: array
          items:
            type: string
        data_types:
          description: |-
            +usage=List of data types that this rule applies to
            +uiType=MultiSelect
            +uiProps={"descriptionInline":true}
            +sort=5
            +label=Data Types
          type: array
          items:
            $ref: '#/components/schemas/GatewayDataType'
          default:
            - traces
            - metrics
      title: DataAccessRuleBase
      x-fern-type-name: DataAccessRuleBase
    GatewayDataMetadataFilter:
      description: +label=Metadata Filter
      type: object
      required:
        - field
        - key
        - operator
        - value
      properties:
        field:
          description: |-
            +value=metadata
            +usage=Filter type
          type: string
          enum:
            - metadata
        key:
          description: |-
            +usage=Metadata key to filter on
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Key
          type: string
        operator:
          description: |-
            +usage=Comparison operator
            +uiProps={"descriptionInline":true}
            +sort=2
            +label=Operator
          type: string
          enum:
            - IN
        value:
          description: |-
            +usage=Values to match against
            +uiProps={"descriptionInline":true}
            +sort=3
            +label=Values
          type: array
          items:
            type: string
      title: GatewayDataMetadataFilter
      x-fern-type-name: GatewayDataMetadataFilter
    GatewayDataCreatedByFilter:
      description: +label=Created By Filter
      type: object
      required:
        - field
        - operator
        - value
      properties:
        field:
          description: |-
            +value=created_by
            +usage=Filter type
          type: string
          enum:
            - created_by
        operator:
          description: |-
            +usage=Comparison operator
            +uiProps={"descriptionInline":true}
            +sort=1
            +label=Operator
          type: string
          enum:
            - IN
        value:
          description: >-
            +usage=List of subjects (users, virtual accounts, teams) to filter
            by

            +uiType=AuthorizedSubjects

            +uiProps={"descriptionInline":true, "optionTypes":
            ["users","virtualAccounts","teams"]}

            +sort=2

            +label=Created By Subjects
          type: array
          items:
            type: string
      title: GatewayDataCreatedByFilter
      x-fern-type-name: GatewayDataCreatedByFilter
    GatewayDataRoutingConfigDestinationMetadataCondition:
      description: |-
        +label=Metadata
        +usage=Metadata condition for data routing destinations.
      type: object
      required:
        - field
        - key
        - operator
        - value
      properties:
        field:
          description: |-
            +value=metadata
            +label=Field
            +usage=Field to evaluate the condition on
            +sort=2
          type: string
          enum:
            - metadata
        key:
          description: |-
            +label=Key
            +usage=Metadata key to match
            +sort=3
          type: string
        operator:
          description: |-
            +label=Operator
            +usage=Operator to use for comparison
            +sort=4
          type: string
          enum:
            - IN
        value:
          description: |-
            +label=Value
            +usage=Value to compare against
            +sort=5
          type: array
          items:
            type: string
      title: GatewayDataRoutingConfigDestinationMetadataCondition
      x-fern-type-name: GatewayDataRoutingConfigDestinationMetadataCondition
    GatewayDataRoutingConfigDestinationCreatedByCondition:
      description: |-
        +label=Created By
        +usage=Created by condition for data routing destinations.
      type: object
      required:
        - field
        - operator
        - value
      properties:
        field:
          description: |-
            +value=created_by
            +label=Field
            +usage=Field to evaluate the condition on
            +sort=2
          type: string
          enum:
            - created_by
        operator:
          description: |-
            +label=Operator
            +usage=Operator to use for comparison
            +sort=3
          type: string
          enum:
            - IN
        value:
          description: |-
            +label=Value
            +usage=Value to compare against
            +sort=4
          type: array
          items:
            type: string
      title: GatewayDataRoutingConfigDestinationCreatedByCondition
      x-fern-type-name: GatewayDataRoutingConfigDestinationCreatedByCondition
    CustomServerHeaderAuth:
      description: |-
        +label=Header Auth
        +usage=Static API key or token authentication via request headers.
      type: object
      required:
        - type
        - headers
      properties:
        type:
          description: +value=header
          type: string
          enum:
            - header
        headers:
          type: object
          additionalProperties:
            type: string
      title: CustomServerHeaderAuth
      x-fern-type-name: CustomServerHeaderAuth
    CustomServerPassthrough:
      description: >-
        +label=Token Passthrough

        +usage=Authenticate using your existing TrueFoundry account (PAT or
        Virtual Accounts). Your credentials are automatically passed to the
        server.
      type: object
      required:
        - type
      properties:
        type:
          description: +value=passthrough
          type: string
          enum:
            - passthrough
      title: CustomServerPassthrough
      x-fern-type-name: CustomServerPassthrough
    A2AFramework:
      description: |-
        +label=A2A
        +sort=500
      type: object
      required:
        - type
        - agent_card_path
      properties:
        type:
          description: |-
            +value=a2a
            +label=Type
            +uiType=Hidden
          type: string
          enum:
            - a2a
        agent_card_path:
          description: |-
            +label=Agent Card Path
            +usage=Path to the agent card JSON, relative to the base URL
          type: string
          default: /.well-known/agent-card.json
      title: A2AFramework
      x-fern-type-name: A2AFramework
    CustomFramework:
      description: |-
        +label=Custom
        +sort=500
      type: object
      required:
        - type
      properties:
        type:
          description: |-
            +value=custom
            +label=Type
            +uiType=Hidden
          type: string
          enum:
            - custom
      title: CustomFramework
      x-fern-type-name: CustomFramework
    AgentSkill:
      type: object
      required:
        - id
        - name
        - description
      properties:
        id:
          type: string
        name:
          type: string
        description:
          type: string
        tags:
          type: array
          items:
            type: string
        examples:
          type: array
          items:
            type: string
        inputModes:
          type: array
          items:
            type: string
        outputModes:
          type: array
          items:
            type: string
      title: AgentSkill
      x-fern-type-name: AgentSkill
    AzureOAuth:
      description: +label=Azure OAuth
      type: object
      required:
        - type
        - tenant_id
        - client_id
        - client_secret
        - subscription_id
      properties:
        type:
          description: +value=oauth
          type: string
          enum:
            - oauth
        tenant_id:
          description: |-
            +label=Azure Tenant ID
            +usage=The Azure Active Directory tenant (directory) ID.
            +sort=100
          type: string
          pattern: ^\S+$
        client_id:
          description: >-
            +label=Azure Client ID

            +usage=The client (application) ID of an App Registration in the
            tenant.

            +sort=200
          type: string
          pattern: ^\S+$
        client_secret:
          description: |-
            +label=Azure Client secret
            +sort=300
            +usage=A client secret that was generated for the App Registration.
            +uiType=SecretSelect
          type: string
          pattern: ^\S+$
        subscription_id:
          description: |-
            +label=Azure Subscription ID
            +usage=The ID of the target subscription. The value must be an UUID.
            +sort=400
          type: string
          pattern: ^\S+$
      title: AzureOAuth
      x-fern-type-name: AzureOAuth
    AzureBasicAuth:
      description: +label=Azure Basic Auth
      type: object
      required:
        - type
        - username
        - password
      properties:
        type:
          description: +value=basic-auth
          type: string
          enum:
            - basic-auth
        username:
          description: |-
            +label=Username
            +usage=The username for the Azure managed identity.
            +sort=100
          type: string
          pattern: ^\S+$
        password:
          description: |-
            +sort=200
            +label=Password
            +usage=The password for the Azure managed identity.
            +uiType=SecretSelect
          type: string
      title: AzureBasicAuth
      x-fern-type-name: AzureBasicAuth
    AzureConnectionStringAuth:
      description: +label=Azure Connection String
      type: object
      required:
        - type
        - connection_string
      properties:
        type:
          description: +value=connection-string
          type: string
          enum:
            - connection-string
        connection_string:
          description: |-
            +label=Connection String
            +usage=The connection string for the Azure managed identity.
            +sort=100
          type: string
          pattern: ^\S+$
      title: AzureConnectionStringAuth
      x-fern-type-name: AzureConnectionStringAuth
    SMTPCredentials:
      description: +label=SMTP Credentials
      type: object
      required:
        - type
        - host
        - port
        - tls
      properties:
        type:
          description: |-
            +value=smtp-credentials
            +sort=250
          type: string
          enum:
            - smtp-credentials
        host:
          description: |-
            +label=Host
            +usage=The SMTP host of the email server.
            +sort=300
          type: string
          pattern: ^\S+$
        username:
          description: |-
            +label=Username
            +usage=The username for the SMTP server.
            +sort=400
          type: string
          minLength: 1
        password:
          description: |-
            +label=Password
            +usage=The password for the SMTP server.
            +sort=500
            +uiType=SecretSelect
          type: string
          minLength: 1
        port:
          description: |-
            +label=Port
            +usage=The port of the SMTP server.
            +sort=600
          type: integer
          default: 587
        tls:
          description: |-
            +label=TLS
            +usage=Whether to use TLS for the SMTP server.
            +sort=700
          type: boolean
          default: true
      title: SMTPCredentials
      x-fern-type-name: SMTPCredentials
    OAuth2LoginProvider:
      type: object
      properties:
        type:
          description: +value=oauth2
          type: string
          enum:
            - oauth2
      allOf:
        - $ref: '#/components/schemas/BaseOAuth2Login'
      required:
        - type
      title: OAuth2LoginProvider
      x-fern-type-name: OAuth2LoginProvider
    SlackWebhookAuth:
      type: object
      required:
        - type
        - webhook_url
      properties:
        type:
          description: +value=auth/slack-webhook
          type: string
          enum:
            - slack-webhook
        webhook_url:
          description: |-
            +label=Webhook URL
            +usage=The URL of the slack webhook
            +sort=100
          type: string
          pattern: >-
            ^https://hooks.slack.com/services/[a-zA-Z0-9]+/[a-zA-Z0-9]+/[a-zA-Z0-9]+$
      title: SlackWebhookAuth
      x-fern-type-name: SlackWebhookAuth
    SlackBotAuth:
      type: object
      required:
        - type
        - bot_token
      properties:
        type:
          description: +value=auth/slack-bot
          type: string
          enum:
            - slack-bot
        bot_token:
          description: >-
            +label=Bot Token

            +usage=The bot token for the slack bot to send messages to the
            channels. This bot should have the following permissions:
            chat:write, chat:write.public

            +sort=100
          type: string
      title: SlackBotAuth
      x-fern-type-name: SlackBotAuth
    WebhookBasicAuth:
      type: object
      required:
        - type
        - username
        - password
      properties:
        type:
          description: +value=auth/webhook-basic
          type: string
          enum:
            - webhook-basic
        username:
          description: |-
            +label=Username
            +usage=Basic auth username
            +sort=100
          type: string
        password:
          description: |-
            +label=Password
            +usage=Basic auth password
            +sort=200
            +uiType=SecretSelect
          type: string
      title: WebhookBasicAuth
      x-fern-type-name: WebhookBasicAuth
    WebhookBearerAuth:
      type: object
      required:
        - type
        - token
        - prefix
      properties:
        type:
          description: +value=auth/webhook-bearer
          type: string
          enum:
            - webhook-bearer
        token:
          description: |-
            +label=Token
            +usage=Bearer token for authentication
            +sort=100
            +uiType=SecretSelect
          type: string
        prefix:
          description: |-
            +label=Prefix
            +usage=Prefix for the token
            +sort=200
          type: string
          default: Bearer
      title: WebhookBearerAuth
      x-fern-type-name: WebhookBearerAuth
    PagerDutyIntegrationKeyAuth:
      description: +label=PagerDuty Credentails
      type: object
      required:
        - type
        - integration_key
        - integration_url
      properties:
        type:
          description: +value=pagerduty-integration-auth
          type: string
          enum:
            - pagerduty-integration-auth
        integration_key:
          description: |-
            +label=Integration Key
            +sort=100
            +usage=The integration key for the PagerDuty integration
            +uiType=SecretSelect
          type: string
        integration_url:
          description: >-
            +label=Integration URL

            +sort=200

            +usage=The integration URL for the PagerDuty integration

            +docs=The integration URL for the PagerDuty integration. This is the
            URL that will be used to send events to PagerDuty. eg.
            https://events.pagerduty.com/generic/2010-04-15/create_event.json

            +placeholder=https://events.pagerduty.com/generic/2010-04-15/create_event.json
          type: string
      title: PagerDutyIntegrationKeyAuth
      x-fern-type-name: PagerDutyIntegrationKeyAuth
    HashicorpTokenAuth:
      description: +label=Vault Token Auth
      type: object
      required:
        - type
        - token
      properties:
        type:
          description: +value=token
          type: string
          enum:
            - token
        token:
          description: >-
            +label=Hashicorp Vault Auth Token

            +usage=Vault authentication token that you want to use for
            authentication.

            +sort=100

            +uiType=SecretSelect
          type: string
          pattern: ^\S+$
      title: HashicorpTokenAuth
      x-fern-type-name: HashicorpTokenAuth
    HashicorpAppRoleAuth:
      description: >-
        +label=AppRole Auth

        +usage=Role ID and secret ID from your AppRole. Token policies must
        allow KV v2 data and metadata on this mount for your root path, secret
        paths, and the connection check.
      type: object
      required:
        - type
        - role_id
        - secret_id
      properties:
        type:
          description: +value=approle
          type: string
          enum:
            - approle
        role_id:
          description: |-
            +label=Role ID
            +usage=AppRole role ID for Vault authentication.
            +sort=100
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
        secret_id:
          description: |-
            +label=Secret ID
            +usage=AppRole secret ID for Vault authentication.
            +sort=200
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: HashicorpAppRoleAuth
      x-fern-type-name: HashicorpAppRoleAuth
    PerThousandTokensCostMetric:
      type: object
      required:
        - metric
        - value
      properties:
        metric:
          description: +value=per_1000_tokens
          type: string
          pattern: ^per_1000_tokens$
        value:
          $ref: '#/components/schemas/InputOutputBasedCostMetricValue'
      title: PerThousandTokensCostMetric
      x-fern-type-name: PerThousandTokensCostMetric
    PerThousandEmbeddingTokensCostMetric:
      type: object
      required:
        - metric
        - value
      properties:
        metric:
          description: +value=per_1000_embedding_tokens
          type: string
          pattern: ^per_1000_embedding_tokens$
        value:
          $ref: '#/components/schemas/NonNegativeFloat'
      title: PerThousandEmbeddingTokensCostMetric
      x-fern-type-name: PerThousandEmbeddingTokensCostMetric
    PublicCostMetric:
      type: object
      required:
        - metric
      properties:
        metric:
          description: +value=public_cost
          type: string
          pattern: ^public_cost$
      title: PublicCostMetric
      x-fern-type-name: PublicCostMetric
    PatronusAnswerRelevanceEvaluator:
      type: object
      required:
        - type
        - criteria
      properties:
        type:
          description: |-
            +label=Evaluator Type
            +usage=The type of the evaluator to use for the Patronus Guardrail.
            +sort=100
          type: string
          enum:
            - answer-relevance
        criteria:
          $ref: '#/components/schemas/PatronusAnswerRelevanceCriteria'
      title: PatronusAnswerRelevanceEvaluator
      x-fern-type-name: PatronusAnswerRelevanceEvaluator
    PatronusGliderEvaluator:
      type: object
      required:
        - type
        - criteria
      properties:
        type:
          description: |-
            +label=Evaluator Type
            +usage=The type of the evaluator to use for the Patronus Guardrail.
            +sort=100
          type: string
          enum:
            - glider
        criteria:
          $ref: '#/components/schemas/PatronusGliderCriteria'
      title: PatronusGliderEvaluator
      x-fern-type-name: PatronusGliderEvaluator
    PatronusJudgeEvaluator:
      type: object
      required:
        - type
        - criteria
      properties:
        type:
          description: |-
            +label=Evaluator Type
            +usage=The type of the evaluator to use for the Patronus Guardrail.
            +sort=100
          type: string
          enum:
            - judge
        criteria:
          $ref: '#/components/schemas/PatronusJudgeCriteria'
      title: PatronusJudgeEvaluator
      x-fern-type-name: PatronusJudgeEvaluator
    PatronusPIIEvaluator:
      type: object
      required:
        - type
        - criteria
      properties:
        type:
          description: |-
            +label=Evaluator Type
            +usage=The type of the evaluator to use for the Patronus Guardrail.
            +sort=100
          type: string
          enum:
            - pii
        criteria:
          $ref: '#/components/schemas/PatronusPIICriteria'
      title: PatronusPIIEvaluator
      x-fern-type-name: PatronusPIIEvaluator
    PatronusPhiEvaluator:
      type: object
      required:
        - type
        - criteria
      properties:
        type:
          description: |-
            +label=Evaluator Type
            +usage=The type of the evaluator to use for the Patronus Guardrail.
            +sort=100
          type: string
          enum:
            - phi
        criteria:
          $ref: '#/components/schemas/PatronusPhiCriteria'
      title: PatronusPhiEvaluator
      x-fern-type-name: PatronusPhiEvaluator
    PatronusToxicityEvaluator:
      type: object
      required:
        - type
        - criteria
      properties:
        type:
          description: |-
            +label=Evaluator Type
            +usage=The type of the evaluator to use for the Patronus Guardrail.
            +sort=100
          type: string
          enum:
            - toxicity
        criteria:
          $ref: '#/components/schemas/PatronusToxicityCriteria'
      title: PatronusToxicityEvaluator
      x-fern-type-name: PatronusToxicityEvaluator
    PresetRegexPatternEnum:
      type: string
      enum:
        - us_ssn
        - us_ssn_no_dash
        - email
        - us_phone
        - visa
        - mastercard
        - amex
        - discover
        - credit_card
        - aws_access_key
        - aws_secret_key
        - github_token
        - slack_token
        - generic_api_key
        - ipv4
        - ipv6
        - url
        - passport_us
        - passport_uk
        - passport_germany
        - passport_france
        - passport_netherlands
        - passport_canada
        - passport_india
        - passport_australia
        - passport_china
        - passport_japan
        - gender_sexual_orientation
        - race_ethnicity_national_origin
        - religion
        - age_discrimination
        - disability
        - marital_family_status
        - military_status
        - public_assistance
        - weapons_firearms
        - weapons_other
        - explosives
        - violence_threats
        - terrorism
        - self_harm_suicide
        - illegal_activities
        - harassment_hate
      title: PresetRegexPatternEnum
      x-fern-type-name: PresetRegexPatternEnum
    OPAHeaderAuth:
      description: |-
        +label=OPA Header Auth
        +usage=Authentication via custom HTTP headers for OPA server
      type: object
      required:
        - type
        - headers
      properties:
        type:
          description: +value=header
          type: string
          enum:
            - header
        headers:
          description: >-
            +label=Headers

            +usage=Custom headers to include in requests to OPA server (e.g.,
            {"Authorization": "Bearer <token>", "X-API-Key": "<key>"})

            +sort=100

            +uiType=KV

            +uiProps={"defaultForCreate": {"": ""}}
          type: object
          additionalProperties:
            type: string
            pattern: ^\S+$
      title: OPAHeaderAuth
      x-fern-type-name: OPAHeaderAuth
    OPABearerAuth:
      description: |-
        +label=OPA Bearer Auth
        +usage=Authentication using a bearer token for OPA server
      type: object
      required:
        - type
        - bearer_token
      properties:
        type:
          description: +value=bearer-auth
          type: string
          enum:
            - bearer-auth
        bearer_token:
          description: |-
            +label=Bearer Token
            +usage=The bearer token for OPA server authentication.
            +sort=100
            +uiType=SecretSelect
            +uiProps={"isPassword": true, "enableNew": true}
          type: string
          pattern: ^\S+$
      title: OPABearerAuth
      x-fern-type-name: OPABearerAuth
    TaskPythonBuild:
      description: |-
        +docs=Describes the configuration for the python build for a task
        +label=Python Buid Spec
        +icon=fa-brands fa-python:#306998
      type: object
      required:
        - type
      properties:
        type:
          description: |-
            +uiType=Hidden
            +value=task-python-build
          type: string
          enum:
            - task-python-build
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page
          type: string
        python_version:
          description: >-
            +label=Python version

            +usage=Python version to run your application. Should be one of the
            tags listed on [Official Python Docker
            Page](https://hub.docker.com/_/python)

            +message=Please enter a valid Python version tag
          type: string
          pattern: ^\d+(\.\d+){1,2}([\-\.a-z0-9]+)?$
        requirements_path:
          description: |-
            `Path to build context`
            +label=Path to requirements
            +usage=Path to `requirements.txt` relative to
            `Path to build context`
          type: string
        pip_packages:
          description: |-
            +label=Pip packages to install
            +usage=Define pip package requirements.
            In Python/YAML E.g. ["fastapi>=0.90,<1.0", "uvicorn"]
            +placeholder=Enter a pip package name E.g. fastapi>=0.90,<1.0
          type: array
          items:
            type: string
        apt_packages:
          description: |-
            +label=List of Debian packages to install.
            +usage=Debian packages to install via `apt get`.
            In Python/YAML E.g. ["git", "ffmpeg", "htop"]
            +placeholder=Enter a debian package name E.g. ffmpeg
          type: array
          items:
            type: string
        cuda_version:
          description: >-
            +label=CUDA Version

            +usage=Version of CUDA Toolkit and CUDNN to install in the image

            These combinations are based off of publically available docker
            images on docker hub

            You can also specify a valid tag of the form
            {cuda_version_number}-cudnn{cudnn_version_number}-{runtime|devel}-ubuntu{ubuntu_version}

            Refer https://hub.docker.com/r/nvidia/cuda/tags for valid set of
            values

            Note: We use deadsnakes ubuntu ppa to add Python that currently
            supports only Ubuntu 18.04, 20.04 and 22.04
          type: string
          pattern: >-
            ^((\d+\.\d+(\.\d+)?-cudnn(\d+)?-(runtime|devel)-ubuntu\d+\.\d+)|11\.0-cudnn8|11\.1-cudnn8|11\.2-cudnn8|11\.3-cudnn8|11\.4-cudnn8|11\.5-cudnn8|11\.6-cudnn8|11\.7-cudnn8|11\.8-cudnn8|12\.0-cudnn8|12\.1-cudnn8|12\.2-cudnn8|12\.3-cudnn9|12\.4-cudnn9|12\.5-cudnn9|12\.6-cudnn9|12\.8-cudnn9|12\.9-cudnn9)$
      title: TaskPythonBuild
      x-fern-type-name: TaskPythonBuild
    TaskDockerFileBuild:
      description: |-
        +docs=Describes the configuration for the docker build for a task
        +label=Docker File
        +icon=fa-brands fa-docker:#0db7ed
      type: object
      required:
        - type
        - dockerfile_path
      properties:
        type:
          description: |-
            +uiType=Hidden
            +value=dockerfile
          type: string
          enum:
            - task-dockerfile-build
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page
          type: string
        dockerfile_path:
          description: >-
            +label=Path to Dockerfile

            +usage=The file path of the Dockerfile relative to project root
            path.
          type: string
          default: ./Dockerfile
        build_args:
          description: +label=Build arguments to pass to docker build
          type: object
          additionalProperties:
            type: string
      title: TaskDockerFileBuild
      x-fern-type-name: TaskDockerFileBuild
    TaskPySparkBuild:
      description: |-
        +docs=Describes the configuration for the PySpark build for a task
        +label=PySpark Build Spec
        +icon=fa-brands fa-python:#306998
      type: object
      required:
        - type
        - spark_version
      properties:
        type:
          description: |-
            +uiType=Hidden
            +value=task-pyspark-build
          type: string
          enum:
            - task-pyspark-build
        spark_version:
          description: >-
            --- Spark Specific Field ---

            +label=Spark Version

            +usage=Spark version should match the spark version installed in the
            image.

            +sort=1001
          type: string
          default: 3.5.2
        container_image:
          description: >-
            +label=Container Image

            +usage=Custom container image URI. If provided, this image is used
            instead of the

            default Spark base image (public.ecr.aws/bitnami/spark). The image
            must be

            Debian-based and have Python and Spark pre-installed.

            +sort=1002
          type: string
          pattern: ^\S+$
        docker_registry:
          description: >-
            +docs=FQN of the container registry. You can the FQN of your desired
            container registry (or add one)

            in the  Integrations
            page[Integrations](https://app.truefoundry.tech/integrations?tab=docker-registry)
            page

            +label=Docker Registry

            +usage=FQN of the container registry. If you can't find your
            registry here,

            add it through the [Integrations](/integrations?tab=docker-registry)
            page

            +uiType=IntegrationSelect

            +uiProps={"descriptionKey":"registry_url","integrationType":"docker-registry"}
          type: string
        requirements_path:
          description: |-
            `Path to build context`
            +label=Path to requirements
            +usage=Path to `requirements.txt` relative to
            `Path to build context`
          type: string
        pip_packages:
          description: |-
            +label=Pip packages to install
            +usage=Define pip package requirements.
            In Python/YAML E.g. ["fastapi>=0.90,<1.0", "uvicorn"]
            +placeholder=Enter a pip package name E.g. fastapi>=0.90,<1.0
          type: array
          items:
            type: string
        apt_packages:
          description: |-
            +label=List of Debian packages to install.
            +usage=Debian packages to install via `apt get`.
            In Python/YAML E.g. ["git", "ffmpeg", "htop"]
            +placeholder=Enter a debian package name E.g. ffmpeg
          type: array
          items:
            type: string
      title: TaskPySparkBuild
      x-fern-type-name: TaskPySparkBuild
    TargetClause:
      type: object
      properties:
        mcpServers:
          $ref: '#/components/schemas/InNotInOperator'
        mcpTools:
          $ref: '#/components/schemas/MCPToolsOperator'
        model:
          $ref: '#/components/schemas/InNotInOperator'
      title: TargetClause
      x-fern-type-name: TargetClause
    SubjectClause:
      type: object
      properties:
        in:
          type: array
          items:
            type: string
        not_in:
          type: array
          items:
            type: string
      title: SubjectClause
      x-fern-type-name: SubjectClause
    GatewayDataType:
      type: string
      enum:
        - traces
        - metrics
      title: GatewayDataType
      x-fern-type-name: GatewayDataType
    BaseOAuth2Login:
      description: |-
        +label=OAuth2 Client Configuration
        +usage=OAuth2 client configuration to get the JWT.
      type: object
      required:
        - client_id
        - authorization_url
        - token_url
        - scopes
        - jwt_source
      properties:
        client_id:
          description: |-
            +label=Client ID
            +usage=client ID for OAuth2.
            +sort=200
          type: string
          pattern: ^\S+$
        client_secret:
          description: >-
            +label=Client Secret

            +usage=Client secret or the TrueFoundry secret containing the client
            secret for OAuth2.

            +sort=300
          type: string
          minLength: 1
        authorization_url:
          description: |-
            +label=Authorization URL
            +usage=URL for the authorization request
            +sort=400
          type: string
          pattern: ^https:\/\/(([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})(\/[^\s]*)?$
        token_url:
          description: |-
            +label=Token URL
            +usage=The endpoint to exchange auth code for tokens.
            +sort=500
          type: string
          pattern: ^https:\/\/(([a-zA-Z0-9-]{1,63}\.)+[a-zA-Z]{2,63})(\/[^\s]*)?$
        scopes:
          description: |-
            +label=Scopes
            +usage=List of scopes to request from the OAuth2 provider.
            +sort=600
          type: array
          items:
            type: string
            minLength: 1
        jwt_source:
          description: |-
            +label=JWT Source
            +usage=Source of the JWT token to be used for verification.
            +sort=700
          type: string
          enum:
            - access_token
            - id_token
          default: access_token
      title: BaseOAuth2Login
      x-fern-type-name: BaseOAuth2Login
    InputOutputBasedCostMetricValue:
      type: object
      required:
        - input
        - output
      properties:
        input:
          $ref: '#/components/schemas/NonNegativeFloat'
        output:
          $ref: '#/components/schemas/NonNegativeFloat'
        cache_read:
          $ref: '#/components/schemas/NonNegativeFloat'
        cache_write:
          $ref: '#/components/schemas/NonNegativeFloat'
        input_tiers:
          description: >-
            +label=Input Tiered Pricing

            +usage=Optional volume-based pricing tiers for input tokens. Tokens
            up to the first tier use base rate above.

            +sort=1000
          type: array
          items:
            $ref: '#/components/schemas/PrivatePricingTier'
        output_tiers:
          description: >-
            +label=Output Tiered Pricing

            +usage=Optional volume-based pricing tiers for output tokens. Tokens
            up to the first tier use base rate above.

            +sort=1100
          type: array
          items:
            $ref: '#/components/schemas/PrivatePricingTier'
        cache_read_tiers:
          description: >-
            +label=Cache Read Tiered Pricing

            +usage=Optional volume-based pricing tiers for cache read tokens.
            Tokens up to the first tier use base rate above.

            +sort=1200
          type: array
          items:
            $ref: '#/components/schemas/PrivatePricingTier'
        cache_write_tiers:
          description: >-
            +label=Cache Write Tiered Pricing

            +usage=Optional volume-based pricing tiers for cache write tokens.
            Tokens up to the first tier use base rate above.

            +sort=1300
          type: array
          items:
            $ref: '#/components/schemas/PrivatePricingTier'
      title: InputOutputBasedCostMetricValue
      x-fern-type-name: InputOutputBasedCostMetricValue
    NonNegativeFloat:
      description: +uiProps={"allowDecimal":true, "allowNegative":false}
      type: number
      minimum: 0
      title: NonNegativeFloat
      x-fern-type-name: NonNegativeFloat
    PatronusAnswerRelevanceCriteria:
      type: string
      enum:
        - patronus:answer-relevance
      title: PatronusAnswerRelevanceCriteria
      x-fern-type-name: PatronusAnswerRelevanceCriteria
    PatronusGliderCriteria:
      type: string
      enum:
        - patronus:is-compliant
        - patronus:is-factually-consistent
        - patronus:is-good-summary
        - patronus:is-harmful-advice
        - patronus:is-informal-tone
      title: PatronusGliderCriteria
      x-fern-type-name: PatronusGliderCriteria
    PatronusJudgeCriteria:
      type: string
      enum:
        - patronus:answer-refusal
        - patronus:no-openai-reference
        - patronus:is-concise
        - patronus:no-apologies
        - patronus:is-polite
        - patronus:is-helpful
        - patronus:clinically-inappropriate-tone
        - patronus:no-harmful-therapeutic-guidance
        - patronus:is-code
        - patronus:is-csv
        - patronus:is-json
        - patronus:no-age-bias
        - patronus:no-gender-bias
        - patronus:no-racial-bias
        - patronus:prompt-injection
      title: PatronusJudgeCriteria
      x-fern-type-name: PatronusJudgeCriteria
    PatronusPIICriteria:
      type: string
      enum:
        - patronus:pii
      title: PatronusPIICriteria
      x-fern-type-name: PatronusPIICriteria
    PatronusPhiCriteria:
      type: string
      enum:
        - patronus:phi
      title: PatronusPhiCriteria
      x-fern-type-name: PatronusPhiCriteria
    PatronusToxicityCriteria:
      type: string
      enum:
        - patronus:toxicity
      title: PatronusToxicityCriteria
      x-fern-type-name: PatronusToxicityCriteria
    InNotInOperator:
      type: object
      required:
        - condition
        - values
      properties:
        condition:
          type: string
          enum:
            - in
            - not_in
          default: in
        values:
          type: array
          items:
            type: string
      title: InNotInOperator
      x-fern-type-name: InNotInOperator
    MCPToolsOperator:
      type: object
      required:
        - condition
        - values
      properties:
        condition:
          type: string
          enum:
            - in
            - not_in
          default: in
        values:
          type: array
          items:
            $ref: '#/components/schemas/MCPToolTarget'
      title: MCPToolsOperator
      x-fern-type-name: MCPToolsOperator
    PrivatePricingTier:
      description: Pricing tier for volume-based pricing (per 1000 tokens)
      type: object
      required:
        - from
        - cost_per_token
      properties:
        from:
          description: >-
            +label=Token Threshold

            +usage=Token count threshold where this pricing tier begins (e.g.,
            200000 for 200k tokens)

            +sort=100
          type: integer
          minimum: 0
        cost_per_token:
          $ref: '#/components/schemas/NonNegativeFloat'
      title: PrivatePricingTier
      x-fern-type-name: PrivatePricingTier
    MCPToolTarget:
      type: object
      required:
        - name
        - enable_all_tools
      properties:
        name:
          description: |-
            +label=MCP Server Name
            +usage=Name of the MCP server
            +sort=1
          type: string
        enable_all_tools:
          description: >-
            +label=Enable All Tools

            +usage=When enabled, all tools from this MCP server are targeted.
            Disable to select specific tools.

            +sort=2
          type: boolean
          default: false
        enabled_tools:
          description: |-
            +label=Enabled Tools
            +usage=List of specific tools to target from this MCP server.
            +sort=3
          type: array
          items:
            type: string
      title: MCPToolTarget
      x-fern-type-name: MCPToolTarget
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````