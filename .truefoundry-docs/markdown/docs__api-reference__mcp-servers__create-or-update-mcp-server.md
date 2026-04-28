> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create or update MCP server

> Creates a new MCP server or updates an existing one.



## OpenAPI

````yaml /openapi.json put /api/svc/v1/mcp-servers
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
  /api/svc/v1/mcp-servers:
    put:
      tags:
        - MCP Servers
      summary: Create or update MCP server
      description: Creates a new MCP server or updates an existing one.
      operationId: MCPServer.createOrUpdate
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrUpdateMCPServerRequestDto'
      responses:
        '200':
          description: MCP server created or updated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetMCPServerResponse'
        '400':
          description: Invalid request body
      security:
        - jwt: []
components:
  schemas:
    CreateOrUpdateMCPServerRequestDto:
      type: object
      properties:
        manifest:
          oneOf:
            - $ref: '#/components/schemas/MCPServerManifest'
          description: MCP Server Manifest
        dryRun:
          type: boolean
          default: false
          description: >-
            Validate and resolve the manifest without persisting or syncing to
            the gateway store
        metadata:
          type: object
          description: >-
            Optional metadata for auto-detected values (e.g.,
            detected_transport)
      required:
        - manifest
      title: CreateOrUpdateMCPServerRequestDto
      x-fern-type-name: CreateOrUpdateMCPServerRequestDto
    GetMCPServerResponse:
      type: object
      properties:
        data:
          description: MCP Server
          allOf:
            - $ref: '#/components/schemas/MCPServer'
      required:
        - data
      title: GetMCPServerResponse
      x-fern-type-name: GetMCPServerResponse
    MCPServerManifest:
      type: object
      oneOf:
        - $ref: '#/components/schemas/RemoteMCPServerManifest'
        - $ref: '#/components/schemas/VirtualMCPServerManifest'
        - $ref: '#/components/schemas/OpenAPIMCPServerManifest'
        - $ref: '#/components/schemas/StdioMCPServerManifest'
      title: MCPServerManifest
      x-fern-type-name: MCPServerManifest
    MCPServer:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        manifest:
          type: object
          additionalProperties: true
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        accountId:
          type: string
        tools:
          type: array
          items:
            type: object
        specEndpoints:
          type: array
          items:
            type: object
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        proxyUrl:
          type: string
          description: Proxy URL for accessing the MCP server
        authStatus:
          description: Authentication status for the MCP server
          allOf:
            - $ref: '#/components/schemas/MCPServerAuthStatusInfo'
        isAuthOverrideSupported:
          type: boolean
      required:
        - fqn
        - manifest
        - createdBySubject
        - accountId
      title: MCPServer
      x-fern-type-name: MCPServer
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
    MCPServerAuthStatusInfo:
      type: object
      properties:
        status:
          type: string
          enum:
            - authenticated
            - unauthenticated
          description: >-
            Whether the user is authenticated or unauthenticated for this MCP
            server
        method:
          type: string
          enum:
            - oauth
            - auth-override
          description: Authentication method used (e.g. OAuth, auth-override)
        message:
          type: string
          description: >-
            Human-readable message providing additional context about the
            authentication status
      required:
        - status
      title: MCPServerAuthStatusInfo
      x-fern-type-name: MCPServerAuthStatusInfo
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
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````