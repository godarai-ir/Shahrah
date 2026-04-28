> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create or Update Agent

> Create or Update an Agent.



## OpenAPI

````yaml /openapi.json put /api/svc/v1/agents
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
  /api/svc/v1/agents:
    put:
      tags:
        - Agent
      summary: Create or Update Agent
      description: Create or Update an Agent.
      operationId: Agent.createOrUpdateAgent
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrUpdateAgentDto'
      responses:
        '200':
          description: Agent created or updated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetAgentResponseDto'
      security:
        - jwt: []
components:
  schemas:
    CreateOrUpdateAgentDto:
      type: object
      properties:
        manifest:
          oneOf:
            - $ref: '#/components/schemas/AgentManifest'
          description: Agent Manifest
        dryRun:
          type: boolean
          default: false
          description: >-
            Validate the manifest and collaborators without persisting or
            syncing to the gateway store
      required:
        - manifest
      title: CreateOrUpdateAgentDto
      x-fern-type-name: CreateOrUpdateAgentDto
    GetAgentResponseDto:
      type: object
      properties:
        data:
          description: Agent
          allOf:
            - $ref: '#/components/schemas/Agent'
      required:
        - data
      title: GetAgentResponseDto
      x-fern-type-name: GetAgentResponseDto
    AgentManifest:
      type: object
      oneOf:
        - $ref: '#/components/schemas/TrueFoundryAgentManifest'
        - $ref: '#/components/schemas/RemoteAgent'
        - $ref: '#/components/schemas/LegacyAgentManifest'
      title: AgentManifest
      x-fern-type-name: AgentManifest
    Agent:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        fqn:
          type: string
        tenantName:
          type: string
        accountId:
          type: string
          nullable: false
        latestVersion:
          type: number
        type:
          type: string
        latestVersionDetails:
          $ref: '#/components/schemas/AgentVersion'
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        createdBy:
          type: string
      required:
        - name
        - fqn
        - tenantName
        - accountId
        - latestVersion
        - type
        - createdBySubject
      title: Agent
      x-fern-type-name: Agent
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
    AgentVersion:
      type: object
      properties:
        id:
          type: string
        agentId:
          type: string
        fqn:
          type: string
        manifest:
          oneOf:
            - $ref: '#/components/schemas/AgentManifest'
          description: Agent Manifest
        version:
          type: number
        metadata:
          type: object
        createdBySubject:
          $ref: '#/components/schemas/Subject'
      required:
        - agentId
        - fqn
        - manifest
        - version
      title: AgentVersion
      x-fern-type-name: AgentVersion
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
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````