> ## Documentation Index
> Fetch the complete documentation index at: https://www.truefoundry.com/llms.txt
> Use this file to discover all available pages before exploring further.

# Get ML Repo

> Get an ML Repo by its ID.



## OpenAPI

````yaml /openapi.json get /api/ml/v1/ml-repos/{id}
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
  /api/ml/v1/ml-repos/{id}:
    get:
      tags:
        - MLRepos
      summary: Get ML Repo
      description: Get an ML Repo by its ID.
      operationId: get_ml_repo_v1_ml_repos__id__get
      parameters:
        - required: true
          schema:
            type: string
            title: Id
          name: id
          in: path
      responses:
        '200':
          description: The ML Repo data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GetMLRepoResponse'
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
    GetMLRepoResponse:
      type: object
      properties:
        data:
          description: ML Repo
          allOf:
            - $ref: '#/components/schemas/MLRepo'
      required:
        - data
      title: GetMLRepoResponse
      x-fern-type-name: GetMLRepoResponse
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