> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List Workspace

> List workspaces associated with the user. Optional filters include clusterId, fqn, and workspace name.



## OpenAPI

````yaml /openapi.json get /api/svc/v1/workspaces
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
    get:
      tags:
        - Workspaces
      summary: List Workspace
      description: >-
        List workspaces associated with the user. Optional filters include
        clusterId, fqn, and workspace name.
      operationId: WorkSpaces.listAllWorkspaces
      parameters:
        - name: limit
          required: false
          in: query
          description: Number of items per page
          schema:
            minimum: 1
            maximum: 1000
            default: 100
            example: 10
            type: integer
        - name: offset
          required: false
          in: query
          description: Number of items to skip
          schema:
            minimum: 0
            default: 0
            example: 0
            type: integer
        - name: clusterId
          required: false
          in: query
          description: ClusterId of the Cluster
          schema:
            type: string
        - name: name
          required: false
          in: query
          description: Workspace Name
          schema:
            type: string
        - name: fqn
          required: false
          in: query
          description: Workspace FQN
          schema:
            type: string
        - name: includeCluster
          required: false
          in: query
          description: When true, each workspace includes cluster information
          schema:
            type: boolean
      responses:
        '200':
          description: >-
            Returns all the workspaces associated with a user and also the
            response includes paginated data.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListWorkspacesResponse'
      security:
        - jwt: []
components:
  schemas:
    ListWorkspacesResponse:
      type: object
      properties:
        data:
          description: Array of Workspace
          type: array
          items:
            $ref: '#/components/schemas/Workspace'
        pagination:
          description: Pagination information
          allOf:
            - $ref: '#/components/schemas/Pagination'
      required:
        - data
        - pagination
      title: ListWorkspacesResponse
      x-fern-type-name: ListWorkspacesResponse
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
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````