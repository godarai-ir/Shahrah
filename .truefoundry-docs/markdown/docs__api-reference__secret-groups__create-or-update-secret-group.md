> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Create or Update Secret Group

> Creates a new secret group or updates an existing one based on the provided manifest.



## OpenAPI

````yaml /openapi.json put /api/svc/v1/secret-groups
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
  /api/svc/v1/secret-groups:
    put:
      tags:
        - Secret Groups
      summary: Create or Update Secret Group
      description: >-
        Creates a new secret group or updates an existing one based on the
        provided manifest.
      operationId: SecretGroup.createOrUpdateSecretGroup
      parameters: []
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrUpdateSecretGroupRequest'
      responses:
        '200':
          description: Secret group created or updated successfully.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetSecretGroupResponse'
        '400':
          description: Bad Request. Invalid input parameters.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '403':
          description: Forbidden. Insufficient permissions.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '404':
          description: Not Found. Tenant not found.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '409':
          description: Conflict. Secret group name already exists in the tenant.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
        '422':
          description: Unprocessable Entity. Manifest validation failed.
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
components:
  schemas:
    CreateOrUpdateSecretGroupRequest:
      type: object
      properties:
        manifest:
          oneOf:
            - $ref: '#/components/schemas/SecretGroupManifest'
          description: Secret Group Manifest
        dryRun:
          type: boolean
          default: false
          description: >-
            Validate the manifest and collaborators without persisting or
            updating authorizations and secret groups
      required:
        - manifest
      title: CreateOrUpdateSecretGroupRequest
      x-fern-type-name: CreateOrUpdateSecretGroupRequest
    GetSecretGroupResponse:
      type: object
      properties:
        data:
          description: Secret Group
          allOf:
            - $ref: '#/components/schemas/SecretGroup'
      required:
        - data
      title: GetSecretGroupResponse
      x-fern-type-name: GetSecretGroupResponse
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
    SecretGroup:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        tenantName:
          type: string
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        associatedSecrets:
          type: array
          items:
            $ref: '#/components/schemas/Secret'
        integrationId:
          type: string
        manifest:
          type: object
        accountId:
          type: string
          nullable: false
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        createdBy:
          type: string
      required:
        - tenantName
        - createdBySubject
        - associatedSecrets
        - accountId
        - createdAt
        - updatedAt
      title: SecretGroup
      x-fern-type-name: SecretGroup
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
    Secret:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        name:
          type: string
        secretGroupId:
          type: string
        value:
          type: string
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        secretVersions:
          type: array
          items:
            $ref: '#/components/schemas/SecretVersion'
        activeDeploymentsCount:
          type: integer
        createdBy:
          type: string
      required:
        - id
        - fqn
        - name
        - secretGroupId
      title: Secret
      x-fern-type-name: Secret
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
    SecretVersion:
      type: object
      properties:
        id:
          type: string
        fqn:
          type: string
        value:
          type: string
        version:
          type: number
        secret:
          $ref: '#/components/schemas/Secret'
        secretId:
          type: string
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
      required:
        - id
        - fqn
      title: SecretVersion
      x-fern-type-name: SecretVersion
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````