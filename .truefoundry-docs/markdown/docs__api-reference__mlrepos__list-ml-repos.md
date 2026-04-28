> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# List ML Repos

> List ML Repos with optional filtering by name.



## OpenAPI

````yaml /openapi.json get /api/ml/v1/ml-repos
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
  /api/ml/v1/ml-repos:
    get:
      tags:
        - MLRepos
      summary: List ML Repos
      description: List ML Repos with optional filtering by name.
      operationId: list_ml_repos_v1_ml_repos_get
      parameters:
        - description: Name of the ML Repo to filter by
          required: false
          schema:
            type: string
            title: Name
            description: Name of the ML Repo to filter by
          name: name
          in: query
        - description: Maximum number of ML Repos to return
          required: false
          schema:
            type: integer
            title: Limit
            description: Maximum number of ML Repos to return
          name: limit
          in: query
        - description: Number of ML Repos to skip for pagination
          required: false
          schema:
            type: integer
            title: Offset
            description: Number of ML Repos to skip for pagination
            default: 0
          name: offset
          in: query
      responses:
        '200':
          description: List of ML Repos matching the query with pagination information
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListMLReposResponse'
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
    ListMLReposResponse:
      properties:
        data:
          items:
            $ref: '#/components/schemas/MLRepo'
          type: array
          title: Data
          description: List of ML Repos matching the query
        pagination:
          allOf:
            - $ref: '#/components/schemas/Pagination'
          title: Pagination
          description: Pagination information including total count, offset, and limit
      type: object
      required:
        - data
        - pagination
      title: ListMLReposResponse
      x-fern-type-name: ListMLReposResponse
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
    MLRepo:
      type: object
      properties:
        tenant_name:
          type: string
          description: Tenant Name
          example: truefoundry
        manifest:
          oneOf:
            - $ref: '#/components/schemas/MLRepoManifest'
          example:
            name: devtest-finetuning-test-420
            type: ml-repo
            description: This is a test MLRepo for fine-tuning experiments.
            collaborators:
              - role_id: mlf-project-admin
                subject: user:user1@example.com
              - role_id: mlf-project-admin
                subject: team:teamb
            storage_integration_fqn: truefoundry:aws:aws-4:blob-storage:devtest-test-1
        created_by_subject:
          example:
            subjectId: cm1xxxp
            subjectSlug: test-user@truefoundry.com
            subjectType: user
            subjectDisplayName: Test User
          allOf:
            - $ref: '#/components/schemas/Subject'
        created_at:
          format: date-time
          type: string
          description: Created At
        account_id:
          type: string
          nullable: false
        id:
          type: string
          description: Experiment Id
          example: '420'
        num_runs:
          type: integer
          description: Number of runs
          example: 10
        artifact_type_counts:
          type: object
          items:
            type: string
          additionalProperties:
            type: integer
        datasets_count:
          type: integer
          description: Number of datasets
          example: 10
      required:
        - tenant_name
        - manifest
        - created_by_subject
        - created_at
        - account_id
        - id
      title: MLRepo
      x-fern-type-name: MLRepo
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