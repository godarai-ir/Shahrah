> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create or Update Workspace

> Creates a new workspace or updates an existing one based on the provided manifest.



## OpenAPI

````yaml /openapi.json put /api/svc/v1/workspaces
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
  /api/svc/v1/workspaces:
    put:
      tags:
        - Workspaces
      summary: Create or Update Workspace
      description: >-
        Creates a new workspace or updates an existing one based on the provided
        manifest.
      operationId: WorkSpaces.createOrUpdateWorkspace
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrUpdateWorkspaceRequest'
      responses:
        '200':
          description: |-
            - Creates or updates a workspace with given manifest
                - Corresponding authorization entry with admin role is made using newly created workspace
                - Attached with the cluster id where the workspace is created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetWorkspaceResponse'
        '400':
          description: |-
            Bad Request. This could happen due to:
                - Invalid clusterId supplied
                - Environment name is required
                - Environment does not exist in cluster
                - Environment does not exist
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '403':
          description: |-
            Forbidden. This could happen due to:
                - Unauthorized to create workspace in the cluster
                - Unauthorized to update workspace with FQN
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: |-
            Not Found. This could happen due to:
                - Invalid clusterId supplied
                - Environment does not exist
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '422':
          description: |-
            Unprocessable Entity. This could happen due to:
                - Manifest validation failed
                - Collaborators cannot be empty for a workspace created by a service account
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    CreateOrUpdateWorkspaceRequest:
      type: object
      properties:
        manifest:
          oneOf:
            - $ref: '#/components/schemas/WorkspaceManifest'
          description: Workspace manifest
        dryRun:
          type: boolean
          default: false
          description: Dry run the request
      required:
        - manifest
      title: CreateOrUpdateWorkspaceRequest
      x-fern-type-name: CreateOrUpdateWorkspaceRequest
    GetWorkspaceResponse:
      type: object
      properties:
        data:
          description: Workspace
          allOf:
            - $ref: '#/components/schemas/Workspace'
      required:
        - data
      title: GetWorkspaceResponse
      x-fern-type-name: GetWorkspaceResponse
    HttpError:
      type: object
      properties:
        statusCode:
          type: integer
          description: HTTP Status Code
        message:
          type: string
          description: Error Message
        code:
          oneOf:
            - type: integer
            - type: string
          description: Error code
        details:
          type: array
          description: Error details
          items:
            type: object
      required:
        - statusCode
        - message
      title: HttpError
      x-fern-type-name: HttpError
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
    Workspace:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        tenantName:
          type: string
        clusterId:
          type: string
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        environmentId:
          type: string
        manifest:
          oneOf:
            - $ref: '#/components/schemas/WorkspaceManifest'
        accountId:
          type: string
          nullable: false
        isSystemWs:
          type: boolean
        createdBy:
          type: string
      required:
        - id
        - fqn
        - tenantName
        - clusterId
        - createdBySubject
        - createdAt
        - updatedAt
        - manifest
        - accountId
      title: Workspace
      x-fern-type-name: Workspace
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
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````