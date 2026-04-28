> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get Team

> Get Team associated with provided team id



## OpenAPI

````yaml /openapi.json get /api/svc/v1/teams/{id}
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
  /api/svc/v1/teams/{id}:
    get:
      tags:
        - Teams
      summary: Get Team
      description: Get Team associated with provided team id
      operationId: Teams.getTeamById
      parameters:
        - name: id
          required: true
          in: path
          description: Team Id
          schema:
            type: string
      responses:
        '200':
          description: Returns the Team associated with provided team id
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetTeamResponse'
        '404':
          description: Not Found. No team found for the given team ID
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/HttpError'
      security:
        - jwt: []
        - jwt: []
components:
  schemas:
    GetTeamResponse:
      type: object
      properties:
        data:
          description: Teams
          allOf:
            - $ref: '#/components/schemas/Team'
      required:
        - data
      title: GetTeamResponse
      x-fern-type-name: GetTeamResponse
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
    Team:
      type: object
      properties:
        id:
          type: string
        description:
          type: string
        tenantName:
          type: string
        accountId:
          type: string
          nullable: false
        createdBySubject:
          $ref: '#/components/schemas/Subject'
        members:
          type: array
          items:
            type: string
        createdAt:
          format: date-time
          type: string
        updatedAt:
          format: date-time
          type: string
        manifest:
          oneOf:
            - $ref: '#/components/schemas/TeamManifest'
        metadata:
          $ref: '#/components/schemas/TeamMetadata'
        isEditable:
          type: boolean
      required:
        - id
        - description
        - tenantName
        - accountId
        - createdBySubject
        - createdAt
        - updatedAt
        - manifest
        - isEditable
      title: Team
      x-fern-type-name: Team
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
    TeamMetadata:
      type: object
      properties:
        createdByScim:
          type: boolean
      title: TeamMetadata
      x-fern-type-name: TeamMetadata
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
  securitySchemes:
    jwt:
      type: http
      scheme: bearer
      bearerFormat: JWT

````